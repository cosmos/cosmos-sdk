package v3

import (
	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// Migrate migrates the x/mint module state from the consensus version 2 to
// version 3.
func Migrate(
	ctx sdk.Context,
	store storetypes.KVStore,
	cdc codec.BinaryCodec,
	params collections.Item[types.Params],
) error {
	currParams, err := params.Get(ctx)
	if err != nil {
		return err
	}

	currParams.MaxSupply = math.NewInt(0)
	if err := currParams.Validate(); err != nil {
		return err
	}

	return params.Set(ctx, currParams)
}
