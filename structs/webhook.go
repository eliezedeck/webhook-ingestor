package structs

type Webhook struct {
	ID      string
	Enabled bool

	Path        string
	ForwardUrls []*ForwardUrl

	FailedRequests []*Request
}

type ForwardUrl struct {
	ID string

	Url                    string
	KeepSuccessfulRequests bool
}
