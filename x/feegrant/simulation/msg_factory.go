package simulation

import (
	"context"
	"github.com/cosmos/cosmos-sdk/simsx/common"
	"github.com/cosmos/cosmos-sdk/simsx/module"

	"cosmossdk.io/x/feegrant"
	"cosmossdk.io/x/feegrant/keeper"
)

func MsgGrantAllowanceFactory(k keeper.Keeper) module.SimMsgFactoryFn[*feegrant.MsgGrantAllowance] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *feegrant.MsgGrantAllowance) {
		granter := testData.AnyAccount(reporter, common.WithSpendableBalance())
		grantee := testData.AnyAccount(reporter, common.ExcludeAccounts(granter))
		if reporter.IsAborted() {
			return nil, nil
		}
		if f, _ := k.GetAllowance(ctx, granter.Address, grantee.Address); f != nil {
			reporter.Skip("fee allowance exists")
			return nil, nil
		}

		coins := granter.LiquidBalance().RandSubsetCoins(reporter, common.WithSendEnabledCoins())
		oneYear := common.BlockTime(ctx).AddDate(1, 0, 0)
		msg, err := feegrant.NewMsgGrantAllowance(
			&feegrant.BasicAllowance{SpendLimit: coins, Expiration: &oneYear},
			granter.AddressBech32,
			grantee.AddressBech32,
		)
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		return []common.SimAccount{granter}, msg
	}
}

func MsgRevokeAllowanceFactory(k keeper.Keeper) module.SimMsgFactoryFn[*feegrant.MsgRevokeAllowance] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *feegrant.MsgRevokeAllowance) {
		var gotGrant *feegrant.Grant
		if err := k.IterateAllFeeAllowances(ctx, func(grant feegrant.Grant) bool {
			gotGrant = &grant
			return true
		}); err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		if gotGrant == nil {
			reporter.Skip("no grant found")
			return nil, nil
		}
		granter := testData.GetAccount(reporter, gotGrant.Granter)
		grantee := testData.GetAccount(reporter, gotGrant.Grantee)
		msg := feegrant.NewMsgRevokeAllowance(granter.AddressBech32, grantee.AddressBech32)
		return []common.SimAccount{granter}, &msg
	}
}
