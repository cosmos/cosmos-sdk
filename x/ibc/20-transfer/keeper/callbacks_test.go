package keeper_test

// TODO: move to 04-channel after CheckOpen implementation in the following PR
/*
func (suite *KeeperTestSuite) TestOnChanOpenInit() {
	invalidOrder := channel.ORDERED

	counterparty := channel.NewCounterparty(testPort2, testChannel2)
	err := suite.app.IBCKeeper.TransferKeeper.OnChanOpenInit(suite.ctx, invalidOrder, []string{testConnection}, testPort1, testChannel1, counterparty, "")
	suite.Error(err) // invalid channel order

	err = suite.app.IBCKeeper.TransferKeeper.OnChanOpenInit(suite.ctx, testChannelOrder, []string{testConnection}, testPort1, testChannel1, counterparty, "")
	suite.Error(err) // invalid counterparty port ID

	counterparty = channel.NewCounterparty(testPort1, testChannel2)
	err = suite.app.IBCKeeper.TransferKeeper.OnChanOpenInit(suite.ctx, testChannelOrder, []string{testConnection}, testPort1, testChannel1, counterparty, testChannelVersion)
	suite.Error(err) // invalid version

	err = suite.app.IBCKeeper.TransferKeeper.OnChanOpenInit(suite.ctx, testChannelOrder, []string{testConnection}, testPort1, testChannel1, counterparty, "")
	suite.NoError(err) // successfully executed
}

func (suite *KeeperTestSuite) TestOnChanOpenTry() {
	invalidOrder := channel.ORDERED

	counterparty := channel.NewCounterparty(testPort2, testChannel2)
	err := suite.app.IBCKeeper.TransferKeeper.OnChanOpenTry(suite.ctx, invalidOrder, []string{testConnection}, testPort1, testChannel1, counterparty, "", "")
	suite.Error(err) // invalid channel order

	err = suite.app.IBCKeeper.TransferKeeper.OnChanOpenTry(suite.ctx, testChannelOrder, []string{testConnection}, testPort1, testChannel1, counterparty, "", "")
	suite.Error(err) // invalid counterparty port ID

	counterparty = channel.NewCounterparty(testPort1, testChannel2)
	err = suite.app.IBCKeeper.TransferKeeper.OnChanOpenTry(suite.ctx, testChannelOrder, []string{testConnection}, testPort1, testChannel1, counterparty, testChannelVersion, "")
	suite.Error(err) // invalid version

	err = suite.app.IBCKeeper.TransferKeeper.OnChanOpenTry(suite.ctx, testChannelOrder, []string{testConnection}, testPort1, testChannel1, counterparty, "", testChannelVersion)
	suite.Error(err) // invalid counterparty version

	err = suite.app.IBCKeeper.TransferKeeper.OnChanOpenTry(suite.ctx, testChannelOrder, []string{testConnection}, testPort1, testChannel1, counterparty, "", "")
	suite.NoError(err) // successfully executed
}

func (suite *KeeperTestSuite) TestOnChanOpenAck() {
	err := suite.app.IBCKeeper.TransferKeeper.OnChanOpenAck(suite.ctx, testPort1, testChannel1, testChannelVersion)
	suite.Error(err) // invalid version

	err = suite.app.IBCKeeper.TransferKeeper.OnChanOpenAck(suite.ctx, testPort1, testChannel1, "")
	suite.NoError(err) // successfully executed
}
*/
