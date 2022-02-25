package types

import (
	"fmt"
)

// Pruning option string constants
const (
	PruningOptionDefault    = "default"
	PruningOptionEverything = "everything"
	PruningOptionNothing    = "nothing"
	PruningOptionCustom     = "custom"
)

var (
	// PruneDefault defines a pruning strategy where the last 362880 heights are
	// kept in addition to every 100th and where to-be pruned heights are pruned
	// at every 10th height. The last 362880 heights are kept assuming the typical
	// block time is 5s and typical unbonding period is 21 days. If these values
	// do not match the applications' requirements, use the "custom" option.
	PruneDefault = NewPruningOptions(362880, 10)

	// PruneEverything defines a pruning strategy where all committed heights are
	// deleted, storing only the current and previous height and where to-be pruned
	// heights are pruned at every 10th height.
	PruneEverything = NewPruningOptions(2, 10)

	// PruneNothing defines a pruning strategy where all heights are kept on disk.
	PruneNothing = NewPruningOptions(0, 0)
)

// PruningOptions defines the pruning strategy used when determining which
// heights are removed from disk when committing state.
type PruningOptions struct {
	// KeepRecent defines how many recent heights to keep on disk.
	KeepRecent uint64

	// Interval defines when the pruned heights are removed from disk.
	Interval uint64
}

func NewPruningOptions(keepRecent, interval uint64) PruningOptions {
	return PruningOptions{
		KeepRecent: keepRecent,
		Interval:   interval,
	}
}

func (po PruningOptions) Validate() error {
	if po.KeepRecent > 0 && po.Interval == 0 {
		return fmt.Errorf("invalid 'Interval' when pruning recent heights: %d", po.Interval)
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
