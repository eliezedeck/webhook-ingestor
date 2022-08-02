package mongodbimpl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/eliezedeck/webhook-ingestor/core"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	client       *mongo.Client
	db           *mongo.Database
	collRequests *mongo.Collection
	collWebhooks *mongo.Collection
}

func NewStorage(uri, dbname string) (*Storage, error) {
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(uri).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	db := client.Database(dbname)

	// Setup collections and their indexes (if any)
	//
	collRequests := db.Collection("requests")
	if err := setupIndex(collRequests, IndexDefinition{
		Fields: []IndexField{
			{Name: "createdAt", Order: OrderASC},
		},
		Name: "date",
	}, false); err != nil {
		return nil, err
	}

	collWebhooks := db.Collection("webhooks")

	return &Storage{
		client:       client,
		db:           db,
		collRequests: collRequests,
		collWebhooks: collWebhooks,
	}, nil
}

func (m *Storage) StoreRequest(request *core.Request) error {
	collection := m.db.Collection("requests")
	_, err := collection.InsertOne(context.Background(), request)
	return err
}

type IndexFieldOrdering int

type IndexField struct {
	Name  string
	Order IndexFieldOrdering
}

const (
	OrderASC  IndexFieldOrdering = 1
	OrderDESC IndexFieldOrdering = -1
)

type IndexDefinition struct {
	Fields []IndexField
	// Name of the index, it is automatically suffixed with "Idx" so no need to add that
	Name string
}

func setupIndex(coll *mongo.Collection, definition IndexDefinition, unique bool) error {
	if len(definition.Fields) == 0 {
		return errors.New("no fields provided for setupIndex()")
	}

	// Build the index model
	keys := bson.D{}
	for _, f := range definition.Fields {
		keys = append(keys, bson.E{Key: f.Name, Value: int(f.Order)})
	}
	idxName := fmt.Sprintf("%sIdx", definition.Name)
	if unique {
		idxName = fmt.Sprintf("%sUniqueIdx", definition.Name)
	}
	mo := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetName(idxName).SetUnique(unique),
	}

	// Apply the index ... it's ok (no error) even if it already exists
	if _, err := coll.Indexes().CreateOne(context.Background(), mo); err != nil {
		return err
	}
	return nil
}
