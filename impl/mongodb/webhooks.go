package mongodbimpl

import (
	"context"
	"fmt"

	"github.com/eliezedeck/gobase/random"
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

	webhooks := make([]*core.Webhook, 0, 10)
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

func (m *Storage) UpdateWebhook(webhook *core.Webhook) error {
	existing, err := m.GetWebhook(webhook.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("webhook with id %s not found", webhook.ID)
	}

	// Disallow mutation of certain fields
	if webhook.Path != existing.Path {
		return fmt.Errorf("cannot update Webhook Path")
	}
	if webhook.Method != existing.Method {
		return fmt.Errorf("cannot update Webhook Method")
	}

	// Update the rest of the fields
	existing.Name = webhook.Name
	existing.Enabled = webhook.Enabled
	for _, f := range webhook.ForwardUrls {
		if f.ID == "" {
			// New forward URL, generate a random ID
			f.ID = random.String(8)
		}
	}
	existing.ForwardUrls = webhook.ForwardUrls

	_, err = m.collWebhooks.UpdateOne(context.Background(), bson.D{{"_id", webhook.ID}}, bson.D{{"$set", existing}})
	return err
}
