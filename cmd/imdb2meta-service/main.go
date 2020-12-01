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

	"github.com/dgraph-io/badger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.etcd.io/bbolt"
)

var (
	bindAddr = flag.String("bindAddr", "localhost", `Local interface address to bind to. "localhost" only allows access from the local host. "0.0.0.0" binds to all network interfaces.`)
	port     = flag.Int("port", 8080, "Port to listen on")

	badgerPath = flag.String("badgerPath", "", "Path to the directory with the BadgerDB files")
	boltPath   = flag.String("boltPath", "", "Path to the bbolt DB file")
)

var (
	imdbBytes = []byte("imdb") // Bucket name for bbolt
)

func main() {
	// Workaround for exiting with 1 despite not using log.Fatal while still running deferred DB close calls.
	exitCode := 1
	defer func() {
		os.Exit(exitCode)
	}()

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

	// Here after we have opened the DB, don't use log.Fatal or os.Exit, as then the DB won't be closed and can end up in a corrupted state.
	// So we log with Print and then return, leading to the deferred DB close and then deferred os.Exit(1) being called.

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
	listenErr := make(chan struct{})
	go func() {
		if err := app.Listen(addr); err != nil {
			if !*stoppingPtr {
				log.Printf("Couldn't start server: %v\n", err)
				close(listenErr)
			} else {
				log.Printf("Error in app.Listen() during server shutdown (probably context deadline expired before the server could shutdown cleanly): %v\n", err)
				close(listenErr)
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
		log.Printf("Couldn't send test request to server: %v\n", err)
		return
	}
	select {
	case <-listenErr:
		return
	case <-time.After(time.Second):
		log.Println("Server started successfully!")
	}

	// Graceful shutdown

	c := make(chan os.Signal, 1)
	// Accept SIGINT (Ctrl+C) and SIGTERM (`docker stop`)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	log.Printf("Received signal %v, shutting down server...\n", sig)
	*stoppingPtr = true
	// Graceful shutdown, waiting for all current requests to finish without accepting new ones.
	if err := app.Shutdown(); err != nil {
		log.Printf("Error shutting down server: %v\n", err)
		return
	}
	log.Println("Finished shutting down server")
	select {
	case <-listenErr:
	default:
		exitCode = 0
	}
}
