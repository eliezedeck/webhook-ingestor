package core

type ConfigStorage interface {
	GetAllWebhooks() ([]*Webhook, error)
	GetWebhook(id string) (*Webhook, error)

	AddWebhook(webhook *Webhook) error
	RemoveWebhook(id string) error
	UpdateWebhook(webhook *Webhook) error
}

type RequestsStorage interface {
	StoreRequest(request *Request) error
	GetOldestRequests(count int) ([]*Request, error)
	GetNewestRequests(count int) ([]*Request, error)
	GetRequest(id string) (*Request, error)
	DeleteRequest(id string) error
}
