package internal

import (
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var (
	tracer          = otel.Tracer("iavlx")
	meter           = otel.Meter("iavlx")
	logger          = otelslog.NewLogger("iavlx")
	leafHashLatency metric.Int64Histogram
	walWriteLatency metric.Int64Histogram
	queryLatency    metric.Int64Histogram
)

func init() {
	var err error

	leafHashLatency, err = meter.Int64Histogram("iavl_leaf_hash_latency_ms",
		metric.WithDescription("The amount of time in commit needed to wait for leaf hashing to complete, before hashing the root"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		panic("failed to create iavl_leaf_hash_latency_ms histogram: " + err.Error())
	}

	walWriteLatency, err = meter.Int64Histogram("iavl_wal_write_latency_ms",
		metric.WithDescription("The amount of time we had to wait in Commit before returning, due to WAL writes"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		panic("failed to create iavl_wal_write_latency_ms histogram: " + err.Error())
	}

	queryLatency, err = meter.Int64Histogram("iavl_query_latency_ms",
		metric.WithDescription("The amount of time we spent in the query method"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		panic("failed to create iavl_query_latency_ms histogram: " + err.Error())
	}
}
