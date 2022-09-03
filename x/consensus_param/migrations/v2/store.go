package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/consensus_param/exported"
	v1 "github.com/cosmos/cosmos-sdk/x/consensus_param/migrations/v1"
	"github.com/cosmos/cosmos-sdk/x/consensus_param/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	tmtypes "github.com/tendermint/tendermint/types"
)

func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec, legacySubspace exported.ParamStore) error {
	store := ctx.KVStore(storeKey)

	consensusParams := new(tmproto.ConsensusParams)

	var bp tmproto.BlockParams
	legacySubspace.Get(ctx, v1.ParamStoreKeyBlockParams, &bp)
	consensusParams.Block = &bp

	var ep tmproto.EvidenceParams
	legacySubspace.Get(ctx, v1.ParamStoreKeyEvidenceParams, &ep)
	consensusParams.Evidence = &ep

	var vp tmproto.ValidatorParams
	legacySubspace.Get(ctx, v1.ParamStoreKeyValidatorParams, &vp)
	consensusParams.Validator = &vp

	var versionParams tmproto.VersionParams
	consensusParams.Version = &versionParams

	if err := types.Validate(tmtypes.ConsensusParamsFromProto(*consensusParams)); err != nil {
		return err
	}

	bz := cdc.MustMarshal(consensusParams)
	store.Set(types.ParamStoreKeyConsensusParams, bz)

	return nil
}
