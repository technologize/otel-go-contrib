package otelhttpmetrics

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

type config struct {
	recordInFlight bool
	recordSize     bool
	recordDuration bool
	groupedStatus  bool
	recorder       Recorder
	attributes     func(*http.Request) []attribute.KeyValue
	shouldRecord   func(*http.Request) bool
}

func defaultConfig() *config {
	return &config{
		recordInFlight: true,
		recordDuration: true,
		recordSize:     true,
		groupedStatus:  true,
		attributes:     DefaultAttributes,
		shouldRecord: func(_ *http.Request) bool {
			return true
		},
	}
}

var DefaultAttributes = func(request *http.Request) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		semconv.HTTPMethodKey.String(request.Method),
	}
	origin := request.Host
	target := request.URL.Path
	if origin != "" {
		attrs = append(attrs, semconv.HTTPHostKey.String(origin))
	}
	if target != "" {
		attrs = append(attrs, semconv.HTTPTargetKey.String(target))
	}
	return attrs
}
