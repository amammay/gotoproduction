package internal

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http"
	"os"
	"testing"
)

type FsTestingClient struct {
	*firestore.Client
	projectID string
	endPoint  string
}

// CreateFirestoreContainer will spin up a new firestore emulator container
func CreateFirestoreContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "ridedott/firestore-emulator:latest",
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithPort("8080"),
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

// ClearData is a util method for clearing all the data in an firestore emulator database
func (f *FsTestingClient) ClearData(t *testing.T) {
	url := fmt.Sprintf("http://%s/emulator/v1/projects/%s/databases/(default)/documents", f.endPoint, f.projectID)
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		t.Errorf("http.NewRequest(): err = %v; want nil", err)
	}

	do, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Errorf("http.DefaultClient.Do() err = %v; want nil", err)
	}
	if do.StatusCode != http.StatusOK {
		t.Errorf("firestore clear data not ok; got %d want %d", do.StatusCode, http.StatusOK)
	}

}

// NewFirestoreTestingClient will setup a connection to the firestore emulator
func NewFirestoreTestingClient(ctx context.Context, t *testing.T, endpoint string) *FsTestingClient {
	err := os.Setenv("FIRESTORE_EMULATOR_HOST", endpoint)
	if err != nil {
		t.Fatalf("os.Setenv() err = %v; want nil", err)
	}
	projectID := "dummy"

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		t.Fatalf("firestore.NewClient() err = %v; want nil", err)
	}
	return &FsTestingClient{
		Client:    client,
		projectID: projectID,
		endPoint:  endpoint,
	}
}
