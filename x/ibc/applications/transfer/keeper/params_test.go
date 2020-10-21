package keeper_test

import "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"

func (suite *KeeperTestSuite) TestParams() {
	expParams := types.DefaultParams()

	params := suite.chainA.App.TransferKeeper.GetParams(suite.chainA.GetContext())
	suite.Require().Equal(expParams, params)

	expParams.SendEnabled = false
	suite.chainA.App.TransferKeeper.SetParams(suite.chainA.GetContext(), expParams)
	params = suite.chainA.App.TransferKeeper.GetParams(suite.chainA.GetContext())
	suite.Require().Equal(expParams, params)
}
