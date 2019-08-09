package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
)

// returns context and an app with updated mint keeper
func newTestApp(t *testing.T) (sdk.Context, *simapp.SimApp) {
	ctx, app := simapp.NewSimAppWithContext(true)
	app.MintKeeper.SetParams(ctx, types.DefaultParams())
	app.MintKeeper.SetMinter(ctx, types.DefaultInitialMinter())

	return ctx, app
}
