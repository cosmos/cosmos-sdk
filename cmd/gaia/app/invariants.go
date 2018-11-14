package app

import (
	"fmt"

	banksim "github.com/cosmos/cosmos-sdk/x/bank/simulation"
	distrsim "github.com/cosmos/cosmos-sdk/x/distribution/simulation"
	govsim "github.com/cosmos/cosmos-sdk/x/gov/simulation"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	slashingsim "github.com/cosmos/cosmos-sdk/x/slashing/simulation"
	stakesim "github.com/cosmos/cosmos-sdk/x/stake/simulation"
)

func (app *GaiaApp) runtimeInvariants() []simulation.Invariant {
	return []simulation.Invariant{
		banksim.NonnegativeBalanceInvariant(app.accountKeeper),
		govsim.AllInvariants(),
		distrsim.AllInvariants(app.distrKeeper, app.stakeKeeper),
		stakesim.AllInvariants(app.bankKeeper, app.stakeKeeper,
			app.feeCollectionKeeper, app.distrKeeper, app.accountKeeper),
		slashingsim.AllInvariants(),
	}
}

func (app *GaiaApp) assertRuntimeInvariants() {
	invariants := app.runtimeInvariants()
	for _, inv := range invariants {
		if err := inv(app.BaseApp); err != nil {
			panic(fmt.Errorf("invariant broken: %s", err))
		}
	}
}
