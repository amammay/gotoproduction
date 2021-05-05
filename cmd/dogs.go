package main

import (
	"github.com/amammay/gotoproduction"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *server) handleGetDog(dogService *gotoproduction.DogService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		dogID, ok := vars["dogID"]
		if !ok {
			s.respond(w, nil, http.StatusBadRequest)
			return
		}
		dog, err := dogService.GetDogByID(r.Context(), dogID)
		if err == gotoproduction.ErrDogNotFound {
			s.respond(w, nil, http.StatusNotFound)
			return
		}
		if err != nil {
			s.respond(w, nil, http.StatusInternalServerError)
			return
		}
		s.respond(w, dog, http.StatusOK)
	}
}

func (s *server) handleFindDog(dogService *gotoproduction.DogService) http.HandlerFunc {
	type DogTypesResponse struct {
		Dogs []*gotoproduction.Dog `json:"dogs"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		dogType := query.Get("type")
		if dogType == "" {
			s.respond(w, nil, http.StatusNotFound)
			return
		}
		byType, err := dogService.FindDogByType(r.Context(), dogType)
		if err != nil {
			s.respond(w, nil, http.StatusInternalServerError)
			return
		}
		response := &DogTypesResponse{Dogs: byType}
		s.respond(w, response, http.StatusOK)
	}
}

func (s *server) handleCreateDog(dogService *gotoproduction.DogService) http.HandlerFunc {
	type CreateDogResponse struct {
		DogID string `json:"dog_id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		request := &gotoproduction.CreateDogRequest{}
		err := s.decode(r, request)
		if err != nil {
			s.respond(w, nil, http.StatusBadRequest)
		}
		defer r.Body.Close()
		if request.Name == "" || request.Type == "" {
			s.respond(w, nil, http.StatusBadRequest)
			return
		}

		dogID, err := dogService.CreateDog(r.Context(), request)
		if err != nil {
			s.respond(w, nil, http.StatusInternalServerError)
			return
		}
		response := &CreateDogResponse{DogID: dogID}
		s.respond(w, response, http.StatusOK)
	}
}
