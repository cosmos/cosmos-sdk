package v043

import (
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

func migrateParamsStore(ctx sdk.Context, paramstore paramtypes.Subspace) {
	paramstore.WithKeyTable(paramtypes.ConsensusParamsKeyTable())
	paramstore.Set(ctx, baseapp.ParamStoreKeyVersionParams, tmproto.VersionParams{})
}

func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey, paramstore paramtypes.Subspace) error {
	migrateParamsStore(ctx, paramstore)
	return nil

}
