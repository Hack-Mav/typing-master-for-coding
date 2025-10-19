package telemetry

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer trace.Tracer
	meter  metric.Meter

	// Metrics
	requestCounter    metric.Int64Counter
	requestDuration   metric.Float64Histogram
	activeConnections metric.Int64UpDownCounter
	errorCounter      metric.Int64Counter
	cacheHitCounter   metric.Int64Counter
	cacheMissCounter  metric.Int64Counter
	dbQueryDuration   metric.Float64Histogram
)

// Config holds telemetry configuration
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Endpoint       string
	Enabled        bool
}

// Initialize sets up OpenTelemetry
func Initialize(cfg Config) (func(context.Context) error, error) {
	if !cfg.Enabled {
		log.Println("Telemetry disabled")
		return func(context.Context) error { return nil }, nil
	}

	ctx := context.Background()

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			attribute.String("environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Setup trace provider
	traceShutdown, err := setupTraceProvider(ctx, res, cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to setup trace provider: %w", err)
	}

	// Setup metric provider
	metricShutdown, err := setupMetricProvider(ctx, res, cfg.Endpoint)
	if err != nil {
		traceShutdown(ctx)
		return nil, fmt.Errorf("failed to setup metric provider: %w", err)
	}

	// Initialize tracer and meter
	tracer = otel.Tracer("typing-master-backend")
	meter = otel.Meter("typing-master-backend")

	// Initialize metrics
	if err := initializeMetrics(); err != nil {
		traceShutdown(ctx)
		metricShutdown(ctx)
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Return combined shutdown function
	shutdown := func(ctx context.Context) error {
		if err := traceShutdown(ctx); err != nil {
			log.Printf("Error shutting down trace provider: %v", err)
		}
		if err := metricShutdown(ctx); err != nil {
			log.Printf("Error shutting down metric provider: %v", err)
		}
		return nil
	}

	log.Println("Telemetry initialized successfully")
	return shutdown, nil
}

// setupTraceProvider creates and registers trace provider
func setupTraceProvider(ctx context.Context, res *resource.Resource, endpoint string) (func(context.Context) error, error) {
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(), // Use WithTLSCredentials in production
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

// setupMetricProvider creates and registers metric provider
func setupMetricProvider(ctx context.Context, res *resource.Resource, endpoint string) (func(context.Context) error, error) {
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithInsecure(), // Use WithTLSCredentials in production
	)
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(30*time.Second))),
		sdkmetric.WithResource(res),
	)

	otel.SetMeterProvider(mp)

	return mp.Shutdown, nil
}

// initializeMetrics creates all metric instruments
func initializeMetrics() error {
	var err error

	requestCounter, err = meter.Int64Counter(
		"http.server.request.count",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return err
	}

	requestDuration, err = meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithDescription("HTTP request duration"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return err
	}

	activeConnections, err = meter.Int64UpDownCounter(
		"http.server.active_connections",
		metric.WithDescription("Number of active HTTP connections"),
		metric.WithUnit("{connection}"),
	)
	if err != nil {
		return err
	}

	errorCounter, err = meter.Int64Counter(
		"http.server.error.count",
		metric.WithDescription("Total number of errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	cacheHitCounter, err = meter.Int64Counter(
		"cache.hit.count",
		metric.WithDescription("Number of cache hits"),
		metric.WithUnit("{hit}"),
	)
	if err != nil {
		return err
	}

	cacheMissCounter, err = meter.Int64Counter(
		"cache.miss.count",
		metric.WithDescription("Number of cache misses"),
		metric.WithUnit("{miss}"),
	)
	if err != nil {
		return err
	}

	dbQueryDuration, err = meter.Float64Histogram(
		"db.query.duration",
		metric.WithDescription("Database query duration"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return err
	}

	return nil
}

// StartSpan starts a new trace span
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tracer.Start(ctx, name, opts...)
}

// RecordRequest records an HTTP request metric
func RecordRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration) {
	attrs := metric.WithAttributes(
		attribute.String("http.method", method),
		attribute.String("http.route", path),
		attribute.Int("http.status_code", statusCode),
	)

	requestCounter.Add(ctx, 1, attrs)
	requestDuration.Record(ctx, float64(duration.Milliseconds()), attrs)
}

// RecordError records an error metric
func RecordError(ctx context.Context, errorType, operation string) {
	attrs := metric.WithAttributes(
		attribute.String("error.type", errorType),
		attribute.String("operation", operation),
	)

	errorCounter.Add(ctx, 1, attrs)
}

// RecordCacheHit records a cache hit
func RecordCacheHit(ctx context.Context, cacheType string) {
	attrs := metric.WithAttributes(
		attribute.String("cache.type", cacheType),
	)

	cacheHitCounter.Add(ctx, 1, attrs)
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss(ctx context.Context, cacheType string) {
	attrs := metric.WithAttributes(
		attribute.String("cache.type", cacheType),
	)

	cacheMissCounter.Add(ctx, 1, attrs)
}

// RecordDBQuery records a database query metric
func RecordDBQuery(ctx context.Context, operation string, duration time.Duration) {
	attrs := metric.WithAttributes(
		attribute.String("db.operation", operation),
	)

	dbQueryDuration.Record(ctx, float64(duration.Milliseconds()), attrs)
}

// IncrementActiveConnections increments active connection count
func IncrementActiveConnections(ctx context.Context) {
	activeConnections.Add(ctx, 1)
}

// DecrementActiveConnections decrements active connection count
func DecrementActiveConnections(ctx context.Context) {
	activeConnections.Add(ctx, -1)
}
