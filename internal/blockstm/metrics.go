package blockstm

import (
	"context"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry/registry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const (
	// Name is the instrument name used in configuration.
	Name = "blockstm"
	// ScopeName is the instrumentation scope name.
	ScopeName = "github.com/cosmos/cosmos-sdk/internal/blockstm"
	// TimingUnit represents the unit of all timing measurements--milliseconds
	TimingUnit = "ms"
)

// inst is the package-level instrument instance, set during Start().
var inst *instrument

func init() {
	registry.Register(&instrument{})
}

type instrument struct {
	Meter metric.Meter
	// MVData metrics
	MVDataRead        metric.Int64Histogram
	MVDataConsolidate metric.Int64Histogram
	// MVView metrics
	MVViewReadWriteSet    metric.Int64Histogram
	MVViewReadMVData      metric.Int64Histogram
	MVViewReadStorage     metric.Int64Histogram
	MVViewWrite           metric.Int64Histogram
	MVViewDelete          metric.Int64Histogram
	MVViewIteratorKeys    metric.Int64Histogram
	MVViewIteratorKeysCnt metric.Int64Counter
	MVViewEstimateWait    metric.Int64Histogram
	// Executor/Transaction metrics
	ExecutedTxs        metric.Int64Counter
	ValidatedTxs       metric.Int64Counter
	DecreaseCount      metric.Int64Counter
	ExecutionRatio     metric.Float64Counter
	TryExecuteTime     metric.Int64Histogram
	TxReadCount        metric.Int64Counter
	TxWriteCount       metric.Int64Counter
	TxNewLocationWrite metric.Int64Counter
}

func (i *instrument) Name() string { return Name }

func (i *instrument) Start(cfg map[string]any) error {
	i.Meter = otel.GetMeterProvider().Meter(
		ScopeName,
	)

	var err error
	i.MVDataRead, err = i.Meter.Int64Histogram(
		"mvdata.read",
		metric.WithDescription("Time to read a value from multi-version data"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.MVDataConsolidate, err = i.Meter.Int64Histogram(
		"mvdata.consolidate",
		metric.WithDescription("Time to consolidate a transaction's write set into multi-version data"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.MVViewReadWriteSet, err = i.Meter.Int64Histogram(
		"mvdata.read.writeset",
		metric.WithDescription("Time to read a value from the transaction's own write set"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.MVViewReadMVData, err = i.Meter.Int64Histogram(
		"mvview.read_mvdata",
		metric.WithDescription("Time to read a value from another transaction's write set via multi-version data"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.MVViewReadStorage, err = i.Meter.Int64Histogram(
		"mvview.read.storage",
		metric.WithDescription("Time to read a value from the underlying storage when not found in multi-version data"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.MVViewWrite, err = i.Meter.Int64Histogram(
		"mvview.write",
		metric.WithDescription("Time to write a value to the transaction's write set"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.MVViewDelete, err = i.Meter.Int64Histogram(
		"mvview.delete",
		metric.WithDescription("Time to delete a key in the transaction's write set"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.MVViewIteratorKeys, err = i.Meter.Int64Histogram(
		"mvview.iterator.keys.read",
		metric.WithDescription("Time spent iterating over keys in the multi-version view"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.MVViewIteratorKeysCnt, err = i.Meter.Int64Counter(
		"mvview.iterator.keys.read.count",
		metric.WithDescription("Total number of keys read by iterators in the multi-version view"),
	)
	if err != nil {
		return err
	}
	i.MVViewEstimateWait, err = i.Meter.Int64Histogram(
		"mvview.estimate.wait",
		metric.WithDescription("Time spent waiting for a blocking transaction to complete due to an estimate marker"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.ExecutedTxs, err = i.Meter.Int64Counter(
		"executed.txs",
		metric.WithDescription("Total number of transaction executions including re-executions"),
	)
	if err != nil {
		return err
	}
	i.ValidatedTxs, err = i.Meter.Int64Counter(
		"validated.txs",
		metric.WithDescription("Total number of transaction validations"),
	)
	if err != nil {
		return err
	}
	i.DecreaseCount, err = i.Meter.Int64Counter(
		"decrease.count",
		metric.WithDescription("Total number of scheduler task count decreases"),
	)
	if err != nil {
		return err
	}
	i.ExecutionRatio, err = i.Meter.Float64Counter(
		"execution.ratio",
		metric.WithDescription("Ratio of total executions to block size, indicating re-execution overhead"),
	)
	if err != nil {
		return err
	}
	i.TryExecuteTime, err = i.Meter.Int64Histogram(
		"try.execute.time",
		metric.WithDescription("Time to execute a transaction and record its results"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.TxReadCount, err = i.Meter.Int64Counter(
		"tx.read.count",
		metric.WithDescription("Total number of keys read across all transaction executions"),
	)
	if err != nil {
		return err
	}
	i.TxWriteCount, err = i.Meter.Int64Counter(
		"tx.write.count",
		metric.WithDescription("Total number of keys written across all transaction executions"),
	)
	if err != nil {
		return err
	}
	i.TxNewLocationWrite, err = i.Meter.Int64Counter(
		"tx.new.location.write",
		metric.WithDescription("Total number of transaction executions that wrote to a new storage location"),
	)
	if err != nil {
		return err
	}

	inst = i
	return nil
}

// measureSince records the duration in milliseconds since start on the given histogram.
// Safe to call when inst is nil.
func measureSince(ctx context.Context, get func() metric.Int64Histogram, start time.Time) {
	if inst == nil {
		return
	}
	get().Record(ctx, time.Since(start).Milliseconds())
}
