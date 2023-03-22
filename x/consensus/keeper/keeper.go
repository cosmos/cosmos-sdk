package keeper

import (
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	storetypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/consensus/exported"
	"github.com/cosmos/cosmos-sdk/x/consensus/types"
)

var _ exported.ConsensusParamSetter = (*Keeper)(nil)

type Keeper struct {
	storeSvc storetypes.KVStoreService
	cdc      codec.BinaryCodec

	authority string
}

func NewKeeper(cdc codec.BinaryCodec, storeSvc storetypes.KVStoreService, authority string) Keeper {
	return Keeper{
		storeSvc:  storeSvc,
		cdc:       cdc,
		authority: authority,
	}
}

func (k *Keeper) GetAuthority() string {
	return k.authority
}

// Get gets the consensus parameters
func (k *Keeper) Get(ctx sdk.Context) (*cmtproto.ConsensusParams, error) {
	store := k.storeSvc.OpenKVStore(ctx)

	cp := &cmtproto.ConsensusParams{}
	bz, err := store.Get(types.ParamStoreKeyConsensusParams)
	if err != nil {
		return nil, err
	}

	if err := k.cdc.Unmarshal(bz, cp); err != nil {
		return nil, err
	}

	return cp, nil
}

func (k *Keeper) Has(ctx sdk.Context) bool {
	store := k.storeSvc.OpenKVStore(ctx)

	has, err := store.Has(types.ParamStoreKeyConsensusParams)
	// should never panic given that key is hardcoded
	if err != nil {
		panic(err)
	}

	return has
}

// Set sets the consensus parameters
func (k *Keeper) Set(ctx sdk.Context, cp *cmtproto.ConsensusParams) {
	store := k.storeSvc.OpenKVStore(ctx)
	err := store.Set(types.ParamStoreKeyConsensusParams, k.cdc.MustMarshal(cp))
	if err != nil {
		panic(err)
	}
}
