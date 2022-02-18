package otelginmetrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/unit"
)

const instrumentationName = "github.com/technologize/otel-go-contrib/otelginmetrics"

// Recorder knows how to record and measure the metrics. This
// has the required methods to be used with the HTTP
// middlewares.
type otelRecorder struct {
	attemptsCounter       metric.Int64UpDownCounter
	totalDuration         metric.Int64Histogram
	activeRequestsCounter metric.Int64UpDownCounter
	requestSize           metric.Int64Histogram
	responseSize          metric.Int64Histogram
}

func GetRecorder(metricsPrefix string) Recorder {
	metricName := func(metricName string) string {
		if len(metricsPrefix) > 0 {
			return metricsPrefix + "." + metricName
		}
		return metricName
	}
	meter := global.Meter(instrumentationName, metric.WithInstrumentationVersion(SemVersion()))
	return &otelRecorder{
		attemptsCounter:       metric.Must(meter).NewInt64UpDownCounter(metricName("http.server.request_count"), metric.WithDescription("Number of Requests"), metric.WithUnit(unit.Dimensionless)),
		totalDuration:         metric.Must(meter).NewInt64Histogram(metricName("http.server.duration"), metric.WithDescription("Time Taken by request"), metric.WithUnit(unit.Milliseconds)),
		activeRequestsCounter: metric.Must(meter).NewInt64UpDownCounter(metricName("http.server.active_requests"), metric.WithDescription("Number of requests inflight"), metric.WithUnit(unit.Dimensionless)),
		requestSize:           metric.Must(meter).NewInt64Histogram(metricName("http.server.request_content_length"), metric.WithDescription("Request Size"), metric.WithUnit(unit.Bytes)),
		responseSize:          metric.Must(meter).NewInt64Histogram(metricName("http.server.response_content_length"), metric.WithDescription("Response Size"), metric.WithUnit(unit.Bytes)),
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
