package impl

import (
	"fmt"
	"time"

	"github.com/eliezedeck/webhook-ingestor/core"
)

// MemoryStorage implements both ConfigStorage and RequestsStorage
type MemoryStorage struct {
	webhooks     []*core.Webhook
	webhooksById map[string]*core.Webhook
	requests     []*core.Request
	requestsById map[string]*core.Request
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		webhooks:     make([]*core.Webhook, 0, 16),
		webhooksById: make(map[string]*core.Webhook, 16),
		requests:     make([]*core.Request, 0, 256),
		requestsById: make(map[string]*core.Request, 256),
	}
}

func (m *MemoryStorage) GetValidWebhooks() ([]*core.Webhook, error) {
	return m.webhooks, nil
}

func (m *MemoryStorage) GetWebhook(id string) (*core.Webhook, error) {
	if w, ok := m.webhooksById[id]; ok {
		return w, nil
	}
	return nil, nil
}

func (m *MemoryStorage) AddWebhook(webhook *core.Webhook) error {
	webhook.Enabled = true
	m.webhooks = append(m.webhooks, webhook)
	m.webhooksById[webhook.ID] = webhook
	return nil
}

func (m *MemoryStorage) RemoveWebhook(id string) error {
	for i, w := range m.webhooks {
		if w.ID == id {
			w.Enabled = false
			m.webhooks = append(m.webhooks[:i], m.webhooks[i+1:]...)
			delete(m.webhooksById, id)
			return nil
		}
	}
	return fmt.Errorf("webhook with id %s not found", id)
}

func (m *MemoryStorage) EnableWebhook(id string) error {
	if w, ok := m.webhooksById[id]; ok {
		w.Enabled = true
		return nil
	}
	return fmt.Errorf("webhook with id %s not found", id)
}

func (m *MemoryStorage) DisableWebhook(id string) error {
	if w, ok := m.webhooksById[id]; ok {
		w.Enabled = false
		return nil
	}
	return fmt.Errorf("webhook with id %s not found", id)
}

func (m *MemoryStorage) StoreRequest(request *core.Request) error {
	if request.CreatedAt.IsZero() {
		request.CreatedAt = time.Now()
	}
	m.requests = append(m.requests, request)
	m.requestsById[request.ID] = request
	return nil
}

func (m *MemoryStorage) GetOldestRequests(count int) ([]*core.Request, error) {
	if count == 0 {
		return nil, nil
	}

	result := make([]*core.Request, 0, count)
	for i := len(m.requests) - 1; i >= 0; i-- {
		result = append(result, m.requests[i])
		if len(result) == count {
			break
		}
	}

	return m.requests, nil
}

func (m *MemoryStorage) GetNewestRequests(count int) ([]*core.Request, error) {
	if count == 0 {
		return nil, nil
	}

	result := make([]*core.Request, 0, count)
	for _, r := range m.requests {
		result = append(result, r)
		if len(result) == count {
			break
		}
	}

	return result, nil
}

func (m *MemoryStorage) GetRequest(id string) (*core.Request, error) {
	if r, ok := m.requestsById[id]; ok {
		return r, nil
	}
	return nil, nil
}

func (m *MemoryStorage) DeleteRequest(id string) error {
	if _, ok := m.requestsById[id]; ok {
		delete(m.requestsById, id)
		for i, rr := range m.requests {
			if rr.ID == id {
				m.requests = append(m.requests[:i], m.requests[i+1:]...)
				break
			}
		}
		return nil
	}
	return fmt.Errorf("request with id %s not found", id)
}
