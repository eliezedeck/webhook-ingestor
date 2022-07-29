package main

import (
	"fmt"
	"net/http"

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
}
