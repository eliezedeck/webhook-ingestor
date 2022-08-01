package main

import (
	"net/http"

	"github.com/eliezedeck/gobase/logging"
	"github.com/eliezedeck/webhook-ingestor/core"
	"github.com/eliezedeck/webhook-ingestor/impl"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func main() {
	// Setup logging
	logging.Init()
	logging.L = logging.L.Named("WebhookIngestor")

	// ... can exit here if user is doing `-help`
	parseFlags()

	// Setup Web server (using Echo)
	e := echo.New()
	e.HidePort = true
	e.HideBanner = true
	e.Use(logging.ZapLoggerForEcho(logging.L))
	e.Use(logging.RecoverWithZapLogging)
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}
		if c.Request().Method == http.MethodHead {
			if err := c.NoContent(http.StatusInternalServerError); err != nil {
				logging.L.Error("Error while returning NoContent for HEAD request", zap.Error(err))
			}
			return
		}
		erro := c.JSON(http.StatusInternalServerError, 500)
		if erro != nil {
			e.Logger.Error(erro)
		}
	}

	// Setup MemoryStorage instance
	// TODO: This will be configurable in the future
	storage := impl.NewMemoryStorage("__admin__")

	// -----------
	setupWebhookPaths(e, storage, storage)

	// -----------
	// Set up the Admin paths
	path, err := storage.GetAdminPath()
	if err != nil {
		panic(err)
	}
	setupAdministration(e, storage, storage, path)

	panic(e.Start(listen))
}

func setupWebhookPaths(e *echo.Echo, config core.ConfigStorage, reqStore core.RequestsStorage) {
	w, err := config.GetValidWebhooks()
	if err != nil {
		panic(err)
	}

	for _, webhook := range w {
		if err = webhook.RegisterWithEcho(e, reqStore); err != nil {
			panic(err)
		}
	}
}
