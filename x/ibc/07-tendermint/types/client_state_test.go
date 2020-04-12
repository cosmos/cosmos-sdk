package types_test

import (
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
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
		clientState    ibctmtypes.ClientState
		consensusState ibctmtypes.ConsensusState
		prefix         commitmenttypes.MerklePrefix
		proof          commitmenttypes.MerkleProof
		expPass        bool
	}{
		// FIXME: uncomment
		// {
		// 	name:        "successful verification",
		// 	clientState: ibctmtypes.NewClientState(chainID, chainID, height),
		// 	consensusState: ibctmtypes.ConsensusState{
		// 		Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.MerklePrefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: ibctmtypes.ClientState{ID: chainID, LastHeader: suite.header, FrozenHeight: height - 1},
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			consensusState: ibctmtypes.ConsensusState{
				Root:         commitmenttypes.NewMerkleRoot(suite.header.AppHash),
				ValidatorSet: suite.valSet,
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:   commitmenttypes.MerkleProof{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyClientConsensusState(
			suite.cdc, tc.consensusState.Root, height, "chainA", tc.consensusState.GetHeight(), tc.prefix, tc.proof, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyConnectionState() {
	counterparty := connection.NewCounterparty("clientB", testConnectionID, commitmenttypes.NewMerklePrefix([]byte("ibc")))
	conn := connection.NewConnectionEnd(connectionexported.OPEN, "clientA", counterparty, []string{"1.0.0"})

	testCases := []struct {
		name           string
		clientState    ibctmtypes.ClientState
		connection     connection.ConnectionEnd
		consensusState ibctmtypes.ConsensusState
		prefix         commitmenttypes.MerklePrefix
		proof          commitmenttypes.MerkleProof
		expPass        bool
	}{
		// FIXME: uncomment
		// {
		// 	name:         "successful verification",
		// 	clientState:  ibctmtypes.NewClientState(chainID, chainID, height),
		// 	connection:   conn,
		// 	consensusState: ibctmtypes.ConsensusState{
		// 		Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			connection:  conn,
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.MerklePrefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			connection:  conn,
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: ibctmtypes.ClientState{ID: chainID, LastHeader: suite.header, FrozenHeight: height - 1},
			connection:  conn,
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			connection:  conn,
			consensusState: ibctmtypes.ConsensusState{
				Root:         commitmenttypes.NewMerkleRoot(suite.header.AppHash),
				ValidatorSet: suite.valSet,
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:   commitmenttypes.MerkleProof{},
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
		clientState    ibctmtypes.ClientState
		channel        channel.Channel
		consensusState ibctmtypes.ConsensusState
		prefix         commitmenttypes.MerklePrefix
		proof          commitmenttypes.MerkleProof
		expPass        bool
	}{
		// FIXME: uncomment
		// {
		// 	name:         "successful verification",
		// 	clientState:  ibctmtypes.NewClientState(chainID, height),
		// 	connection:   conn,
		// 	consensusState: ibctmtypes.ConsensusState{
		// 		Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			channel:     ch,
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.MerklePrefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			channel:     ch,
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: ibctmtypes.ClientState{ID: chainID, LastHeader: suite.header, FrozenHeight: height - 1},
			channel:     ch,
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			channel:     ch,
			consensusState: ibctmtypes.ConsensusState{
				Root:         commitmenttypes.NewMerkleRoot(suite.header.AppHash),
				ValidatorSet: suite.valSet,
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:   commitmenttypes.MerkleProof{},
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
		clientState    ibctmtypes.ClientState
		commitment     []byte
		consensusState ibctmtypes.ConsensusState
		prefix         commitmenttypes.MerklePrefix
		proof          commitmenttypes.MerkleProof
		expPass        bool
	}{
		// FIXME: uncomment
		// {
		// 	name:         "successful verification",
		// 	clientState:  ibctmtypes.NewClientState(chainID, height),
		// 	connection:   conn,
		// 	consensusState: ibctmtypes.ConsensusState{
		// 		Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			commitment:  []byte{},
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.MerklePrefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			commitment:  []byte{},
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: ibctmtypes.ClientState{ID: chainID, LastHeader: suite.header, FrozenHeight: height - 1},
			commitment:  []byte{},
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			commitment:  []byte{},
			consensusState: ibctmtypes.ConsensusState{
				Root:         commitmenttypes.NewMerkleRoot(suite.header.AppHash),
				ValidatorSet: suite.valSet,
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:   commitmenttypes.MerkleProof{},
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
		clientState    ibctmtypes.ClientState
		ack            []byte
		consensusState ibctmtypes.ConsensusState
		prefix         commitmenttypes.MerklePrefix
		proof          commitmenttypes.MerkleProof
		expPass        bool
	}{
		// FIXME: uncomment
		// {
		// 	name:         "successful verification",
		// 	clientState:  ibctmtypes.NewClientState(chainID, chainID, height),
		// 	connection:   conn,
		// 	consensusState: ibctmtypes.ConsensusState{
		// 		Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			ack:         []byte{},
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.MerklePrefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			ack:         []byte{},
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: ibctmtypes.ClientState{ID: chainID, LastHeader: suite.header, FrozenHeight: height - 1},
			ack:         []byte{},
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			ack:         []byte{},
			consensusState: ibctmtypes.ConsensusState{
				Root:         commitmenttypes.NewMerkleRoot(suite.header.AppHash),
				ValidatorSet: suite.valSet,
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:   commitmenttypes.MerkleProof{},
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
		clientState    ibctmtypes.ClientState
		consensusState ibctmtypes.ConsensusState
		prefix         commitmenttypes.MerklePrefix
		proof          commitmenttypes.MerkleProof
		expPass        bool
	}{
		// FIXME: uncomment
		// {
		// 	name:         "successful verification",
		// 	clientState:  ibctmtypes.NewClientState(chainID, chainID, height),
		// 	connection:   conn,
		// 	consensusState: ibctmtypes.ConsensusState{
		// 		Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.MerklePrefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: ibctmtypes.ClientState{ID: chainID, LastHeader: suite.header, FrozenHeight: height - 1},
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			consensusState: ibctmtypes.ConsensusState{
				Root:         commitmenttypes.NewMerkleRoot(suite.header.AppHash),
				ValidatorSet: suite.valSet,
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:   commitmenttypes.MerkleProof{},
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
		clientState    ibctmtypes.ClientState
		consensusState ibctmtypes.ConsensusState
		prefix         commitmenttypes.MerklePrefix
		proof          commitmenttypes.MerkleProof
		expPass        bool
	}{
		// FIXME: uncomment
		// {
		// 	name:         "successful verification",
		// 	clientState:  ibctmtypes.NewClientState(chainID, chainID, height),
		// 	connection:   conn,
		// 	consensusState: ibctmtypes.ConsensusState{
		// 		Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.MerklePrefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: ibctmtypes.ClientState{ID: chainID, LastHeader: suite.header, FrozenHeight: height - 1},
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header),
			consensusState: ibctmtypes.ConsensusState{
				Root:         commitmenttypes.NewMerkleRoot(suite.header.AppHash),
				ValidatorSet: suite.valSet,
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:   commitmenttypes.MerkleProof{},
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
