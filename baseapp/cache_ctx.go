package baseapp

import (
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// This function lets you run the function f, but if there's an error or panic
// drop the state machine change and log the error.
// If there is no error, proceeds as normal (but with some slowdown due to SDK store weirdness)
// Try to avoid usage of iterators in f.
//
// If its an out of gas panic, this function will also panic like in normal tx execution flow.
// This is still safe for beginblock / endblock code though, as they do not have out of gas panics.
func ApplyFuncIfNoError(ctx sdk.Context, f func(ctx sdk.Context) error) (err error) {
	// Add a panic safeguard
	defer func() {
		if recoveryError := recover(); recoveryError != nil {
			if isErr, _ := IsOutOfGasError(recoveryError); isErr {
				// We panic with the same error, to replicate the normal tx execution flow.
				panic(recoveryError)
			} else {
				PrintPanicRecoveryError(ctx, recoveryError)
				err = errors.New("panic occurred during execution")
			}
		}
	}()
	// makes a new cache context, which all state changes get wrapped inside of.
	cacheCtx, write := ctx.CacheContext()
	err = f(cacheCtx)
	if err != nil {
		ctx.Logger().Error(err.Error())
	} else {
		// no error, write the output of f
		write()
	}
	return err
}

// Frustratingly, this has to return the error descriptor, not an actual error itself
// because the SDK errors here are not actually errors. (They don't implement error interface)
func IsOutOfGasError(err any) (bool, string) {
	switch e := err.(type) {
	case storetypes.ErrorOutOfGas:
		return true, e.Descriptor
	case storetypes.ErrorGasOverflow:
		return true, e.Descriptor
	default:
		return false, ""
	}
}

// PrintPanicRecoveryError error logs the recoveryError, along with the stacktrace, if it can be parsed.
// If not emits them to stdout.
func PrintPanicRecoveryError(ctx sdk.Context, recoveryError interface{}) {
	errStackTrace := string(debug.Stack())
	switch e := recoveryError.(type) {
	case storetypes.ErrorOutOfGas:
		ctx.Logger().Debug("out of gas error inside panic recovery block: " + e.Descriptor)
		return
	case string:
		ctx.Logger().Error("Recovering from (string) panic: " + e)
	case runtime.Error:
		ctx.Logger().Error("recovered (runtime.Error) panic: " + e.Error())
	case error:
		ctx.Logger().Error("recovered (error) panic: " + e.Error())
	default:
		ctx.Logger().Error("recovered (default) panic. Could not capture logs in ctx, see stdout")
		fmt.Println("Recovering from panic ", recoveryError)
		debug.PrintStack()
		return
	}
	ctx.Logger().Error("stack trace: " + errStackTrace)
}
