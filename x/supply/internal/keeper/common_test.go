package keeper_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/cosmos/cosmos-sdk/x/supply/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/types"
)

// nolint: deadcode unused
var (
	multiPerm  = "multiple permissions account"
	randomPerm = "random permission"
	holder     = "holder"
)

// nolint: deadcode unused
func createTestApp(t *testing.T, isCheckTx bool) (sdk.Context, *simapp.SimApp) {
	db := dbm.NewMemDB()
	app := simapp.NewSimApp(log.NewNopLogger(), db, nil, true, 0)

	if !isCheckTx {
		// must start chain or will panic when trying to create a context
		genesisState := simapp.NewDefaultGenesisState()
		stateBytes, err := codec.MarshalJSONIndent(app.Cdc, genesisState)
		require.NoError(t, err)

		// Initialize the chain
		app.InitChain(
			abci.RequestInitChain{
				Validators:    []abci.ValidatorUpdate{},
				AppStateBytes: stateBytes,
			},
		)
	}

	ctx := app.BaseApp.NewContext(isCheckTx, abci.Header{})
	return modifyTestKeeper(ctx, app)
}

// initialize a new supply keeper with more module accounts added
func modifyTestKeeper(ctx sdk.Context, app *simapp.SimApp) (sdk.Context, *simapp.SimApp) {
	maccPerms := simapp.MaccPerms
	maccPerms[holder] = nil
	maccPerms[types.Burner] = []string{types.Burner}
	maccPerms[types.Minter] = []string{types.Minter}
	maccPerms[multiPerm] = []string{types.Burner, types.Minter, types.Staking}
	maccPerms[randomPerm] = []string{"random"}

	app.SupplyKeeper = NewKeeper(app.Cdc, app.Keys[types.StoreKey], app.AccountKeeper, app.BankKeeper, maccPerms)
	app.SupplyKeeper.SetSupply(ctx, types.NewSupply(sdk.NewCoins()))
	return ctx, app
}
