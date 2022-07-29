package structs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/eliezedeck/gobase/logging"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Webhook struct {
	ID          string        `json:"id"`
	Enabled     bool          `json:"enabled"`
	Method      string        `json:"method"`
	Path        string        `json:"path"`
	ForwardUrls []*ForwardUrl `json:"forwardUrls"`
}

type ForwardUrl struct {
	ID                     string        `json:"id"`
	Url                    string        `json:"url"`
	KeepSuccessfulRequests bool          `json:"keepSuccessfulRequests"`
	Timeout                time.Duration `json:"timeout"`
	ReturnAsResponse       bool          `json:"returnAsResponse"`
	WaitTillCompletion     bool          `json:"waitForCompletion"`
	SavedRequests          []*Request    `json:"savedRequests"`
}

var (
	httpClient = &http.Client{}
)

func transferHeaders(dest, source http.Header) {
	for key, oheader := range source {
		if strings.ToLower(key) == "host" {
			continue
		}
		for _, h := range oheader {
			dest.Add(key, h)
		}
	}
}

func (w *Webhook) RegisterWithEcho(e *echo.Echo) error {
	// Verify that there is at least one forward url
	if len(w.ForwardUrls) == 0 {
		logging.L.Error("Webhook has no forward urls", zap.String("id", w.ID))
		return fmt.Errorf("webhook has no forward urls")
	}

	// There must be exactly one forward url with the returnAsResponse flag set to true
	returnAsResponseCount := 0
	for _, furl := range w.ForwardUrls {
		if furl.ReturnAsResponse {
			returnAsResponseCount++
			if returnAsResponseCount > 1 {
				logging.L.Error("Webhook has more than one forward url with returnAsResponse set to true", zap.String("id", w.ID))
				return fmt.Errorf("webhook has more than one forward url with returnAsResponse set to true")
			}
		}
	}
	if returnAsResponseCount == 0 {
		logging.L.Error("Webhook has no forward url with returnAsResponse set to true", zap.String("id", w.ID))
		return fmt.Errorf("webhook has no forward url with returnAsResponse set to true")
	}

	e.Add(w.Method, w.Path, func(c echo.Context) error {
		if !w.Enabled {
			return c.String(http.StatusNotFound, "404 Disabled")
		}

		// Get the full body of the request
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.String(http.StatusInternalServerError, "500 Internal Server Error")
		}

		// Forward the request to all the ForwardUrls
		responseErr := make(chan error)
		for _, furl := range w.ForwardUrls {
			go func(furl *ForwardUrl) {
				ctx, cancel := context.WithTimeout(context.Background(), furl.Timeout)
				defer cancel()

				// Prepare a new request, transfer the headers
				request, _ := http.NewRequestWithContext(ctx, w.Method, furl.Url, bytes.NewReader(body))
				transferHeaders(request.Header, c.Request().Header)

				// Execute the request
				response, err := httpClient.Do(request)
				if err != nil {
					responseErr <- err
					return
				}
				defer response.Body.Close()

				if furl.ReturnAsResponse {
					fbody, err := io.ReadAll(response.Body)
					if err != nil {
						responseErr <- err
						return
					}

					// Write back to the Webhook caller
					transferHeaders(c.Response().Header(), response.Header)
					c.Response().WriteHeader(response.StatusCode)
					_, err = c.Response().Write(fbody)
					responseErr <- err
				}
			}(furl)
		}

		return <-responseErr
	})

	logging.L.Info(fmt.Sprintf("Registered webhook: %s %s", w.Method, w.Path))
	return nil
}
