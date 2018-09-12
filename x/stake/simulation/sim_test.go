package simulation

import (
	"encoding/json"
	"math/rand"
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// TestStakeWithRandomMessages
func TestStakeWithRandomMessages(t *testing.T) {
	mapp := mock.NewApp()

	bank.RegisterWire(mapp.Cdc)
	mapper := mapp.AccountMapper
	bankKeeper := bank.NewBaseKeeper(mapper)
	stakeKey := sdk.NewKVStoreKey("stake")
	stakeKeeper := stake.NewKeeper(mapp.Cdc, stakeKey, bankKeeper, stake.DefaultCodespace)
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

	appStateFn := func(r *rand.Rand, keys []crypto.PrivKey, accs []sdk.AccAddress) json.RawMessage {
		mock.RandomSetGenesis(r, mapp, accs, []string{"stake"})
		return json.RawMessage("{}")
	}

	simulation.Simulate(
		t, mapp.BaseApp, appStateFn,
		[]simulation.WeightedOperation{
			{10, SimulateMsgCreateValidator(mapper, stakeKeeper)},
			{5, SimulateMsgEditValidator(stakeKeeper)},
			{15, SimulateMsgDelegate(mapper, stakeKeeper)},
			{10, SimulateMsgBeginUnbonding(mapper, stakeKeeper)},
			{3, SimulateMsgCompleteUnbonding(stakeKeeper)},
			{10, SimulateMsgBeginRedelegate(mapper, stakeKeeper)},
			{3, SimulateMsgCompleteRedelegate(stakeKeeper)},
		}, []simulation.RandSetup{
			Setup(mapp, stakeKeeper),
		}, []simulation.Invariant{
			AllInvariants(bankKeeper, stakeKeeper, mapp.AccountMapper),
		}, 10, 100,
		false,
	)
}
