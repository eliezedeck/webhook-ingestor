package main

import (
	"net/http"

	"github.com/eliezedeck/gobase/logging"
	"github.com/eliezedeck/webhook-ingestor/core"
	"github.com/eliezedeck/webhook-ingestor/impl"
	mongodbimpl "github.com/eliezedeck/webhook-ingestor/impl/mongodb"
	"github.com/eliezedeck/webhook-ingestor/parameters"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func main() {
	// Setup logging
	logging.Init()
	logging.L = logging.L.Named("WebhookIngestor")

	// ... can exit here if user is doing `-help`
	parameters.ParseFlags()

	// Setup Web server (using Echo)
	e := buildEcho()

	// Setup MemoryStorage instance
	var (
		configStorage core.ConfigStorage
		reqStorage    core.RequestsStorage
	)
	switch parameters.ParamStorage {
	case "memory":
		storage := impl.NewMemoryStorage()
		configStorage = storage
		reqStorage = storage
		logging.L.Info("Using in-memory storage")
	case "mongo":
		storage, err := mongodbimpl.NewStorage(parameters.ParamStorageMongoUri, parameters.ParamStorageMongoDb)
		if err != nil {
			panic(err)
		}
		configStorage = storage
		reqStorage = storage
		logging.L.Info("Using MongoDB as storage")
	default:
		panic("invalid -storage parameter, valid values are 'memory' and 'mongo'")
	}

	// -----------
	setupWebhookPaths(e, configStorage, reqStorage)

	// -----------
	// Set up the Admin paths
	if parameters.ParamListen == parameters.ParamAdminListen {
		core.SetupAdministration(e, configStorage, reqStorage, parameters.ParamAdminPath)
	} else {
		a := buildEcho()
		core.SetupAdministration(a, configStorage, reqStorage, parameters.ParamAdminPath)
		go func() {
			panic(a.Start(parameters.ParamListen))
		}()
	}
	panic(e.Start(parameters.ParamListen))
}

func buildEcho() *echo.Echo {
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
	return e
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
