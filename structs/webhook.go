package structs

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Webhook struct {
	ID          string        `json:"id"`
	Enabled     bool          `json:"enabled"`
	Method      string        `json:"method"`
	Path        string        `json:"path"`
	ForwardUrls []*ForwardUrl `json:"forwardUrls"`
}

type ForwardUrl struct {
	ID                     string     `json:"id"`
	Url                    string     `json:"url"`
	KeepSuccessfulRequests bool       `json:"keepSuccessfulRequests"`
	PendingRequests        []*Request `json:"pendingRequests"`
}

func (w *Webhook) RegisterWithEcho(e *echo.Echo) {
	e.Add(w.Method, w.Path, func(c echo.Context) error {
		if !w.Enabled {
			return c.String(http.StatusNotFound, "404")
		}
		return c.String(http.StatusOK, "OK. To be implemented.")
	})
}
