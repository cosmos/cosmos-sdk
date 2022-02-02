package keeper

// Config is a config struct used for intialising the group module to avoid using globals.
type Config struct {
	// MaxMetadataLen defines the max length of the metadata bytes field for various entities within the group module. Defaults to 255 if not explicitly set.
	MaxMetadataLen uint64
}

// DefaultConfig returns the default config for group.
func DefaultConfig() Config {
	return Config{
		MaxMetadataLen: 255,
	}
}
