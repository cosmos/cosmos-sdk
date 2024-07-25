package store

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
// app.toml config options
type PruningOption struct {
	// KeepRecent sets the number of recent versions to keep.
	KeepRecent uint64 `mapstructure:"keep-recent" toml:"keep-recent" comment:"Number of recent heights to keep on disk."`

	// Interval sets the number of how often to prune.
	// If set to 0, no pruning will be done.
	Interval uint64 `mapstructure:"interval" toml:"interval" comment:"Height interval at which pruned heights are removed from disk."`
}

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

// NewPruningOptionWithCustom returns a new PruningOption based on the given parameters.
func NewPruningOptionWithCustom(keepRecent, interval uint64) *PruningOption {
	return &PruningOption{
		KeepRecent: keepRecent,
		Interval:   interval,
	}
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
