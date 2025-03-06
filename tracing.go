package bobotel

import (
	"context"
	"fmt"
	"os"
	"sync"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

var (
	lock                   sync.RWMutex
	singletonTraceProvider trace.TracerProvider
)

// NewTracer creates an open-telemetry tracer with the given name and options. NewTracer must be called after
// InitializeTraceProvider in order to not receive a no-op tracer.
func NewTracer(tracerName string, options ...trace.TracerOption) trace.Tracer {
	lock.RLock()
	defer lock.RUnlock()

	if singletonTraceProvider != nil {
		return singletonTraceProvider.Tracer(tracerName, options...)
	} else {
		return NewNoopTracer(tracerName, options...)
	}
}

// NewNoopTracer creates a no-op tracer with the given name.
func NewNoopTracer(tracerName string, options ...trace.TracerOption) trace.Tracer {
	return noop.NewTracerProvider().Tracer(tracerName, options...)
}

// InitializeTraceProvider initializes an open-telemetry trace provider configured via the given TracerConfig.
func InitializeTraceProvider(c *TracerConfig) error {
	providerResource, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(c.AppName),
			semconv.ServiceInstanceIDKey.String(c.AppID),
		),
	)
	if err != nil {
		return fmt.Errorf("problem creating tracer provider resources: %w", err)
	}

	opts := []sdktrace.TracerProviderOption{sdktrace.WithResource(providerResource)}

	if len(c.OtelExporters) < 1 {
		lock.Lock()
		defer lock.Unlock()

		singletonTraceProvider = noop.NewTracerProvider()

		return nil
	}

	for _, exporter := range c.OtelExporters {
		switch exporter {
		case "console":
			consoleExporter, err := newConsoleExporter(c)
			if err != nil {
				return fmt.Errorf("problem creating tracer console exporter: %w", err)
			}

			opts = append(opts, sdktrace.WithBatcher(consoleExporter))
		case "otlp":
			otlpExporter, err := newOtlpExporter(c)
			if err != nil {
				return fmt.Errorf("problem creating tracer otlp exporter: %w", err)
			}

			opts = append(opts, sdktrace.WithBatcher(otlpExporter))
		default:
			return fmt.Errorf("unsupported exporter found: %s", exporter)
		}
	}

	lock.Lock()
	defer lock.Unlock()

	singletonTraceProvider = sdktrace.NewTracerProvider(opts...)

	return nil
}

// ShutdownTraceProvider ...
func ShutdownTraceProvider(ctx context.Context) error {
	lock.Lock()
	defer lock.Unlock()

	if sdkTraceProvider, ok := singletonTraceProvider.(*sdktrace.TracerProvider); ok {
		_ = sdkTraceProvider.ForceFlush(ctx)

		if err := sdkTraceProvider.Shutdown(ctx); err != nil {
			return fmt.Errorf("problem shutting down trace provider: %w", err)
		}

		return nil
	}

	return nil
}

// RecordError is a helper function that attaches an error to a span.
func RecordError(span trace.Span, err error) {
	if span == nil || !span.IsRecording() {
		return
	}

	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

func newConsoleExporter(c *TracerConfig) (sdktrace.SpanExporter, error) {
	if c.OtelConsoleFormat == "production" {
		return stdouttrace.New(
			stdouttrace.WithWriter(os.Stdout),
		)
	}

	return stdouttrace.New(
		stdouttrace.WithWriter(os.Stdout),
		stdouttrace.WithPrettyPrint(),
	)
}

func newOtlpExporter(c *TracerConfig) (sdktrace.SpanExporter, error) {
	// NOTE: default http port is 4318, default grpc port is 4317
	var exporter sdktrace.SpanExporter
	var err error

	switch c.OtlpEndpointKind {
	case "http":
		exporter, err = otlptracehttp.New(
			context.Background(),
			otlptracehttp.WithEndpoint(fmt.Sprintf("%s:%d", c.OtlpHost, c.OtlpPort)),
		)
	case "grpc":
		exporter, err = otlptracegrpc.New(
			context.Background(),
			otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:%d", c.OtlpHost, c.OtlpPort)),
		)
	default:
		return nil, fmt.Errorf("unsupported otlp endpoint kind: %s", c.OtlpEndpointKind)
	}

	if err != nil {
		return nil, fmt.Errorf("problem creating otlp exporter: %w", err)
	}

	return exporter, nil
}
