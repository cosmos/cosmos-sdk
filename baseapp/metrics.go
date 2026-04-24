package baseapp

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/cosmos/cosmos-sdk/telemetry/registry"
)

const (
	// InstrumentName is the baseapp instrument name used in telemetry config.
	InstrumentName = "baseapp"
	// ScopeName is the instrumentation scope name.
	ScopeName = "github.com/cosmos/cosmos-sdk/baseapp"
	// TimingUnit represents the unit of all timing measurements.
	TimingUnit = "ms"
)

var (
	tracer = otel.Tracer(ScopeName)
	inst   *instrument
)

func init() {
	registry.Register(&instrument{})
}

type instrument struct {
	Meter metric.Meter

	BlockCount              metric.Int64Counter
	TxCount                 metric.Int64Counter
	OEAborted               metric.Int64Counter
	OETime                  metric.Int64Histogram
	NonOEInternalFinalize   metric.Int64Histogram
	WorkingHashTime         metric.Int64Histogram
	StreamingListenerTime   metric.Int64Histogram
	OEAbortIfNeededTime     metric.Int64Histogram
	FinalizeBlockTime       metric.Int64Histogram
	InternalFinalizeTime    metric.Int64Histogram
	ExecuteWithExecutorTime metric.Int64Histogram
	GetFinalizeStateTime    metric.Int64Histogram
	PreBlockTime            metric.Int64Histogram
	BeginBlockTime          metric.Int64Histogram
	EndBlockTime            metric.Int64Histogram
}

func (i *instrument) Name() string { return InstrumentName }

func (i *instrument) Start(cfg map[string]any) error {
	i.Meter = otel.GetMeterProvider().Meter(ScopeName)

	var err error
	i.BlockCount, err = i.Meter.Int64Counter(
		"block.count",
		metric.WithDescription("Total number of committed blocks"),
	)
	if err != nil {
		return err
	}
	i.TxCount, err = i.Meter.Int64Counter(
		"tx.count",
		metric.WithDescription("Total number of successfully finalized transactions"),
	)
	if err != nil {
		return err
	}
	i.OEAborted, err = i.Meter.Int64Counter(
		"oe.aborted",
		metric.WithDescription("Total number of optimistic execution aborts"),
	)
	if err != nil {
		return err
	}
	i.OETime, err = i.Meter.Int64Histogram(
		"oe.time",
		metric.WithDescription("Time spent waiting for optimistic execution to finish"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.NonOEInternalFinalize, err = i.Meter.Int64Histogram(
		"finalize.non_oe_internal.time",
		metric.WithDescription("Time spent executing internalFinalizeBlock outside optimistic execution"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.WorkingHashTime, err = i.Meter.Int64Histogram(
		"working_hash.time",
		metric.WithDescription("Time spent computing the working app hash"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.StreamingListenerTime, err = i.Meter.Int64Histogram(
		"streaming_listener.time",
		metric.WithDescription("Time spent invoking streaming finalize block listeners"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.OEAbortIfNeededTime, err = i.Meter.Int64Histogram(
		"oe.abort_if_needed.time",
		metric.WithDescription("Time spent in optimistic execution abort checks"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.FinalizeBlockTime, err = i.Meter.Int64Histogram(
		"finalize.block.time",
		metric.WithDescription("Time spent in the FinalizeBlock ABCI method"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.InternalFinalizeTime, err = i.Meter.Int64Histogram(
		"internal_finalize.time",
		metric.WithDescription("Time spent in internalFinalizeBlock"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.ExecuteWithExecutorTime, err = i.Meter.Int64Histogram(
		"execute_with_executor.time",
		metric.WithDescription("Time spent executing transactions with the configured executor"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.GetFinalizeStateTime, err = i.Meter.Int64Histogram(
		"get_finalize_state.time",
		metric.WithDescription("Time spent getting or initializing finalize block state"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.PreBlockTime, err = i.Meter.Int64Histogram(
		"pre_block.time",
		metric.WithDescription("Time spent executing preBlock"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.BeginBlockTime, err = i.Meter.Int64Histogram(
		"begin_block.time",
		metric.WithDescription("Time spent executing beginBlock"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}
	i.EndBlockTime, err = i.Meter.Int64Histogram(
		"end_block.time",
		metric.WithDescription("Time spent executing endBlock"),
		metric.WithUnit(TimingUnit),
	)
	if err != nil {
		return err
	}

	inst = i
	return nil
}

func measureSince(ctx context.Context, get func() metric.Int64Histogram, start time.Time) {
	if inst == nil {
		return
	}
	get().Record(ctx, time.Since(start).Milliseconds())
}

func (app *BaseApp) metricsCtx() context.Context {
	if finalizeState := app.stateManager.GetState(execModeFinalize); finalizeState != nil {
		return finalizeState.Context()
	}
	return context.Background()
}
