package main

import (
	"net/http"

	"github.com/eliezedeck/gobase/logging"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func main() {
	// Setup logging
	logging.Init()
	logging.L = logging.L.Named("WebhookIngestor")

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

	panic(e.Start(":8080"))
}
