package simulation

import (
	"fmt"
	"math/rand"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/keeper"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

var TypeFundCommunityPool = sdk.MsgTypeURL(&types.MsgFundCommunityPool{})

// Simulation operation weights constants
const (
	OpWeightMsgFundCommunityPool = "op_weight_msg_fund_community_pool"

	DefaultWeightMsgFundCommunityPool = 100
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams,
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simulation.WeightedOperations {
	var weightMsgFundCommunityPool int

	appParams.GetOrGenerate(OpWeightMsgFundCommunityPool, &weightMsgFundCommunityPool, nil,
		func(_ *rand.Rand) {
			weightMsgFundCommunityPool = DefaultWeightMsgFundCommunityPool
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgFundCommunityPool,
			SimulateMsgFundCommunityPool(txGen, ak, bk, k),
		),
	}
}

func SimulateMsgFundCommunityPool(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	_ keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())
		// choose 10% - 70% of spendable
		// we do not want to use the full balance so that fees can be covered
		decimalVal := r.Intn(6) + 1 // [1:7]
		spendableSubAmount := keeper.PercentageCoinMul(math.LegacyMustNewDecFromStr(fmt.Sprintf("0.%d", decimalVal)), spendable)
		if spendableSubAmount.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, TypeFundCommunityPool, "no balance"), nil, nil
		}

		msg := &types.MsgFundCommunityPool{
			Amount:    spendableSubAmount,
			Depositor: account.GetAddress().String(),
		}

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           txGen,
			Cdc:             nil,
			Msg:             msg,
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendableSubAmount,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
