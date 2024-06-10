package simulation

import (
	"context"

	"cosmossdk.io/x/authz"
	banktypes "cosmossdk.io/x/bank/types"
	"github.com/cosmos/cosmos-sdk/simsx"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func MsgSendFactory() simsx.SimMsgFactoryFn[*authz.MsgGrant] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, sdk.Msg) {
		granter := testData.AnyAccount(reporter, simsx.WithSpendableBalance())
		grantee := testData.AnyAccount(reporter, simsx.ExcludeAccounts(granter))
		spendLimit := granter.LiquidBalance().RandSubsetCoins(reporter)
		expiration := testData.Rand().Timestamp()
		var randomAuthz authz.Authorization
		if testData.Rand().Bool() {
			randomAuthz = banktypes.NewSendAuthorization(spendLimit, nil, nil)
		} else {
			randomAuthz = authz.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{}))
		}

		msg, err := authz.NewMsgGrant(granter.AddressBech32, grantee.AddressBech32, randomAuthz, &expiration)
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		return []simsx.SimAccount{granter}, msg
	}
}
