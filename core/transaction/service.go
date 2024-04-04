package transaction

import "context"

// ExecMode defines the execution mode which can be set on a Context.
type ExecMode uint8

// All possible execution modes.
const (
	ExecModeCheck ExecMode = iota
	_
	ExecModeSimulate
	_
	_
	_
	_
	ExecModeFinalize
)

// Service creates a transaction service.
// This service is used to get information about which context is used to execute a transaction.
type Service interface {
	ExecMode(ctx context.Context) ExecMode
}
