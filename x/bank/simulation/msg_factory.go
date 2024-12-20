package simulation

import (
	"context"
	"github.com/cosmos/cosmos-sdk/simsx/common"
	"github.com/cosmos/cosmos-sdk/simsx/module"
	"slices"

	"cosmossdk.io/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func MsgSendFactory() module.SimMsgFactoryFn[*types.MsgSend] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *types.MsgSend) {
		from := testData.AnyAccount(reporter, common.WithSpendableBalance())
		to := testData.AnyAccount(reporter, common.ExcludeAccounts(from))
		coins := from.LiquidBalance().RandSubsetCoins(reporter, common.WithSendEnabledCoins())
		return []common.SimAccount{from}, types.NewMsgSend(from.AddressBech32, to.AddressBech32, coins)
	}
}

func MsgMultiSendFactory() module.SimMsgFactoryFn[*types.MsgMultiSend] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *types.MsgMultiSend) {
		r := testData.Rand()
		var (
			sending              = make([]types.Input, 1)
			receiving            = make([]types.Output, r.Intn(3)+1)
			senderAcc            = make([]common.SimAccount, len(sending))
			totalSentCoins       sdk.Coins
			uniqueAccountsFilter = common.UniqueAccounts()
		)
		for i := range sending {
			// generate random input fields, ignore to address
			from := testData.AnyAccount(reporter, common.WithSpendableBalance(), uniqueAccountsFilter)
			if reporter.IsAborted() {
				return nil, nil
			}
			coins := from.LiquidBalance().RandSubsetCoins(reporter, common.WithSendEnabledCoins())

			// set signer privkey
			senderAcc[i] = from

			// set next input and accumulate total sent coins
			sending[i] = types.NewInput(from.AddressBech32, coins)
			totalSentCoins = totalSentCoins.Add(coins...)
		}

		for i := range receiving {
			receiver := testData.AnyAccount(reporter)
			if reporter.IsAborted() {
				return nil, nil
			}

			var outCoins sdk.Coins
			// split total sent coins into random subsets for output
			if i == len(receiving)-1 {
				// last one receives remaining amount
				outCoins = totalSentCoins
			} else {
				// take random subset of remaining coins for output
				// and update remaining coins
				outCoins = r.SubsetCoins(totalSentCoins)
				totalSentCoins = totalSentCoins.Sub(outCoins...)
			}

			receiving[i] = types.NewOutput(receiver.AddressBech32, outCoins)
		}

		// remove any entries that have no coins
		receiving = slices.DeleteFunc(receiving, func(o types.Output) bool {
			return o.Address == "" || o.Coins.Empty()
		})
		return senderAcc, &types.MsgMultiSend{Inputs: sending, Outputs: receiving}
	}
}

// MsgUpdateParamsFactory creates a gov proposal for param updates
func MsgUpdateParamsFactory() module.SimMsgFactoryFn[*types.MsgUpdateParams] {
	return func(_ context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *types.MsgUpdateParams) {
		params := types.DefaultParams()
		params.DefaultSendEnabled = testData.Rand().Intn(2) == 0
		return nil, &types.MsgUpdateParams{
			Authority: testData.ModuleAccountAddress(reporter, "gov"),
			Params:    params,
		}
	}
}
