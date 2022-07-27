package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestSubmitProposal_InitialDeposit(t *testing.T) {
	const meetsDepositValue = baseDepositTestAmount * baseDepositTestPercent / 100
	var baseDepositRatioDec = sdk.NewDec(baseDepositTestPercent).Quo(sdk.NewDec(100))

	testcases := map[string]struct {
		minDeposit             sdk.Coins
		minInitialDepositRatio sdk.Dec
		initialDeposit         sdk.Coins
		accountBalance         sdk.Coins

		expectError bool
	}{
		"meets initial deposit, enough balance - success": {
			minDeposit:             sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseDepositTestAmount))),
			minInitialDepositRatio: baseDepositRatioDec,
			initialDeposit:         sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(meetsDepositValue))),
			accountBalance:         sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(meetsDepositValue))),
		},
		"does not meet initial deposit, enough balance - error": {
			minDeposit:             sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseDepositTestAmount))),
			minInitialDepositRatio: baseDepositRatioDec,
			initialDeposit:         sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(meetsDepositValue-1))),
			accountBalance:         sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(meetsDepositValue))),

			expectError: true,
		},
		"meets initial deposit, not enough balance - error": {
			minDeposit:             sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseDepositTestAmount))),
			minInitialDepositRatio: baseDepositRatioDec,
			initialDeposit:         sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(meetsDepositValue))),
			accountBalance:         sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(meetsDepositValue-1))),

			expectError: true,
		},
		"does not meet initial deposit and not enough balance - error": {
			minDeposit:             sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseDepositTestAmount))),
			minInitialDepositRatio: baseDepositRatioDec,
			initialDeposit:         sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(meetsDepositValue-1))),
			accountBalance:         sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(meetsDepositValue-1))),

			expectError: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			// Setup
			app := simapp.Setup(false)
			ctx := app.BaseApp.NewContext(false, tmproto.Header{})
			govKeeper := app.GovKeeper
			msgServer := keeper.NewMsgServerImpl(govKeeper)

			params := types.DefaultDepositParams()
			params.MinDeposit = tc.minDeposit
			params.MinInitialDepositRatio = tc.minInitialDepositRatio
			govKeeper.SetDepositParams(ctx, params)

			address := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(0))[0]
			simapp.FundAccount(app.BankKeeper, ctx, address, tc.accountBalance)

			msg, err := types.NewMsgSubmitProposal(TestProposal, tc.initialDeposit, address)
			require.NoError(t, err)

			// System under test
			_, err = msgServer.SubmitProposal(sdk.WrapSDKContext(ctx), msg)

			// Assertions
			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
