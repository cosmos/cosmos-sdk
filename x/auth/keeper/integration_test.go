// nolint
package keeper_test

// DONTCOVER

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// returns context and app with params set on account keeper
func newTestApp(t *testing.T) (sdk.Context, *simapp.SimApp) {
	ctx, app := simapp.NewSimAppWithContext(true)
	app.AccountKeeper.SetParams(ctx, types.DefaultParams())

	return ctx, app
}
