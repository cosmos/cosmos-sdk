package keeper_test

func (suite *KeeperTestSuite) TestSendTransfer() {
	err := suite.keeper.SendTransfer(suite.ctx, "", "", nil, nil, nil, true)
	suite.Error(err)
}
