package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/golang/mock/gomock"
)

func (suite *KeeperTestSuite) TestMigrate1to2() {
	ctx := suite.ctx
	for _, addr := range []string{"Alice", "Bob", "Carol"} {
		bacc, origCoins := initBaseAccount(sdk.AccAddress(addr))
		cva := types.NewContinuousVestingAccount(bacc, origCoins, suite.now.Unix(), suite.endTime.Unix())

		suite.accountKeeper.EXPECT().SetAccount(ctx, cva)
		suite.accountKeeper.SetAccount(ctx, cva)
	}

	var vestingAccounts []exported.VestingAccount
	suite.vestingKeeper.IterateVestingAccounts(ctx, func(account exported.VestingAccount) (stop bool) {
		vestingAccounts = append(vestingAccounts, account)
		return false
	})
	suite.Require().Empty(vestingAccounts)

	suite.accountKeeper.EXPECT().IterateAccounts(ctx, gomock.Any()).Do(func(ctx sdk.Context, _ func(authtypes.AccountI) bool) {
		for _, addr := range []string{"Alice", "Bob", "Carol"} {
			suite.vestingKeeper.AddVestingAccount(ctx, sdk.AccAddress(addr))
		}
	})
	migrator := keeper.NewMigrator(suite.vestingKeeper, suite.accountKeeper)
	err := migrator.Migrate1to2(suite.ctx)
	suite.Require().NoError(err)

	suite.vestingKeeper.IterateVestingAccounts(ctx, func(account exported.VestingAccount) (stop bool) {
		vestingAccounts = append(vestingAccounts, account)
		return false
	})
	suite.Require().Equal(3, len(vestingAccounts))
}
