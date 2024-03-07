package types

import (
	fmt "fmt"
	"strconv"
	"context"

	"github.com/hashicorp/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

type EpochHooks interface {
	// the first block whose timestamp is after the duration is counted as the end of the epoch
	AfterEpochEnd(ctx context.Context, epochIdentifier string, epochNumber int64) error
	// new epoch is next block of epoch end block
	BeforeEpochStart(ctx context.Context, epochIdentifier string, epochNumber int64) error
	// Returns the name of the module implementing epoch hook.
	GetModuleName() string
}

const (
	// flag indicating whether this is a before epoch hook
	isBeforeEpoch = true
)

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
	for _, hook := range h {
		panicCatchingEpochHook(ctx, hook.AfterEpochEnd, epochIdentifier, epochNumber, h.GetModuleName(), !isBeforeEpoch)
	}
	return nil
}

// BeforeEpochStart is called when epoch is going to be started, epochNumber is the number of epoch that is starting.
func (h MultiEpochHooks) BeforeEpochStart(ctx context.Context, epochIdentifier string, epochNumber int64) error {
	for _, hook := range h {
		panicCatchingEpochHook(ctx, hook.BeforeEpochStart, epochIdentifier, epochNumber, hook.GetModuleName(), isBeforeEpoch)
	}
	return nil
}

func panicCatchingEpochHook(
	ctx context.Context,
	hookFn func(ctx context.Context, epochIdentifier string, epochNumber int64) error,
	epochIdentifier string,
	epochNumber int64,
	moduleName string,
	isBeforeEpoch bool,
) {
	wrappedHookFn := func(ctx sdk.Context) error {
		return hookFn(ctx, epochIdentifier, epochNumber)
	}
	// TODO: Thread info for which hook this is, may be dependent on larger hook system refactoring
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	err := baseapp.ApplyFuncIfNoError(sdkCtx, wrappedHookFn)
	if err != nil {
		telemetry.IncrCounterWithLabels([]string{}, 1, []metrics.Label{
			{
				Name:  "module_name",
				Value: moduleName,
			},
			{
				Name:  "error",
				Value: err.Error(),
			},
			{
				Name:  "is_before_hook",
				Value: strconv.FormatBool(isBeforeEpoch),
			},
		})

		sdkCtx.Logger().Error(fmt.Sprintf("error in epoch hook %v", err))
	}
}
