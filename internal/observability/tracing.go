package observability

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	tracer           trace.Tracer
	isTracingEnabled bool
)

// InitTracer initializes the tracer with the given service name, logger, and configuration
func InitTracer(ctx context.Context, serviceName string, logger *logrus.Logger) (func(), error) {
	isTracingEnabled = true
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		logger.WithError(err).Error("Failed to create resource")
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	exporter, err := otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint("localhost:4317"),
			otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
			otlptracegrpc.WithTimeout(5*time.Second),
		),
	)
	if err != nil {
		logger.WithError(err).Error("Failed to create OTLP exporter")
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(tracerProvider)
	tracer = tracerProvider.Tracer(serviceName)

	logger.Info("Tracer initialized successfully")

	return func() {
		if !isTracingEnabled {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tracerProvider.Shutdown(ctx); err != nil {
			logger.WithError(err).Error("Error shutting down tracer provider")
		}
	}, nil
}

// StartSpan starts a new span with the given name
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	if !isTracingEnabled {
		return ctx, trace.SpanFromContext(ctx)
	}
	return tracer.Start(ctx, name)
}

// AddAttribute adds an attribute to the span in the given context
func AddAttribute(ctx context.Context, key string, value interface{}) {
	if !isTracingEnabled {
		return
	}
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", value)))
}
