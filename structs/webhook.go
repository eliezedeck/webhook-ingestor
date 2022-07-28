package structs

type Webhook struct {
	ID             string        `json:"id"`
	Enabled        bool          `json:"enabled"`
	Path           string        `json:"path"`
	ForwardUrls    []*ForwardUrl `json:"forwardUrls"`
	FailedRequests []*Request    `json:"failedRequests"`
}

type ForwardUrl struct {
	ID                     string `json:"id"`
	Url                    string `json:"url"`
	KeepSuccessfulRequests bool   `json:"keepSuccessfulRequests"`
}
