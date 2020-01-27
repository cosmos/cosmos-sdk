package tendermint_test

import (
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

const (
	testConnectionID = "connectionid"
	testPortID       = "testportid"
	testChannelID    = "testchannelid"
	testSequence     = 1
)

func (suite *TendermintTestSuite) TestVerifyClientConsensusState() {
	testCases := []struct {
		name           string
		clientState    tendermint.ClientState
		consensusState tendermint.ConsensusState
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
		// 	prefix:  commitment.NewPrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: tendermint.NewClientState(chainID, height),
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			prefix:  commitment.Prefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: tendermint.NewClientState(chainID, height-1),
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: height - 1},
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
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
			prefix:  commitment.NewPrefix([]byte("ibc")),
			proof:   commitment.Proof{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyClientConsensusState(
			suite.cdc, height, tc.prefix, tc.proof, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyConnectionState() {
	counterparty := connection.NewCounterparty("clientB", testConnectionID, commitment.NewPrefix([]byte("ibc")))
	conn := connection.NewConnectionEnd(connectionexported.OPEN, "clientA", counterparty, []string{"1.0.0"})

	testCases := []struct {
		name           string
		clientState    tendermint.ClientState
		connection     connection.ConnectionEnd
		consensusState tendermint.ConsensusState
		prefix         commitment.Prefix
		proof          commitment.Proof
		expPass        bool
	}{
		// {
		// 	name:         "successful verification",
		// 	clientState:  tendermint.NewClientState(chainID, height),
		// 	connection:   conn,
		// 	consensusState: tendermint.ConsensusState{
		// 		Root: commitment.NewRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitment.NewPrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: tendermint.NewClientState(chainID, height),
			connection:  conn,
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			prefix:  commitment.Prefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: tendermint.NewClientState(chainID, height-1),
			connection:  conn,
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: height - 1},
			connection:  conn,
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: tendermint.NewClientState(chainID, height),
			connection:  conn,
			consensusState: tendermint.ConsensusState{
				Root:             commitment.NewRoot(suite.header.AppHash),
				ValidatorSetHash: suite.valSet.Hash(),
			},
			prefix:  commitment.NewPrefix([]byte("ibc")),
			proof:   commitment.Proof{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyConnectionState(
			suite.cdc, height, tc.prefix, tc.proof, testConnectionID, tc.connection, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyChannelState() {
	counterparty := channel.NewCounterparty(testPortID, testChannelID)
	ch := channel.NewChannel(channelexported.OPEN, channelexported.ORDERED, counterparty, []string{testConnectionID}, "1.0.0")

	testCases := []struct {
		name           string
		clientState    tendermint.ClientState
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
		// 	connection:   conn,
		// 	consensusState: tendermint.ConsensusState{
		// 		Root: commitment.NewRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitment.NewPrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: tendermint.NewClientState(chainID, height),
			channel:     ch,
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			prefix:  commitment.Prefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: tendermint.NewClientState(chainID, height-1),
			channel:     ch,
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: height - 1},
			channel:     ch,
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: tendermint.NewClientState(chainID, height),
			channel:     ch,
			consensusState: tendermint.ConsensusState{
				Root:             commitment.NewRoot(suite.header.AppHash),
				ValidatorSetHash: suite.valSet.Hash(),
			},
			prefix:  commitment.NewPrefix([]byte("ibc")),
			proof:   commitment.Proof{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyChannelState(
			suite.cdc, height, tc.prefix, tc.proof, testPortID, testChannelID, tc.channel, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyPacketCommitment() {
	testCases := []struct {
		name           string
		clientState    tendermint.ClientState
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
		// 	connection:   conn,
		// 	consensusState: tendermint.ConsensusState{
		// 		Root: commitment.NewRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitment.NewPrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: tendermint.NewClientState(chainID, height),
			commitment:  []byte{},
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			prefix:  commitment.Prefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: tendermint.NewClientState(chainID, height-1),
			commitment:  []byte{},
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: height - 1},
			commitment:  []byte{},
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: tendermint.NewClientState(chainID, height),
			commitment:  []byte{},
			consensusState: tendermint.ConsensusState{
				Root:             commitment.NewRoot(suite.header.AppHash),
				ValidatorSetHash: suite.valSet.Hash(),
			},
			prefix:  commitment.NewPrefix([]byte("ibc")),
			proof:   commitment.Proof{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyPacketCommitment(
			height, tc.prefix, tc.proof, testPortID, testChannelID, testSequence, tc.commitment, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyPacketAcknowledgement() {
	testCases := []struct {
		name           string
		clientState    tendermint.ClientState
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
		// 	connection:   conn,
		// 	consensusState: tendermint.ConsensusState{
		// 		Root: commitment.NewRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitment.NewPrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: tendermint.NewClientState(chainID, height),
			ack:         []byte{},
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			prefix:  commitment.Prefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: tendermint.NewClientState(chainID, height-1),
			ack:         []byte{},
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: height - 1},
			ack:         []byte{},
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: tendermint.NewClientState(chainID, height),
			ack:         []byte{},
			consensusState: tendermint.ConsensusState{
				Root:             commitment.NewRoot(suite.header.AppHash),
				ValidatorSetHash: suite.valSet.Hash(),
			},
			prefix:  commitment.NewPrefix([]byte("ibc")),
			proof:   commitment.Proof{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyPacketAcknowledgement(
			height, tc.prefix, tc.proof, testPortID, testChannelID, testSequence, tc.ack, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyPacketAcknowledgementAbsence() {
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
		// 	name:         "successful verification",
		// 	clientState:  tendermint.NewClientState(chainID, height),
		// 	connection:   conn,
		// 	consensusState: tendermint.ConsensusState{
		// 		Root: commitment.NewRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitment.NewPrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: tendermint.NewClientState(chainID, height),
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			prefix:  commitment.Prefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: tendermint.NewClientState(chainID, height-1),
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: height - 1},
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
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
			prefix:  commitment.NewPrefix([]byte("ibc")),
			proof:   commitment.Proof{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyPacketAcknowledgementAbsence(
			height, tc.prefix, tc.proof, testPortID, testChannelID, testSequence, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyNextSeqRecv() {
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
		// 	name:         "successful verification",
		// 	clientState:  tendermint.NewClientState(chainID, height),
		// 	connection:   conn,
		// 	consensusState: tendermint.ConsensusState{
		// 		Root: commitment.NewRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitment.NewPrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: tendermint.NewClientState(chainID, height),
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			prefix:  commitment.Prefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: tendermint.NewClientState(chainID, height-1),
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: height - 1},
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
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
			prefix:  commitment.NewPrefix([]byte("ibc")),
			proof:   commitment.Proof{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyNextSequenceRecv(
			height, tc.prefix, tc.proof, testPortID, testChannelID, testSequence, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
