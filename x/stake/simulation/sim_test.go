package simulation

import (
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/cosmos/cosmos-sdk/x/stake"
	abci "github.com/tendermint/tendermint/abci/types"
)

// TestStakeWithRandomMessages
func TestStakeWithRandomMessages(t *testing.T) {
	mapp := mock.NewApp()

	bank.RegisterWire(mapp.Cdc)
	mapper := mapp.AccountMapper
	coinKeeper := bank.NewKeeper(mapper)
	stakeKey := sdk.NewKVStoreKey("stake")
	stakeKeeper := stake.NewKeeper(mapp.Cdc, stakeKey, coinKeeper, stake.DefaultCodespace)
	mapp.Router().AddRoute("stake", stake.NewHandler(stakeKeeper))
	mapp.SetEndBlocker(func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		validatorUpdates := stake.EndBlocker(ctx, stakeKeeper)
		return abci.ResponseEndBlock{
			ValidatorUpdates: validatorUpdates,
		}
	})

	err := mapp.CompleteSetup([]*sdk.KVStoreKey{stakeKey})
	if err != nil {
		panic(err)
	}

	simulation.Simulate(
		t, mapp.BaseApp, json.RawMessage("{}"),
		[]simulation.TestAndRunTx{
			SimulateMsgCreateValidator(mapper, stakeKeeper),
			SimulateMsgEditValidator(stakeKeeper),
			SimulateMsgDelegate(mapper, stakeKeeper),
			// XXX TODO
			// SimulateMsgBeginUnbonding(mapper, stakeKeeper),
			SimulateMsgCompleteUnbonding(stakeKeeper),
			SimulateMsgBeginRedelegate(mapper, stakeKeeper),
			SimulateMsgCompleteRedelegate(stakeKeeper),
		}, []simulation.RandSetup{
			SimulationSetup(mapp, stakeKeeper),
		}, []simulation.Invariant{
			AllInvariants(coinKeeper, stakeKeeper, mapp.AccountMapper),
		}, 10, 100, 500,
	)
}
