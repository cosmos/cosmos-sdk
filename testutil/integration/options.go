package integration

// Config is the configuration for the integration app.
type Config struct {
	AutomaticBlockCreation bool
	AutomaticCommit        bool
}

// Option is a function that can be used to configure the integration app.
type Option func(*Config)

// WithAutomaticBlockCreation enables automatic block creation.
// This means that the integration app will automatically create a new block after each msgs.
func WithAutomaticBlockCreation() Option {
	return func(cfg *Config) {
		cfg.AutomaticBlockCreation = true
	}
}

// WithAutomaticCommit enables automatic commit.
// This means that the integration app will automatically commit the state after each msgs.
func WithAutomaticCommit() Option {
	return func(cfg *Config) {
		cfg.AutomaticCommit = true
	}
}
