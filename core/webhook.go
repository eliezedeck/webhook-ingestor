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

func (w *Webhook) Verify() error {
	// There must be exactly one forward url with the returnAsResponse flag set to true
	returnAsResponseCount := 0
	for _, furl := range w.ForwardUrls {
		if furl.ReturnAsResponse >= 1 {
			returnAsResponseCount++
			if returnAsResponseCount > 1 {
				return fmt.Errorf("webhook has more than one forward url with returnAsResponse set to true")
			}
		}
	}
	if returnAsResponseCount == 0 {
		return fmt.Errorf("webhook has no forward url with returnAsResponse set to true")
	}
	return nil
}

var (
	webhooksCache   = make(map[string]*Webhook)
	webhooksCacheMu = &sync.Mutex{}
)

func (w *Webhook) RegisterWithEcho(e *echo.Echo, storage RequestsStorage) error {
	if err := w.Verify(); err != nil {
		return err
	}

	// Cache this Webhook
	// - Upon Webhook update, this makes sure that handler will use the updated version, not the initial one
	// - This is used to ensure that the same Webhook is not registered twice
	key := fmt.Sprintf("%s %s", w.Method, w.Path)
	webhooksCacheMu.Lock()
	_, found := webhooksCache[key]
	webhooksCache[key] = w
	webhooksCacheMu.Unlock()
	if found {
		logging.L.Error("Webhook already registered, only Cache entry is updated", zap.String("id", w.ID), zap.String("key", key))
		return nil
	}

	e.Add(w.Method, w.Path, func(c echo.Context) error {
		reqId := fmt.Sprintf("r-%s", random.String(16))

		// Always get the freshest version of the webhook from the Cache
		webhooksCacheMu.Lock()
		currentWebhook := webhooksCache[key]
		webhooksCacheMu.Unlock()

		L := logging.L.Named(fmt.Sprintf("Webhook[%s:%s]", currentWebhook.ID, currentWebhook.Path)).With(
			zap.String("requestId", reqId),
			zap.Time("time", time.Now()),
			zap.Any("headers", c.Request().Header))

		//
		// Webhook has been called
		//

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
				FromWebhookId: currentWebhook.ID,
				CreatedAt:     time.Now(),

				ReplayPayload: &Replay{
					RequestId:       reqId,
					WebhookId:       currentWebhook.ID,
					ForwardUrlId:    furl.ID,
					DeleteOnSuccess: 0,
				},
			}
			if err := storage.StoreRequest(request); err != nil {
				L.Error("Error saving request", zap.Error(err), zap.String("webhookId", currentWebhook.ID))
			} else {
				L.Info("Request has been saved", zap.String("id", request.ID))
			}
		}

		responseErr := make(chan error, 1)
		if currentWebhook.Enabled && len(currentWebhook.ForwardUrls) > 0 {
			wg := &sync.WaitGroup{}
			for _, furl := range currentWebhook.ForwardUrls {
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
					request, _ := http.NewRequestWithContext(ctx, currentWebhook.Method, furl.Url, bytes.NewReader(body))
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
			saveRequest(nil)
			responseErr <- web.OK(c)
		}

		err = <-responseErr
		if err != nil {
			L.Error("Unsuccessful Webhook handling: {error}", zap.Error(err))
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
