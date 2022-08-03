package core

import (
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/eliezedeck/gobase/logging"
	"github.com/eliezedeck/gobase/random"
	"github.com/eliezedeck/gobase/validation"
	"github.com/eliezedeck/gobase/web"
	"github.com/eliezedeck/webhook-ingestor/parameters"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func SetupAdministration(echoForWebhooks, echoForAdmin *echo.Echo, config ConfigStorage, reqStore RequestsStorage, path string) {
	a := echoForAdmin.Group(path, middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if subtle.ConstantTimeCompare([]byte(username), []byte(parameters.ParamAdminUsername)) == 1 && subtle.ConstantTimeCompare([]byte(password), []byte(parameters.ParamAdminPassword)) == 1 {
			return true, nil
		}
		return false, nil
	}))

	// ---
	a.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello Admin")
	})

	// --- Webhooks: List
	a.GET("/webhooks", func(c echo.Context) error {
		webhooks, err := config.GetAllWebhooks()
		if err != nil {
			return web.Error(c, err.Error())
		}
		return c.JSON(http.StatusOK, webhooks)
	})

	// --- Webhook: Add
	a.POST("/webhooks", func(c echo.Context) error {
		webhook := &Webhook{}
		webhook.ID = fmt.Sprintf("w-%s", random.String(11))
		webhook.Enabled = 1 // enabled by default
		webhook.CreatedAt = time.Now()
		if _, err := validation.ValidateJSONBody(c.Request().Body, webhook); err != nil {
			return web.BadRequestError(c, "Invalid JSON body")
		}

		// Ensure that this Webhook doesn't already exist (using the Method and Path)
		webhooks, err := config.GetAllWebhooks()
		if err != nil {
			return web.Error(c, err.Error())
		}
		for _, w := range webhooks {
			if w.Method == webhook.Method && w.Path == webhook.Path {
				return web.BadRequestError(c, "Webhook already exists")
			}
		}
		for _, furl := range webhook.ForwardUrls {
			// Set IDs for each of the new forward URLs
			furl.ID = fmt.Sprintf("f-%s", random.String(11))
		}

		if err := config.AddWebhook(webhook); err != nil {
			return web.BadRequestError(c, err.Error())
		}

		// Immediately register the route so that it's available for requests
		if err := webhook.RegisterWithEcho(echoForWebhooks, reqStore); err != nil {
			return web.Error(c, err.Error())
		}

		return c.JSON(http.StatusOK, webhook)
	})

	// --- Webhook: Remove
	a.DELETE("/webhooks/:id", func(c echo.Context) error {
		if err := config.RemoveWebhook(c.Param("id")); err != nil {
			return web.Error(c, err.Error())
		}

		return web.OK(c)
	})

	// --- Webhook: Update
	a.PUT("/webhooks", func(c echo.Context) error {
		webhook := &Webhook{}
		if _, err := validation.ValidateJSONBody(c.Request().Body, webhook); err != nil {
			return web.BadRequestError(c, "Invalid JSON body")
		}
		if webhook.ID == "" {
			return web.BadRequestError(c, "Webhook ID is required")
		}

		// - This doesn't re-register the handler, simply update the cache that is going to be used by the handler.
		// - This also makes a verification to ensure that it remains a valid Webhook.
		if err := webhook.RegisterWithEcho(echoForWebhooks, reqStore); err != nil {
			return web.Error(c, err.Error())
		}

		if err := config.UpdateWebhook(webhook); err != nil {
			return web.Error(c, err.Error())
		}
		return web.OK(c)
	})

	// --- Requests: List from newest
	a.GET("/requests/newest", func(c echo.Context) error {
		var err error

		count := uint64(100)
		countStr := strings.TrimSpace(c.QueryParam("count"))
		if countStr != "" {
			count, err = strconv.ParseUint(countStr, 10, 64)
			if err != nil {
				return web.BadRequestError(c, "Invalid count parameter")
			}
			if count > 1000 {
				return web.BadRequestError(c, "Count parameter must be less than 1000")
			}
		}

		requests, err := reqStore.GetNewestRequests(int(count))
		if err != nil {
			return web.Error(c, err.Error())
		}
		return c.JSON(http.StatusOK, requests)
	})

	// --- Requests: List from oldest
	a.GET("/requests/oldest", func(c echo.Context) error {
		var err error

		count := uint64(100)
		countStr := strings.TrimSpace(c.QueryParam("count"))
		if countStr != "" {
			count, err = strconv.ParseUint(countStr, 10, 64)
			if err != nil {
				return web.BadRequestError(c, "Invalid count parameter")
			}
			if count > 1000 {
				return web.BadRequestError(c, "Count parameter must be less than 1000")
			}
		}

		requests, err := reqStore.GetOldestRequests(int(count))
		if err != nil {
			return web.Error(c, err.Error())
		}
		return c.JSON(http.StatusOK, requests)
	})

	// --- Requests: Replay
	a.POST("/requests/replay", func(c echo.Context) error {
		// We only allow replay to a single Forward URL, pretty much any URL that's already registered
		wreq := Replay{}
		if _, err := validation.ValidateJSONBody(c.Request().Body, &wreq); err != nil {
			return web.BadRequestError(c, "Invalid JSON body")
		}

		// Get the request and the webhook and the selected Forward URL
		oreq, err := reqStore.GetRequest(wreq.RequestId)
		if err != nil {
			return err
		}
		webhook, err := config.GetWebhook(wreq.WebhookId)
		if err != nil {
			return err
		}
		if oreq == nil || webhook == nil {
			return web.BadRequestError(c, "Invalid request or webhook")
		}
		var furl *ForwardUrl
		for _, furl = range webhook.ForwardUrls {
			if furl.ID == wreq.ForwardUrlId {
				break
			}
		}
		if furl == nil {
			return web.BadRequestError(c, "Invalid forward URL")
		}

		// Craft the request based on the saved Request instance
		req, err := http.NewRequest(oreq.Method, furl.Url, strings.NewReader(oreq.Body))
		if err != nil {
			return web.Error(c, err.Error())
		}
		TransferHeaders(req.Header, oreq.Headers)

		// Execute the request
		response, err := ForwardHttpClient.Do(req)
		if err != nil {
			return web.Error(c, err.Error())
		}
		defer response.Body.Close()

		TransferHeaders(c.Response().Header(), response.Header)
		c.Response().WriteHeader(response.StatusCode)
		if _, err = io.Copy(c.Response(), response.Body); err != nil {
			return web.Error(c, err.Error())
		}

		// Should we delete the successful request?
		if wreq.DeleteOnSuccess >= 1 {
			if err = reqStore.DeleteRequest(wreq.RequestId); err != nil {
				return web.Error(c, err.Error())
			}
		}

		return nil // success
	})

	// --- Requests: Delete by ID
	a.DELETE("/requests/:id", func(c echo.Context) error {
		if err := reqStore.DeleteRequest(c.Param("id")); err != nil {
			return web.Error(c, err.Error())
		}
		return web.OK(c)
	})

	logging.L.Info("Administration setup complete", zap.String("path", path))
}
