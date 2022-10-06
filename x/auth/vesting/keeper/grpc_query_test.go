package keeper_test

import (
	gocontext "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

func (suite *KeeperTestSuite) TestQueryVestingAccounts() {
	ctx, queryClient := suite.ctx, suite.queryClient

	res, err := queryClient.VestingAccounts(gocontext.Background(), &types.QueryVestingAccountsRequest{})
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Empty(res.Accounts)

	res, err = queryClient.VestingAccounts(gocontext.Background(),
		&types.QueryVestingAccountsRequest{Pagination: &query.PageRequest{Limit: 1}})
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Empty(res.Accounts)

	for _, addr := range []string{"Alice", "Bob", "Carol"} {
		bacc, origCoins := initBaseAccount(sdk.AccAddress(addr))
		cva := types.NewContinuousVestingAccount(bacc, origCoins, suite.now.Unix(), suite.endTime.Unix())

		suite.accountKeeper.EXPECT().SetAccount(ctx, cva)
		suite.accountKeeper.SetAccount(ctx, cva)
		suite.vestingKeeper.AddVestingAccount(ctx, cva.GetAddress())
	}

	res, err = queryClient.VestingAccounts(gocontext.Background(),
		&types.QueryVestingAccountsRequest{Pagination: &query.PageRequest{Limit: 2}})
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(2, len(res.Accounts))
}
