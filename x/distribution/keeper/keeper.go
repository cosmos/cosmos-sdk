package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	wire "github.com/tendermint/go-wire"
)

// keeper of the stake store
type Keeper struct {
	storeKey    sdk.StoreKey
	storeTKey   sdk.StoreKey
	cdc         *wire.Codec
	coinKeeper  types.CoinKeeper
	stakeKeeper types.StakeKeeper

	// codespace
	codespace sdk.CodespaceType
}

func NewKeeper(cdc *wire.Codec, key, tkey sdk.StoreKey, ck types.CoinKeeper,
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

// set the global fee pool distribution info
func (k Keeper) GetFeePool(ctx sdk.Context, proposerPK sdk.PubKey) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(proposerPK)
	store.Set(ProposerKey, b)
}
