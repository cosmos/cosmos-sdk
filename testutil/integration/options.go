package integration

import cmtabcitypes "github.com/cometbft/cometbft/abci/types"

// Config is the configuration for the integration app.
type Config struct {
	AutomaticBeginEndBlock bool
	AutomaticCommit        bool
	CustomBeginBlock       cmtabcitypes.RequestBeginBlock
}

// Option is a function that can be used to configure the integration app.
type Option func(*Config)

// WithCustomBeginBlock takes additional params into the configuration.
func WithCustomBeginBlock(req cmtabcitypes.RequestBeginBlock) Option {
	return func(cfg *Config) {
		cfg.AutomaticBeginEndBlock = true
		cfg.CustomBeginBlock = req
	}
}

// WithAutomaticBlockCreation enables begin/end block calls.
func WithAutomaticBeginEndBlock() Option {
	return func(cfg *Config) {
		cfg.AutomaticBeginEndBlock = true
	}
}

// WithAutomaticCommit enables automatic commit.
// This means that the integration app will automatically commit the state after each msgs.
func WithAutomaticCommit() Option {
	return func(cfg *Config) {
		cfg.AutomaticCommit = true
	}
}
