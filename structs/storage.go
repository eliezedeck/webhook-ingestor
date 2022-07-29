package structs

type ConfigStorage interface {
	GetAdminPath() (string, error)
	SetAdminPath(path string) error

	GetValidWebhooks() ([]*Webhook, error)
	GetWebhook(id string) (*Webhook, error)

	AddWebhook(webhook *Webhook) error
	RemoveWebhook(id string) error
	EnableWebhook(id string) error
	DisableWebhook(id string) error
}

type RequestsStorage interface {
	StoreRequest(request *Request) error
	GetOldestRequests(count int) ([]*Request, error)
	GetNewestRequests(count int) ([]*Request, error)
	GetRequest(id string) (*Request, error)
}
