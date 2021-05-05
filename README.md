# My Golang Production App

Reference source code found [here on github](https://github.com/amammay/gotoproduction)

## Rest api app structure

Supporting resources

- https://pace.dev/blog/2018/05/09/how-I-write-http-services-after-eight-years.html

As a note this is a somewhat opinionated section...

### Server struct

App server level dependencies are fields on our server struct (think something like a database, or router).

```go
package main

type server struct {
	router    *mux.Router
	firestore *firestore.Client
}
```

Main func just calls our run function, so we can handle fatal errors in a nice, easy to consume manor.

```go
package main

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "run(): %w\n", err)
	}
}

// run is a nice function that does all our server setup and starts listening for requests
func run() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	client, err := firestore.NewClient(context.Background(), "TODO-PROJECT")
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

```

Our server implements the http.Handler interface that way we could easily swap router implementations.

```go
package main

func (s *server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.router.ServeHTTP(writer, request)
}

```

#### Handlers/Middleware

Our handlers just hang off the server struct as a way to provide a closure for a handler func.

```go
package main

func (s *server) handleFindDog(someService *dogs.DogService) http.HandlerFunc {
	//define our request/response needed structs within the closure 
	type DogTypesResponse struct {
		Dogs []*dogs.Dog `json:"dogs"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		//...Do request processing, maybe use the DogService we have access to as well :)
		response := &DogTypesResponse{Dogs: byType}
		s.respond(w, response, http.StatusOK)
	}
}

```

In addition, middleware is just regular old go code, nothing special here. Just do our middleware processing and either
end the request chain or continue to the next handler.

```go
package main

func (s *server) isAllowed(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// if user is not allowed, then just send a 404
		if !userAllowed(r) {
			http.NotFound(w, r)
			return
		}
		// otherwise process the request to the next function
		next(w, r)
	}
}
```

Applying the middleware would be something as easy as

```go
package main

// apply our middleware
r.HandleFunc("/find", s.isAllowed(s.handleFindDog(dogService))).Methods(http.MethodGet)


```

#### Routing

Our requests routing is defined all withing a `routes` method that hangs off our server struct, that way you can easily
identify how all the routes are structured for our application

```go
package main

// as a fyi we are using github.com/gorilla/mux as our router
func (s *server) routes() {

	dogService := dogs.NewDogService(s.firestore)

	func(r *mux.Router) {
		r.HandleFunc("/find", s.handleFindDog(dogService)).Methods(http.MethodGet)
		r.HandleFunc("/{dogID}", s.handleGetDog(dogService)).Methods(http.MethodGet)
		r.HandleFunc("", s.handleCreateDog(dogService)).Methods(http.MethodPost)
	}(s.router.PathPrefix("/dogs").Subrouter())

}
```

### Logging

Using one of these products should get the job done

- https://github.com/uber-go/zap
- https://github.com/sirupsen/logrus

## Testing

Supporting resources

- https://testwithgo.com/

Testing is a very important part to the development of an application. Knowing what kind of testing to do, at what time
can help boost your productivity short term and long term.

The three main types of testing we will focus on are

- Unit Testing - Testing the smallest unit of code, mocking all external deps.
- Integration Testing - Testing how code integrates with 3rd party services (database as an example)
- Functional Testing - Testing how your end user would interact with your application all together.

Now before we dive in to testing tips and common things you can do to help your testing journey, we will just go over
some Opinionated rules that i generally follow.

### Naming Conventions

Generally i follow [this article](https://ieftimov.com/post/testing-in-go-naming-conventions/) for how i name my
files/variables/functions and so on for testing.

The main bullet points that i think are the most useful would be the following

**General File/Package naming conventions.**

- `xxx_test.go` is the file name for a test
- `example_xxx_test.go` is the file name for examples
- Tests within a package that is non-executable should most likely prefer doing black box testing. (
  ie. `package dog_test`)
    - The only reason for this is for the sake of testing an api layer that closely resonates to other consumers of that
      package.
    - In addition, you should use your best judgment for the statements above, and feel free to diverge where you seem
      its necessary or required

Function Naming Conventions

- `func TestXxx(*testing.T)` this is at a mininum enforce by the go tool
- A common pattern is `func TestStruct_Method_Params(t *testing.T)`

Now all together!

**Go Code**

```go
package dogs

import "fmt"

type Dog struct {
	Name string
	Age  int
}

func NewDog(name string, age int) Dog {
	return Dog{Name: name, Age: age}
}

func (d Dog) GetFormattedDesc() string {
	return fmt.Sprintf("%s is %d years old", d.Name, d.Age)
}

```

**Test Code**

```go
package dogs

import (
	"testing"
)

func TestDog_GetFormattedDesc(t *testing.T) {
	type fields struct {
		Name string
		Age  int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		//Table driven test
		{name: "Properly formatted name", want: "Oscar is 1 years old", fields: fields{Name: "Oscar", Age: 1}},
		//add as many test cases you want :)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Dog{
				Name: tt.fields.Name,
				Age:  tt.fields.Age,
			}
			if got := d.GetFormattedDesc(); got != tt.want {
				t.Errorf("GetFormattedDesc() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

**Variable Naming Conventions**

standard syntax for failed assertion would be

Regular Failed Assertion
`t.Errorf("FunctionCall(%q) = %s; want %s",arg,got,want)`

Error Handling
`t.Errorf("FunctionCall() = %v; want nil",err)`

```go

package dogs

import "testing"

func TestThing(t *testing.T) {

	// any special args that you would to showcase
	arg := "foo"
	// call the function/method and store it as the result
	got := SomeFuncCall(arg)
	// the expected result
	want := "some other thing"

	//assertion
	if got != want {
		t.Errorf("Thing(%q) = %s; want %s", arg, got, want)
	}

}



```

### Integration Testing

Before we dive into integration testing examples some important notes that I think is important to call out.

Peter Bourgon said it
best [here](https://peter.bourgon.org/blog/2021/04/02/dont-use-build-tags-for-integration-tests.html)

If for any reason the tests need extra OS level setup done prior to them running, skip the tests, and print out some
useful piece of information around why it was skipped. This can help get new developers get up and running with a
project alot easier.

Reasoning is that once a developer clones the code base and runs a `go test ./...` from the root of the repo, they will
immediately see what else the tests need to fully run, sort of like self documenting the expected dependencies of the
system right of the bat!

For an example of integration testing we will be using [test containers](github.com/testcontainers/testcontainers-go) to
provide implementations of our dependent services.

Reference code is found [here](./dogs/dogservice.go).

```go

package dogs_test

import (
	"context"
	"github.com/amammay/gotoproduction/dogs"
	"github.com/amammay/gotoproduction/internal"
	"github.com/matryer/is"
	"github.com/testcontainers/testcontainers-go"
	"testing"
)

// integration testing a service with a db interaction
func TestDogService(t *testing.T) {

	// skip test if docker is not available, very important that go test always work!
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx := context.Background()

	// set up out testing container
	fsContainer, err := internal.CreateFirestoreContainer(ctx)
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
	fsClient := internal.NewFirestoreTestingClient(ctx, t, endpoint)

	service := dogs.NewDogService(fsClient.Client)
	t.Run("Create", testDogService_CreateDog(service, fsClient))
	t.Run("Find By Type", testDogService_FindDogByType(service, fsClient))
	t.Run("GetDog By ID", testDogService_GetDogByID(service, fsClient))
}

func testDogService_CreateDog(ds *dogs.DogService, fsClient *internal.FsTestingClient) func(t *testing.T) {
	return func(t *testing.T) {
		// test our service
		...
	}
}

func testDogService_FindDogByType(ds *dogs.DogService, fsClient *internal.FsTestingClient) func(t *testing.T) {
	return func(t *testing.T) {
		// test our service
		...
	}
}

func testDogService_GetDogByID(ds *dogs.DogService, fsClient *internal.FsTestingClient) func(t *testing.T) {
	return func(t *testing.T) {
		// test our service
		...
	}
}

```

#### A bit more about testing containers

Leverage testing containers for easily spinning up external resources needed for testing

`go get -u github.com/testcontainers/testcontainers-go`

```go

package internal

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// CreateFirestoreContainer will spin up a new firestore emulator container
func CreateFirestoreContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "ridedott/firestore-emulator:latest",
		ExposedPorts: []string{"8080/tcp"},
		// An example for waiting for some endpoint in the container to be healthy before proceeding. 
		WaitingFor: wait.ForHTTP("/").WithPort("8080"),
		// An example for waiting for a certain log to be produced from the container
		WaitingFor: wait.ForLog("Dev App Server is now running."),
		// combining multiple strategies together
		WaitingFor: wait.ForAll(wait.ForHTTP("/").WithPort("8080"), wait.ForLog("Dev App Server is now running.")),
	}
	fsContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("testcontainers.GenericContainer(): %w", err)
	}

	return fsContainer, nil
}


```

### Functional Testing

So from the functional requirements of our system, we can say that we have a web service that consumes and produces json
for some crud'y operations for our dogs db.

The approach for the functional test is very similar to our integration test in terms of test setup.

Our main course of divergence is that instead of creating an instance of our service, we are working against our server
struct.

We are able to test our routing logic since our app server implements the `ServeHTTP` interface, and we invoke that with
our test recorder and test requests. In addition, we are able to test that our handler is processing the correct
payloads, and returning the correct json responses.

There are a couple different ways to approach the functional tests for http services, but that is for a different
segment.

```go

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/amammay/gotoproduction/dogs"
	"github.com/amammay/gotoproduction/internal"
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
	fsContainer, err := internal.CreateFirestoreContainer(ctx)
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
	fsClient := internal.NewFirestoreTestingClient(ctx, t, endpoint)
	// construct a instance of our server application
	s := newServer(fsClient.Client)
	t.Run("create dog handler", test_handleCreateDog(s, fsClient))
	t.Run("get dog handler", test_handleGetDog(s, fsClient))
	t.Run("find dog handler", test_handleFindDog(s, fsClient))
	t.Run("find dog handler, no dogs found", test_handleFindDog_nonFound(s, fsClient))
}

func test_handleCreateDog(s *server, fsClient *internal.FsTestingClient) func(t *testing.T) {
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

func test_handleGetDog(s *server, fsClient *internal.FsTestingClient) func(t *testing.T) {
	return func(t *testing.T) {
		...
	}

}

func test_handleFindDog(s *server, fsClient *internal.FsTestingClient) func(t *testing.T) {
	return func(t *testing.T) {
		...
	}

}

func test_handleFindDog_nonFound(s *server, fsClient *internal.FsTestingClient) func(t *testing.T) {
	return func(t *testing.T) {
		...
	}

}



```
