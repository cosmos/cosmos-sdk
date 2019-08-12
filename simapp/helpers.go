package simapp

import (
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Setup initializes a new SimApp and context. A Nop logger is set in SimApp.
func Setup(isCheckTx bool) (app *SimApp, ctx sdk.Context) {
	db := dbm.NewMemDB()
	app = NewSimApp(log.NewNopLogger(), db, nil, true, 0)
	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		genesisState := NewDefaultGenesisState()
		stateBytes, err := codec.MarshalJSONIndent(app.Cdc, genesisState)
		if err != nil {
			panic(err)
		}

		// Initialize the chain
		app.InitChain(
			abci.RequestInitChain{
				Validators:    []abci.ValidatorUpdate{},
				AppStateBytes: stateBytes,
			},
		)
	}

	ctx = app.BaseApp.NewContext(isCheckTx, abci.Header{})
	return app, ctx
}
