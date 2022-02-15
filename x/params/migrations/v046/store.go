package v046

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

// MigrateStore performs in-place store migrations from v0.45 to v0.46. The
// migration includes:
//
// - Setting the version param in the paramstore
func MigrateStore(ctx sdk.Context, paramstore paramtypes.Subspace) error {

	migrateParamsStore(ctx, paramstore)

	return nil
}
