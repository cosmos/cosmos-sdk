package gas

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/space"
)

// Keys for parameter access
const (
	DefaultParamSpace = "GasConfig"
)

// nolint - Key generators for parameter space
func KVStoreKey() params.Key        { return params.NewKey("KVStore") }
func TransientStoreKey() params.Key { return params.NewKey("TransientStore") }

// Cached parameter keys
var (
	kvStoreKey        = KVStoreKey()
	transientStoreKey = TransientStoreKey()
)

// NewAnteHandler returns AnteHandler
// that overrides existing gasconfig in the context
// with the gasconfig spaced in the paramspace
func NewAnteHandler(space params.Space) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx) (sdk.Context, sdk.Result, bool) {
		if space.Has(ctx, kvStoreKey) {
			var config sdk.GasConfig
			space.Get(ctx, kvStoreKey, &config)
			ctx = ctx.WithKVGasConfig(config)
		}
		if space.Has(ctx, transientStoreKey) {
			var config sdk.GasConfig
			space.Get(ctx, transientStoreKey, &config)
			ctx = ctx.WithTransientGasConfig(config)
		}
		return ctx, sdk.Result{}, false
	}
}
