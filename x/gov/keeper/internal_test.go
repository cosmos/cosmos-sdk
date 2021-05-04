package keeper

import "github.com/cosmos/cosmos-sdk/x/gov/types"

func UpdateHooks(k *Keeper, h types.GovHooks) *Keeper {
	k.hooks = h
	return k
}
