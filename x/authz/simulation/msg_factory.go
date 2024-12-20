package simulation

import (
	"context"
	"github.com/cosmos/cosmos-sdk/simsx/common"
	"github.com/cosmos/cosmos-sdk/simsx/module"
	"time"

	"cosmossdk.io/x/authz"
	"cosmossdk.io/x/authz/keeper"
	banktype "cosmossdk.io/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func MsgGrantFactory() module.SimMsgFactoryFn[*authz.MsgGrant] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *authz.MsgGrant) {
		granter := testData.AnyAccount(reporter, common.WithSpendableBalance())
		grantee := testData.AnyAccount(reporter, common.ExcludeAccounts(granter))
		spendLimit := granter.LiquidBalance().RandSubsetCoins(reporter, common.WithSendEnabledCoins())

		r := testData.Rand()
		var expiration *time.Time
		if t1 := r.Timestamp(); !t1.Before(common.BlockTime(ctx)) {
			expiration = &t1
		}
		// pick random authorization
		authorizations := []authz.Authorization{
			banktype.NewSendAuthorization(spendLimit, nil, testData.AddressCodec()),
			authz.NewGenericAuthorization(sdk.MsgTypeURL(&banktype.MsgSend{})),
		}
		randomAuthz := common.OneOf(r, authorizations)

		msg, err := authz.NewMsgGrant(granter.AddressBech32, grantee.AddressBech32, randomAuthz, expiration)
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		return []common.SimAccount{granter}, msg
	}
}

func MsgExecFactory(k keeper.Keeper) module.SimMsgFactoryFn[*authz.MsgExec] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *authz.MsgExec) {
		bankSendOnlyFilter := func(a authz.Authorization) bool {
			_, ok := a.(*banktype.SendAuthorization)
			return ok
		}
		granterAddr, granteeAddr, gAuthz := findGrant(ctx, k, reporter, bankSendOnlyFilter)
		granter := testData.GetAccountbyAccAddr(reporter, granterAddr)
		grantee := testData.GetAccountbyAccAddr(reporter, granteeAddr)
		if reporter.IsAborted() {
			return nil, nil
		}
		amount := granter.LiquidBalance().RandSubsetCoins(reporter, common.WithSendEnabledCoins())
		amount = amount.Min(gAuthz.(*banktype.SendAuthorization).SpendLimit)

		payloadMsg := []sdk.Msg{banktype.NewMsgSend(granter.AddressBech32, grantee.AddressBech32, amount)}
		msgExec := authz.NewMsgExec(grantee.AddressBech32, payloadMsg)
		return []common.SimAccount{grantee}, &msgExec
	}
}

func MsgRevokeFactory(k keeper.Keeper) module.SimMsgFactoryFn[*authz.MsgRevoke] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *authz.MsgRevoke) {
		granterAddr, granteeAddr, auth := findGrant(ctx, k, reporter)
		granter := testData.GetAccountbyAccAddr(reporter, granterAddr)
		grantee := testData.GetAccountbyAccAddr(reporter, granteeAddr)
		if reporter.IsAborted() {
			return nil, nil
		}
		msgExec := authz.NewMsgRevoke(granter.AddressBech32, grantee.AddressBech32, auth.MsgTypeURL())
		return []common.SimAccount{granter}, &msgExec
	}
}

func findGrant(
	ctx context.Context,
	k keeper.Keeper,
	reporter common.SimulationReporter,
	acceptFilter ...func(a authz.Authorization) bool,
) (granterAddr, granteeAddr sdk.AccAddress, auth authz.Authorization) {
	err := k.IterateGrants(ctx, func(granter, grantee sdk.AccAddress, grant authz.Grant) (bool, error) {
		a, err2 := grant.GetAuthorization()
		if err2 != nil {
			return true, err2
		}
		for _, filter := range acceptFilter {
			if !filter(a) {
				return false, nil
			}
		}
		granterAddr, granteeAddr, auth = granter, grantee, a
		return true, nil
	})
	if err != nil {
		reporter.Skip(err.Error())
		return
	}
	if auth == nil {
		reporter.Skip("no grant found")
	}
	return
}
