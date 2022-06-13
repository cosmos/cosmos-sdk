package runtime

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/depinject"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/x/nft/testutil"
)

// Setup initializes a new runtime.App. A Nop logger is set in runtime.App.
func Setup(t *testing.T, appConfig depinject.Config, extraInject ...interface{}) *App {
	t.Helper()

	var appBuilder *AppBuilder
	var msgServiceRouter *baseapp.MsgServiceRouter

	if err := depinject.Inject(
		testutil.AppConfig,
		append(extraInject, &appBuilder, &msgServiceRouter)...,
	); err != nil {
		t.Fatal("failed to inject dependencies")
	}

	app := appBuilder.Build(log.NewNopLogger(), dbm.NewMemDB(), nil, msgServiceRouter)
	require.NoError(t, app.Load(true))

	// init chain must be called to stop deliverState from being nil
	stateBytes, err := tmjson.MarshalIndent(appBuilder.DefaultGenesis(), "", " ")
	require.NoError(t, err)

	// init chain will set the validator set and initialize the genesis accounts
	app.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simtestutil.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)

	return app
}
