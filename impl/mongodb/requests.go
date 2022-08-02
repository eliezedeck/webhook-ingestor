package mongodbimpl

import (
	"context"

	"github.com/eliezedeck/webhook-ingestor/core"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *Storage) StoreRequest(request *core.Request) error {
	collection := m.db.Collection("requests")
	_, err := collection.InsertOne(context.Background(), request)
	return err
}

func (m *Storage) GetOldestRequests(count int) ([]*core.Request, error) {
	opts := options.Find().SetSort(bson.D{{"createdAt", OrderASC}}).SetLimit(int64(count))
	cur, err := m.collRequests.Find(context.Background(), bson.D{}, opts)
	if err != nil {
		return nil, err
	}

	requests := make([]*core.Request, 0, count)
	if err := cur.All(context.Background(), &requests); err != nil {
		return nil, err
	}
	return requests, err
}

func (m *Storage) GetNewestRequests(count int) ([]*core.Request, error) {
	opts := options.Find().SetSort(bson.D{{"createdAt", OrderDESC}}).SetLimit(int64(count))
	cur, err := m.collRequests.Find(context.Background(), bson.D{}, opts)
	if err != nil {
		return nil, err
	}

	requests := make([]*core.Request, 0, count)
	if err := cur.All(context.Background(), &requests); err != nil {
		return nil, err
	}
	return requests, err
}

func (m *Storage) GetRequest(id string) (*core.Request, error) {
	var request core.Request
	err := m.collRequests.FindOne(context.Background(), bson.D{{"_id", id}}).Decode(&request)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &request, err
}

func (m *Storage) DeleteRequest(id string) error {
	_, err := m.collRequests.DeleteOne(context.Background(), bson.D{{"_id", id}})
	return err
}
