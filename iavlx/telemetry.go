package iavlx

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	meter = otel.Meter("comsos-sdk/iavlx")

	nodeCacheHitCounter metric.Int64Counter
	nodeReadLatency     metric.Int64Histogram
	operationTiming     metric.Int64Histogram
)

func init() {
	var err error
	nodeCacheHitCounter, err = meter.Int64Counter("node.cache_hit")
	if err != nil {
		panic(err)
	}
	nodeReadLatency, err = meter.Int64Histogram("node.read_latency")
	if err != nil {
		panic(err)
	}
	operationTiming, err = meter.Int64Histogram("operation.timing")
	if err != nil {
		panic(err)
	}
}

func recordOperationTiming(start time.Time, op string) {
	operationTiming.Record(context.Background(), time.Since(start).Milliseconds(), metric.WithAttributes(attribute.String("op", op)))
}
