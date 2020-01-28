package keeper_test

func (suite *KeeperTestSuite) TestSendPacket() {
	// Packet passes/fails validate basic packet
	// Channel found/not found
	// Channel closed/not CLOSED
	// if packet.GetDestPort() != channel.Counterparty.PortID {}
	// if packet.GetDestChannel() != channel.Counterparty.ChannelID {}
	// Connection found/not found
	// Connection initiated/uninitialized
	// Client state found/not found
	// if clientState.GetLatestHeight() >= packet.GetTimeoutHeight() {}
	// Next sequence found/not found
	// if packet.GetSequence() != nextSequenceSend {}
	// Success
}

func (suite *KeeperTestSuite) TestRecvPacket() {
	// Channel found/not found
	// Channel closed/not CLOSED
	// if packet.GetSourcePort() != channel.Counterparty.PortID {}
	// if packet.GetSourceChannel() != channel.Counterparty.ChannelID {}
	// Connection found/not found
	// Connection initiated/uninitialized
	// if uint64(ctx.BlockHeight()) >= packet.GetTimeoutHeight() {}
	// Client state found/not found
	// Success/fail on verify packet commitment
}

func (suite *KeeperTestSuite) TestPacketExecuted() {
	// Channel found/not found
	// Channel closed/not CLOSED
	// if acknowledgement != nil || channel.Ordering == exported.UNORDERED {}
	// if channel.Ordering == exported.ORDERED {
	// Ensure next sequence recieve is found
	// if packet.GetSequence() != nextSequenceRecv {}
	// }
	// Success, packet recieved and acknowledged
}

func (suite *KeeperTestSuite) TestAcknowledgePacket() {
	// Channel found/not found
	// Channel closed/not CLOSED
	// if packet.GetSourcePort() != channel.Counterparty.PortID {}
	// if packet.GetSourceChannel() != channel.Counterparty.ChannelID {}
	// Connection found/not found
	// Connection initiated/uninitialized
	// if !bytes.Equal(commitment, types.CommitPacket(packet.GetData())) {}
	// Client state found/not found
	// Success/fail on verify packet commitment
}

func (suite *KeeperTestSuite) TestAcknowledgementExecuted() {
	// Delete non existent packet commitment
	// Create packet commitment
	// Delete that packet commitment
	// Ensure packet commitment deleted
}

func (suite *KeeperTestSuite) TestCleanupPacket() {
	// Channel found/not found
	// Channel closed/not CLOSED
	// if packet.GetSourcePort() != channel.Counterparty.PortID {}
	// if packet.GetSourceChannel() != channel.Counterparty.ChannelID {}
	// Connection found/not found
	// Connection initiated/uninitialized
	// if nextSequenceRecv <= packet.GetSequence() {}
	// if !bytes.Equal(commitment, types.CommitPacket(packet.GetData())) {}
	// Client state found/not found
	// Success/fail on verify ORDERED packet commitment
	// Success/fail on verify UNORDERED packet commitment
	// Invalid ordering packet failure
}
