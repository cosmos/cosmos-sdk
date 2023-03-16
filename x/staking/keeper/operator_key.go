package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k Keeper) GetValidatorOperatorKeyRotationsHistory(ctx sdk.Context,
	valAddr sdk.ValAddress) []types.OperatorKeyRotationHistory {

	var historyObjs []types.OperatorKeyRotationHistory
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store,
		append([]byte(types.ValidatorOperatorKeyRotationHistoryKey), valAddr...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var history types.OperatorKeyRotationHistory

		k.cdc.MustUnmarshal(iterator.Value(), &history)
		historyObjs = append(historyObjs, history)
	}
	return historyObjs
}

func (k Keeper) SetOperatorKeyRotationHistory(ctx sdk.Context, curValAddr, newValAddr sdk.ValAddress, height int64) error {
	historyObj := types.OperatorKeyRotationHistory{
		OperatorAddress:    newValAddr.String(),
		OldOperatorAddress: curValAddr.String(),
		Height:             uint64(ctx.BlockHeight()),
	}

	store := ctx.KVStore(k.storeKey)
	key := types.GetOperKeyRotationHistoryKey(newValAddr, uint64(height))
	bz, err := k.cdc.Marshal(&historyObj)

	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

func (k Keeper) SetVORQueue(ctx sdk.Context) {

}
