package params

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState defines initial activated msg types
type GenesisState struct {
	ActivatedTypes []string `json:"activated-types"`
}

// ActivatedParamKey - paramstore key for msg type activation
func ActivatedParamKey(ty string) string {
	return "Activated/" + ty
}

// InitGenesis stores activated type to param store
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	for _, ty := range data.ActivatedTypes {
		k.set(ctx, ActivatedParamKey(ty), true)
	}
}

// NewAnteHandler returns an AnteHandler that checks
// whether msg type is activate or not
func NewAnteHandler(k Keeper) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx) (sdk.Context, sdk.Result, bool) {
		for _, msg := range tx.GetMsgs() {
			ok := k.Getter().GetBoolWithDefault(ctx, ActivatedParamKey(msg.Type()), false)
			if !ok {
				return ctx, sdk.ErrUnauthorized("deactivated msg type").Result(), true
			}
		}
		return ctx, sdk.Result{}, false
	}
}
