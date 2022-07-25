package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// GetSupplyOffset retrieves the SupplyOffset from store for a specific denom
func (k BaseViewKeeper) GetSupplyOffset(ctx sdk.Context, denom string) sdk.Int {
	store := ctx.KVStore(k.storeKey)
	supplyOffsetStore := prefix.NewStore(store, types.SupplyOffsetKey)

	bz := supplyOffsetStore.Get([]byte(denom))
	if bz == nil {
		return sdk.NewInt(0)
	}

	var amount sdk.Int
	err := amount.Unmarshal(bz)
	if err != nil {
		panic(fmt.Errorf("unable to unmarshal supply offset value %v", err))
	}

	return amount
}

// setSupplyOffset sets the supply offset for the given denom
func (k BaseKeeper) setSupplyOffset(ctx sdk.Context, denom string, offsetAmount sdk.Int) {
	intBytes, err := offsetAmount.Marshal()
	if err != nil {
		panic(fmt.Errorf("unable to marshal amount value %v", err))
	}

	store := ctx.KVStore(k.storeKey)
	supplyOffsetStore := prefix.NewStore(store, types.SupplyOffsetKey)

	// Bank invariants and IBC requires to remove zero coins.
	if offsetAmount.IsZero() {
		supplyOffsetStore.Delete([]byte(denom))
	} else {
		supplyOffsetStore.Set([]byte(denom), intBytes)
	}
}

// AddSupplyOffset adjusts the current supply offset of a denom by the inputted offsetAmount
func (k BaseKeeper) AddSupplyOffset(ctx sdk.Context, denom string, offsetAmount sdk.Int) {
	k.setSupplyOffset(ctx, denom, k.GetSupplyOffset(ctx, denom).Add(offsetAmount))
}

// GetSupplyWithOffset retrieves the Supply of a denom and offsets it by SupplyOffset
// If SupplyWithOffset is negative, it returns 0.  This is because sdk.Coin is not valid
// with a negative amount
func (k BaseKeeper) GetSupplyWithOffset(ctx sdk.Context, denom string) sdk.Coin {
	supply := k.GetSupply(ctx, denom)
	supply.Amount = supply.Amount.Add(k.GetSupplyOffset(ctx, denom))

	if supply.Amount.IsNegative() {
		supply.Amount = sdk.ZeroInt()
	}

	return supply
}

// GetPaginatedTotalSupplyWithOffsets queries for the supply with offset, ignoring 0 coins, with a given pagination
func (k BaseKeeper) GetPaginatedTotalSupplyWithOffsets(ctx sdk.Context, pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	supplyStore := prefix.NewStore(store, types.SupplyKey)

	supply := sdk.NewCoins()

	pageRes, err := query.Paginate(supplyStore, pagination, func(key, value []byte) error {
		denom := string(key)

		var amount sdk.Int
		err := amount.Unmarshal(value)
		if err != nil {
			return fmt.Errorf("unable to convert amount string to Int %v", err)
		}

		amount = amount.Add(k.GetSupplyOffset(ctx, denom))

		// `Add` omits the 0 coins addition to the `supply`.
		supply = supply.Add(sdk.NewCoin(denom, amount))
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return supply, pageRes, nil
}

// IterateTotalSupplyWithOffsets iterates over the total supply with offsets calling the given cb (callback) function
// with the balance of each coin.
// The iteration stops if the callback returns true.
func (k BaseViewKeeper) IterateTotalSupplyWithOffsets(ctx sdk.Context, cb func(sdk.Coin) bool) {
	store := ctx.KVStore(k.storeKey)
	supplyStore := prefix.NewStore(store, types.SupplyKey)

	iterator := supplyStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var amount sdk.Int
		err := amount.Unmarshal(iterator.Value())
		if err != nil {
			panic(fmt.Errorf("unable to unmarshal supply value %v", err))
		}

		balance := sdk.Coin{
			Denom:  string(iterator.Key()),
			Amount: amount.Add(k.GetSupplyOffset(ctx, string(iterator.Key()))),
		}

		if cb(balance) {
			break
		}
	}
}

// getGenesisSupplyOffsets returns supply offset for genesis, encoded with denom in state
func (k BaseViewKeeper) getGenesisSupplyOffsets(ctx sdk.Context) []types.GenesisSupplyOffset {
	store := ctx.KVStore(k.storeKey)
	supplyStore := prefix.NewStore(store, types.SupplyKey)

	iterator := supplyStore.Iterator(nil, nil)
	defer iterator.Close()

	supplyOffsets := []types.GenesisSupplyOffset{}
	for ; iterator.Valid(); iterator.Next() {
		supplyOffset := types.GenesisSupplyOffset{
			Denom:  string(iterator.Key()),
			Offset: k.GetSupplyOffset(ctx, string(iterator.Key())),
		}
		supplyOffsets = append(supplyOffsets, supplyOffset)
	}
	return supplyOffsets
}
