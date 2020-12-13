package main

import (
	"errors"
	"flag"
	"log"
	"net"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/deflix-tv/imdb2meta/pb"
)

var (
	bindAddr = flag.String("bindAddr", "localhost", `Local interface address to bind to. "localhost" only allows access from the local host. "0.0.0.0" binds to all network interfaces.`)
	httpPort = flag.Int("httpPort", 8080, "Port to listen on for HTTP requests")
	grpcPort = flag.Int("grpcPort", 8081, "Port to listen on for gRPC requests")

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

	metaStore := &metaStore{
		badgerDB: badgerDB,
		boltDB:   boltDB,
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
	app.Get("/meta/:id", createMetaHandler(metaStore))

	// Start HTTP server

	log.Println("Starting HTTP server...")
	stopping := false
	stoppingPtr := &stopping
	addr := *bindAddr + ":" + strconv.Itoa(*httpPort)
	listenErr := make(chan struct{})
	go func() {
		if err := app.Listen(addr); err != nil {
			if !*stoppingPtr {
				log.Printf("Couldn't start HTTP server: %v\n", err)
				close(listenErr)
			} else {
				log.Printf("Error in app.Listen() during HTTP server shutdown (probably context deadline expired before the server could shutdown cleanly): %v\n", err)
				close(listenErr)
			}
		}
	}()
	// Send HTTP request to the health endpoint to ensure the server was started successfully
	healthURL := "http://"
	if *bindAddr == "0.0.0.0" {
		healthURL += "localhost"
	} else {
		healthURL += *bindAddr
	}
	healthURL += ":" + strconv.Itoa(*httpPort) + "/health"
	_, err = http.Get(healthURL) // Note: No timeout by default
	if err != nil {
		log.Printf("Couldn't send test request to HTTP server: %v\n", err)
		return
	}
	// Check listenErr additionally, because the health request could otherwise be successful for example due to an already running service on the same port.
	select {
	case <-listenErr:
		return
	case <-time.After(time.Second):
		log.Println("HTTP server started successfully!")
	}

	// Start gRPC server

	addr = *bindAddr + ":" + strconv.Itoa(*grpcPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("failed to listen: %v", err)
		return
	}
	s := grpc.NewServer()
	metaServer := createGRPCserver(metaStore)
	pb.RegisterMetaFetcherServer(s, metaServer)
	// Register reflection service on gRPC server for dynamic clients to discover services and types.
	reflection.Register(s)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Printf("Failed to serve gRPC: %v\n", err)
			// TODO: Check if there are cases where the channel could already be closed
			close(listenErr)
		}
	}()
	select {
	case <-listenErr:
		return
	case <-time.After(time.Second):
		log.Println("gRPC server started successfully!")
	}

	// Graceful shutdown

	c := make(chan os.Signal, 1)
	// Accept SIGINT (Ctrl+C) and SIGTERM (`docker stop`)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	log.Printf("Received signal %v, shutting down HTTP and gRPC server...\n", sig)
	*stoppingPtr = true
	// Graceful shutdown, waiting for all current requests to finish without accepting new ones.
	httpShutdownErr := false
	if err := app.Shutdown(); err != nil {
		log.Printf("Error shutting down HTTP server: %v\n", err)
		httpShutdownErr = true
	}
	s.GracefulStop()
	if httpShutdownErr {
		return
	}
	log.Println("Finished shutting down HTTP and gRPC server")
	select {
	case <-listenErr:
	default:
		exitCode = 0
	}
}
