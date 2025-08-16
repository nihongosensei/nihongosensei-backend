package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/processors/minsev"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	otrace "go.opentelemetry.io/otel/trace"
)

// SetupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTraceProvider(ctx)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider(ctx)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	loggerProvider, err := newLoggerProvider(ctx)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	// Add bridge for otel to slog.
	handler := otelslog.NewHandler("default", otelslog.WithLoggerProvider(loggerProvider))
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(ctx context.Context) (*trace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}

	provider := trace.NewTracerProvider(trace.WithBatcher(exporter))
	return provider, nil
}

func newMeterProvider(ctx context.Context) (*metric.MeterProvider, error) {
	exporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return nil, err
	}

	provider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exporter)))
	return provider, nil
}

var getSeverity = sync.OnceValue(func() log.Severity {
	conv := map[string]log.Severity{
		"":      log.SeverityInfo, // Default to SeverityInfo for unset.
		"debug": log.SeverityDebug,
		"info":  log.SeverityInfo,
		"warn":  log.SeverityWarn,
		"error": log.SeverityError,
	}
	// log.SeverityUndefined for unknown values.
	return conv[strings.ToLower(os.Getenv("LOG_LEVEL"))]
})

type EnvSeverity struct{}

func (EnvSeverity) Severity() log.Severity { return getSeverity() }

func newLoggerProvider(ctx context.Context) (*logsdk.LoggerProvider, error) {
	exporter, err := otlploghttp.New(ctx)
	if err != nil {
		return nil, err
	}

	processor := minsev.NewLogProcessor(logsdk.NewBatchProcessor(exporter), EnvSeverity{})
	provider := logsdk.NewLoggerProvider(logsdk.WithProcessor(processor))

	return provider, nil
}

// responseWriter is a wrapper to capture HTTP status codes
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// TracingMiddleware records http errors to the current active span.
func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// Capture errors in response
		rw := &responseWriter{ResponseWriter: w}
		next.ServeHTTP(rw, r.WithContext(ctx))

		// Record the error in the span if an error status was returned
		if rw.statusCode >= http.StatusBadRequest {
			span := otrace.SpanFromContext(ctx)

			if span != nil {
				span.SetStatus(codes.Error, http.StatusText(rw.statusCode))
			}
		}
	})
}
