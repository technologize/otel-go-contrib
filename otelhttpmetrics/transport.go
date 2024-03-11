package otelhttpmetrics

import (
	"net/http"
	"time"

	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

type transport struct {
	rt  http.RoundTripper
	cfg *config
}

func NewTransport(base http.RoundTripper, options ...Option) *transport {
	if base == nil {
		base = http.DefaultTransport
	}

	cfg := defaultConfig()
	for _, option := range options {
		option.apply(cfg)
	}
	if cfg.recorder == nil {
		cfg.recorder = GetRecorder("")
	}

	t := transport{
		rt:  base,
		cfg: cfg,
	}

	return &t
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	start := time.Now()
	cfg := t.cfg
	recorder := cfg.recorder
	if !cfg.shouldRecord(r) {
		return t.rt.RoundTrip(r)
	}
	reqAttributes := cfg.attributes(r)

	if cfg.recordInFlight {
		recorder.AddInflightRequests(r.Context(), 1, reqAttributes)
		defer recorder.AddInflightRequests(r.Context(), -1, reqAttributes)
	}

	res, err := t.rt.RoundTrip(r)

	defer func() {
		if err != nil {
			return
		}

		resAttributes := append(reqAttributes[0:0], reqAttributes...)

		if cfg.groupedStatus {
			code := int(res.StatusCode/100) * 100
			resAttributes = append(resAttributes, semconv.HTTPStatusCodeKey.Int(code))
		} else {
			resAttributes = append(resAttributes, semconv.HTTPAttributesFromHTTPStatusCode(res.StatusCode)...)
		}

		recorder.AddRequests(r.Context(), 1, resAttributes)

		if cfg.recordSize {
			requestSize := computeApproximateRequestSize(r)
			recorder.ObserveHTTPRequestSize(r.Context(), requestSize, resAttributes)
			recorder.ObserveHTTPResponseSize(r.Context(), int64(res.ContentLength), resAttributes)
		}

		if cfg.recordDuration {
			recorder.ObserveHTTPRequestDuration(r.Context(), time.Since(start), resAttributes)
		}
	}()
	return res, err
}

func computeApproximateRequestSize(r *http.Request) int64 {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return int64(s)
}
