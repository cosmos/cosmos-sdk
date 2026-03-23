package blockstm

import (
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

func init() {
	registry.Register(&instrument{})
}

type instrument struct {
	Meter metric.Meter
	// MVData metrics
	MVDataRead        metric.Int64ObservableCounter
	MVDataConsolidate metric.Int64ObservableCounter
	// MVView metrics
	MVViewReadWriteSet    metric.Int64ObservableCounter
	MVViewReadMVData      metric.Int64ObservableCounter
	MVViewReadStorage     metric.Int64ObservableCounter
	MVViewWrite           metric.Int64ObservableCounter
	MVViewDelete          metric.Int64ObservableCounter
	MVViewIteratorKeys    metric.Int64ObservableCounter
	MVViewIteratorKeysCnt metric.Int64Counter
	MVViewEstimateWait    metric.Int64ObservableCounter
	// Executor/Transaction metrics
	ExecutedTxs        metric.Int64Counter
	ValidatedTxs       metric.Int64Counter
	DecreaseCount      metric.Int64Counter
	TryExecuteTime     metric.Int64ObservableCounter
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
	i.MVDataRead, err = i.Meter.Int64ObservableCounter(
		"mvdata.read",
		metric.WithDescription(""),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.MVDataConsolidate, err = i.Meter.Int64ObservableCounter(
		"mvdata.consolidate",
		metric.WithDescription(""),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.MVViewReadWriteSet, err = i.Meter.Int64ObservableCounter(
		"mvdata.read.writeset",
		metric.WithDescription(""),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.MVViewReadMVData, err = i.Meter.Int64ObservableCounter(
		"mvview.read_mvdata",
		metric.WithDescription(""),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.MVViewReadStorage, err = i.Meter.Int64ObservableCounter(
		"mvview.read.storage",
		metric.WithDescription(""),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.MVViewWrite, err = i.Meter.Int64ObservableCounter(
		"mvview.write",
		metric.WithDescription(""),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.MVViewDelete, err = i.Meter.Int64ObservableCounter(
		"mvview.delete",
		metric.WithDescription(""),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.MVViewIteratorKeys, err = i.Meter.Int64ObservableCounter(
		"mvview.iterator.keys.read",
		metric.WithDescription(""),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.MVViewIteratorKeysCnt, err = i.Meter.Int64Counter(
		"mvview.iterator.keys.read.count",
		metric.WithDescription(""),
	)
	if err != nil {
		return err
	}
	i.MVViewEstimateWait, err = i.Meter.Int64ObservableCounter(
		"mvview.estimate.wait",
		metric.WithDescription(""),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.ExecutedTxs, err = i.Meter.Int64Counter(
		"executed.txs",
		metric.WithDescription(""),
	)
	if err != nil {
		return err
	}
	i.ValidatedTxs, err = i.Meter.Int64Counter(
		"validated.txs",
		metric.WithDescription(""),
	)
	if err != nil {
		return err
	}
	i.DecreaseCount, err = i.Meter.Int64Counter(
		"decrease.count",
		metric.WithDescription(""),
	)
	if err != nil {
		return err
	}
	i.DecreaseCount, err = i.Meter.Int64Counter(
		"decrease.count",
		metric.WithDescription(""),
	)
	if err != nil {
		return err
	}

	return err
}
