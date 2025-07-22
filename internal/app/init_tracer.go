package app

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"log"
)

func initTracer(ctx context.Context) func() {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(
		jaeger.WithEndpoint("http://localhost:14268/api/traces"),
	))
	if err != nil {
		log.Fatalf("failed to initialize Jaeger exporter: %v", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("pvz-app"),
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown TracerProvider: %v", err)
		}
	}
}
