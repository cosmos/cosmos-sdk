package types

import "fmt"

// Pruning option string constants
const (
	PruningOptionDefault    = "default"
	PruningOptionEverything = "everything"
	PruningOptionNothing    = "nothing"
	PruningOptionCustom     = "custom"
)

var (
	// PruneDefault defines a pruning strategy where the last 100,000 heights are
	// kept where to-be pruned heights are pruned at every 10th height.
	// The last 100000 heights are kept(approximately 1 week worth of state) assuming the typical
	// block time is 6s. If these values
	// do not match the applications' requirements, use the "custom" option.
	PruneDefault = NewPruningOptions(100_000, 0, 100)

	// PruneEverything defines a pruning strategy where all committed heights are
	// deleted, storing only the current height and last 10 states. To-be pruned heights are
	// pruned at every 10th height.
	PruneEverything = NewPruningOptions(10, 0, 10)

	// PruneNothing defines a pruning strategy where all heights are kept on disk.
	PruneNothing = NewPruningOptions(0, 1, 0)
)

// PruningOptions defines the pruning strategy used when determining which
// heights are removed from disk when committing state.
type PruningOptions struct {
	// KeepRecent defines how many recent heights to keep on disk.
	KeepRecent uint64

	// KeepEvery defines how many offset heights are kept on disk past KeepRecent.
	KeepEvery uint64

	// Interval defines when the pruned heights are removed from disk.
	Interval uint64
}

func NewPruningOptions(keepRecent, keepEvery, interval uint64) PruningOptions {
	return PruningOptions{
		KeepRecent: keepRecent,
		KeepEvery:  keepEvery,
		Interval:   interval,
	}
}

func (po PruningOptions) Validate() error {
	if po.KeepEvery == 0 && po.Interval == 0 {
		return fmt.Errorf("invalid 'Interval' when pruning everything: %d", po.Interval)
	}
	if po.KeepEvery == 1 && po.Interval != 0 { // prune nothing
		return fmt.Errorf("invalid 'Interval' when pruning nothing: %d", po.Interval)
	}
	if po.KeepEvery > 1 && po.Interval == 0 {
		return fmt.Errorf("invalid 'Interval' when pruning: %d", po.Interval)
	}

	return nil
}

func NewPruningOptionsFromString(strategy string) PruningOptions {
	switch strategy {
	case PruningOptionEverything:
		return PruneEverything

	case PruningOptionNothing:
		return PruneNothing

	case PruningOptionDefault:
		return PruneDefault

	default:
		return PruneDefault
	}
}
