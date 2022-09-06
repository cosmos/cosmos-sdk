package v1

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/consensus/exported"
	"github.com/cosmos/cosmos-sdk/x/consensus/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	tmtypes "github.com/tendermint/tendermint/types"
)

func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec, legacySubspace exported.ParamStore) error {
	store := ctx.KVStore(storeKey)

	consensusParams := new(tmproto.ConsensusParams)

	var bp tmproto.BlockParams
	legacySubspace.Get(ctx, ParamStoreKeyBlockParams, &bp)
	consensusParams.Block = &bp

	var ep tmproto.EvidenceParams
	legacySubspace.Get(ctx, ParamStoreKeyEvidenceParams, &ep)
	consensusParams.Evidence = &ep

	var vp tmproto.ValidatorParams
	legacySubspace.Get(ctx, ParamStoreKeyValidatorParams, &vp)
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
