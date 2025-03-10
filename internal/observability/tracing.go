package observability

import (
	"context"
	"fmt"

	"github.com/shaharia-lab/goai/observability"

	"github.com/shaharia-lab/mcp-kit/internal/config"

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

var tracer trace.Tracer
var isEnabled bool

// TracingService handles all tracing operations
type TracingService struct {
	config    config.TracingConfig
	logger    observability.Logger
	tracer    trace.Tracer
	provider  *sdktrace.TracerProvider
	isEnabled bool
}

// NewTracingService creates a new instance of TracingService
func NewTracingService(config config.TracingConfig, logger observability.Logger) *TracingService {
	return &TracingService{
		config:    config,
		logger:    logger,
		isEnabled: config.Enabled,
	}
}

// Initialize sets up the tracing infrastructure
func (ts *TracingService) Initialize(ctx context.Context) error {
	if !ts.isEnabled {
		ts.logger.Info("Tracing is disabled")
		return nil
	}

	// Create resource with service information and custom attributes
	resourceAttrs := []attribute.KeyValue{
		semconv.ServiceName(ts.config.ServiceName),
		semconv.DeploymentEnvironment(ts.config.Environment),
	}

	res, err := resource.New(ctx, resource.WithAttributes(resourceAttrs...))
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// Configure OTLP exporter
	exporter, err := ts.createExporter(ctx)
	if err != nil {
		return fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Configure batch span processor
	bsp := sdktrace.NewBatchSpanProcessor(
		exporter,
		sdktrace.WithBatchTimeout(ts.config.BatchTimeout),
	)

	// Create tracer provider
	ts.provider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(ts.config.SamplingRate)),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(ts.provider)
	tracer = ts.provider.Tracer(ts.config.ServiceName)
	isEnabled = true

	ts.logger.Info("Tracing initialized successfully")
	return nil
}

// createExporter creates and configures the OTLP exporter
func (ts *TracingService) createExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	return otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(ts.config.EndpointAddress),
			otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
			otlptracegrpc.WithTimeout(ts.config.Timeout),
		),
	)
}

// StartSpan starts a new span with the given name
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	if !isEnabled {
		return ctx, trace.SpanFromContext(ctx)
	}

	return tracer.Start(ctx, name)
}

// AddAttribute adds an attribute to the current span
func AddAttribute(ctx context.Context, key string, value interface{}) {
	if !isEnabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", value)))
}

// Shutdown gracefully shuts down the tracing service
func (ts *TracingService) Shutdown(ctx context.Context) error {
	if !ts.isEnabled || ts.provider == nil {
		return nil
	}

	return ts.provider.Shutdown(ctx)
}

// SetEnabled enables or disables tracing at runtime
func (ts *TracingService) SetEnabled(enabled bool) {
	ts.isEnabled = enabled
}

// IsEnabled returns whether tracing is currently enabled
func (ts *TracingService) IsEnabled() bool {
	return ts.isEnabled
}
