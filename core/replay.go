package core

// Replay is the set of information required to replay any request. It is not stored directly in the database but as a child/nested object.
type Replay struct {
	RequestId       string `json:"requestId"         validate:"required"`
	WebhookId       string `json:"webhookId"         validate:"required"`
	ForwardUrlId    string `json:"forwardUrlId"      validate:"required"`
	DeleteOnSuccess int    `json:"deleteOnSuccess"`
}
