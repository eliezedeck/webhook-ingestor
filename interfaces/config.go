package interfaces

import "github.com/eliezedeck/webhook-ingestor/core"

type ConfigStorage interface {
	GetAdminPath() (string, error)
	SetAdminPath(path string) error

	GetValidWebhooks() ([]*core.Webhook, error)
	GetWebhook(id string) (*core.Webhook, error)

	AddWebhook(webhook *core.Webhook) error
	RemoveWebhook(id string) error
	EnableWebhook(id string) error
	DisableWebhook(id string) error
}
