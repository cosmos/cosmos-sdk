package store

import (
	"errors"
	"fmt"
)

// PruneOptions defines the pruning configuration.
type PruneOptions struct {
	// KeepRecent sets the number of recent versions to keep.
	KeepRecent uint64

	// Interval sets the number of how often to prune.
	// If set to 0, no pruning will be done.
	Interval uint64
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

func NewPruneOptions(pruningOption string) *PruneOptions {
	switch pruningOption {
	case PruningOptionDefault:
		return &PruneOptions{
			KeepRecent: 362880,
			Interval:   10,
		}
	case PruningOptionEverything:
		return &PruneOptions{
			KeepRecent: pruneEverythingKeepRecent,
			Interval:   pruneEverythingInterval,
		}
	case PruningOptionNothing:
		return &PruneOptions{
			KeepRecent: 0,
			Interval:   0,
		}
	default:
		return &PruneOptions{} 
	}
}

func NewCustomPruneOptions(keepRecent, interval uint64) *PruneOptions {
	return &PruneOptions{
		KeepRecent: keepRecent,
		Interval:   interval,
	}
}

func (po *PruneOptions) Validate() error {
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

// DefaultPruneOptions returns the default pruning options.
// Interval is set to 0, which means no pruning will be done.
func DefaultPruneOptions() *PruneOptions {
	return &PruneOptions{
		KeepRecent: 0,
		Interval:   0,
	}
}

// ShouldPrune returns true if the given version should be pruned.
// If true, it also returns the version to prune up to.
// NOTE: The current version is not pruned.
func (opts *PruneOptions) ShouldPrune(version uint64) (bool, uint64) {
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
