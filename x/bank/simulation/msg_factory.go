package simulation

import (
	"context"
	"slices"

	"github.com/cosmos/cosmos-sdk/simsx"

	"cosmossdk.io/x/bank/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/exp/maps"
)

func MsgSendFactory() simsx.SimMsgFactoryFn[*types.MsgSend] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, sdk.Msg) {
		from := testData.AnyAccount(reporter, simsx.WithSpendableBalance())
		to := testData.AnyAccount(reporter, simsx.ExcludeAccounts(from))
		coins := from.LiquidBalance().RandSubsetCoins(reporter, simsx.WithSendEnabledCoins())
		return []simsx.SimAccount{from}, types.NewMsgSend(from.AddressBech32, to.AddressBech32, coins)
	}
}

func MsgSendToModuleAccountFactory() simsx.SimMsgFactoryFn[*types.MsgSend] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, sdk.Msg) {
		from := testData.AnyAccount(reporter, simsx.WithSpendableBalance())
		toStr := testData.ModuleAccountAddress(reporter, "distribution")
		coins := from.LiquidBalance().RandSubsetCoins(reporter, simsx.WithSendEnabledCoins())
		return []simsx.SimAccount{from}, types.NewMsgSend(from.AddressBech32, toStr, coins)
	}
}

func MsgMultiSendFactory() simsx.SimMsgFactoryFn[*types.MsgMultiSend] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, sdk.Msg) {
		r := testData.Rand()
		// random number of inputs/outputs between [1, 3]
		inputs := make([]types.Input, r.Intn(1)+1) //nolint:staticcheck // SA4030: (*math/rand.Rand).Intn(n) generates a random value 0 <= x < n; that is, the generated values don't include n; r.Intn(1) therefore always returns 0
		outputs := make([]types.Output, r.Intn(3)+1)
		senderAcc := make([]simsx.SimAccount, len(inputs))
		// use map to check if address already exists as input
		usedAddrs := make(map[string]struct{})

		var totalSentCoins sdk.Coins
		for i := range inputs {
			// generate random input fields, ignore to address
			from := testData.AnyAccount(reporter, simsx.WithSpendableBalance(), simsx.ExcludeAddresses(maps.Keys(usedAddrs)...))
			if reporter.IsSkipped() {
				return nil, nil
			}
			coins := from.LiquidBalance().RandSubsetCoins(reporter, simsx.WithSendEnabledCoins())
			fromAddr := from.AddressBech32

			// set input address in used address map
			usedAddrs[fromAddr] = struct{}{}

			// set signer privkey
			senderAcc[i] = from

			// set next input and accumulate total sent coins
			inputs[i] = types.NewInput(fromAddr, coins)
			totalSentCoins = totalSentCoins.Add(coins...)
		}

		for i := range outputs {
			out := testData.AnyAccount(reporter)
			outAddr := out.AddressBech32
			if reporter.IsSkipped() {
				return nil, nil
			}

			var outCoins sdk.Coins
			// split total sent coins into random subsets for output
			if i == len(outputs)-1 {
				outCoins = totalSentCoins
			} else {
				// take random subset of remaining coins for output
				// and update remaining coins
				outCoins = r.SubsetCoins(totalSentCoins)
				totalSentCoins = totalSentCoins.Sub(outCoins...)
			}

			outputs[i] = types.NewOutput(outAddr, outCoins)
		}

		// remove any output that has no coins
		slices.DeleteFunc(outputs, func(o types.Output) bool {
			return o.Coins.Empty()
		})
		return senderAcc, &types.MsgMultiSend{Inputs: inputs, Outputs: outputs}
	}
}
