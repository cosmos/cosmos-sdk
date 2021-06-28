package v044

import (
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// migrateParamsStore adds application version parameters to baseapp parameters
func migrateParamsStore(ctx sdk.Context, paramstore paramtypes.Subspace) {
	if paramstore.HasKeyTable() {
		paramstore.Set(ctx, baseapp.ParamStoreKeyVersionParams, tmproto.VersionParams{})
	} else {
		paramstore.WithKeyTable(paramtypes.ConsensusParamsKeyTable())
		paramstore.Set(ctx, baseapp.ParamStoreKeyVersionParams, tmproto.VersionParams{})
	}
}

// MigrateStore adds a new key, versionParams, for baseapp consensus parameters.
func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey, paramstore paramtypes.Subspace) error {
	migrateParamsStore(ctx, paramstore)
	return nil

}
