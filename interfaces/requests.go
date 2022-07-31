package interfaces

import "github.com/eliezedeck/webhook-ingestor/core"

type RequestsStorage interface {
	StoreRequest(request *core.Request) error
	GetOldestRequests(count int) ([]*core.Request, error)
	GetNewestRequests(count int) ([]*core.Request, error)
	GetRequest(id string) (*core.Request, error)
}
