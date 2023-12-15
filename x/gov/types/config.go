package types

// Config is a config struct used for initializing the gov module to avoid using globals.
type Config struct {
	// MaxTitleLen defines the amount of characters that can be used for proposal title
	MaxTitleLen uint64
	// MaxMetadataLen defines the amount of characters that can be used for proposal metadata.
	MaxMetadataLen uint64
	// MaxSummaryLen defines the amount of characters that can be used for proposal summary
	MaxSummaryLen uint64
}

// DefaultConfig returns the default config for gov.
func DefaultConfig() Config {
	return Config{
		MaxTitleLen:    255,
		MaxMetadataLen: 255,
		MaxSummaryLen:  10200,
	}
}
