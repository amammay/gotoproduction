package main

import (
	"context"
	"fmt"
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/amammay/propagationgcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func initTracing(ctx context.Context, projectID string) (func(), error) {

	exporter, err := texporter.NewExporter(texporter.WithProjectID(projectID))
	if err != nil {
		return nil, fmt.Errorf("texporter.NewExporter(): %v", err)
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
	otel.SetTracerProvider(tp)

	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
		propagationgcp.HTTPFormat{},
	)
	otel.SetTextMapPropagator(propagator)
	return func() {
		defer tp.ForceFlush(ctx) // flushes any pending spans
	}, nil
}
