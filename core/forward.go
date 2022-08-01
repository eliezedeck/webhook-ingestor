package core

import (
	"net/http"
	"strings"
	"time"
)

type ForwardUrl struct {
	ID                     string        `json:"id"`
	Url                    string        `json:"url"                      validate:"required"`
	KeepSuccessfulRequests bool          `json:"keepSuccessfulRequests"`
	Timeout                time.Duration `json:"timeout"                  validate:"required"`
	ReturnAsResponse       int           `json:"returnAsResponse"         validate:"required"`
	WaitTillCompletion     int           `json:"waitForCompletion"        validate:"required"`
}

var (
	ForwardHttpClient = &http.Client{}
)

func TransferHeaders(dest, source http.Header) {
	for key, oheader := range source {
		if strings.ToLower(key) == "host" {
			continue
		}
		for _, h := range oheader {
			dest.Add(key, h)
		}
	}
}
