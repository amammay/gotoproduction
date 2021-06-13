package main

import (
	"github.com/amammay/gotoproduction"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *server) handleGetDog(dogService *gotoproduction.DogService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := s.appLogger.WrapTraceContext(ctx)
		vars := mux.Vars(r)
		dogID, ok := vars["dogID"]
		logger.Infof("searching for dog %s", dogID)
		if !ok {
			s.respond(w, nil, http.StatusBadRequest)
			return
		}
		dog, err := dogService.GetDogByID(ctx, dogID)
		if err == gotoproduction.ErrDogNotFound {
			s.respond(w, nil, http.StatusNotFound)
			return
		}
		if err != nil {
			s.respond(w, nil, http.StatusInternalServerError)
			return
		}
		logger.Infof("search found dog: %s", dog.ID)
		s.respond(w, dog, http.StatusOK)
	}
}

func (s *server) handleFindDog(dogService *gotoproduction.DogService) http.HandlerFunc {
	type DogTypesResponse struct {
		Dogs []*gotoproduction.Dog `json:"dogs"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := s.appLogger.WrapTraceContext(ctx)

		query := r.URL.Query()
		dogType := query.Get("type")
		logger.Infof("searching for dog type %s", dogType)

		if dogType == "" {
			s.respond(w, nil, http.StatusNotFound)
			return
		}
		byType, err := dogService.FindDogByType(ctx, dogType)
		if err != nil {
			s.respond(w, nil, http.StatusInternalServerError)
			return
		}
		logger.Infof("found %d dogs for %s", len(byType), dogType)
		response := &DogTypesResponse{Dogs: byType}
		s.respond(w, response, http.StatusOK)
	}
}

func (s *server) handleCreateDog(dogService *gotoproduction.DogService) http.HandlerFunc {
	type createDogResponse struct {
		DogID string `json:"dog_id"`
	}

	type createDogRequest struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
		Type string `json:"type"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := s.appLogger.WrapTraceContext(ctx)

		request := &createDogRequest{}
		err := s.decode(r, request)
		if err != nil {
			s.respond(w, nil, http.StatusBadRequest)
		}

		if request.Name == "" || request.Type == "" {
			s.respond(w, nil, http.StatusBadRequest)
			return
		}
		logger.Infow("incoming dog request", "name", request.Name, "type", request.Type, "age", request.Age)

		dogID, err := dogService.CreateDog(ctx, &gotoproduction.CreateDogRequest{
			Name: request.Name,
			Age:  request.Age,
			Type: request.Type,
		})
		if err != nil {
			s.respond(w, nil, http.StatusInternalServerError)
			return
		}
		logger.Infof("created dog: %s ", dogID)

		response := &createDogResponse{DogID: dogID}
		s.respond(w, response, http.StatusOK)
	}
}
