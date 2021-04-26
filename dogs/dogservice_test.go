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

// simple test case, just creates a dog
func testDogService_CreateDog(ds *dogs.DogService, fsClient *internal.FsTestingClient) func(t *testing.T) {
	return func(t *testing.T) {
		fsClient.ClearData(t)
		ctx := context.Background()
		is := is.New(t)

		createDogRequest := &dogs.CreateDogRequest{
			Name: "Oscar",
			Age:  1,
			Type: "Golden Doodle",
		}
		dog, err := ds.CreateDog(ctx, createDogRequest)
		is.NoErr(err) // ds.CreateDog error
		got := len(dog)
		want := 20
		is.Equal(got, want) // must have a document id

	}
}

// a bit more complex test case, creates a dog and then attempts to find that dog we created
func testDogService_FindDogByType(ds *dogs.DogService, fsClient *internal.FsTestingClient) func(t *testing.T) {
	return func(t *testing.T) {
		fsClient.ClearData(t)

		ctx := context.Background()
		is := is.New(t)

		createDogRequest := &dogs.CreateDogRequest{
			Name: "Oscar",
			Age:  1,
			Type: "Golden Doodle",
		}
		_, err := ds.CreateDog(ctx, createDogRequest)
		is.NoErr(err) // ds.CreateDog error

		byType, err := ds.FindDogByType(ctx, createDogRequest.Type)
		is.NoErr(err) // ds.FindDogByType error
		got := len(byType)
		want := 1
		is.Equal(got, want) // must find one dog by type

	}
}

// table driven test example, create a dog then attempt to find it by its ID, also test to see if our custom error is thrown when it cant find the dog by id
func testDogService_GetDogByID(ds *dogs.DogService, fsClient *internal.FsTestingClient) func(t *testing.T) {

	return func(t *testing.T) {
		fsClient.ClearData(t)
		ctx := context.Background()
		is := is.New(t)

		createDogRequest := &dogs.CreateDogRequest{
			Name: "Oscar",
			Age:  1,
			Type: "Golden Doodle",
		}
		createDog, err := ds.CreateDog(ctx, createDogRequest)
		is.NoErr(err) // ds.CreateDog error

		tests := []struct {
			args    string
			want    *dogs.Dog
			wantErr error
		}{
			{args: createDog, wantErr: nil},
			{args: "999", wantErr: dogs.ErrDogNotFound},
		}

		for _, tt := range tests {
			t.Run(tt.args, func(t *testing.T) {
				is := is.New(t)
				_, err := ds.GetDogByID(ctx, tt.args)
				if tt.wantErr != nil {
					is.Equal(tt.wantErr, err) // not matching errors
					return
				}
				is.NoErr(err) // ds.GetDogByID error

			})
		}

	}
}
