package core

import (
	"net/http"
	"strings"
	"time"
)

type ForwardUrl struct {
	ID                     string        `bson:"_id"                     json:"id"`
	Url                    string        `bson:"url"                     json:"url"                      validate:"required"`
	KeepSuccessfulRequests int           `bson:"keepSuccessfulRequests"  json:"keepSuccessfulRequests"`
	Timeout                time.Duration `bson:"timeout"                 json:"timeout"                  validate:"required"`
	ReturnAsResponse       int           `bson:"returnAsResponse"        json:"returnAsResponse"         validate:"required"`
	WaitTillCompletion     int           `bson:"waitTillCompletion"      json:"waitForCompletion"        validate:"required"`
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
