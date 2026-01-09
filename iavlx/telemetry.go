package iavlx

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var (
	meter = otel.Meter("comsos-sdk/iavlx")

	nodeCacheHitCounter metric.Int64Counter
	nodeReadLatency     metric.Float64Histogram
	operationTiming     metric.Float64Histogram
)

func init() {
	var err error
	nodeCacheHitCounter, err = meter.Int64Counter("node.cache_hit")
	if err != nil {
		panic(err)
	}
	nodeReadLatency, err = meter.Float64Histogram("node.read_latency")
	if err != nil {
		panic(err)
	}
	operationTiming, err = meter.Float64Histogram("operation.timing")
	if err != nil {
		panic(err)
	}
}
