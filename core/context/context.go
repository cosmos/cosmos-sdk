package appmodule

// ExecMode defines the execution mode which can be set on a Context.
type ExecMode uint8

// All possible execution modes.
const (
	ExecModeCheck ExecMode = iota
	ExecModeReCheck
	ExecModeSimulate
	ExecModePrepareProposal
	ExecModeProcessProposal
	ExecModeVoteExtension
	ExecModeVerifyVoteExtension
	ExecModeFinalize
)
