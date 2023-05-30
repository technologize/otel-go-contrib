package otelhttpmetrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

const instrumentationName = "github.com/technologize/otel-go-contrib/otelhttpmetrics"

// Recorder knows how to record and measure the metrics. This
// has the required methods to be used with the HTTP
// middlewares.
type otelRecorder struct {
	attemptsCounter       instrument.Int64UpDownCounter
	totalDuration         instrument.Int64Histogram
	activeRequestsCounter instrument.Int64UpDownCounter
	requestSize           instrument.Int64Histogram
	responseSize          instrument.Int64Histogram
}

func GetRecorder(metricsPrefix string) Recorder {
	metricName := func(metricName string) string {
		if len(metricsPrefix) > 0 {
			return metricsPrefix + "." + metricName
		}
		return metricName
	}
	meter := global.MeterProvider().Meter(instrumentationName, metric.WithInstrumentationVersion(SemVersion()))
	attemptsCounter, _ := meter.Int64UpDownCounter(metricName("http.client.request_count"), instrument.WithDescription("Number of Requests"))
	totalDuration, _ := meter.Int64Histogram(metricName("http.client.duration"), instrument.WithDescription("Time Taken by request"), instrument.WithUnit("ms"))
	activeRequestsCounter, _ := meter.Int64UpDownCounter(metricName("http.client.active_requests"), instrument.WithDescription("Number of requests inflight"))
	requestSize, _ := meter.Int64Histogram(metricName("http.client.request_content_length"), instrument.WithDescription("Request Size"), instrument.WithUnit("by"))
	responseSize, _ := meter.Int64Histogram(metricName("http.client.response_content_length"), instrument.WithDescription("Response Size"), instrument.WithUnit("by"))
	return &otelRecorder{
		attemptsCounter:       attemptsCounter,
		totalDuration:         totalDuration,
		activeRequestsCounter: activeRequestsCounter,
		requestSize:           requestSize,
		responseSize:          responseSize,
	}
}

// AddRequests increments the number of requests being processed.
func (r *otelRecorder) AddRequests(ctx context.Context, quantity int64, attributes []attribute.KeyValue) {
	r.attemptsCounter.Add(ctx, quantity, attributes...)
}

// ObserveHTTPRequestDuration measures the duration of an HTTP request.
func (r *otelRecorder) ObserveHTTPRequestDuration(ctx context.Context, duration time.Duration, attributes []attribute.KeyValue) {
	r.totalDuration.Record(ctx, int64(duration/time.Millisecond), attributes...)
}

// ObserveHTTPRequestSize measures the size of an HTTP request in bytes.
func (r *otelRecorder) ObserveHTTPRequestSize(ctx context.Context, sizeBytes int64, attributes []attribute.KeyValue) {
	r.requestSize.Record(ctx, sizeBytes, attributes...)
}

// ObserveHTTPResponseSize measures the size of an HTTP response in bytes.
func (r *otelRecorder) ObserveHTTPResponseSize(ctx context.Context, sizeBytes int64, attributes []attribute.KeyValue) {
	r.responseSize.Record(ctx, sizeBytes, attributes...)
}

// AddInflightRequests increments and decrements the number of inflight request being processed.
func (r *otelRecorder) AddInflightRequests(ctx context.Context, quantity int64, attributes []attribute.KeyValue) {
	r.activeRequestsCounter.Add(ctx, quantity, attributes...)
}
