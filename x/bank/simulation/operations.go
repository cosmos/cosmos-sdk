package simulation

import (
	"math/rand"

	"cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// Simulation operation weights constants
const (
	DistributionModuleName = "distribution"
)

// GenerateMsgSend generates a randomized single send transaction.
func GenerateMsgSend(r *rand.Rand, ctx sdk.Context, clientCtx client.Context, moniker string, to, from simtypes.Account, bk keeper.Keeper, ak types.AccountKeeper) sdk.Tx {
	spendable := bk.SpendableCoins(ctx, from.Address)

	coins := simtypes.RandSubsetCoins(r, spendable)

	msg := types.NewMsgSend(from.Address, to.Address, coins)

	remaining := spendable.Sub(msg.Amount...)
	fees, _ := simtypes.RandomFees(r, remaining)

	acc := ak.GetAccount(ctx, from.Address)
	tx, err := simtestutil.GenSignedMockTx(
		r,
		clientCtx.TxConfig,
		[]sdk.Msg{msg},
		fees,
		simtestutil.DefaultGenTxGas,
		ctx.ChainID(),
		[]uint64{acc.GetAccountNumber()},
		[]uint64{acc.GetSequence()},
		[]cryptotypes.PrivKey{from.PrivKey}...,
	)
	if err != nil {
		panic(err)
	}
	return tx
}

// GenerateMsgMultiSend generates a randomized multi send transaction.
func GenerateMsgMultiSend(r *rand.Rand, ctx sdk.Context, clientCtx client.Context, moniker string, accs []simtypes.Account, bk keeper.Keeper, ak types.AccountKeeper) sdk.Tx {
	// random number of inputs/outputs between [1, 3]
	inputs := make([]types.Input, r.Intn(1)+1) //nolint:staticcheck // SA4030: (*math/rand.Rand).Intn(n) generates a random value 0 <= x < n; that is, the generated values don't include n; r.Intn(1) therefore always returns 0
	outputs := make([]types.Output, r.Intn(3)+1)

	// collect signer privKeys
	privs := make([]cryptotypes.PrivKey, len(inputs))

	// use map to check if address already exists as input
	usedAddrs := make(map[string]bool)

	var totalSentCoins sdk.Coins
	for i := range inputs {
		from, _ := simtypes.RandomAcc(r, accs)
		to, _ := simtypes.RandomAcc(r, accs)

		for usedAddrs[from.Address.String()] {
			from, _ = simtypes.RandomAcc(r, accs)
		}

		// disallow sending money to yourself
		for from.PubKey.Equals(to.PubKey) {
			to, _ = simtypes.RandomAcc(r, accs)
		}

		spendable := bk.SpendableCoins(ctx, from.Address)

		coins := simtypes.RandSubsetCoins(r, spendable)

		// set input address in used address map
		usedAddrs[from.Address.String()] = true

		// set signer privkey
		privs[i] = from.PrivKey

		// set next input and accumulate total sent coins
		inputs[i] = types.NewInput(from.Address, coins)
		totalSentCoins = totalSentCoins.Add(coins...)
	}

	for o := range outputs {
		outAddr, _ := simtypes.RandomAcc(r, accs)

		var outCoins sdk.Coins
		// split total sent coins into random subsets for output
		if o == len(outputs)-1 {
			outCoins = totalSentCoins
		} else {
			// take random subset of remaining coins for output
			// and update remaining coins
			outCoins = simtypes.RandSubsetCoins(r, totalSentCoins)
			totalSentCoins = totalSentCoins.Sub(outCoins...)
		}

		outputs[o] = types.NewOutput(outAddr.Address, outCoins)
	}

	return generateMsgMultiSend(r, ctx, clientCtx.TxConfig, bk, ak, privs, &types.MsgMultiSend{
		Inputs:  inputs,
		Outputs: outputs,
	})
}

func generateMsgMultiSend(r *rand.Rand, ctx sdk.Context, txGen client.TxConfig, bk keeper.Keeper, ak types.AccountKeeper, privs []cryptotypes.PrivKey, msg *types.MsgMultiSend) sdk.Tx {
	// remove any output that has no coins

	for i := 0; i < len(msg.Outputs); {
		if msg.Outputs[i].Coins.Empty() {
			msg.Outputs[i] = msg.Outputs[len(msg.Outputs)-1]
			msg.Outputs = msg.Outputs[:len(msg.Outputs)-1]
		} else {
			// continue onto next coin
			i++
		}
	}

	accountNumbers := make([]uint64, len(msg.Inputs))
	sequenceNumbers := make([]uint64, len(msg.Inputs))
	for i := 0; i < len(msg.Inputs); i++ {
		addr, err := ak.AddressCodec().StringToBytes(msg.Inputs[i].Address)
		if err != nil {
			panic(err)
		}

		acc := ak.GetAccount(ctx, addr)
		accountNumbers[i] = acc.GetAccountNumber()
		sequenceNumbers[i] = acc.GetSequence()
	}

	addr, err := ak.AddressCodec().StringToBytes(msg.Inputs[0].Address)
	if err != nil {
		panic(err)
	}
	// feePayer is the first signer, i.e. first input address
	feePayer := ak.GetAccount(ctx, addr)
	spendable := bk.SpendableCoins(ctx, feePayer.GetAddress())
	coins := spendable.Sub(msg.Inputs[0].Coins...)
	fees, _ := simtypes.RandomFees(r, coins)
	tx, err := simtestutil.GenSignedMockTx(
		r,
		txGen,
		[]sdk.Msg{msg},
		fees,
		simtestutil.DefaultGenTxGas,
		ctx.ChainID(),
		accountNumbers,
		sequenceNumbers,
		privs...,
	)
	if err != nil {
		panic(err)
	}
	return tx
}

// GenerateMsgMultiSend generates a randomized multi send transaction.
func GenerateMsgMultiSendToModuleAccount(r *rand.Rand, ctx sdk.Context, clientCtx client.Context, moniker string, accs []simtypes.Account, bk keeper.Keeper, ak types.AccountKeeper, moduleAccCount int) sdk.Tx {
	inputs := make([]types.Input, 2)
	outputs := make([]types.Output, moduleAccCount)
	// collect signer privKeys
	privs := make([]cryptotypes.PrivKey, len(inputs))
	var totalSentCoins sdk.Coins
	for i := range inputs {
		sender := accs[i]
		privs[i] = sender.PrivKey
		spendable := bk.SpendableCoins(ctx, sender.Address)
		coins := simtypes.RandSubsetCoins(r, spendable)
		inputs[i] = types.NewInput(sender.Address, coins)
		totalSentCoins = totalSentCoins.Add(coins...)
	}
	moduleAccounts := getModuleAccounts(ak, ctx, moduleAccCount)
	for i := range outputs {
		var outCoins sdk.Coins
		// split total sent coins into random subsets for output
		if i == len(outputs)-1 {
			outCoins = totalSentCoins
		} else {
			// take random subset of remaining coins for output
			// and update remaining coins
			outCoins = simtypes.RandSubsetCoins(r, totalSentCoins)
			totalSentCoins = totalSentCoins.Sub(outCoins...)
		}
		outputs[i] = types.NewOutput(moduleAccounts[i].Address, outCoins)
	}
	return generateMsgMultiSend(r, ctx, clientCtx.TxConfig, bk, ak, privs, &types.MsgMultiSend{
		Inputs:  inputs,
		Outputs: outputs,
	})
}

func getModuleAccounts(ak types.AccountKeeper, ctx sdk.Context, moduleAccCount int) []simtypes.Account {
	moduleAccounts := make([]simtypes.Account, moduleAccCount)

	acc := ak.GetModuleAccount(ctx, DistributionModuleName)
	for i := 0; i < moduleAccCount; i++ {
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
