package simulation

import (
	"context"

	"cosmossdk.io/x/protocolpool/types"
	"github.com/cosmos/cosmos-sdk/simsx"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func MsgCreateValidatorFactory() simsx.SimMsgFactoryFn[*types.MsgFundCommunityPool] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, sdk.Msg) {
		funder := testData.AnyAccount(reporter, simsx.WithSpendableBalance())
		fundAmount := funder.LiquidBalance().RandSubsetCoins(reporter)
		msg := types.NewMsgFundCommunityPool(fundAmount, funder.AddressBech32)
		return []simsx.SimAccount{funder}, msg
	}
}
