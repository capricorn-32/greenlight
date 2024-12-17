package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct{ dsn string }
}

type application struct {
	config config
	logger *log.Logger
}

func main() {
	var cfg config

	// Define and parse command line flags for server configuration
	// Port for the API server
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	// Environment the application is running in (development, staging, or production)
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	// Data Source Name (DSN) for connecting to a PostgreSQL database
	flag.StringVar(&cfg.db.dsn, "dsn", "postgres://greenlight:password@localhost/greenlight", "PostgreSQL DSN")

	// Parse the command line flags provided
	flag.Parse()

	// Initialize a new logger that writes to standard output
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Open a database connection using the provided configuration
	db, err := openDB(cfg)
	if err != nil {
		// Log fatal error and terminate the application if database connection fails
		logger.Fatal(err)
	}
	defer db.Close() // Ensure the database connection is closed when main exits

	// Log a message indicating that the database connection pool has been established
	logger.Printf("database connection pool established")

	// Create an instance of the application with the configuration and logger
	app := &application{
		config: cfg,
		logger: logger,
	}

	// Configure the HTTP server with address, handlers, and timeout settings
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port), // Set server address using the configured port
		Handler:      app.routes(),                 // Set the HTTP request handler
		IdleTimeout:  time.Minute,                  // Set idle timeout duration
		ReadTimeout:  10 * time.Second,             // Set read timeout duration
		WriteTimeout: 30 * time.Second,             // Set write timeout duration
	}

	// Start the HTTP server and log the environment and address details
	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	if err = srv.ListenAndServe(); err != nil {
		// Log a fatal error and terminate the application if the server fails to start
		logger.Fatal(err)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	// Use sql.Open() to create an empty connection pool, using the dsn from the
	// config struct.
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// create a context with a 5second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use PingContext() to establish a new connection to the database, passing in the
	// context we created as a parameter. If the connection couldn't be established
	// successfully within the 5 sec deadline, then this will return an error.
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	// returns the sql.DB connection pool
	return db, nil
}
