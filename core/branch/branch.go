// Package branch contains the core branch service interface.
package branch

import "context"

// Service is the branch service interface. It can be used to execute
// code paths in an isolated execution context that can be reverted.
// A revert typically means a rollback on events and state changes.
type Service interface {
	// Execute executes the given function in an isolated context. If the
	// `f` function returns an error, the execution is considered failed,
	// and every change made affecting the execution context is rolled back.
	// If the function returns nil, the execution is considered successful, and
	// committed.
	// The context.Context passed to the `f` function is a child of the context
	// passed to the Execute function, and is what should be used with other
	// core services in order to ensure the execution remains isolated.
	Execute(ctx context.Context, f func(ctx context.Context) error) error
}
