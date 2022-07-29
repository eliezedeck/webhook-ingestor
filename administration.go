package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/eliezedeck/gobase/random"
	"github.com/eliezedeck/gobase/validation"
	"github.com/eliezedeck/gobase/web"
	"github.com/eliezedeck/webhook-ingestor/structs"
	"github.com/labstack/echo/v4"
)

func setupAdministration(e *echo.Echo, config structs.ConfigStorage, reqStore structs.RequestsStorage, path string) {
	// TODO: Add a authentication middleware
	a := e.Group(path)

	// ---
	a.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello Admin")
	})

	// --- Webhook: Add
	a.POST("/webhooks", func(c echo.Context) error {
		webhook := &structs.Webhook{}
		webhook.ID = fmt.Sprintf("w-%s", random.String(11))
		webhook.Enabled = true // enabled by default
		if _, err := validation.ValidateJSONBody(c.Request().Body, webhook); err != nil {
			return web.BadRequestError(c, "Invalid JSON body")
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
		wreq := struct {
			RequestId       string `json:"requestId" validate:"required"`
			WebhookId       string `json:"webhookId" validate:"required"`
			DeleteOnSuccess int    `json:"deleteOnSuccess"`
		}{}
		if _, err := validation.ValidateJSONBody(c.Request().Body, &wreq); err != nil {
			return web.BadRequestError(c, "Invalid JSON body")
		}

		// Get the request and the webhook
		oreq, err := reqStore.GetRequest(wreq.RequestId)
		if err != nil {
			return err
		}
		webhook, err := config.GetWebhook(wreq.WebhookId)
		if err != nil {
			return err
		}

		// Send the request to all the Forward URLs set in the Webhook's configuration
		wg := &sync.WaitGroup{}
		for _, furl := range webhook.ForwardUrls {
			if furl.WaitTillCompletion {
				wg.Add(1)
			}
			go func(furl *structs.ForwardUrl) {
				defer func() {
					if furl.WaitTillCompletion {
						wg.Done()
					}
				}()

				// Craft the request based on the saved Request instance
				req, err := http.NewRequest(oreq.Method, furl.Url, strings.NewReader(oreq.Body))

				// FIXME: implement the rest ...
			}(furl)
		}

		return c.String(http.StatusBadRequest, "Not yet implemented")
	})
}
