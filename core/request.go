package core

import "time"

type Request struct {
	ID            string              `json:"id"`
	Method        string              `json:"method"`
	Path          string              `json:"path"`
	Headers       map[string][]string `json:"headers"`
	Body          string              `json:"body"`
	ForwardUrl    *ForwardUrl         `json:"forwardUrl"`
	FromWebhookId string              `json:"fromWebhookId"`
	CreatedAt     time.Time           `json:"createdAt"`
}
