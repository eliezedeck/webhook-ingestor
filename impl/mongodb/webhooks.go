package mongodbimpl

import (
	"context"

	"github.com/eliezedeck/webhook-ingestor/core"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *Storage) GetValidWebhooks() ([]*core.Webhook, error) {
	opts := options.Find().SetSort(bson.D{{"createdAt", OrderASC}})
	cur, err := m.collWebhooks.Find(context.Background(), bson.D{{"enabled", true}}, opts)
	if err != nil {
		return nil, err
	}

	var webhooks []*core.Webhook
	if err := cur.All(context.Background(), &webhooks); err != nil {
		return nil, err
	}
	return webhooks, err
}

func (m *Storage) GetWebhook(id string) (*core.Webhook, error) {
	var webhook core.Webhook
	err := m.collWebhooks.FindOne(context.Background(), bson.D{{"_id", id}}).Decode(&webhook)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &webhook, err
}

func (m *Storage) AddWebhook(webhook *core.Webhook) error {
	_, err := m.collWebhooks.InsertOne(context.Background(), webhook)
	return err
}

func (m *Storage) RemoveWebhook(id string) error {
	_, err := m.collWebhooks.DeleteOne(context.Background(), bson.D{{"_id", id}})
	return err
}

func (m *Storage) EnableWebhook(id string) error {
	_, err := m.collWebhooks.UpdateOne(context.Background(), bson.D{{"_id", id}}, bson.D{{"$set", bson.D{{"enabled", true}}}})
	return err
}

func (m *Storage) DisableWebhook(id string) error {
	_, err := m.collWebhooks.UpdateOne(context.Background(), bson.D{{"_id", id}}, bson.D{{"$set", bson.D{{"enabled", false}}}})
	return err
}
