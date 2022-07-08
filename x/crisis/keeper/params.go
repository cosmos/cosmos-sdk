package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/crisis/types"
)

// GetConstantFee get's the constant fee from the paramSpace
func (k *Keeper) GetConstantFee(ctx sdk.Context) (constantFee sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ConstantFee)
	if bz == nil {
		return constantFee
	}
	k.cdc.MustUnmarshal(bz, &constantFee)
	return
}

// GetConstantFee set's the constant fee in the paramSpace
func (k *Keeper) SetConstantFee(ctx sdk.Context, constantFee sdk.Coin) error {
	fmt.Printf("\"0\": %v\n", "0")
	fmt.Printf("k.storeKey: %v\n", k.storeKey)

	store := ctx.KVStore(k.storeKey)
	fmt.Println("1")
	bz, err := k.cdc.Marshal(&constantFee)
	fmt.Println("2")

	if err != nil {
		fmt.Printf("\"3\": %v\n", "3")

		return err
	}
	fmt.Printf("\"4\": %v\n", "4")
	store.Set(types.ConstantFee, bz)
	fmt.Printf("\"5\": %v\n", "5")
	return nil
}
