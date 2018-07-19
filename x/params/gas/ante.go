package gas

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/store"
)

const (
	DefaultParamSpace = "GasConfig"
)

func KVStoreKey() params.Key        { return params.NewKey("KVStore") }
func TransientStoreKey() params.Key { return params.NewKey("TransientStore") }

var (
	kvStoreKey        = KVStoreKey()
	transientStoreKey = TransientStoreKey()
)

func NewAnteHandler(store params.Store) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx) (sdk.Context, sdk.Result, bool) {
		if store.Has(ctx, kvStoreKey) {
			var config sdk.GasConfig
			store.Get(ctx, kvStoreKey, &config)
			ctx = ctx.WithKVGasConfig(config)
		}
		if store.Has(ctx, transientStoreKey) {
			var config sdk.GasConfig
			store.Get(ctx, transientStoreKey, &config)
			ctx = ctx.WithTransientGasConfig(config)
		}
		return ctx, sdk.Result{}, false
	}
}
