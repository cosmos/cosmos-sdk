package app

import (
	"fmt"
	"time"

	banksim "github.com/cosmos/cosmos-sdk/x/bank/simulation"
	distrsim "github.com/cosmos/cosmos-sdk/x/distribution/simulation"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	stakesim "github.com/cosmos/cosmos-sdk/x/stake/simulation"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (app *GaiaApp) runtimeInvariants() []simulation.Invariant {
	return []simulation.Invariant{
		banksim.NonnegativeBalanceInvariant(app.accountKeeper),
		distrsim.ValAccumInvariants(app.distrKeeper, app.stakeKeeper),
		stakesim.SupplyInvariants(app.bankKeeper, app.stakeKeeper,
			app.feeCollectionKeeper, app.distrKeeper, app.accountKeeper),
		stakesim.PositivePowerInvariant(app.stakeKeeper),
	}
}

func (app *GaiaApp) assertRuntimeInvariants() {
	invariants := app.runtimeInvariants()
	start := time.Now()
	ctx := app.NewContext(false, abci.Header{Height: app.LastBlockHeight() + 1})
	for _, inv := range invariants {
		if err := inv(ctx); err != nil {
			panic(fmt.Errorf("invariant broken: %s", err))
		}
	}
	end := time.Now()
	diff := end.Sub(start)
	app.BaseApp.Logger.With("module", "invariants").Info("Asserted all invariants", "duration", diff)
}
