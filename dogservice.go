package gotoproduction

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	"fmt"
	"github.com/amammay/gotoproduction/internal/logx"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

const dogCollectionName = "dogs"

// ErrDogNotFound represents when a dog cannot be found
var ErrDogNotFound = errors.New("dog not found")

type Dog struct {
	Name             string    `json:"name" firestore:"name"`
	Age              int       `json:"age" firestore:"age"`
	Type             string    `json:"type" firestore:"type"`
	ID               string    `json:"id" firestore:"id"`
	CreatedTimestamp time.Time `json:"created_timestamp" firestore:"created_timestamp,serverTimestamp"`
}

type CreateDogRequest struct {
	Name string `json:"name" firestore:"name"`
	Age  int    `json:"age" firestore:"age"`
	Type string `json:"type" firestore:"type"`
}

type DogService struct {
	db        *firestore.Client
	appLogger *logx.AppLogger
}

func NewDogService(db *firestore.Client, logger *logx.AppLogger) *DogService {
	return &DogService{db: db, appLogger: logger}
}

// GetDogByID retrieves 1 dog by its id
func (ds *DogService) GetDogByID(ctx context.Context, id string) (*Dog, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "DogService.GetDogByID")
	defer span.End()
	logger := ds.appLogger.WrapTraceContext(ctx)
	dogPath := fmt.Sprintf("%s/%s", dogCollectionName, id)
	logger.Debugw("searching firestore", "path", dogPath)
	docRefSnap, err := ds.db.Doc(dogPath).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return nil, ErrDogNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("ds.db.Doc(%q): %w", dogPath, err)
	}
	dog := &Dog{}
	err = docRefSnap.DataTo(dog)
	if err != nil {
		return nil, fmt.Errorf("docRefSnap.DataTo(): %w", err)
	}
	return dog, nil
}

// FindDogByType will return all the dogs by a given type
func (ds *DogService) FindDogByType(ctx context.Context, dogType string) ([]*Dog, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "DogService.FindDogByType")
	defer span.End()
	logger := ds.appLogger.WrapTraceContext(ctx)
	logger.Debugw("searching firestore", "collection", dogCollectionName, "type", dogType)

	all, err := ds.db.Collection(dogCollectionName).Where("type", "==", dogType).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("ds.db.Collection(): %w", err)
	}
	var dogs []*Dog
	for _, snapshot := range all {
		dog := &Dog{}
		err := snapshot.DataTo(dog)
		if err != nil {
			return nil, fmt.Errorf("snapshot.DataTo(): %w", err)
		}
		dogs = append(dogs, dog)
	}
	return dogs, nil
}

// CreateDog will create a new dog entry
func (ds *DogService) CreateDog(ctx context.Context, request *CreateDogRequest) (string, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "DogService.CreateDog")
	defer span.End()
	logger := ds.appLogger.WrapTraceContext(ctx)

	doc := ds.db.Collection(dogCollectionName).NewDoc()
	logger.Debugw("creating firestore doc", "collection", dogCollectionName, "id", doc.ID)
	dog := &Dog{
		Name: request.Name,
		Age:  request.Age,
		Type: request.Type,
		ID:   doc.ID,
	}
	_, err := doc.Create(ctx, dog)
	if err != nil {
		return "", fmt.Errorf("doc.Create(): %w", err)
	}
	return doc.ID, nil
}
