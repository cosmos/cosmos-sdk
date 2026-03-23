package tree

import (
	"context"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry/registry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const (
	TreeName       = "blockstm_btree"
	TreeScopeName  = "github.com/cosmos/cosmos-sdk/internal/blockstm/tree"
	TreeTimingUnit = "ms"
)

var treeInst *treeInstrument

func init() {
	registry.Register(&treeInstrument{})
}

type treeInstrument struct {
	Meter        metric.Meter
	Get          metric.Int64Histogram
	Set          metric.Int64Histogram
	Delete       metric.Int64Histogram
	ReverseSeek  metric.Int64Histogram
	Scan         metric.Int64Histogram
	GetOrDefault metric.Int64Histogram
}

func (i *treeInstrument) Name() string { return TreeName }

func (i *treeInstrument) Start(cfg map[string]any) error {
	i.Meter = otel.GetMeterProvider().Meter(
		TreeScopeName,
	)

	var err error
	i.Get, err = i.Meter.Int64Histogram(
		"get",
		metric.WithDescription("Time to look up an item in the B-tree"),
		metric.WithUnit(TreeTimingUnit),
	)
	if err != nil {
		return err
	}
	i.Set, err = i.Meter.Int64Histogram(
		"set",
		metric.WithDescription("Time to insert or update an item in the B-tree via copy-on-write"),
		metric.WithUnit(TreeTimingUnit),
	)
	if err != nil {
		return err
	}
	i.Delete, err = i.Meter.Int64Histogram(
		"delete",
		metric.WithDescription("Time to delete an item from the B-tree via copy-on-write"),
		metric.WithUnit(TreeTimingUnit),
	)
	if err != nil {
		return err
	}
	i.ReverseSeek, err = i.Meter.Int64Histogram(
		"reverse_seek",
		metric.WithDescription("Time to find the first item less than or equal to a pivot in the B-tree"),
		metric.WithUnit(TreeTimingUnit),
	)
	if err != nil {
		return err
	}
	i.Scan, err = i.Meter.Int64Histogram(
		"scan",
		metric.WithDescription("Time to scan all items in the B-tree in ascending order"),
		metric.WithUnit(TreeTimingUnit),
	)
	if err != nil {
		return err
	}
	i.GetOrDefault, err = i.Meter.Int64Histogram(
		"get_or_default",
		metric.WithDescription("Time to look up an item in the B-tree, inserting a default if not found"),
		metric.WithUnit(TreeTimingUnit),
	)
	if err != nil {
		return err
	}

	treeInst = i
	return nil
}

func measureSince(ctx context.Context, get func() metric.Int64Histogram, start time.Time) {
	if treeInst == nil {
		return
	}
	get().Record(ctx, time.Since(start).Milliseconds())
}
