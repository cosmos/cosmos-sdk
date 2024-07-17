package store

import (
	"errors"
	"fmt"
)

type PruningStrategy int

const (
	// PruningDefault defines a pruning strategy where the last 362880 heights are
	// kept where to-be pruned heights are pruned at every 10th height.
	// The last 362880 heights are kept(approximately 3.5 weeks worth of state) assuming the typical
	// block time is 6s. If these values do not match the applications' requirements, use the "custom" option.
	PruningDefault PruningStrategy = iota
	// PruningEverything defines a pruning strategy where all committed heights are
	// deleted, storing only the current height and last 2 states. To-be pruned heights are
	// pruned at every 10th height.
	PruningEverything
	// PruningNothing defines a pruning strategy where all heights are kept on disk.
	// This is the only stretegy where KeepEvery=1 is allowed with state-sync snapshots disabled.
	PruningNothing
)

// PruningOption defines the pruning configuration.
type PruningOption struct {
	// KeepRecent sets the number of recent versions to keep.
	KeepRecent uint64 `mapstructure:"keep-recent" toml:"keep-recent"`

	// Interval sets the number of how often to prune.
	// If set to 0, no pruning will be done.
	Interval uint64   `mapstructure:"interval" toml:"interval"`
}

// Pruning option string constants
const (
	PruningOptionDefault    = "default"
	PruningOptionEverything = "everything"
	PruningOptionNothing    = "nothing"
	PruningOptionCustom     = "custom"
)

const (
	pruneEverythingKeepRecent = 2
	pruneEverythingInterval   = 10
)

var (
	ErrPruningIntervalZero       = errors.New("'pruning-interval' must not be 0. If you want to disable pruning, select pruning = \"nothing\"")
	ErrPruningIntervalTooSmall   = fmt.Errorf("'pruning-interval' must not be less than %d. For the most aggressive pruning, select pruning = \"everything\"", pruneEverythingInterval)
	ErrPruningKeepRecentTooSmall = fmt.Errorf("'pruning-keep-recent' must not be less than %d. For the most aggressive pruning, select pruning = \"everything\"", pruneEverythingKeepRecent)
)

// NewPruningOption returns a new PruningOption instance based on the given pruning strategy.
func NewPruningOption(pruningStrategy PruningStrategy) *PruningOption {
	switch pruningStrategy {
	case PruningDefault:
		return &PruningOption{
			KeepRecent: 362880,
			Interval:   10,
		}
	case PruningEverything:
		return &PruningOption{
			KeepRecent: 2,
			Interval:   10,
		}
	case PruningNothing:
		return &PruningOption{
			KeepRecent: 0,
			Interval:   0,
		}
	default:
		return nil
	}
}

// NewPruningOption returns a new PruningOption instance based on the given pruning strategy.
func NewPruningOptionFromString(pruningStrategy string) *PruningOption {
	switch pruningStrategy {
	case PruningOptionDefault:
		return &PruningOption{
			KeepRecent: 362880,
			Interval:   10,
		}
	case PruningOptionEverything:
		return &PruningOption{
			KeepRecent: 2,
			Interval:   10,
		}
	case PruningOptionNothing:
		return &PruningOption{
			KeepRecent: 0,
			Interval:   0,
		}
	default:
		return nil
	}
}

// NewPruningOptionWithCustom returns a new PruningOption based on the given parameters.
func NewPruningOptionWithCustom(keepRecent, interval uint64) *PruningOption {
	return &PruningOption{
		KeepRecent: keepRecent,
		Interval:   interval,
	}
}

func (po *PruningOption) Validate() error {
	if po.Interval == 0 {
		return ErrPruningIntervalZero
	}
	if po.Interval < pruneEverythingInterval {
		return ErrPruningIntervalTooSmall
	}
	if po.KeepRecent < pruneEverythingKeepRecent {
		return ErrPruningKeepRecentTooSmall
	}
	return nil
}

// ShouldPrune returns true if the given version should be pruned.
// If true, it also returns the version to prune up to.
// NOTE: The current version is not pruned.
func (opts *PruningOption) ShouldPrune(version uint64) (bool, uint64) {
	if opts.Interval == 0 {
		return false, 0
	}

	if version <= opts.KeepRecent {
		return false, 0
	}

	if version%opts.Interval == 0 {
		return true, version - opts.KeepRecent - 1
	}

	return false, 0
}

// DBOptions defines the interface of a database options.
type DBOptions interface {
	Get(string) interface{}
}
