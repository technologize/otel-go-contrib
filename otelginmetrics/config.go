package otelginmetrics

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
)

type RequestURLLabelMappingFn func(c *gin.Context) string
type config struct {
	recordInFlight bool
	recordSize     bool
	recordDuration bool
	groupedStatus  bool
	attributes     []attribute.KeyValue
	recorder       Recorder
}

func defaultConfig() *config {
	return &config{
		recordInFlight: true,
		recordDuration: true,
		recordSize:     true,
		groupedStatus:  true,
	}
}
