package core

import "time"

type Request struct {
	ID            string              `bson:"_id"            json:"id"`
	Method        string              `bson:"method"         json:"method"`
	Path          string              `bson:"path"           json:"path"`
	Headers       map[string][]string `bson:"headers"        json:"headers"`
	Body          string              `bson:"body"           json:"body"`
	ForwardUrl    *ForwardUrl         `bson:"forwardUrl"     json:"forwardUrl"`
	FromWebhookId string              `bson:"fromWebhookId"  json:"fromWebhookId"`
	CreatedAt     time.Time           `bson:"createdAt"      json:"createdAt"`

	ReplayPayload *Replay `bson:"replayPayload" json:"replayPayload"`
}
