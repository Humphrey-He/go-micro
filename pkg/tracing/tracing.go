package tracing

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go-micro/pkg/config"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func Init(serviceName string) (func(context.Context) error, error) {
	enabled := config.GetEnv("OTEL_ENABLED", "false")
	if enabled != "true" {
		tracer = otel.Tracer(serviceName)
		return func(ctx context.Context) error { return nil }, nil
	}

	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			attribute.String("environment", config.GetEnv("APP_ENV", "development")),
		)),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer = tp.Tracer(serviceName)

	return tp.Shutdown, nil
}

func Tracer() trace.Tracer {
	if tracer == nil {
		tracer = otel.Tracer("go-micro")
	}
	return tracer
}

func SpanName(operation string) string {
	return fmt.Sprintf("%s", operation)
}

func AddSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

func RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}

func ContextWithTraceID(ctx context.Context) context.Context {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		return trace.ContextWithSpan(ctx, span)
	}
	return ctx
}

func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

func Middleware(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}
