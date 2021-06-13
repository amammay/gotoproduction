package main

import (
	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/amammay/gotoproduction/internal/logx"
	"github.com/gorilla/mux"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//define our ENV variable keys up here so its easy for somebody to see what they can set
const (
	portEnv = "PORT"

	defaultPortValue = "8080"
	defaultHostValue = "127.0.0.1"
)

type server struct {
	router    *mux.Router
	firestore *firestore.Client
	appLogger *logx.AppLogger
}

func newServer(client *firestore.Client, logger *logx.AppLogger) *server {
	s := &server{router: mux.NewRouter(), firestore: client, appLogger: logger}
	s.routes()
	return s
}

func (s *server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.router.ServeHTTP(writer, request)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "run(): %v\n", err)
		os.Exit(1)
	}
}

// run is an little abstraction layer so we can do our server startup and init deps and handle any setup/http server related errors in an easy fashion
func run() error {
	// create our base context that we will use for our all our server setup operations
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// init open telemetry

	onGCE := metadata.OnGCE()
	projectID := "a-mammay-website"
	if onGCE {
		id, err := metadata.ProjectID()
		if err != nil {
			return fmt.Errorf("metadata.ProjectID(): %v", err)
		}
		projectID = id
	}
	shutdownTracer, err := initTracing(ctx, projectID)
	if err != nil {
		return fmt.Errorf("initTracing(): %v", err)
	}
	defer shutdownTracer()

	logger, err := logx.NewProdLogger(projectID)
	if err != nil {
		return fmt.Errorf("logx.NewProdLogger(): %v", err)
	}
	defer logger.Sync()

	fsClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("firestore.NewClient(): %w", err)
	}

	port := os.Getenv(portEnv)
	if port == "" {
		port = defaultPortValue
	}
	host := ""
	if !onGCE {
		var err error
		logger, err = logx.NewDevLogger(projectID)
		if err != nil {
			return fmt.Errorf("logx.NewProdLogger(): %v", err)
		}
		host = defaultHostValue
	}

	s := newServer(fsClient, logger)

	httpServer := http.Server{
		Addr:         fmt.Sprintf("%s:%s", host, port),
		Handler:      s,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	httpServer.RegisterOnShutdown(cancel)

	// setup our shutdown signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(
		shutdown,
		os.Interrupt,    // Capture ctrl + c events
		syscall.SIGTERM, // Capture actual sig term event (kill command).
	)

	// setup our errgroup is listen for shutdown signal, from there attempt to shutdown our http server and capture any errors during shutdown
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		o := <-shutdown
		logger.Infof("sig: %s - starting shutting down sequence...", o)
		// we need to use a fresh context.Background() because the parent ctx we have in our current scope will be cancelled during the Shutdown method call
		graceFull, cancel := context.WithTimeout(context.Background(), 9*time.Second)
		defer cancel()
		// Shutdown the server with a timeout
		if err := httpServer.Shutdown(graceFull); err != nil {
			return fmt.Errorf("httpServer.Shutdown(): %w", err)
		}
		logger.Info("server has shutdown gracefully")
		return nil
	})
	logger.Infof("starting server on %q", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("httpServer.ListenAndServe(): %v", err)
	}
	return g.Wait()
}
