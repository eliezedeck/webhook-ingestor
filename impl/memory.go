package impl

import (
	"fmt"
	"time"

	"github.com/eliezedeck/webhook-ingestor/structs"
)

type MemoryStorage struct {
	AdminPath string

	webhooks []*structs.Webhook
	requests []*structs.Request
}

func (m *MemoryStorage) GetAdminPath() (string, error) {
	return m.AdminPath, nil
}

func (m *MemoryStorage) SetAdminPath(path string) error {
	m.AdminPath = path
	return nil
}

func (m *MemoryStorage) GetValidWebhooks() ([]*structs.Webhook, error) {
	return m.webhooks, nil
}

func (m *MemoryStorage) AddWebhook(webhook *structs.Webhook) error {
	webhook.Enabled = true
	m.webhooks = append(m.webhooks, webhook)
	return nil
}

func (m *MemoryStorage) RemoveWebhook(id string) error {
	for i, w := range m.webhooks {
		if w.ID == id {
			m.webhooks = append(m.webhooks[:i], m.webhooks[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("webhook with id %s not found", id)
}

func (m *MemoryStorage) EnableWebhook(id string) error {
	for _, w := range m.webhooks {
		if w.ID == id {
			w.Enabled = true
			return nil
		}
	}
	return fmt.Errorf("webhook with id %s not found", id)
}

func (m *MemoryStorage) DisableWebhook(id string) error {
	for _, w := range m.webhooks {
		if w.ID == id {
			w.Enabled = false
			return nil
		}
	}
	return fmt.Errorf("webhook with id %s not found", id)
}

func (m *MemoryStorage) StoreRequest(request *structs.Request) error {
	if request.CreatedAt.IsZero() {
		request.CreatedAt = time.Now()
	}
	m.requests = append(m.requests, request)
	return nil
}

func (m *MemoryStorage) GetOldestRequests(count int) ([]*structs.Request, error) {
	if count == 0 {
		return nil, nil
	}

	result := make([]*structs.Request, 0, count)
	for i := len(m.requests) - 1; i >= 0; i-- {
		result = append(result, m.requests[i])
		if len(result) == count {
			break
		}
	}

	return m.requests, nil
}

func (m *MemoryStorage) GetNewestRequests(count int) ([]*structs.Request, error) {
	if count == 0 {
		return nil, nil
	}

	result := make([]*structs.Request, 0, count)
	for _, r := range m.requests {
		result = append(result, r)
		if len(result) == count {
			break
		}
	}

	return result, nil
}
