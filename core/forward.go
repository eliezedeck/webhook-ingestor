package core

import (
	"net/http"
	"strings"
	"time"
)

type ForwardUrl struct {
	ID                     string        `json:"id"`
	Url                    string        `json:"url"`
	KeepSuccessfulRequests bool          `json:"keepSuccessfulRequests"`
	Timeout                time.Duration `json:"timeout"`
	ReturnAsResponse       bool          `json:"returnAsResponse"`
	WaitTillCompletion     bool          `json:"waitForCompletion"`
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
