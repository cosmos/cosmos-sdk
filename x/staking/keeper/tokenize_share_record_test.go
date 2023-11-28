package keeper_test

func (suite *KeeperTestSuite) TestGetLastTokenizeShareRecordId() {
	ctx, keeper := suite.ctx, suite.stakingKeeper
	lastTokenizeShareRecordID := keeper.GetLastTokenizeShareRecordID(ctx)
	suite.Equal(lastTokenizeShareRecordID, uint64(0))
	keeper.SetLastTokenizeShareRecordID(ctx, 100)
	lastTokenizeShareRecordID = keeper.GetLastTokenizeShareRecordID(ctx)
	suite.Equal(lastTokenizeShareRecordID, uint64(100))
}

func (suite *KeeperTestSuite) TestGetTokenizeShareRecord() {
	// TODO add LSM test
}
