package keeper

import (
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/consensus/exported"
	"github.com/cosmos/cosmos-sdk/x/consensus/types"
)

var _ exported.ConsensusParamSetter = (*Keeper)(nil)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	authority string
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, authority string) Keeper {
	return Keeper{
		storeKey:  storeKey,
		cdc:       cdc,
		authority: authority,
	}
}

func (k *Keeper) GetAuthority() string {
	return k.authority
}

// Get gets the consensus parameters
func (k *Keeper) Get(ctx sdk.Context) (*cmtproto.ConsensusParams, error) {
	store := ctx.KVStore(k.storeKey)

	cp := &cmtproto.ConsensusParams{}
	bz := store.Get(types.ParamStoreKeyConsensusParams)

	if err := k.cdc.Unmarshal(bz, cp); err != nil {
		return nil, err
	}

	return cp, nil
}

func (k *Keeper) Has(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.ParamStoreKeyConsensusParams)
}

// Set sets the consensus parameters
func (k *Keeper) Set(ctx sdk.Context, cp *cmtproto.ConsensusParams) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ParamStoreKeyConsensusParams, k.cdc.MustMarshal(cp))
}
