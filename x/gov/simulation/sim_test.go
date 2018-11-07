package simulation

import (
	"encoding/json"
	"math/rand"
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// TestGovWithRandomMessages
func TestGovWithRandomMessages(t *testing.T) {
	mapp := mock.NewApp()

	bank.RegisterCodec(mapp.Cdc)
	gov.RegisterCodec(mapp.Cdc)
	mapper := mapp.AccountKeeper

	bankKeeper := bank.NewBaseKeeper(mapper)
	stakeKey := sdk.NewKVStoreKey("stake")
	stakeTKey := sdk.NewTransientStoreKey("transient_stake")
	paramKey := sdk.NewKVStoreKey("params")
	paramTKey := sdk.NewTransientStoreKey("transient_params")
	paramKeeper := params.NewKeeper(mapp.Cdc, paramKey, paramTKey)
	stakeKeeper := stake.NewKeeper(mapp.Cdc, stakeKey, stakeTKey, bankKeeper, paramKeeper.Subspace(stake.DefaultParamspace), stake.DefaultCodespace)
	govKey := sdk.NewKVStoreKey("gov")
	govKeeper := gov.NewKeeper(mapp.Cdc, govKey, paramKeeper, paramKeeper.Subspace(gov.DefaultParamspace), bankKeeper, stakeKeeper, gov.DefaultCodespace)
	mapp.Router().AddRoute("gov", gov.NewHandler(govKeeper))
	mapp.SetEndBlocker(func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		gov.EndBlocker(ctx, govKeeper)
		return abci.ResponseEndBlock{}
	})

	err := mapp.CompleteSetup(stakeKey, stakeTKey, paramKey, paramTKey, govKey)
	if err != nil {
		panic(err)
	}

	appStateFn := func(r *rand.Rand, accs []simulation.Account) json.RawMessage {
		simulation.RandomSetGenesis(r, mapp, accs, []string{"stake"})
		return json.RawMessage("{}")
	}

	setup := func(r *rand.Rand, accs []simulation.Account) {
		ctx := mapp.NewContext(false, abci.Header{})
		stake.InitGenesis(ctx, stakeKeeper, stake.DefaultGenesisState())
		gov.InitGenesis(ctx, govKeeper, gov.DefaultGenesisState())
	}

	// Test with unscheduled votes
	simulation.Simulate(
		t, mapp.BaseApp, appStateFn,
		[]simulation.WeightedOperation{
			{2, SimulateMsgSubmitProposal(govKeeper)},
			{3, SimulateMsgDeposit(govKeeper)},
			{20, SimulateMsgVote(govKeeper)},
		}, []simulation.RandSetup{
			setup,
		}, []simulation.Invariant{
			AllInvariants(),
		}, 10, 100,
		false,
	)

	// Test with scheduled votes
	simulation.Simulate(
		t, mapp.BaseApp, appStateFn,
		[]simulation.WeightedOperation{
			{10, SimulateSubmittingVotingAndSlashingForProposal(govKeeper, stakeKeeper)},
			{5, SimulateMsgDeposit(govKeeper)},
		}, []simulation.RandSetup{
			setup,
		}, []simulation.Invariant{
			AllInvariants(),
		}, 10, 100,
		false,
	)
}
