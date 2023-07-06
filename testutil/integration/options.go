package integration

import cmtabcitypes "github.com/cometbft/cometbft/abci/types"

// Config is the configuration for the integration app.
type Config struct {
	AutomaticProcessProposal bool
	AutomaticCommit          bool
	FinalizeBlockRequest     *cmtabcitypes.RequestFinalizeBlock
}

// Option is a function that can be used to configure the integration app.
type Option func(*Config)

// WithFinalizeBlockRequest sets the finalize block request that will be used
// as a base for both ProcessProposal and FinalizeBlock.
func WithFinalizeBlockRequest(req *cmtabcitypes.RequestFinalizeBlock) Option {
	return func(cfg *Config) {
		cfg.FinalizeBlockRequest = req
	}
}

// WithAutomaticProcessProposal calls ABCI process proposal.
func WithAutomaticProcessProposal() Option {
	return func(cfg *Config) {
		cfg.AutomaticProcessProposal = true
	}
}

// WithAutomaticCommit enables automatic commit.
// This means that the integration app will automatically commit the state after each msgs.
func WithAutomaticCommit() Option {
	return func(cfg *Config) {
		cfg.AutomaticCommit = true
	}
}
