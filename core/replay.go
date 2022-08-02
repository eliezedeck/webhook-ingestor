package core

// Replay is the set of information required to replay any request. It is not stored directly in the database but as a child/nested object.
type Replay struct {
	RequestId       string `bson:"requestId"        json:"requestId"         validate:"required"`
	WebhookId       string `bson:"webhookId"        json:"webhookId"         validate:"required"`
	ForwardUrlId    string `bson:"forwardUrlId"     json:"forwardUrlId"      validate:"required"`
	DeleteOnSuccess int    `bson:"deleteOnSuccess"  json:"deleteOnSuccess"`
}
