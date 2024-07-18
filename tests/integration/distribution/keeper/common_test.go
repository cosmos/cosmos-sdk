package keeper_test

import (
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	PKS = simtestutil.CreateTestPubKeys(3)

	valConsPk0 = PKS[0]

	// The default power validators are initialized to have within tests
	initAmt   = sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	initCoins = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))
)
