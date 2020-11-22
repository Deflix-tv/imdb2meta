package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/deflix-tv/imdb2meta/pb"
	"github.com/dgraph-io/badger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	bindAddr = flag.String("bindAddr", "localhost", `Local interface address to bind to. "localhost" only allows access from the local host. "0.0.0.0" binds to all network interfaces.`)
	port     = flag.Int("port", 8080, "Port to listen on")

	badgerPath = flag.String("badgerPath", "", "Path to the directory with the BadgerDB files")
	boltPath   = flag.String("boltPath", "", "Path to the bbolt DB file")
)

var (
	imdbBytes = []byte("imdb") // Bucket name for bbolt

	errNotFound = errors.New("Not found")
)

func main() {
	flag.Parse()

	// CLI argument check
	if *badgerPath == "" && *boltPath == "" {
		log.Fatalln(`Missing an argument for the DB: Either "-badgerPath" or "-boltPath".`)
	} else if *badgerPath != "" && *boltPath != "" {
		log.Fatalln(`You can only use either "-badgerPath" or "-boltPath", but not both at the same time`)
	}

	// Set up DB

	log.Println("Setting up DB...")
	var badgerDB *badger.DB
	var boltDB *bbolt.DB
	var err error
	if *badgerPath != "" {
		opts := badger.DefaultOptions(*badgerPath).
			WithLoggingLevel(badger.WARNING)
		badgerDB, err = badger.Open(opts)
		if err != nil {
			log.Fatalf("Couldn't open BadgerDB: %v\n", err)
		}
		defer badgerDB.Close()
	} else {
		boltDB, err = bbolt.Open(*boltPath, 0666, nil)
		if err != nil {
			log.Fatalf("Couldn't open bbolt DB: %v\n", err)
		}
		defer boltDB.Close()
		err = boltDB.View(func(tx *bbolt.Tx) error {
			if tx.Bucket(imdbBytes) == nil {
				return errors.New(`bbolt bucket "imdb" doesn't exist`)
			}
			return nil
		})
		if err != nil {
			log.Fatalf("Error during bbolt DB check: %v\n", err)
		}
	}

	// Set up HTTP service

	log.Println("Setting up HTTP service...")
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				log.Printf("Fiber's error handler was called: %v\n", e)
			} else {
				log.Printf("Fiber's error handler was called: %v\n", err)
			}
			c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
			return c.Status(code).SendString("An internal server error occurred")
		},
		DisableStartupMessage: true,
		BodyLimit:             0,
		ReadTimeout:           5 * time.Second,
		// Docker stop only gives us 10s. We want to close all connections before that.
		WriteTimeout: 9 * time.Second,
		IdleTimeout:  9 * time.Second,
	})
	// Middlewares
	app.Use(recover.New())
	app.Use(logger.New())
	// Endpoints
	app.Get("/health", healthHandler)
	app.Get("/meta/:id", createMetaHandler(badgerDB, boltDB))

	// Start server

	log.Println("Starting server...")
	stopping := false
	stoppingPtr := &stopping
	addr := *bindAddr + ":" + strconv.Itoa(*port)
	go func() {
		if err := app.Listen(addr); err != nil {
			if !*stoppingPtr {
				log.Fatalf("Couldn't start server: %v\n", err)
			} else {
				log.Fatalf("Error in app.Listen() during server shutdown (probably context deadline expired before the server could shutdown cleanly): %v\n", err)
			}
		}
	}()
	healthURL := "http://"
	if *bindAddr == "0.0.0.0" {
		healthURL += "localhost"
	} else {
		healthURL += *bindAddr
	}
	healthURL += ":" + strconv.Itoa(*port) + "/health"
	_, err = http.Get(healthURL) // Note: No timeout by default
	if err != nil {
		log.Fatalf("Couldn't send test request to server: %v\n", err)
	}
	log.Println("Server started successfully!")

	// Graceful shutdown

	c := make(chan os.Signal, 1)
	// Accept SIGINT (Ctrl+C) and SIGTERM (`docker stop`)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	log.Printf("Received signal %v, shutting down server...\n", sig)
	*stoppingPtr = true
	// Graceful shutdown, waiting for all current requests to finish without accepting new ones.
	if err := app.Shutdown(); err != nil {
		log.Fatalf("Error shutting down server: %v\n", err)
	}
	log.Println("Finished shutting down server")
}

var healthHandler fiber.Handler = func(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func createMetaHandler(badgerDB *badger.DB, boltDB *bbolt.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		var err error
		var metaBytes []byte

		if badgerDB != nil {
			err = badgerDB.View(func(txn *badger.Txn) error {
				item, err := txn.Get([]byte(id))
				if err != nil {
					if err == badger.ErrKeyNotFound {
						return errNotFound
					}
					return err
				}
				metaBytes, err = item.ValueCopy(nil)
				return err
			})
		} else {
			err = boltDB.View(func(tx *bbolt.Tx) error {
				txBytes := tx.Bucket(imdbBytes).Get([]byte(id))
				if txBytes == nil {
					return errNotFound
				}
				copy(metaBytes, txBytes)
				return nil
			})
		}
		if err != nil {
			if err == errNotFound {
				log.Printf("Key not found in DB: %v\n", err)
				return c.SendStatus(fiber.StatusNotFound)
			}
			log.Printf("Couldn't get data from DB: %v\n", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		meta := &pb.Meta{}
		err = proto.Unmarshal(metaBytes, meta)
		if err != nil {
			log.Printf("Couldn't unmarshal protocol buffer into object: %v\n", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		metaJSON, err := protojson.Marshal(meta)
		if err != nil {
			log.Printf("Couldn't marshal object into JSON: %v\n", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
		return c.Send(metaJSON)
	}
}
