package integration

// Config is the configuration for the application.
type Config struct {
	AutomaticBlockCreation bool
}

// Option is a function that can be used to configure the application.
type Option func(*Config)

// WithAutomaticBlockCreation enables automatic block creation.
// This means that the application will automatically create a new block after each msgs.
func WithAutomaticBlockCreation() Option {
	return func(cfg *Config) {
		cfg.AutomaticBlockCreation = true
	}
}
