package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/golang/mock/gomock"
)

func (suite *KeeperTestSuite) TestInitGenesis() {
	ctx := suite.ctx

	// vesting accouts slice should be empty before we initialize the state.
	var vestingAccounts []exported.VestingAccount
	suite.vestingKeeper.IterateVestingAccounts(ctx, func(account exported.VestingAccount) (stop bool) {
		vestingAccounts = append(vestingAccounts, account)
		return false
	})
	suite.Require().Empty(vestingAccounts)

	// initialize accountkeeper.
	for _, addr := range []string{"Alice", "Bob", "Carol"} {
		bacc, origCoins := initBaseAccount(sdk.AccAddress(addr))
		cva := types.NewContinuousVestingAccount(bacc, origCoins, suite.now.Unix(), suite.endTime.Unix())

		suite.accountKeeper.EXPECT().SetAccount(ctx, cva)
		suite.accountKeeper.SetAccount(ctx, cva)
	}

	// initialize vestingkeeper.
	suite.accountKeeper.EXPECT().IterateAccounts(ctx, gomock.Any()).Do(func(ctx sdk.Context, _ func(authtypes.AccountI) bool) {
		for _, addr := range []string{"Alice", "Bob", "Carol"} {
			suite.vestingKeeper.AddVestingAccount(ctx, sdk.AccAddress(addr))
		}
	})
	suite.vestingKeeper.InitGenesis(ctx)
	suite.vestingKeeper.IterateVestingAccounts(ctx, func(account exported.VestingAccount) (stop bool) {
		vestingAccounts = append(vestingAccounts, account)
		return false
	})
	suite.Require().Equal(3, len(vestingAccounts))
}
