package keeper

import "time"

// Config is a config struct used for intialising the group module to avoid using globals.
type Config struct {
	// ExecutionPeriod defines the max duration after a proposal's voting
	// period ends that members can send a MsgExec to execute the proposal.
	ExecutionPeriod time.Duration
	// MaxMetadataLen defines the max length of the metadata bytes field for various entities within the group module. Defaults to 255 if not explicitly set.
	MaxMetadataLen uint64
}

// DefaultConfig returns the default config for group.
func DefaultConfig() Config {
	return Config{
		ExecutionPeriod: time.Hour * 24 * 7, // One week.
		MaxMetadataLen:  255,
	}
}
