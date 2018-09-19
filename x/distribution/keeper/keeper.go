package keeper

import (
	wire "github.com/tendermint/go-wire"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// nolint
const (
	ParamStoreKeyCommunityTax = "distr/community-tax"
)

// keeper of the stake store
type Keeper struct {
	storeKey    sdk.StoreKey
	storeTKey   sdk.StoreKey
	cdc         *wire.Codec
	ps          params.Setter
	coinKeeper  types.CoinKeeper
	stakeKeeper types.StakeKeeper

	// codespace
	codespace sdk.CodespaceType
}

func NewKeeper(cdc *wire.Codec, key, tkey sdk.StoreKey, ps params.Setter, ck types.CoinKeeper,
	sk types.StakeKeeper, codespace sdk.CodespaceType) Keeper {

	keeper := Keeper{
		storeKey:   key,
		storeTKey:  tkey,
		cdc:        cdc,
		coinKeeper: ck,
		codespace:  codespace,
	}
	return keeper
}

//______________________________________________________________________

// get the global fee pool distribution info
func (k Keeper) GetFeePool(ctx sdk.Context) (feePool types.FeePool) {
	store := ctx.KVStore(k.storeKey)

	b := store.Get(FeePoolKey)
	if b == nil {
		panic("Stored fee pool should not have been nil")
	}

	k.cdc.MustUnmarshalBinary(b, &feePool)
	return
}

// set the global fee pool distribution info
func (k Keeper) SetFeePool(ctx sdk.Context, feePool types.FeePool) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(feePool)
	store.Set(FeePoolKey, b)
}

//______________________________________________________________________

// set the proposer public key for this block
func (k Keeper) GetProposerConsAddr(ctx sdk.Context) (consAddr sdk.ConsAddress) {
	store := ctx.KVStore(k.storeKey)

	b := store.Get(ProposerKey)
	if b == nil {
		panic("Stored fee pool should not have been nil")
	}

	k.cdc.MustUnmarshalBinary(b, &consAddr)
	return
}

// get the proposer public key for this block
func (k Keeper) SetProposerConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(consAddr)
	store.Set(ProposerKey, b)
}

//______________________________________________________________________

// Returns the current Deposit Procedure from the global param store
// nolint: errcheck
func (k Keeper) GetCommunityTax(ctx sdk.Context) sdk.Dec {
	var communityTax sdk.Dec
	keeper.ps.Get(ctx, ParamStoreKeyCommunityTax, &communityTax)
	return communityTax
}

// nolint: errcheck
func (k Keeper) setCommunityTax(ctx sdk.Context, communityTax sdk.Dec) {
	keeper.ps.Set(ctx, ParamStoreKeyCommunityTax, &communityTax)
}
