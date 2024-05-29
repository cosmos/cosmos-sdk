package internal

import "cosmossdk.io/core/transaction"

// All possible transaction execution modes.
// For backwards compatibility and easier casting, the ExecMode values must be:
// 1) set equivalent to cosmos/cosmos-sdk/types package.
// 2) a superset of core/transaction/service.go:ExecMode with same numeric values.
const (
	ExecModeCheck transaction.ExecMode = iota
	ExecModeReCheck
	ExecModeSimulate
	ExecModePrepareProposal
	ExecModeProcessProposal
	ExecModeVoteExtension
	ExecModeVerifyVoteExtension
	ExecModeFinalize
)
