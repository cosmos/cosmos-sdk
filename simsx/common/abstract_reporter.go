package common

import "github.com/cosmos/cosmos-sdk/types"

// SimulationReporter is an interface to handle simulation reporting, tracking status, errors, comments, and outcomes.
type SimulationReporter interface {
	// Skip set skip flag to abort further execution
	Skip(comment string)
	// Skipf set skip flag to abort further execution
	Skipf(comment string, args ...any)
	// Success complete with success
	Success(msg types.Msg, comments ...string)
	// Fail complete with failure
	Fail(err error, comments ...string)
	// IsAborted returns true when skipped or an error was recorded
	IsAborted() bool
	// IsSkipped
	// Deprecated: will be removed. use IsSkipped() instead
	IsSkipped() bool
}

// SimulationReporterRuntime is an interface to handle simulation reporting, tracking status, errors, comments, and outcomes.
type SimulationReporterRuntime interface {
	SimulationReporter
	WithScope(msg types.Msg, optionalSkipHook ...SkipHook) SimulationReporterRuntime
	// Scope returns module and msg url
	Scope() (string, string)
	// Status returns the current status of the reporter as a ReporterStatus, reflecting the simulation's state outcome.
	Status() ReporterStatus

	// Comment returns any skip or completion comment
	Comment() string
	// Error returns any error recorded
	Error() error
	// Close end recording and return error captured on fail
	Close() error
}

// ReporterStatus represents the status of a reporter in a simulation, used to track state transitions and outcomes.
type ReporterStatus uint8

const (
	ReporterStatusUndefined ReporterStatus = iota
	ReporterStatusSkipped   ReporterStatus = iota
	ReporterStatusCompleted ReporterStatus = iota
)

func (s ReporterStatus) String() string {
	switch s {
	case ReporterStatusSkipped:
		return "skipped"
	case ReporterStatusCompleted:
		return "completed"
	default:
		return "undefined"
	}
}

// SkipHook is an interface that represents a callback hook used triggered on skip operations.
// It provides a single method `Skip` that accepts variadic arguments. This interface is implemented
// by Go stdlib testing.T and testing.B
type SkipHook interface {
	Skip(args ...any)
}

type SkipHookFn func(args ...any)

func (s SkipHookFn) Skip(args ...any) {
	s(args...)
}
