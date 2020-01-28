package keeper_test

// TODO: Actually write tests

func (suite *KeeperTestSuite) TestTimeoutPacket() {
	// Create packet with appropriate channel and port, test wrong port/channel
	// Test with an unopened channel
	// if packet.GetDestPort() != channel.Counterparty.PortID {}
	// if packet.GetDestChannel() != channel.Counterparty.ChannelID {}
	// Ensure connection assocated with channel exists, test fail/pass
	// if nextSequenceRecv >= packet.GetSequence() {}
	// if !bytes.Equal(commitment, types.CommitPacket(packet.GetData())) {}
	// Ensure that consensus state is found
	// Test ordered channel
	// Test unordered channel

}

func (suite *KeeperTestSuite) TimeoutExecuted() {
	// Create packet with appropriate channel and port, test wrong port/channel
	// Ensure that ordered channel that is created is closed once TimeoutExecuted is called
}

func (suite *KeeperTestSuite) TimeoutOnClose() {
	// Create packet with appropriate channel and port, test wrong port/channel
	// if packet.GetDestPort() != channel.Counterparty.PortID {}
	// if packet.GetDestChannel() != channel.Counterparty.ChannelID {}
	// Ensure connection assocated with channel exists, test fail/pass
	// if !bytes.Equal(commitment, types.CommitPacket(packet.GetData())) {}
	// Ensure that consensus state is found
	// Ensure that opposite channel has closed
	// Test ordered channel
	// Test unordered channel
}
