package keeper

import (
	"context"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ApplyRekey exposes applyRekey to the keeper_test package.
func (ak AccountKeeper) ApplyRekey(ctx context.Context, acc sdk.AccountI, newPk cryptotypes.PubKey) error {
	return ak.applyRekey(ctx, acc, newPk)
}
