package structs

import "time"

type Request struct {
	ID               string              `json:"id"`
	Method           string              `json:"method"`
	Path             string              `json:"path"`
	Headers          map[string][]string `json:"headers"`
	Body             string              `json:"body"`
	CreatedAt        time.Time           `json:"createdAt"`
	FailedForwardUrl *ForwardUrl         `json:"failedForwardUrl"`
}
