package context

// ExecMode defines the execution mode which can be set on a Context.
type ExecMode uint8

// All possible execution modes.
const (
	ExecModeCheck ExecMode = iota
	ExecModeReCheck
	ExecModePrepareProposal
	ExecModeProcessProposal
	ExecModeSimulate
	ExecModeFinalize
)

// TODO: remove
type ContextKey string

// TODO: remove
const CometInfoKey ContextKey = "comet-info"
