package otelginmetrics

import (
	"go.opentelemetry.io/otel/attribute"
)

// Option applies a configuration to the given config
type Option interface {
	apply(cfg *config)
}

type optionFunc func(cfg *config)

func (fn optionFunc) apply(cfg *config) {
	fn(cfg)
}

// WithAdditionalAttributes sets a list of attribute.KeyValue labels for all metrics associated with this round tripper
func WithAdditionalAttributes(attributes map[string]string) Option {
	return optionFunc(func(cfg *config) {
		attr := make([]attribute.KeyValue, 0, len(attributes))
		for k, v := range attributes {
			attr = append(attr, attribute.String(k, v))
		}
		cfg.attributes = attr
	})
}

// WithRecordInFlight determines whether to record In Flight Requests or not
// By default the recordInFlight is true
func WithRecordInFlightDisabled(recordInFlight bool) Option {
	return optionFunc(func(cfg *config) {
		cfg.recordInFlight = false
	})
}

// WithRecordDuration determines whether to record Duration of Requests or not
// By default the recordDuration is true
func WithRecordDurationDisabled(recordDuration bool) Option {
	return optionFunc(func(cfg *config) {
		cfg.recordDuration = false
	})
}

// WithRecordSize determines whether to record Size of Requests and Responses or not
// By default the recordSize is true
func WithRecordSizeDisabled(recordSize bool) Option {
	return optionFunc(func(cfg *config) {
		cfg.recordSize = false
	})
}

// WithGroupedStatus determines whether to group the response status codes or not. If true 2xx, 3xx will be stored
// By default the groupedStatus is true
func WithGroupedStatusDisabled() Option {
	return optionFunc(func(cfg *config) {
		cfg.groupedStatus = false
	})
}

func WithRecorder(recorder Recorder) Option {
	return optionFunc(func(cfg *config) {
		cfg.recorder = recorder
	})
}
