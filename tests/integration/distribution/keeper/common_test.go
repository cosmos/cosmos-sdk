package keeper_test

import (
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

var (
	PKS = simtestutil.CreateTestPubKeys(5)

	valConsPk0 = PKS[0]
	valConsPk1 = PKS[1]
	valConsPk2 = PKS[2]

	valConsAddr0 = sdk.ConsAddress(valConsPk0.Address())

	distrAcc = authtypes.NewEmptyModuleAccount(types.ModuleName)
)
