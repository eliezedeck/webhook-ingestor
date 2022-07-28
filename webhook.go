package main

type Webhook struct {
	ID      string
	Enabled bool

	Path        string
	ForwardUrls []*ForwardUrl
}

type ForwardUrl struct {
	URL string

	KeepSuccessfulRequests bool
}
