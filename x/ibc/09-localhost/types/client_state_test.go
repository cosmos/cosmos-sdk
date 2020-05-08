package types_test

import (
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

const (
	testConnectionID = "connectionid"
	testPortID       = "testportid"
	testChannelID    = "testchannelid"
	testSequence     = 1
)

func (suite *LocalhostTestSuite) TestValidate() {
	testCases := []struct {
		name        string
		clientState types.ClientState
		expPass     bool
	}{
		{
			name:        "valid client",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			expPass:     true,
		},
		{
			name:        "invalid chain id",
			clientState: types.NewClientState(suite.store, " ", 10),
			expPass:     false,
		},
		{
			name:        "invalid height",
			clientState: types.NewClientState(suite.store, "chainID", 0),
			expPass:     false,
		},
		{
			name:        "invalid store",
			clientState: types.NewClientState(nil, "chainID", 10),
			expPass:     false,
		},
	}

	for _, tc := range testCases {
		err := tc.clientState.Validate()
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

func (suite *LocalhostTestSuite) TestVerifyClientConsensusState() {
	testCases := []struct {
		name        string
		clientState types.ClientState
		prefix      commitmenttypes.MerklePrefix
		proof       commitmenttypes.MerkleProof
		expPass     bool
	}{
		{
			name:        "ApplyPrefix failed",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			prefix:      commitmenttypes.MerklePrefix{},
			expPass:     false,
		},
		{
			name:        "proof verification failed",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:       commitmenttypes.MerkleProof{},
			expPass:     false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyClientConsensusState(
			suite.aminoCdc, nil, height, "chainA", 0, tc.prefix, tc.proof, nil,

			// suite.cdc, height, tc.prefix, tc.proof, nil,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *LocalhostTestSuite) TestVerifyConnectionState() {
	counterparty := connection.NewCounterparty("clientB", testConnectionID, commitmenttypes.NewMerklePrefix([]byte("ibc")))
	conn := connection.NewConnectionEnd(ibctypes.OPEN, testConnectionID, "clientA", counterparty, []string{"1.0.0"})

	testCases := []struct {
		name        string
		clientState types.ClientState
		connection  connection.End
		prefix      commitmenttypes.MerklePrefix
		proof       commitmenttypes.MerkleProof
		expPass     bool
	}{
		{
			name:        "ApplyPrefix failed",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			connection:  conn,
			prefix:      commitmenttypes.MerklePrefix{},
			expPass:     false,
		},
		{
			name:        "proof verification failed",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			connection:  conn,
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:       commitmenttypes.MerkleProof{},
			expPass:     false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyConnectionState(
			suite.cdc, height, tc.prefix, tc.proof, testConnectionID, &tc.connection, nil,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *LocalhostTestSuite) TestVerifyChannelState() {
	counterparty := channel.NewCounterparty(testPortID, testChannelID)
	ch := channel.NewChannel(ibctypes.OPEN, ibctypes.ORDERED, counterparty, []string{testConnectionID}, "1.0.0")

	testCases := []struct {
		name        string
		clientState types.ClientState
		channel     channel.Channel
		prefix      commitmenttypes.MerklePrefix
		proof       commitmenttypes.MerkleProof
		expPass     bool
	}{
		{
			name:        "ApplyPrefix failed",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			channel:     ch,
			prefix:      commitmenttypes.MerklePrefix{},
			expPass:     false,
		},
		{
			name:        "latest client height < height",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			channel:     ch,
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass:     false,
		},
		{
			name:        "proof verification failed",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			channel:     ch,
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:       commitmenttypes.MerkleProof{},
			expPass:     false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyChannelState(
			suite.cdc, height, tc.prefix, tc.proof, testPortID, testChannelID, &tc.channel, nil,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *LocalhostTestSuite) TestVerifyPacketCommitment() {
	testCases := []struct {
		name        string
		clientState types.ClientState
		commitment  []byte
		prefix      commitmenttypes.MerklePrefix
		proof       commitmenttypes.MerkleProof
		expPass     bool
	}{
		{
			name:        "ApplyPrefix failed",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			commitment:  []byte{},
			prefix:      commitmenttypes.MerklePrefix{},
			expPass:     false,
		},
		{
			name:        "latest client height < height",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			commitment:  []byte{},
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass:     false,
		},
		{
			name:        "client is frozen",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			commitment:  []byte{},
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass:     false,
		},
		{
			name:        "proof verification failed",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			commitment:  []byte{},
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:       commitmenttypes.MerkleProof{},
			expPass:     false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyPacketCommitment(
			height, tc.prefix, tc.proof, testPortID, testChannelID, testSequence, tc.commitment, nil,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *LocalhostTestSuite) TestVerifyPacketAcknowledgement() {
	testCases := []struct {
		name        string
		clientState types.ClientState
		ack         []byte
		prefix      commitmenttypes.MerklePrefix
		proof       commitmenttypes.MerkleProof
		expPass     bool
	}{
		{
			name:        "ApplyPrefix failed",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			ack:         []byte{},
			prefix:      commitmenttypes.MerklePrefix{},
			expPass:     false,
		},
		{
			name:        "latest client height < height",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			ack:         []byte{},
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass:     false,
		},
		{
			name:        "client is frozen",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			ack:         []byte{},
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass:     false,
		},
		{
			name:        "proof verification failed",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			ack:         []byte{},
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:       commitmenttypes.MerkleProof{},
			expPass:     false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyPacketAcknowledgement(
			height, tc.prefix, tc.proof, testPortID, testChannelID, testSequence, tc.ack, nil,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *LocalhostTestSuite) TestVerifyPacketAcknowledgementAbsence() {
	testCases := []struct {
		name        string
		clientState types.ClientState
		prefix      commitmenttypes.MerklePrefix
		proof       commitmenttypes.MerkleProof
		expPass     bool
	}{
		{
			name:        "ApplyPrefix failed",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			prefix:      commitmenttypes.MerklePrefix{},
			expPass:     false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyPacketAcknowledgementAbsence(
			height, tc.prefix, tc.proof, testPortID, testChannelID, testSequence, nil,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *LocalhostTestSuite) TestVerifyNextSeqRecv() {
	testCases := []struct {
		name        string
		clientState types.ClientState
		prefix      commitmenttypes.MerklePrefix
		proof       commitmenttypes.MerkleProof
		expPass     bool
	}{
		{
			name:        "ApplyPrefix failed",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			prefix:      commitmenttypes.MerklePrefix{},
			expPass:     false,
		},
		{
			name:        "latest client height < height",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass:     false,
		},
		{
			name:        "client is frozen",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass:     false,
		},
		{
			name:        "proof verification failed",
			clientState: types.NewClientState(suite.store, "chainID", 10),
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:       commitmenttypes.MerkleProof{},
			expPass:     false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyNextSequenceRecv(
			height, tc.prefix, tc.proof, testPortID, testChannelID, testSequence, nil,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
