package tendermint_test

import (
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func (suite *TendermintTestSuite) TestVerifyClientConsensusState() {
	testCases := []struct {
		name           string
		clientState    tendermint.ClientState
		consensusState tendermint.ConsensusState
		height         uint64
		prefix         commitment.Prefix
		proof          commitment.Proof
		expPass        bool
	}{
		// {
		// 	name:        "successful verification",
		// 	clientState: tendermint.NewClientState(chainID, height),
		// 	consensusState: tendermint.ConsensusState{
		// 		Root: commitment.NewRoot(suite.header.AppHash),
		// 	},
		// 	height:  height,
		// 	prefix:  commitment.NewPrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: tendermint.NewClientState(chainID, height),
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.Prefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: tendermint.NewClientState(chainID, height-1),
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: height - 1},
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: tendermint.NewClientState(chainID, height),
			consensusState: tendermint.ConsensusState{
				Root:             commitment.NewRoot(suite.header.AppHash),
				ValidatorSetHash: suite.valSet.Hash(),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			proof:   commitment.Proof{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyClientConsensusState(
			suite.cdc, tc.height, tc.prefix, tc.proof, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyConnectionState() {
	testConnectionID := "connectionid"
	counterparty := connection.NewCounterparty("clientB", testConnectionID, commitment.NewPrefix([]byte("ibc")))
	conn := connection.NewConnectionEnd(connectionexported.OPEN, "clientA", counterparty, []string{"1.0.0"})

	testCases := []struct {
		name           string
		clientState    tendermint.ClientState
		connectionID   string
		connection     connection.ConnectionEnd
		consensusState tendermint.ConsensusState
		height         uint64
		prefix         commitment.Prefix
		proof          commitment.Proof
		expPass        bool
	}{
		// {
		// 	name:         "successful verification",
		// 	clientState:  tendermint.NewClientState(chainID, height),
		// 	connectionID: testConnectionID,
		// 	connection:   conn,
		// 	consensusState: tendermint.ConsensusState{
		// 		Root: commitment.NewRoot(suite.header.AppHash),
		// 	},
		// 	height:  height,
		// 	prefix:  commitment.NewPrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:         "ApplyPrefix failed",
			clientState:  tendermint.NewClientState(chainID, height),
			connectionID: testConnectionID,
			connection:   conn,
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.Prefix{},
			expPass: false,
		},
		{
			name:         "latest client height < height",
			clientState:  tendermint.NewClientState(chainID, height-1),
			connectionID: testConnectionID,
			connection:   conn,
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:         "client is frozen",
			clientState:  tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: height - 1},
			connectionID: testConnectionID,
			connection:   conn,
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:         "proof verification failed",
			clientState:  tendermint.NewClientState(chainID, height),
			connectionID: testConnectionID,
			connection:   conn,
			consensusState: tendermint.ConsensusState{
				Root:             commitment.NewRoot(suite.header.AppHash),
				ValidatorSetHash: suite.valSet.Hash(),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			proof:   commitment.Proof{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyConnectionState(
			suite.cdc, tc.height, tc.prefix, tc.proof, tc.connectionID, tc.connection, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyChannelState() {
	testConnectionID := "connectionid"
	testPortID := "testportid"
	testChannelID := "testchannelid"
	counterparty := channel.NewCounterparty(testPortID, testChannelID)
	ch := channel.NewChannel(channelexported.OPEN, channelexported.ORDERED, counterparty, []string{testConnectionID}, "1.0.0")

	testCases := []struct {
		name           string
		clientState    tendermint.ClientState
		portID         string
		channelID      string
		channel        channel.Channel
		consensusState tendermint.ConsensusState
		height         uint64
		prefix         commitment.Prefix
		proof          commitment.Proof
		expPass        bool
	}{
		// {
		// 	name:         "successful verification",
		// 	clientState:  tendermint.NewClientState(chainID, height),
		// 	connectionID: testConnectionID,
		// 	connection:   conn,
		// 	consensusState: tendermint.ConsensusState{
		// 		Root: commitment.NewRoot(suite.header.AppHash),
		// 	},
		// 	height:  height,
		// 	prefix:  commitment.NewPrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: tendermint.NewClientState(chainID, height),
			portID:      testPortID,
			channelID:   testChannelID,
			channel:     ch,
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.Prefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: tendermint.NewClientState(chainID, height-1),
			portID:      testPortID,
			channelID:   testChannelID,
			channel:     ch,
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: height - 1},
			portID:      testPortID,
			channelID:   testChannelID,
			channel:     ch,
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: tendermint.NewClientState(chainID, height),
			portID:      testPortID,
			channelID:   testChannelID,
			channel:     ch,
			consensusState: tendermint.ConsensusState{
				Root:             commitment.NewRoot(suite.header.AppHash),
				ValidatorSetHash: suite.valSet.Hash(),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			proof:   commitment.Proof{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyChannelState(
			suite.cdc, tc.height, tc.prefix, tc.proof, tc.portID, tc.channelID, tc.channel, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyPacketCommitment() {
	testPortID := "testportid"
	testChannelID := "testchannelid"
	testSequence := uint64(1)

	testCases := []struct {
		name           string
		clientState    tendermint.ClientState
		portID         string
		channelID      string
		seq            uint64
		commitment     []byte
		consensusState tendermint.ConsensusState
		height         uint64
		prefix         commitment.Prefix
		proof          commitment.Proof
		expPass        bool
	}{
		// {
		// 	name:         "successful verification",
		// 	clientState:  tendermint.NewClientState(chainID, height),
		// 	connectionID: testConnectionID,
		// 	connection:   conn,
		// 	consensusState: tendermint.ConsensusState{
		// 		Root: commitment.NewRoot(suite.header.AppHash),
		// 	},
		// 	height:  height,
		// 	prefix:  commitment.NewPrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: tendermint.NewClientState(chainID, height),
			portID:      testPortID,
			channelID:   testChannelID,
			seq:         testSequence,
			commitment:  []byte{},
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.Prefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: tendermint.NewClientState(chainID, height-1),
			portID:      testPortID,
			channelID:   testChannelID,
			seq:         testSequence,
			commitment:  []byte{},
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: height - 1},
			portID:      testPortID,
			channelID:   testChannelID,
			seq:         testSequence,
			commitment:  []byte{},
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: tendermint.NewClientState(chainID, height),
			portID:      testPortID,
			channelID:   testChannelID,
			seq:         testSequence,
			commitment:  []byte{},
			consensusState: tendermint.ConsensusState{
				Root:             commitment.NewRoot(suite.header.AppHash),
				ValidatorSetHash: suite.valSet.Hash(),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			proof:   commitment.Proof{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyPacketCommitment(
			tc.height, tc.prefix, tc.proof, tc.portID, tc.channelID, tc.seq, tc.commitment, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyPacketAcknowledgement() {
	testPortID := "testportid"
	testChannelID := "testchannelid"
	testSequence := uint64(1)

	testCases := []struct {
		name           string
		clientState    tendermint.ClientState
		portID         string
		channelID      string
		seq            uint64
		ack            []byte
		consensusState tendermint.ConsensusState
		height         uint64
		prefix         commitment.Prefix
		proof          commitment.Proof
		expPass        bool
	}{
		// {
		// 	name:         "successful verification",
		// 	clientState:  tendermint.NewClientState(chainID, height),
		// 	connectionID: testConnectionID,
		// 	connection:   conn,
		// 	consensusState: tendermint.ConsensusState{
		// 		Root: commitment.NewRoot(suite.header.AppHash),
		// 	},
		// 	height:  height,
		// 	prefix:  commitment.NewPrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: tendermint.NewClientState(chainID, height),
			portID:      testPortID,
			channelID:   testChannelID,
			seq:         testSequence,
			ack:         []byte{},
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.Prefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: tendermint.NewClientState(chainID, height-1),
			portID:      testPortID,
			channelID:   testChannelID,
			seq:         testSequence,
			ack:         []byte{},
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: height - 1},
			portID:      testPortID,
			channelID:   testChannelID,
			seq:         testSequence,
			ack:         []byte{},
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: tendermint.NewClientState(chainID, height),
			portID:      testPortID,
			channelID:   testChannelID,
			seq:         testSequence,
			ack:         []byte{},
			consensusState: tendermint.ConsensusState{
				Root:             commitment.NewRoot(suite.header.AppHash),
				ValidatorSetHash: suite.valSet.Hash(),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			proof:   commitment.Proof{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyPacketAcknowledgement(
			tc.height, tc.prefix, tc.proof, tc.portID, tc.channelID, tc.seq, tc.ack, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyPacketAcknowledgementAbsence() {
	testPortID := "testportid"
	testChannelID := "testchannelid"
	testSequence := uint64(1)

	testCases := []struct {
		name           string
		clientState    tendermint.ClientState
		portID         string
		channelID      string
		seq            uint64
		consensusState tendermint.ConsensusState
		height         uint64
		prefix         commitment.Prefix
		proof          commitment.Proof
		expPass        bool
	}{
		// {
		// 	name:         "successful verification",
		// 	clientState:  tendermint.NewClientState(chainID, height),
		// 	connectionID: testConnectionID,
		// 	connection:   conn,
		// 	consensusState: tendermint.ConsensusState{
		// 		Root: commitment.NewRoot(suite.header.AppHash),
		// 	},
		// 	height:  height,
		// 	prefix:  commitment.NewPrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: tendermint.NewClientState(chainID, height),
			portID:      testPortID,
			channelID:   testChannelID,
			seq:         testSequence,
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.Prefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: tendermint.NewClientState(chainID, height-1),
			portID:      testPortID,
			channelID:   testChannelID,
			seq:         testSequence,
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: height - 1},
			portID:      testPortID,
			channelID:   testChannelID,
			seq:         testSequence,
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: tendermint.NewClientState(chainID, height),
			portID:      testPortID,
			channelID:   testChannelID,
			seq:         testSequence,
			consensusState: tendermint.ConsensusState{
				Root:             commitment.NewRoot(suite.header.AppHash),
				ValidatorSetHash: suite.valSet.Hash(),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			proof:   commitment.Proof{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyPacketAcknowledgementAbsence(
			tc.height, tc.prefix, tc.proof, tc.portID, tc.channelID, tc.seq, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyNextSeqRecv() {
	testPortID := "testportid"
	testChannelID := "testchannelid"
	testSequence := uint64(1)

	testCases := []struct {
		name             string
		clientState      tendermint.ClientState
		portID           string
		channelID        string
		nextSequenceRecv uint64
		consensusState   tendermint.ConsensusState
		height           uint64
		prefix           commitment.Prefix
		proof            commitment.Proof
		expPass          bool
	}{
		// {
		// 	name:         "successful verification",
		// 	clientState:  tendermint.NewClientState(chainID, height),
		// 	connectionID: testConnectionID,
		// 	connection:   conn,
		// 	consensusState: tendermint.ConsensusState{
		// 		Root: commitment.NewRoot(suite.header.AppHash),
		// 	},
		// 	height:  height,
		// 	prefix:  commitment.NewPrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:             "ApplyPrefix failed",
			clientState:      tendermint.NewClientState(chainID, height),
			portID:           testPortID,
			channelID:        testChannelID,
			nextSequenceRecv: testSequence,
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.Prefix{},
			expPass: false,
		},
		{
			name:             "latest client height < height",
			clientState:      tendermint.NewClientState(chainID, height-1),
			portID:           testPortID,
			channelID:        testChannelID,
			nextSequenceRecv: testSequence,
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:             "client is frozen",
			clientState:      tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: height - 1},
			portID:           testPortID,
			channelID:        testChannelID,
			nextSequenceRecv: testSequence,
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:             "proof verification failed",
			clientState:      tendermint.NewClientState(chainID, height),
			portID:           testPortID,
			channelID:        testChannelID,
			nextSequenceRecv: testSequence,
			consensusState: tendermint.ConsensusState{
				Root:             commitment.NewRoot(suite.header.AppHash),
				ValidatorSetHash: suite.valSet.Hash(),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			proof:   commitment.Proof{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyNextSequenceRecv(
			tc.height, tc.prefix, tc.proof, tc.portID, tc.channelID, tc.nextSequenceRecv, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
