package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/eliezedeck/gobase/logging"
	"github.com/eliezedeck/gobase/random"
	"github.com/eliezedeck/gobase/validation"
	"github.com/eliezedeck/gobase/web"
	"github.com/eliezedeck/webhook-ingestor/core"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func setupAdministration(e *echo.Echo, config core.ConfigStorage, reqStore core.RequestsStorage, path string) {
	// TODO: Add a authentication middleware
	a := e.Group(path)

	// ---
	a.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello Admin")
	})

	// --- Webhooks: List
	a.GET("/webhooks", func(c echo.Context) error {
		webhooks, err := config.GetValidWebhooks()
		if err != nil {
			return err // HTTP 500
		}
		return c.JSON(http.StatusOK, webhooks)
	})

	// --- Webhook: Add
	a.POST("/webhooks", func(c echo.Context) error {
		webhook := &core.Webhook{}
		webhook.ID = fmt.Sprintf("w-%s", random.String(11))
		webhook.Enabled = true // enabled by default
		if _, err := validation.ValidateJSONBody(c.Request().Body, webhook); err != nil {
			return web.BadRequestError(c, "Invalid JSON body")
		}

		// Set IDs for each of the new forward URLs
		for _, furl := range webhook.ForwardUrls {
			furl.ID = fmt.Sprintf("f-%s", random.String(11))
		}

		if err := config.AddWebhook(webhook); err != nil {
			return web.BadRequestError(c, err.Error())
		}

		// Immediately register the route so that it's available for requests
		if err := webhook.RegisterWithEcho(e, reqStore); err != nil {
			return err // HTTP  500
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

	// --- Requests: Replay
	a.POST("/requests/replay", func(c echo.Context) error {
		// We only allow replay to a single Forward URL, pretty much any URL that's already registered
		wreq := struct {
			RequestId       string `json:"requestId" validate:"required"`
			WebhookId       string `json:"webhookId" validate:"required"`
			ForwardUrlId    string `json:"forwardUrlId" validate:"required"`
			DeleteOnSuccess int    `json:"deleteOnSuccess"`
		}{}
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
		var furl *core.ForwardUrl
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
			return err // HTTP 500
		}
		core.TransferHeaders(req.Header, oreq.Headers)

		// Execute the request
		response, err := core.ForwardHttpClient.Do(req)
		if err != nil {
			return err // HTTP 500
		}
		defer response.Body.Close()

		core.TransferHeaders(c.Response().Header(), response.Header)
		c.Response().WriteHeader(response.StatusCode)
		if _, err = io.Copy(c.Response(), response.Body); err != nil {
			return err // HTTP 500
		}
		return nil // success
	})

	logging.L.Info("Administration setup complete", zap.String("path", path))
}
