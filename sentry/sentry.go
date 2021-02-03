package sentry

import (
	"github.com/getsentry/sentry-go"
	"log"
)

type Sentry struct {
	merchantID string
	pointID    string
}

// NewSentry иницилизация сентри
func NewSentry(merchantID, pointID, revision string) *Sentry {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:     "https://821d9f7ed9794a248b2073e9fb81415a@sentry.pharmecosystem.ru/4",
		Release: revision,
		Debug:   true,
	})

	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	extra := make(map[string]interface{})
	extra["merchant_id"] = merchantID
	extra["point_id"] = pointID
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetExtras(extra)
	})

	return &Sentry{
		merchantID: merchantID,
		pointID:    pointID,
	}
}

func (s *Sentry) Message(msg string) {
	sentry.CaptureMessage(msg)
}

func (s *Sentry) Error(err error) {
	sentry.CaptureException(err)
}
