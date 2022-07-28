package impl

import "github.com/eliezedeck/webhook-ingestor/structs"

type MemoryStorage struct {
	webhooks []*structs.Webhook
	requests []*structs.Request
}

func (m *MemoryStorage) GetValidWebhooks() []*structs.Webhook {
	return m.webhooks
}

func (m *MemoryStorage) AddWebhook(webhook *structs.Webhook) {
	webhook.Enabled = true
	m.webhooks = append(m.webhooks, webhook)
}

func (m *MemoryStorage) RemoveWebhook(id string) {
	for i, w := range m.webhooks {
		if w.ID == id {
			m.webhooks = append(m.webhooks[:i], m.webhooks[i+1:]...)
			return
		}
	}
}

func (m *MemoryStorage) EnableWebhook(id string) {
	for _, w := range m.webhooks {
		if w.ID == id {
			w.Enabled = true
			return
		}
	}
}

func (m *MemoryStorage) DisableWebhook(id string) {
	for _, w := range m.webhooks {
		if w.ID == id {
			w.Enabled = false
			return
		}
	}
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
