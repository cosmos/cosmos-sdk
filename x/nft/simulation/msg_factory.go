package simulation

import (
	"context"

	"github.com/cosmos/cosmos-sdk/testutil/simsx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft"        //nolint:staticcheck // deprecated and to be removed
	"github.com/cosmos/cosmos-sdk/x/nft/keeper" //nolint:staticcheck // deprecated and to be removed
)

func MsgSendFactory(k keeper.Keeper) simsx.SimMsgFactoryFn[*nft.MsgSend] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *nft.MsgSend) {
		from := testData.AnyAccount(reporter, simsx.WithSpendableBalance())
		to := testData.AnyAccount(reporter, simsx.ExcludeAccounts(from))
		if reporter.IsSkipped() {
			return nil, nil
		}
		n, err := randNFT(sdk.UnwrapSDKContext(ctx), testData.Rand().Rand, k, from.Address)
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		msg := &nft.MsgSend{
			ClassId:  n.ClassId,
			Id:       n.Id,
			Sender:   from.AddressBech32,
			Receiver: to.AddressBech32,
		}

		return []simsx.SimAccount{from}, msg
	}
}
