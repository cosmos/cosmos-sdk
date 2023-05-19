package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/crisis/types"
)

// GetConstantFee get's the constant fee from the store
func (k *Keeper) GetConstantFee(ctx context.Context) (constantFee sdk.Coin, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ConstantFeeKey)
	if bz == nil || err != nil {
		return constantFee, err
	}
	err = k.cdc.Unmarshal(bz, &constantFee)
	return constantFee, err
}

// GetConstantFee set's the constant fee in the store
func (k *Keeper) SetConstantFee(ctx context.Context, constantFee sdk.Coin) error {
	if !constantFee.IsValid() || constantFee.IsNegative() {
		return errorsmod.Wrapf(errors.ErrInvalidCoins, "negative or invalid constant fee: %s", constantFee)
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&constantFee)
	if err != nil {
		return err
	}

	return store.Set(types.ConstantFeeKey, bz)
}
