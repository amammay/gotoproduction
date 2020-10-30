# My Golang production checklist/guide

Supporting resources

- https://testwithgo.com/

## Unit Testing

### File Naming Conventions

- `export_test.go` to access unexported variables in external tests
- `xxx_internal_test.go` for internal tests
- `example_xxx_test.go` for examples in isolated files

### Function Naming Conventions

- `func TestStruct_Method_params(t *testing.T)` most likely if you are using `Method_params` then ya need to refactor to table driven testing.

Source Code

```go
package dog

type Dog struct {
    Name string
    Age int
}

func (d Dog) Bark(muzzled bool){
    if muzzled {
        fmt.Println("woof")
    } else {
        fmt.Println("WOOF!!")
    }
}

func Speak(lang string){
    switch lang {
    case "spanish":
        fmt.Println("Hola")
    default:
        fmt.Println("Hello")
    }

}

```

Test File

```go
package dog

//test type
func TestDog(t *testing.t){

}

//test method on type
func TestDog_Bark(t *testing.t){

}

//test method on type with param
// TestStruct_Method_params
func TestDog_Bark_muzzled(t *testing.t){

}

//test method on type with param
// TestStruct_Method_params
func TestDog_Bark_unmuzzled(t *testing.t){

}

//test function
func TestSpeak(t *testing.t){

}

```

Table Driven Tests

```go

func TestCircle(t *testing.T) {
    type args struct {
        r float64
    }
    tests := []struct {
        name string
        args args
        want float64
    }{
        {"circle", args{1}, 3.141592653589793},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := Circle(tt.args.r); got != tt.want {
                t.Errorf("Circle() = %v, want %v", got, tt.want)
            }
        })
    }
}

```

### Variable Naming Conventions

standard syntax for failed assertion would be

Regular Failed Assertion
`t.Errorf("FunctionCall(%q) = %s; want %s",arg,got,want)`

Error Handling
`t.Errorf("FunctionCall() = err; want nil",got,want)`

```go

func TestThing(t *testing.T) {

    // any special args that you would to showcase
    arg := "foo"
    // call the function/method and store it as the result
    got := Thing(arg)
    // the expected result
    want := "some other thing"

    //assertion
    if got != want {
        t.Errorf("Thing(%q) = %s; want %s",arg, got, want)
    }

}

```

## Integration Test

### Build Tags

Use build tags to seperate intergration test from normal unit tests, invoke with `- go test --tags=integration`

`myfile_integration_test.go`

```go
// +build integration

package mypackage

import "testing"


```

### Testing Containers

Leverage testing containers for easily spinning up external resources needed for testing

`go get -u github.com/testcontainers/testcontainers-go`

```go
package tasks

import (
    "context"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
    "os"
    "testing"
)

func TestMain(m *testing.M) {

    ctx := context.Background()
    req := testcontainers.ContainerRequest{
        Image:        "ridedott/firestore-emulator",
        ExposedPorts: []string{"8080:8080"},
        WaitingFor:   wait.ForLog("Dev App Server is now running."),
    }
    firestoreC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })

    if err != nil {
        panic(err)
    }

    run := m.Run()
    firestoreC.Terminate(ctx)
    os.Exit(run)
}

```

### Http Testing

#### Testing a non live server

Testing indvidual routes

```go
func TestHome(t *testing.T) {
    //response writer recorder
    w := httptest.NewRecorder()
    // httptest request
    r := httptest.NewRequest(http.MethodGet, "/", nil)
    // send the request into the application
    app.Home(w, r)
    // capture response
    response := w.Result()
    defer response.Body.Close()
    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        t.Errorf("ioutil.ReadAll() err = %s; want nil", err)
    }
    want := "<h1>Welcome!</h1>"
    got := string(body)
    if got != want {
        t.Errorf("GET / = %s; want %s", got, want)
    }

}
```

#### Testing against test server

Also usefull for spinning up a stub server for upstreams apis that you might have limited control over or they are not relialble.

```go
func TestLive(t *testing.T) {
    // create our test server, it will have a unique url and port
    serve := httptest.NewServer(&app.Server{})
    //defer the closing
    defer serve.Close()
    //since the server has a unique url, use a regular http request
    response, err := http.Get(serve.URL)
    if err != nil {
        t.Fatalf("GET / err = %s; want nil", err)
    }

    defer response.Body.Close()
    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        t.Errorf("ioutil.ReadAll() err = %s; want nil", err)
    }
    want := "<h1>Welcome!</h1>"
    got := string(body)
    if got != want {
        t.Errorf("GET / = %s; want %s", got, want)
    }

}
```
