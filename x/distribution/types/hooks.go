package types

import (
	context "context"
)

type DistributionHook interface {
	BeforeFeeCollectorSend(ctx context.Context) error
}

// combine multiple staking hooks, all hook functions are run in array sequence
var _ DistributionHook = &MultiDistributionHooks{}

type MultiDistributionHooks []DistributionHook

func NewMultiDistributionHooks(hooks ...DistributionHook) MultiDistributionHooks {
	return hooks
}

func (h MultiDistributionHooks) BeforeFeeCollectorSend(ctx context.Context) error {
	for i := range h {
		if err := h[i].BeforeFeeCollectorSend(ctx); err != nil {
			return err
		}
	}

	return nil
}
