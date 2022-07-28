package structs

type ConfigStorage interface {
	GetValidWebhooks() ([]*Webhook, error)

	AddWebhook(webhook *Webhook) error
	RemoveWebhook(id string) error
	EnableWebhook(id string) error
	DisableWebhook(id string) error
}

type RequestsStorage interface {
	StoreRequest(request *Request) error
	GetOldestRequests(count int) ([]*Request, error)
	GetNewestRequests(count int) ([]*Request, error)
}
