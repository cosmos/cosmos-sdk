package keeper

import (
	store2 "github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/crisis/types"
)

func (k *Keeper) decodeConstantFee(bz []byte) (sdk.Coin, error) {
	var constantFee sdk.Coin
	if bz == nil {
		return constantFee, nil
	}
	k.cdc.MustUnmarshal(bz, &constantFee)
	return constantFee, nil
}

// GetConstantFee get's the constant fee from the store
func (k *Keeper) GetConstantFee(ctx sdk.Context) (constantFee sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	constantFee, _ = store2.GetAndDecode(store, k.decodeConstantFee, types.ConstantFeeKey)
	return constantFee
}

// GetConstantFee set's the constant fee in the store
func (k *Keeper) SetConstantFee(ctx sdk.Context, constantFee sdk.Coin) error {
	if !constantFee.IsValid() || constantFee.IsNegative() {
		return errors.Wrapf(errors.ErrInvalidCoins, "negative or invalid constant fee: %s", constantFee)
	}

	store := store2.NewStoreAPI(ctx.KVStore(k.storeKey))
	bz, err := k.cdc.Marshal(&constantFee)
	if err != nil {
		return err
	}

	store.Set(types.ConstantFeeKey, bz)
	return nil
}
