package fee_delegation

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
}

func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return Keeper{storeKey, cdc}
}

func FeeAllowanceKey(grantee sdk.AccAddress, granter sdk.AccAddress) []byte {
	return []byte(fmt.Sprintf("f/%x/%x", grantee, granter))
}

func (k Keeper) DelegateFeeAllowance(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, allowance FeeAllowance) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(allowance)
	store.Set(FeeAllowanceKey(grantee, granter), bz)
}

func (k Keeper) RevokeFeeAllowance(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(FeeAllowanceKey(grantee, granter))
}

type FeeAllowanceGrant struct {
	Allowance FeeAllowance   `json:"allowance"`
	Grantee   sdk.AccAddress `json:"grantee"`
	Granter   sdk.AccAddress `json:"granter"`
}

func (k Keeper) GetFeeAllowances(ctx sdk.Context, grantee sdk.AccAddress) []FeeAllowanceGrant {
	prefix := fmt.Sprintf("g/%x/", grantee)
	prefixBytes := []byte(prefix)
	store := ctx.KVStore(k.storeKey)
	var grants []FeeAllowanceGrant
	iter := sdk.KVStorePrefixIterator(store, prefixBytes)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		granter, _ := sdk.AccAddressFromHex(string(iter.Key()[len(prefix):]))
		bz := iter.Value()
		var allowance FeeAllowance
		k.cdc.MustUnmarshalBinaryBare(bz, &allowance)
		grants = append(grants, FeeAllowanceGrant{
			Allowance: allowance,
			Grantee:   grantee,
			Granter:   granter,
		})
	}
	return grants
}

func (k Keeper) AllowDelegatedFees(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, fee sdk.Coins) bool {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(FeeAllowanceKey(grantee, granter))
	if len(bz) == 0 {
		return false
	}
	var allowance FeeAllowance
	k.cdc.MustUnmarshalBinaryBare(bz, &allowance)
	if allowance == nil {
		return false
	}
	allow, updated, delete := allowance.Accept(fee, ctx.BlockHeader())
	if allow == false {
		return false
	}
	if delete {
		k.RevokeFeeAllowance(ctx, grantee, granter)
	} else if updated != nil {
		k.DelegateFeeAllowance(ctx, grantee, granter, updated)
	}
	return true
}
