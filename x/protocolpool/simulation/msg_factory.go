package simulation

import (
	"context"
	"github.com/cosmos/cosmos-sdk/simsx/common"
	"github.com/cosmos/cosmos-sdk/simsx/module"

	"cosmossdk.io/x/protocolpool/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func MsgFundCommunityPoolFactory() module.SimMsgFactoryFn[*types.MsgFundCommunityPool] {
	return func(_ context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *types.MsgFundCommunityPool) {
		funder := testData.AnyAccount(reporter, common.WithSpendableBalance())
		fundAmount := funder.LiquidBalance().RandSubsetCoins(reporter)
		msg := types.NewMsgFundCommunityPool(fundAmount, funder.AddressBech32)
		return []common.SimAccount{funder}, msg
	}
}

// MsgCommunityPoolSpendFactory creates a gov proposal to send tokens from the community pool to a random account
func MsgCommunityPoolSpendFactory() module.SimMsgFactoryFn[*types.MsgCommunityPoolSpend] {
	return func(_ context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *types.MsgCommunityPoolSpend) {
		return nil, &types.MsgCommunityPoolSpend{
			Authority: testData.ModuleAccountAddress(reporter, "gov"),
			Recipient: testData.AnyAccount(reporter).AddressBech32,
			Amount:    must(sdk.ParseCoinsNormalized("100stake,2testtoken")),
		}
	}
}

func must[T any](r T, err error) T {
	if err != nil {
		panic(err)
	}
	return r
}
