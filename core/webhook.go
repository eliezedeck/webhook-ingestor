package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/eliezedeck/gobase/logging"
	"github.com/eliezedeck/gobase/random"
	"github.com/eliezedeck/gobase/web"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Webhook struct {
	ID          string        `bson:"_id"          json:"id"`
	Name        string        `bson:"name"         json:"name"         validate:"required"`
	Enabled     bool          `bson:"enabled"      json:"enabled"`
	Method      string        `bson:"method"       json:"method"       validate:"required"`
	Path        string        `bson:"path"         json:"path"         validate:"required"`
	ForwardUrls []*ForwardUrl `bson:"forwardUrls"  json:"forwardUrls"  validate:"required"`

	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}

func (w *Webhook) RegisterWithEcho(e *echo.Echo, storage RequestsStorage) error {
	// There must be exactly one forward url with the returnAsResponse flag set to true
	returnAsResponseCount := 0
	for _, furl := range w.ForwardUrls {
		if furl.ReturnAsResponse >= 1 {
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
		reqId := fmt.Sprintf("r-%s", random.String(16))
		L := logging.L.Named(fmt.Sprintf("Webhook[%s:%s]", w.ID, w.Path)).With(
			zap.String("requestId", reqId),
			zap.Time("time", time.Now()),
			zap.Any("headers", c.Request().Header))

		//
		// Webhook has been called
		//

		if !w.Enabled {
			// Don't save the request here because it's not enabled
			L.Warn("Attempt to use disabled Webhook route", zap.String("path", w.Path))
			return c.String(http.StatusNotFound, "404 Disabled")
		}

		// Get the full body of the request
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			L.Error("Could not read the body of the request", zap.Error(err))
			return c.String(http.StatusInternalServerError, "500 Internal Server Error")
		}
		L.Info("Request body", zap.ByteString("body", body))

		//
		// Webhook body is now available
		//

		saveRequest := func(furl *ForwardUrl) {
			// Save the request
			request := &Request{
				ID:            reqId,
				Method:        c.Request().Method,
				Path:          c.Request().URL.Path,
				Headers:       c.Request().Header,
				Body:          string(body),
				ForwardUrl:    furl,
				FromWebhookId: w.ID,
				CreatedAt:     time.Now(),

				ReplayPayload: &Replay{
					RequestId:       reqId,
					WebhookId:       w.ID,
					ForwardUrlId:    furl.ID,
					DeleteOnSuccess: 1,
				},
			}
			if err := storage.StoreRequest(request); err != nil {
				L.Error("Error saving request", zap.Error(err), zap.String("webhookId", w.ID))
			} else {
				L.Info("Request has been saved", zap.String("id", request.ID))
			}
		}

		//
		// Forward the request to each of the ForwardUrls
		//
		responseErr := make(chan error, 1)
		wg := &sync.WaitGroup{}
		if len(w.ForwardUrls) > 0 {
			for _, furl := range w.ForwardUrls {
				if furl.WaitTillCompletion >= 1 {
					wg.Add(1)
				}

				go func(furl *ForwardUrl) {
					ctx, cancel := context.WithTimeout(context.Background(), furl.Timeout)
					defer func() {
						cancel()
						if furl.WaitTillCompletion >= 1 {
							wg.Done()
						}
					}()

					// Prepare a new request, transfer the headers
					request, _ := http.NewRequestWithContext(ctx, w.Method, furl.Url, bytes.NewReader(body))
					TransferHeaders(request.Header, c.Request().Header)

					// Execute the request
					response, err := ForwardHttpClient.Do(request)
					if err != nil {
						// Error executing: Rebuilt request -> Forwarded host
						saveRequest(furl)
						if furl.ReturnAsResponse >= 1 {
							responseErr <- err
						}
						return
					}
					defer func() {
						_ = response.Body.Close()
					}()

					// Always fully read the body
					fbody, err := io.ReadAll(response.Body)
					if err != nil {
						// Error reading: Body <- Forwarded host
						saveRequest(furl)
						if furl.ReturnAsResponse >= 1 {
							responseErr <- err
						}
						return
					}

					if furl.ReturnAsResponse >= 1 {
						// Body from Forwarded host -> Webhook caller
						TransferHeaders(c.Response().Header(), response.Header)
						c.Response().WriteHeader(response.StatusCode)
						_, err = c.Response().Write(fbody)
						if err != nil {
							saveRequest(furl)
							responseErr <- err
							return
						} else {
							// Success
							responseErr <- nil
						}
					}

					if furl.KeepSuccessfulRequests >= 1 {
						// Save the request
						saveRequest(furl)
					}
				}(furl)
			}

			wg.Wait()
		} else {
			// Simply save the request
			saveRequest(nil)
			responseErr <- web.OK(c)
		}

		err = <-responseErr
		if err != nil {
			L.Error("Unsuccessful request: {error}", zap.Error(err))
		}
		return err
	})

	logging.L.Info("Webhook has been registered: {method} {path} — {name}",
		zap.String("id", w.ID),
		zap.String("method", w.Method),
		zap.String("path", w.Path),
		zap.String("name", w.Name))
	return nil
}
