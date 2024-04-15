package types

import (
	"context"
	"errors"
)

type EpochHooks interface {
	// the first block whose timestamp is after the duration is counted as the end of the epoch
	AfterEpochEnd(ctx context.Context, epochIdentifier string, epochNumber int64) error
	// new epoch is next block of epoch end block
	BeforeEpochStart(ctx context.Context, epochIdentifier string, epochNumber int64) error
	// Returns the name of the module implementing epoch hook.
	GetModuleName() string
}

var _ EpochHooks = MultiEpochHooks{}

// combine multiple gamm hooks, all hook functions are run in array sequence.
type MultiEpochHooks []EpochHooks

// GetModuleName implements EpochHooks.
func (MultiEpochHooks) GetModuleName() string {
	return ModuleName
}

func NewMultiEpochHooks(hooks ...EpochHooks) MultiEpochHooks {
	return hooks
}

// AfterEpochEnd is called when epoch is going to be ended, epochNumber is the number of epoch that is ending.
func (h MultiEpochHooks) AfterEpochEnd(ctx context.Context, epochIdentifier string, epochNumber int64) error {
	var errs error
	for i := range h {
		errs = errors.Join(errs, h[i].AfterEpochEnd(ctx, epochIdentifier, epochNumber))
	}
	return errs
}

// BeforeEpochStart is called when epoch is going to be started, epochNumber is the number of epoch that is starting.
func (h MultiEpochHooks) BeforeEpochStart(ctx context.Context, epochIdentifier string, epochNumber int64) error {
	var errs error
	for i := range h {
		errs = errors.Join(errs, h[i].BeforeEpochStart(ctx, epochIdentifier, epochNumber))
	}
	return errs
}

// StakingHooksWrapper is a wrapper for modules to inject StakingHooks using depinject.
type EpochHooksWrapper struct{ EpochHooks }

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (EpochHooksWrapper) IsOnePerModuleType() {}
