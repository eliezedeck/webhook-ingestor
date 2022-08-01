package core

type Replay struct {
	RequestId       string `json:"requestId"         validate:"required"`
	WebhookId       string `json:"webhookId"         validate:"required"`
	ForwardUrlId    string `json:"forwardUrlId"      validate:"required"`
	DeleteOnSuccess int    `json:"deleteOnSuccess"`
}
