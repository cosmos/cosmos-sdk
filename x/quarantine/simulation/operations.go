package simulation

import (
	"bytes"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	"github.com/cosmos/cosmos-sdk/x/quarantine/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Quarantine message types.
var (
	TypeMsgOptIn               = sdk.MsgTypeURL(&quarantine.MsgOptIn{})
	TypeMsgOptOut              = sdk.MsgTypeURL(&quarantine.MsgOptOut{})
	TypeMsgAccept              = sdk.MsgTypeURL(&quarantine.MsgAccept{})
	TypeMsgDecline             = sdk.MsgTypeURL(&quarantine.MsgDecline{})
	TypeMsgUpdateAutoResponses = sdk.MsgTypeURL(&quarantine.MsgUpdateAutoResponses{})
)

// Simulation operation weights constants.
const (
	OpMsgOptIn               = "op_weight_quarantine_msg_opt_in"
	OpMsgOptOut              = "op_weight_quarantine_msg_opt_out"
	OpMsgAccept              = "op_weight_quarantine_msg_accept"
	OpMsgDecline             = "op_weight_quarantine_msg_decline"
	OpMsgUpdateAutoResponses = "op_weight_quarantine_msg_update_auto_responses"
)

// Default weights.
const (
	WeightMsgOptIn               = 100
	WeightMsgOptOut              = 50
	WeightMsgAccept              = 50
	WeightMsgDecline             = 20
	WeightMsgUpdateAutoResponses = 50
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec,
	ak quarantine.AccountKeeper, bk quarantine.BankKeeper, k keeper.Keeper, appCdc cdctypes.AnyUnpacker,
) simulation.WeightedOperations {
	var (
		weightMsgOptIn               int
		weightMsgOptOut              int
		weightMsgAccept              int
		weightMsgDecline             int
		weightMsgUpdateAutoResponses int
	)

	appParams.GetOrGenerate(cdc, OpMsgOptIn, &weightMsgOptIn, nil,
		func(_ *rand.Rand) { weightMsgOptIn = WeightMsgOptIn })
	appParams.GetOrGenerate(cdc, OpMsgOptOut, &weightMsgOptOut, nil,
		func(_ *rand.Rand) { weightMsgOptOut = WeightMsgOptOut })
	appParams.GetOrGenerate(cdc, OpMsgAccept, &weightMsgAccept, nil,
		func(_ *rand.Rand) { weightMsgAccept = WeightMsgAccept })
	appParams.GetOrGenerate(cdc, OpMsgDecline, &weightMsgDecline, nil,
		func(_ *rand.Rand) { weightMsgDecline = WeightMsgDecline })
	appParams.GetOrGenerate(cdc, OpMsgUpdateAutoResponses, &weightMsgUpdateAutoResponses, nil,
		func(_ *rand.Rand) { weightMsgUpdateAutoResponses = WeightMsgUpdateAutoResponses })

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(weightMsgOptIn, SimulateMsgOptIn(ak, bk)),
		simulation.NewWeightedOperation(weightMsgOptOut, SimulateMsgOptOut(ak, bk, k)),
		simulation.NewWeightedOperation(weightMsgAccept, SimulateMsgAccept(ak, bk, k)),
		simulation.NewWeightedOperation(weightMsgDecline, SimulateMsgDecline(ak, bk, k)),
		simulation.NewWeightedOperation(weightMsgUpdateAutoResponses, SimulateMsgUpdateAutoResponses(ak, bk, k)),
	}
}

// SimulateMsgOptIn opts an account into quarantine.
func SimulateMsgOptIn(ak quarantine.AccountKeeper, bk quarantine.BankKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		acct, _ := simtypes.RandomAcc(r, accs)
		msg := &quarantine.MsgOptIn{
			ToAddress: acct.Address.String(),
		}
		msgType := TypeMsgOptIn

		spendableCoins := bk.SpendableCoins(ctx, acct.Address)
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "fee error"), nil, err
		}

		account := ak.GetAccount(ctx, acct.Address)

		encCfg := simappparams.MakeTestEncodingConfig()
		tx, err := helpers.GenSignedMockTx(
			r,
			encCfg.TxConfig,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acct.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", codec.NewProtoCodec(encCfg.InterfaceRegistry)), nil, nil
	}
}

// SimulateMsgOptOut opts an account out of quarantine.
func SimulateMsgOptOut(ak quarantine.AccountKeeper, bk quarantine.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msg := &quarantine.MsgOptOut{}
		msgType := TypeMsgOptOut

		// 3 in 4 chance of using a quarantined account.
		// 1 in 4 chance of using a random account.
		var acct simtypes.Account
		if r.Intn(4) != 0 {
			addr := randomQuarantinedAccount(ctx, r, k)
			if len(addr) == 0 {
				return simtypes.NoOpMsg(msgType, msgType, "no addresses opted in yet"), nil, nil
			}
			acctInd := findAccount(accs, addr)
			if acctInd < 0 {
				return simtypes.NoOpMsg(msgType, msgType, "account not found for quarantined address"), nil, nil
			}
			acct = accs[acctInd]
			msg.ToAddress = addr.String()
		}

		if len(msg.ToAddress) == 0 {
			acct, _ = simtypes.RandomAcc(r, accs)
			msg.ToAddress = acct.Address.String()
		}

		spendableCoins := bk.SpendableCoins(ctx, acct.Address)
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "fee error"), nil, err
		}

		account := ak.GetAccount(ctx, acct.Address)

		encCfg := simappparams.MakeTestEncodingConfig()
		tx, err := helpers.GenSignedMockTx(
			r,
			encCfg.TxConfig,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acct.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", codec.NewProtoCodec(encCfg.InterfaceRegistry)), nil, nil
	}
}

// SimulateMsgAccept accepts quarantined funds.
func SimulateMsgAccept(ak quarantine.AccountKeeper, bk quarantine.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msg := &quarantine.MsgAccept{}
		msgType := TypeMsgAccept

		// 3 in 4 chance of accepting actually quarantined funds (if any exist).
		// 1 in 4 chance of the accept being meaningless.
		// Then, 1 in 5 chance of it being permanent.
		var acct simtypes.Account
		if r.Intn(4) != 0 {
			funds := randomQuarantinedFunds(ctx, r, k)
			if funds == nil {
				return simtypes.NoOpMsg(msgType, msgType, "no funds yet quarantined"), nil, nil
			}
			msg.ToAddress = funds.ToAddress
			msg.FromAddresses = funds.UnacceptedFromAddresses
			addr, err := sdk.AccAddressFromBech32(msg.ToAddress)
			if err != nil {
				return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "invalid to address in quarantined funds"), nil, err
			}
			acctInd := findAccount(accs, addr)
			if acctInd < 0 {
				return simtypes.NoOpMsg(msgType, msgType, "account not found for to address"), nil, nil
			}
			acct = accs[acctInd]
		}

		if len(msg.ToAddress) == 0 {
			acct, _ = simtypes.RandomAcc(r, accs)
			fromAcct, _ := simtypes.RandomAcc(r, accs)
			msg.ToAddress = acct.Address.String()
			msg.FromAddresses = append(msg.FromAddresses, fromAcct.Address.String())
		}

		msg.Permanent = r.Intn(5) == 0

		spendableCoins := bk.SpendableCoins(ctx, acct.Address)
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "fee error"), nil, err
		}

		account := ak.GetAccount(ctx, acct.Address)

		encCfg := simappparams.MakeTestEncodingConfig()
		tx, err := helpers.GenSignedMockTx(
			r,
			encCfg.TxConfig,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acct.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", codec.NewProtoCodec(encCfg.InterfaceRegistry)), nil, nil
	}
}

// SimulateMsgDecline declines quarantined funds.
func SimulateMsgDecline(ak quarantine.AccountKeeper, bk quarantine.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msg := &quarantine.MsgDecline{}
		msgType := TypeMsgDecline

		// 3 in 4 chance of declining actually quarantined funds (if any exist).
		// 1 in 4 chance of the decline being meaningless.
		// Then, 1 in 5 chance of it being permanent.
		var acct simtypes.Account
		if r.Intn(4) != 0 {
			funds := randomQuarantinedFunds(ctx, r, k)
			if funds == nil {
				return simtypes.NoOpMsg(msgType, msgType, "no funds yet quarantined"), nil, nil
			}
			msg.ToAddress = funds.ToAddress
			msg.FromAddresses = funds.UnacceptedFromAddresses
			addr, err := sdk.AccAddressFromBech32(msg.ToAddress)
			if err != nil {
				return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "invalid to address in quarantined funds"), nil, err
			}
			acctInd := findAccount(accs, addr)
			if acctInd < 0 {
				return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "account not found for to address"), nil, nil
			}
			acct = accs[acctInd]
		}

		if len(msg.ToAddress) == 0 {
			acct, _ = simtypes.RandomAcc(r, accs)
			fromAcct, _ := simtypes.RandomAcc(r, accs)
			msg.ToAddress = acct.Address.String()
			msg.FromAddresses = append(msg.FromAddresses, fromAcct.Address.String())
		}

		msg.Permanent = r.Intn(5) == 0

		spendableCoins := bk.SpendableCoins(ctx, acct.Address)
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "fee error"), nil, err
		}

		account := ak.GetAccount(ctx, acct.Address)

		encCfg := simappparams.MakeTestEncodingConfig()
		tx, err := helpers.GenSignedMockTx(
			r,
			encCfg.TxConfig,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acct.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", codec.NewProtoCodec(encCfg.InterfaceRegistry)), nil, nil
	}
}

// SimulateMsgUpdateAutoResponses updates an accounts auto-responses
func SimulateMsgUpdateAutoResponses(ak quarantine.AccountKeeper, bk quarantine.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msg := &quarantine.MsgUpdateAutoResponses{}
		msgType := TypeMsgUpdateAutoResponses

		// 3 in 4 chance of using a quarantined account.
		// 1 to 4 entries.
		// Each entry: 75% accept, 25% decline, 5% unspecified.
		var acct simtypes.Account
		if r.Intn(4) != 0 {
			addr := randomQuarantinedAccount(ctx, r, k)
			if len(addr) == 0 {
				return simtypes.NoOpMsg(msgType, msgType, "no addresses opted in yet"), nil, nil
			}
			acctInd := findAccount(accs, addr)
			if acctInd < 0 {
				return simtypes.NoOpMsg(msgType, msgType, "account not found for quarantined address"), nil, nil
			}
			acct = accs[acctInd]
			msg.ToAddress = addr.String()
		}

		if len(msg.ToAddress) == 0 {
			acct, _ = simtypes.RandomAcc(r, accs)
			msg.ToAddress = acct.Address.String()
		}

		entryCount := r.Intn(3) + 1
		for len(msg.Updates) < entryCount {
			entry := &quarantine.AutoResponseUpdate{}
			acct, _ := simtypes.RandomAcc(r, accs)
			entry.FromAddress = acct.Address.String()
			respR := r.Intn(20)
			switch respR {
			case 0:
				entry.Response = quarantine.AUTO_RESPONSE_UNSPECIFIED
			case 1, 2, 3, 4, 5:
				entry.Response = quarantine.AUTO_RESPONSE_DECLINE
			case 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19:
				entry.Response = quarantine.AUTO_RESPONSE_ACCEPT
			default:
				err := sdkerrors.ErrLogic.Wrapf("response type random number case %d not present in switch", respR)
				return simtypes.NoOpMsg(quarantine.ModuleName, msgType, ""), nil, err
			}
			msg.Updates = append(msg.Updates, entry)
		}

		spendableCoins := bk.SpendableCoins(ctx, acct.Address)
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "fee error"), nil, err
		}

		account := ak.GetAccount(ctx, acct.Address)

		encCfg := simappparams.MakeTestEncodingConfig()
		tx, err := helpers.GenSignedMockTx(
			r,
			encCfg.TxConfig,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acct.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", codec.NewProtoCodec(encCfg.InterfaceRegistry)), nil, nil
	}
}

func randomQuarantinedAccount(ctx sdk.Context, r *rand.Rand, k keeper.Keeper) sdk.AccAddress {
	allQuarantinedAddrs := []*sdk.AccAddress{}
	k.IterateQuarantinedAccounts(ctx, func(toAddr sdk.AccAddress) bool {
		allQuarantinedAddrs = append(allQuarantinedAddrs, &toAddr)
		return false
	})
	rv := randomEntry(r, allQuarantinedAddrs)
	if rv == nil || len(*rv) == 0 {
		return nil
	}
	return *rv
}

func randomQuarantinedFunds(ctx sdk.Context, r *rand.Rand, k keeper.Keeper) *quarantine.QuarantinedFunds {
	return randomEntry(r, k.GetAllQuarantinedFunds(ctx))
}

func randomEntry[V any](r *rand.Rand, addrs []*V) *V {
	if len(addrs) == 0 {
		return nil
	}
	return addrs[r.Intn(len(addrs))]
}

func findAccount(accounts []simtypes.Account, addr sdk.AccAddress) int {
	for i := range accounts {
		if bytes.Equal(addr, accounts[i].Address) {
			return i
		}
	}
	return -1
}
