package main

import (
	"encoding/json"
	"github.com/amammay/gotoproduction"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"net/http"
)

func (s *server) routes() {

	dogService := gotoproduction.NewDogService(s.firestore, s.appLogger)

	s.router.Use(otelmux.Middleware("gotoproduction"))

	func(r *mux.Router) {
		r.HandleFunc("/find", s.handleFindDog(dogService)).Methods(http.MethodGet)
		r.HandleFunc("/{dogID}", s.handleGetDog(dogService)).Methods(http.MethodGet)
		r.HandleFunc("", s.handleCreateDog(dogService)).Methods(http.MethodPost)
	}(s.router.PathPrefix("/dogs").Subrouter())

}

func (s *server) respond(w http.ResponseWriter, data interface{}, status int) {
	if data != nil {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(status)
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		return
	}

	w.WriteHeader(status)

}

func (s *server) decode(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
