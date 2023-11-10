package types

// Config is a config struct used for intialising the gov module to avoid using globals.
type Config struct {
	// MaxTitleLen defines the amount of characters that can be used for proposal title
	MaxTitleLen int
	// MaxMetadataLen defines the amount of characters that can be used for proposal metadata.
	MaxMetadataLen int
	// MaxSummaryLen defines the amount of characters that can be used for proposal summary
	MaxSummaryLen int
}

// DefaultConfig returns the default config for gov.
func DefaultConfig() Config {
	return Config{
		MaxTitleLen:    100,
		MaxMetadataLen: 255,
		MaxSummaryLen:  10200,
	}
}
