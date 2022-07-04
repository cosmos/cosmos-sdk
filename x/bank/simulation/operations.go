package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants
//nolint:gosec // these are not hardcoded credentials.
const (
	OpWeightMsgSend = "op_weight_msg_send"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper, bk keeper.Keeper,
) simulation.WeightedOperations {
	var weightMsgSend int
	appParams.GetOrGenerate(cdc, OpWeightMsgSend, &weightMsgSend, nil,
		func(_ *rand.Rand) {
			weightMsgSend = simappparams.DefaultWeightMsgSend
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgSend,
			SimulateMsgSend(ak, bk),
		),
	}
}

// SimulateMsgSend tests and runs a single msg send where both
// accounts already exist.
func SimulateMsgSend(ak types.AccountKeeper, bk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, to, coins, skip := randomSendFields(r, ctx, accs, bk, ak)

		// if coins slice is empty, we can not create valid types.MsgSend
		if len(coins) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSend, "empty coins slice"), nil, nil
		}

		// Check send_enabled status of each coin denom
		if err := bk.IsSendEnabledCoins(ctx, coins...); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSend, err.Error()), nil, nil
		}

		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSend, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgSend(from.Address, to.Address, coins)

		err := sendMsgSend(r, app, bk, ak, msg, ctx, chainID, []cryptotypes.PrivKey{from.PrivKey})
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "invalid transfers"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

// SimulateMsgSendToModuleAccount tests and runs a single msg send where both
// accounts already exist.
func SimulateMsgSendToModuleAccount(ak types.AccountKeeper, bk keeper.Keeper, moduleAccCount int) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from := accs[0]

		to := getModuleAccounts(ak, ctx, moduleAccCount)[0]

		spendable := bk.SpendableCoins(ctx, from.Address)
		coins := simtypes.RandSubsetCoins(r, spendable)
		// if coins slice is empty, we can not create valid types.MsgSend
		if len(coins) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSend, "empty coins slice"), nil, nil
		}

		// Check send_enabled status of each coin denom
		if err := bk.IsSendEnabledCoins(ctx, coins...); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSend, err.Error()), nil, nil
		}

		msg := types.NewMsgSend(from.Address, to.Address, coins)

		err := sendMsgSend(r, app, bk, ak, msg, ctx, chainID, []cryptotypes.PrivKey{from.PrivKey})
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "invalid transfers"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

// sendMsgSend sends a transaction with a MsgSend from a provided random account.
func sendMsgSend(
	r *rand.Rand, app *baseapp.BaseApp, bk keeper.Keeper, ak types.AccountKeeper,
	msg *types.MsgSend, ctx sdk.Context, chainID string, privkeys []cryptotypes.PrivKey,
) error {
	var (
		fees sdk.Coins
		err  error
	)

	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return err
	}

	account := ak.GetAccount(ctx, from)
	spendable := bk.SpendableCoins(ctx, account.GetAddress())

	coins, hasNeg := spendable.SafeSub(msg.Amount...)
	if !hasNeg {
		fees, err = simtypes.RandomFees(r, ctx, coins)
		if err != nil {
			return err
		}
	}
	txGen := simappparams.MakeTestEncodingConfig().TxConfig
	tx, err := simtestutil.GenSignedMockTx(
		r,
		txGen,
		[]sdk.Msg{msg},
		fees,
		simtestutil.DefaultGenTxGas,
		chainID,
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		privkeys...,
	)
	if err != nil {
		return err
	}

	_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
	if err != nil {
		return err
	}

	return nil
}

// randomSendFields returns the sender and recipient simulation accounts as well
// as the transferred amount.
func randomSendFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, bk keeper.Keeper, ak types.AccountKeeper,
) (simtypes.Account, simtypes.Account, sdk.Coins, bool) {
	from, _ := simtypes.RandomAcc(r, accs)
	to, _ := simtypes.RandomAcc(r, accs)

	// disallow sending money to yourself
	for from.PubKey.Equals(to.PubKey) {
		to, _ = simtypes.RandomAcc(r, accs)
	}

	acc := ak.GetAccount(ctx, from.Address)
	if acc == nil {
		return from, to, nil, true
	}

	spendable := bk.SpendableCoins(ctx, acc.GetAddress())

	sendCoins := simtypes.RandSubsetCoins(r, spendable)
	if sendCoins.Empty() {
		return from, to, nil, true
	}

	return from, to, sendCoins, false
}

func getModuleAccounts(ak types.AccountKeeper, ctx sdk.Context, moduleAccCount int) []simtypes.Account {
	moduleAccounts := make([]simtypes.Account, moduleAccCount)

	for i := 0; i < moduleAccCount; i++ {
		addr := ak.GetModuleAddress(distributiontypes.ModuleName)
		acc := ak.GetAccount(ctx, addr)
		mAcc := simtypes.Account{
			Address: acc.GetAddress(),
			PrivKey: nil,
			ConsKey: nil,
			PubKey:  acc.GetPubKey(),
		}
		moduleAccounts[i] = mAcc
	}

	return moduleAccounts
}
