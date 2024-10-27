package keeper

import (
	"testing"

	"cosmossdk.io/simapp"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func TestParams(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	expParams := types.DefaultParams()

	// check that the empty keeper loads the default
	resParams := app.StakingKeeper.GetParams(ctx)
	require.True(t, expParams.Equal(resParams))

	// modify a params, save, and retrieve
	expParams.MaxValidators = 777
	app.StakingKeeper.SetParams(ctx, expParams)
	resParams = app.StakingKeeper.GetParams(ctx)
	require.True(t, expParams.Equal(resParams))
}
