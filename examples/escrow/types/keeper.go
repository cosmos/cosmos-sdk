package types

import (
	"bytes"
	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
	"strconv"
	"strings"
)

type Keeper struct {
	covStoreKey sdk.StoreKey
	bankKeeper  bank.Keeper
	cdc         *wire.Codec
}

func NewKeeper(cdc *wire.Codec, covKey sdk.StoreKey, bk bank.Keeper) Keeper {
	return Keeper{
		covStoreKey: covKey,
		bankKeeper:  bk,
		cdc:         cdc,
	}
}

func (keeper Keeper) createCovenant(ctx sdk.Context, Sender sdk.Address,
	Settlers []sdk.Address, Receivers []sdk.Address,
	Amount sdk.Coins) (int64, bool) {

	if keeper.bankKeeper.HasCoins(ctx, Sender, Amount) {
		cov := Covenant{Settlers, Receivers, Amount}
		covID := keeper.storeCovenant(ctx, cov)
		return covID, true
	}
	return 0, false

}

func (keeper Keeper) settleCovenant(ctx sdk.Context, covID int64,
	Settler sdk.Address, Receiver sdk.Address) bool {
	cov := keeper.getCovenant(ctx, covID)
	validSettler := false
	validReceiver := false
	for _, s := range cov.Settlers {
		if bytes.Equal(s, Settler) {
			validSettler = true
		}
	}
	for _, r := range cov.Receivers {
		if bytes.Equal(r, Receiver) {
			validReceiver = true
		}
	}
	if validSettler && validReceiver {
		keeper.bankKeeper.AddCoins(ctx, Receiver, cov.Amount)
		keeper.deleteCovenant(ctx, covID)
		return true
	} else {
		return false
	}
}

func prefixArrayKey(name string, index int64) []byte {
	return []byte(strings.Join([]string{"arrays", name, strconv.FormatInt(index, 10)}, ":"))
}
func prefixVariableKey(name string) []byte {
	return []byte(strings.Join([]string{"variables", name}, ":"))
}
func (keeper Keeper) getCovenant(ctx sdk.Context, covID int64) Covenant {
	store := ctx.KVStore(keeper.covStoreKey)
	covKey := prefixArrayKey("covenants", covID)
	bz := store.Get(covKey)
	var cov Covenant
	keeper.cdc.UnmarshalBinary(bz, cov)
	return cov
}

func (keeper Keeper) storeCovenant(ctx sdk.Context, cov Covenant) int64 {
	covID := keeper.getNewCovenantID(ctx)
	store := ctx.KVStore(keeper.covStoreKey)
	covKey := prefixArrayKey("covenants", covID)
	bz, err := keeper.cdc.MarshalBinary(cov)
	store.Set(covKey, bz)
	return covID
}

func (keeper Keeper) getNewCovenantID(ctx sdk.Context) int64 {
	store := ctx.KVStore(keeper.covStoreKey)
	bz := store.Get(prefixVariableKey("nextCovenantID"))
	nextCovID := int64(0)
	if bz != nil {
		keeper.cdc.UnmarshalBinary(bz, nextCovID)
	}
	bz, _ = keeper.cdc.MarshalBinary(nextCovID + 1)
	store.Set(prefixVariableKey("nextCovenantID"), bz)
	return nextCovID
}
