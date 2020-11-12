package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
)

func (suite *KeeperTestSuite) TestParams() {
	expParams := types.DefaultParams()

	params := suite.chainA.App.IBCKeeper.ClientKeeper.GetParams(suite.chainA.GetContext())
	suite.Require().Equal(expParams, params)

	expParams.AllowedClients = []string{}
	suite.chainA.App.IBCKeeper.ClientKeeper.SetParams(suite.chainA.GetContext(), expParams)
	params = suite.chainA.App.IBCKeeper.ClientKeeper.GetParams(suite.chainA.GetContext())
	suite.Require().Empty(expParams.AllowedClients)
}
