package main

type ConfigStorage interface {
	GetValidWebhooks() ([]*Webhook, error)

	AddWebhook(webhook *Webhook) error
	RemoveWebhook(id string) error
	EnableWebhook(id string) error
	DisableWebhook(id string) error
}

type RequestsStorage interface {
	GetOldestRequests(count int) ([]*Request, error)
	GetNewestRequests(count int) ([]*Request, error)
}
