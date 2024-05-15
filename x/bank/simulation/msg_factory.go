package simulation

import (
	"context"
	"slices"

	"cosmossdk.io/x/bank/keeper"
	types "cosmossdk.io/x/bank/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"golang.org/x/exp/maps"
)

func MsgSendFactory(bk keeper.Keeper) SimMsgFactory {
	return func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) ([]SimAccount, sdk.Msg) {
		from := testData.AnyAccount(reporter, WithSpendableBalance())
		to := testData.AnyAccount(reporter, ExcludeAccounts(from))
		coins := from.LiquidBalance().RandSubsetCoins()

		// Check send_enabled status of each coin denom
		if err := bk.IsSendEnabledCoins(ctx, coins...); err != nil {
			reporter.Skipf("not sendable coins: %s", coins.Denoms())
			return nil, nil
		}
		return []SimAccount{from}, types.NewMsgSend(from.AddressString(), to.AddressString(), coins)
	}
}

func MsgSendToModuleAccountFactory(bk keeper.Keeper) SimMsgFactory {
	return func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) ([]SimAccount, sdk.Msg) {
		from := testData.AnyAccount(reporter, WithSpendableBalance())
		toStr := testData.ModuleAccountAddress(reporter, "distribution")
		coins := from.LiquidBalance().RandSubsetCoins()
		// Check send_enabled status of each coin denom
		if err := bk.IsSendEnabledCoins(ctx, coins...); err != nil {
			reporter.Skipf("not sendable coins: %s", coins.Denoms())
			return nil, nil
		}
		return []SimAccount{from}, types.NewMsgSend(from.AddressString(), toStr, coins)
	}
}

func MsgMultiSendFactory(bk keeper.Keeper) SimMsgFactory {
	return func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) ([]SimAccount, sdk.Msg) {
		r := testData.Rand()
		// random number of inputs/outputs between [1, 3]
		inputs := make([]types.Input, r.Intn(1)+1) //nolint:staticcheck // SA4030: (*math/rand.Rand).Intn(n) generates a random value 0 <= x < n; that is, the generated values don't include n; r.Intn(1) therefore always returns 0
		outputs := make([]types.Output, r.Intn(3)+1)
		senderAcc := make([]SimAccount, len(inputs))
		// use map to check if address already exists as input
		usedAddrs := make(map[string]struct{})

		var totalSentCoins sdk.Coins
		for i := range inputs {
			// generate random input fields, ignore to address
			from := testData.AnyAccount(reporter, WithSpendableBalance(), ExcludeAddresses(maps.Keys(usedAddrs)...))
			if reporter.IsSkipped() {
				return nil, nil
			}
			coins := from.LiquidBalance().RandSubsetCoins()
			fromAddr := from.AddressString()

			// set input address in used address map
			usedAddrs[fromAddr] = struct{}{}

			// set signer privkey
			senderAcc[i] = from

			// set next input and accumulate total sent coins
			inputs[i] = types.NewInput(fromAddr, coins)
			totalSentCoins = totalSentCoins.Add(coins...)
		}

		// Check send_enabled status of each sent coin denom
		if err := bk.IsSendEnabledCoins(ctx, totalSentCoins...); err != nil {
			reporter.Skipf("not sendable coins: %s", totalSentCoins.Denoms())
			return nil, nil
		}

		for i := range outputs {
			out := testData.AnyAccount(reporter)
			outAddr := out.AddressString()
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
				outCoins = simulation.RandSubsetCoins(r, totalSentCoins)
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
