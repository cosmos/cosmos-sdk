package simulation

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
)

// SendTx tests and runs a single msg send where both
// accounts already exist.
func SendMsg(mapper auth.AccountKeeper, bk bank.Keeper) simulation.Operation {
	handler := bank.NewHandler(bk)
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, event func(string)) (action string, fOps []simulation.FutureOperation, err error) {
		fromAcc, action, msg, abort := createSendMsg(r, ctx, accs, mapper)
		if abort {
			return action, nil, nil
		}
		err = sendAndVerifyMsgSend(app, mapper, msg, ctx, []crypto.PrivKey{fromAcc.PrivKey}, handler)
		if err != nil {
			return "", nil, err
		}
		event("bank/sendAndVerifyTxSend/ok")

		return action, nil, nil
	}
}

// SendTx tests and runs a single tx send, with auth where both
// accounts already exist.
func SendTx(mapper auth.AccountKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, event func(string)) (action string, fOps []simulation.FutureOperation, err error) {
		fromAcc, action, msg, abort := createSendMsg(r, ctx, accs, mapper)
		if abort {
			return action, nil, nil
		}
		err = sendAndVerifyMsgSend(app, mapper, msg, ctx, []crypto.PrivKey{fromAcc.PrivKey}, nil)
		if err != nil {
			return "", nil, err
		}
		event("bank/sendAndVerifyTxSend/ok")

		return action, nil, nil
	}
}

func createSendMsg(r *rand.Rand, ctx sdk.Context, accs []simulation.Account, mapper auth.AccountKeeper) (fromAcc simulation.Account, action string, msg bank.MsgSend, abort bool) {
	fromAcc = simulation.RandomAcc(r, accs)
	toAcc := simulation.RandomAcc(r, accs)
	// Disallow sending money to yourself
	for {
		if !fromAcc.PubKey.Equals(toAcc.PubKey) {
			break
		}
		toAcc = simulation.RandomAcc(r, accs)
	}
	toAddr := toAcc.Address
	initFromCoins := mapper.GetAccount(ctx, fromAcc.Address).SpendableCoins(ctx.BlockHeader().Time)

	if len(initFromCoins) == 0 {
		return fromAcc, "skipping, no coins at all", msg, true
	}

	denomIndex := r.Intn(len(initFromCoins))
	amt, goErr := randPositiveInt(r, initFromCoins[denomIndex].Amount)
	if goErr != nil {
		return fromAcc, "skipping bank send due to account having no coins of denomination " + initFromCoins[denomIndex].Denom, msg, true
	}

	action = fmt.Sprintf("%s is sending %s %s to %s",
		fromAcc.Address.String(),
		amt.String(),
		initFromCoins[denomIndex].Denom,
		toAddr.String(),
	)

	coins := sdk.Coins{sdk.NewCoin(initFromCoins[denomIndex].Denom, amt)}
	msg = bank.NewMsgSend(fromAcc.Address, toAcc.Address, coins)
	return
}

// Sends and verifies the transition of a msg send.
func sendAndVerifyMsgSend(app *baseapp.BaseApp, mapper auth.AccountKeeper, msg bank.MsgSend, ctx sdk.Context, privkeys []crypto.PrivKey, handler sdk.Handler) error {
	fromAcc := mapper.GetAccount(ctx, msg.FromAddress)
	AccountNumbers := []uint64{fromAcc.GetAccountNumber()}
	SequenceNumbers := []uint64{fromAcc.GetSequence()}
	initialFromAddrCoins := fromAcc.GetCoins()

	toAcc := mapper.GetAccount(ctx, msg.ToAddress)
	initialToAddrCoins := toAcc.GetCoins()

	if handler != nil {
		res := handler(ctx, msg)
		if !res.IsOK() {
			if res.Code == bank.CodeSendDisabled {
				return nil
			}
			// TODO: Do this in a more 'canonical' way
			return fmt.Errorf("handling msg failed %v", res)
		}
	} else {
		tx := mock.GenTx([]sdk.Msg{msg},
			AccountNumbers,
			SequenceNumbers,
			privkeys...)
		res := app.Deliver(tx)
		if !res.IsOK() {
			// TODO: Do this in a more 'canonical' way
			return fmt.Errorf("Deliver failed %v", res)
		}
	}

	fromAcc = mapper.GetAccount(ctx, msg.FromAddress)
	toAcc = mapper.GetAccount(ctx, msg.ToAddress)

	if !initialFromAddrCoins.Minus(msg.Amount).IsEqual(fromAcc.GetCoins()) {
		return fmt.Errorf("fromAddress %s had an incorrect amount of coins", fromAcc.GetAddress())
	}

	if !initialToAddrCoins.Plus(msg.Amount).IsEqual(toAcc.GetCoins()) {
		return fmt.Errorf("toAddress %s had an incorrect amount of coins", toAcc.GetAddress())
	}

	return nil
}

// SingleInputSendTx tests and runs a single msg multisend w/ auth, with one input and one output, where both
// accounts already exist.
func SingleInputMultiSendTx(mapper auth.AccountKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, event func(string)) (action string, fOps []simulation.FutureOperation, err error) {
		fromAcc, action, msg, abort := createSingleInputMsgMultiSend(r, ctx, accs, mapper)
		if abort {
			return action, nil, nil
		}
		err = sendAndVerifyMsgMultiSend(app, mapper, msg, ctx, []crypto.PrivKey{fromAcc.PrivKey}, nil)
		if err != nil {
			return "", nil, err
		}
		event("bank/sendAndVerifyTxMultiSend/ok")

		return action, nil, nil
	}
}

// SingleInputSendMsg tests and runs a single msg multisend, with one input and one output, where both
// accounts already exist.
func SingleInputMsgMultiSend(mapper auth.AccountKeeper, bk bank.Keeper) simulation.Operation {
	handler := bank.NewHandler(bk)
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, event func(string)) (action string, fOps []simulation.FutureOperation, err error) {
		fromAcc, action, msg, abort := createSingleInputMsgMultiSend(r, ctx, accs, mapper)
		if abort {
			return action, nil, nil
		}
		err = sendAndVerifyMsgMultiSend(app, mapper, msg, ctx, []crypto.PrivKey{fromAcc.PrivKey}, handler)
		if err != nil {
			return "", nil, err
		}
		event("bank/sendAndVerifyMsgMultiSend/ok")

		return action, nil, nil
	}
}

func createSingleInputMsgMultiSend(r *rand.Rand, ctx sdk.Context, accs []simulation.Account, mapper auth.AccountKeeper) (fromAcc simulation.Account, action string, msg bank.MsgMultiSend, abort bool) {
	fromAcc = simulation.RandomAcc(r, accs)
	toAcc := simulation.RandomAcc(r, accs)
	// Disallow sending money to yourself
	for {
		if !fromAcc.PubKey.Equals(toAcc.PubKey) {
			break
		}
		toAcc = simulation.RandomAcc(r, accs)
	}
	toAddr := toAcc.Address
	initFromCoins := mapper.GetAccount(ctx, fromAcc.Address).SpendableCoins(ctx.BlockHeader().Time)

	if len(initFromCoins) == 0 {
		return fromAcc, "skipping, no coins at all", msg, true
	}

	denomIndex := r.Intn(len(initFromCoins))
	amt, goErr := randPositiveInt(r, initFromCoins[denomIndex].Amount)
	if goErr != nil {
		return fromAcc, "skipping bank send due to account having no coins of denomination " + initFromCoins[denomIndex].Denom, msg, true
	}

	action = fmt.Sprintf("%s is sending %s %s to %s",
		fromAcc.Address.String(),
		amt.String(),
		initFromCoins[denomIndex].Denom,
		toAddr.String(),
	)

	coins := sdk.Coins{sdk.NewCoin(initFromCoins[denomIndex].Denom, amt)}
	msg = bank.MsgMultiSend{
		Inputs:  []bank.Input{bank.NewInput(fromAcc.Address, coins)},
		Outputs: []bank.Output{bank.NewOutput(toAddr, coins)},
	}
	return
}

// Sends and verifies the transition of a msg multisend. This fails if there are repeated inputs or outputs
// pass in handler as nil to handle txs, otherwise handle msgs
func sendAndVerifyMsgMultiSend(app *baseapp.BaseApp, mapper auth.AccountKeeper, msg bank.MsgMultiSend, ctx sdk.Context, privkeys []crypto.PrivKey, handler sdk.Handler) error {
	initialInputAddrCoins := make([]sdk.Coins, len(msg.Inputs))
	initialOutputAddrCoins := make([]sdk.Coins, len(msg.Outputs))
	AccountNumbers := make([]uint64, len(msg.Inputs))
	SequenceNumbers := make([]uint64, len(msg.Inputs))

	for i := 0; i < len(msg.Inputs); i++ {
		acc := mapper.GetAccount(ctx, msg.Inputs[i].Address)
		AccountNumbers[i] = acc.GetAccountNumber()
		SequenceNumbers[i] = acc.GetSequence()
		initialInputAddrCoins[i] = acc.GetCoins()
	}
	for i := 0; i < len(msg.Outputs); i++ {
		acc := mapper.GetAccount(ctx, msg.Outputs[i].Address)
		initialOutputAddrCoins[i] = acc.GetCoins()
	}
	if handler != nil {
		res := handler(ctx, msg)
		if !res.IsOK() {
			if res.Code == bank.CodeSendDisabled {
				return nil
			}
			// TODO: Do this in a more 'canonical' way
			return fmt.Errorf("handling msg failed %v", res)
		}
	} else {
		tx := mock.GenTx([]sdk.Msg{msg},
			AccountNumbers,
			SequenceNumbers,
			privkeys...)
		res := app.Deliver(tx)
		if !res.IsOK() {
			// TODO: Do this in a more 'canonical' way
			return fmt.Errorf("Deliver failed %v", res)
		}
	}

	for i := 0; i < len(msg.Inputs); i++ {
		terminalInputCoins := mapper.GetAccount(ctx, msg.Inputs[i].Address).GetCoins()
		if !initialInputAddrCoins[i].Minus(msg.Inputs[i].Coins).IsEqual(terminalInputCoins) {
			return fmt.Errorf("input #%d had an incorrect amount of coins", i)
		}
	}
	for i := 0; i < len(msg.Outputs); i++ {
		terminalOutputCoins := mapper.GetAccount(ctx, msg.Outputs[i].Address).GetCoins()
		if !terminalOutputCoins.IsEqual(initialOutputAddrCoins[i].Plus(msg.Outputs[i].Coins)) {
			return fmt.Errorf("output #%d had an incorrect amount of coins", i)
		}
	}
	return nil
}

func randPositiveInt(r *rand.Rand, max sdk.Int) (sdk.Int, error) {
	if !max.GT(sdk.OneInt()) {
		return sdk.Int{}, errors.New("max too small")
	}
	max = max.Sub(sdk.OneInt())
	return sdk.NewIntFromBigInt(new(big.Int).Rand(r, max.BigInt())).Add(sdk.OneInt()), nil
}
