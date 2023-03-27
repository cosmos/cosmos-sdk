package keeper

import (
	"context"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	storetypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/consensus/exported"
	"github.com/cosmos/cosmos-sdk/x/consensus/types"
)

var _ exported.ConsensusParamSetter = (*Keeper)(nil)

type Keeper struct {
	storeService storetypes.KVStoreService
	cdc          codec.BinaryCodec

	authority string
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService, authority string) Keeper {
	return Keeper{
		storeService: storeService,
		cdc:          cdc,
		authority:    authority,
	}
}

func (k *Keeper) GetAuthority() string {
	return k.authority
}

// Get gets the consensus parameters
func (k *Keeper) Get(ctx context.Context) (*cmtproto.ConsensusParams, error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(types.ParamStoreKeyConsensusParams)
	if err != nil {
		return nil, err
	}

	cp := &cmtproto.ConsensusParams{}
	if err := k.cdc.Unmarshal(bz, cp); err != nil {
		return nil, err
	}

	return cp, nil
}

func (k *Keeper) Has(ctx context.Context) (bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	return store.Has(types.ParamStoreKeyConsensusParams)
}

// Set sets the consensus parameters
func (k *Keeper) Set(ctx context.Context, cp *cmtproto.ConsensusParams) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.ParamStoreKeyConsensusParams, k.cdc.MustMarshal(cp))
}
