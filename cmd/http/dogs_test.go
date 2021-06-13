package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/amammay/gotoproduction"
	"github.com/amammay/gotoproduction/internal/logx"
	"github.com/amammay/gotoproduction/internal/testx"
	"github.com/matryer/is"
	"github.com/testcontainers/testcontainers-go"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// more of a functional style of test that tests the rest endpoint + dog api layer
func Test_server_dogs(t *testing.T) {
	// skip test if docker is not available, very important that go test always work!
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx := context.Background()

	// set up out testing container
	fsContainer, err := testx.CreateFirestoreContainer(ctx)
	if err != nil {
		t.Fatalf("createFirestoreContainer() err = %v; want nil", err)
	}

	// run a clean up method to terminate the container when this test and all sub tests are completed
	t.Cleanup(func() {
		err := fsContainer.Terminate(ctx)
		if err != nil {
			t.Fatalf("fsContainer.Terminate() err = %v; want nil", err)
		}
	})

	// cool way to dynamically allocate an endpoint for us to use, we could have multiple integration test running in parallel that would never have a port collision
	endpoint, err := fsContainer.Endpoint(ctx, "")
	if err != nil {
		t.Errorf("fsContainer.Endpoint() err = %v; want nil", err)
	}

	// home made firestore testing client that has a util method for clearing all data
	fsClient := testx.NewFirestoreTestingClient(ctx, t, endpoint)

	s := newServer(fsClient.Client, logx.NewTesterLogger(t))

	t.Run("create dog handler", test_handleCreateDog(s, fsClient))
	t.Run("get dog handler", test_handleGetDog(s, fsClient))
	t.Run("find dog handler", test_handleFindDog(s, fsClient))
	t.Run("find dog handler, no dogs found", test_handleFindDog_nonFound(s, fsClient))
}

func test_handleCreateDog(s *server, fsClient *testx.FsTestingClient) func(t *testing.T) {
	type dogCreateReq struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	return func(t *testing.T) {
		fsClient.ClearData(t)
		is := is.New(t)

		dogReq := &dogCreateReq{
			Name: "Oscar",
			Type: "Golden Doodle",
		}
		jsonBytes, err := json.Marshal(dogReq)
		is.NoErr(err) // json.Marshal error
		buffer := bytes.NewBuffer(jsonBytes)

		request := httptest.NewRequest(http.MethodPost, "/dogs", buffer)
		recorder := httptest.NewRecorder()
		s.ServeHTTP(recorder, request)

		result := recorder.Result()
		is.Equal(result.StatusCode, http.StatusOK) // correct status code set
		body, err := io.ReadAll(result.Body)
		is.NoErr(err) // io.ReadAll error

		is.True(strings.Contains(string(body), `"dog_id":`)) // verify we got back an ID

	}
}

func test_handleGetDog(s *server, fsClient *testx.FsTestingClient) func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)
		fsClient.ClearData(t)

		dogService := gotoproduction.NewDogService(fsClient.Client, logx.NewTesterLogger(t))
		dog, err := dogService.CreateDog(context.Background(), &gotoproduction.CreateDogRequest{
			Name: "Oscar",
			Age:  1,
			Type: "Golden Doodle",
		})
		if err != nil {
			t.Fatalf("dogService.CreateDog() err = %v; want nil", err)
		}

		request := httptest.NewRequest(http.MethodGet, "/dogs/"+dog, nil)
		recorder := httptest.NewRecorder()
		s.ServeHTTP(recorder, request)
		result := recorder.Result()
		is.Equal(result.StatusCode, http.StatusOK) // correct status code set
		body, err := io.ReadAll(result.Body)
		is.NoErr(err)                                                       // io.ReadAll error
		is.True(strings.Contains(string(body), `{"name":"Oscar","age":1,`)) // verify we got back an ID
	}

}

func test_handleFindDog(s *server, fsClient *testx.FsTestingClient) func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)
		fsClient.ClearData(t)

		dogService := gotoproduction.NewDogService(fsClient.Client, logx.NewTesterLogger(t))
		dogType := "Golden Doodle"
		_, err := dogService.CreateDog(context.Background(), &gotoproduction.CreateDogRequest{
			Name: "Oscar",
			Age:  1,
			Type: dogType,
		})
		if err != nil {
			t.Fatalf("dogService.CreateDog() err = %v; want nil", err)
		}

		request := httptest.NewRequest(http.MethodGet, "/dogs/find", nil)
		query := request.URL.Query()
		query.Set("type", dogType)
		request.URL.RawQuery = query.Encode()
		recorder := httptest.NewRecorder()
		s.ServeHTTP(recorder, request)
		result := recorder.Result()
		is.Equal(result.StatusCode, http.StatusOK) // correct status code set
		body, err := io.ReadAll(result.Body)
		is.NoErr(err)                                                        // io.ReadAll error
		is.True(strings.Contains(string(body), `[{"name":"Oscar","age":1,`)) // verify we got back an ID
	}

}

func test_handleFindDog_nonFound(s *server, fsClient *testx.FsTestingClient) func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)
		fsClient.ClearData(t)

		dogService := gotoproduction.NewDogService(fsClient.Client, logx.NewTesterLogger(t))
		dogType := "Golden Doodle"
		_, err := dogService.CreateDog(context.Background(), &gotoproduction.CreateDogRequest{
			Name: "Oscar",
			Age:  1,
			Type: dogType,
		})
		if err != nil {
			t.Fatalf("dogService.CreateDog() err = %v; want nil", err)
		}

		request := httptest.NewRequest(http.MethodGet, "/dogs/find", nil)
		query := request.URL.Query()
		query.Set("type", dogType)
		request.URL.RawQuery = query.Encode()
		recorder := httptest.NewRecorder()
		s.ServeHTTP(recorder, request)
		result := recorder.Result()
		is.Equal(result.StatusCode, http.StatusOK) // correct status code set
		body, err := io.ReadAll(result.Body)
		is.NoErr(err)                                                        // io.ReadAll error
		is.True(strings.Contains(string(body), `[{"name":"Oscar","age":1,`)) // verify we got back an ID
	}

}
