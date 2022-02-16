package keeper_test

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
	// The default power validators are initialized to have within tests
	InitTokens = sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
)
