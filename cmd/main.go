package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"time"
)

type server struct {
	router    *mux.Router
	firestore *firestore.Client
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "run(): %v\n", err)
	}
}

func run() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	client, err := firestore.NewClient(context.Background(), "a-mammay-website")
	if err != nil {
		return fmt.Errorf("firestore.NewClient(): %w", err)
	}
	s := newServer(client)

	server := http.Server{
		Addr:         ":" + port,
		Handler:      s,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	return server.ListenAndServe()
}

func newServer(client *firestore.Client) *server {
	s := &server{router: mux.NewRouter(), firestore: client}
	s.routes()
	return s
}

func (s *server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.router.ServeHTTP(writer, request)
}
