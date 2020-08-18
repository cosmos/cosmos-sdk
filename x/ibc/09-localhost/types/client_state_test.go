package types_test

import (
	"time"

	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

const (
	testConnectionID = "connectionid"
	testPortID       = "testportid"
	testChannelID    = "testchannelid"
	testSequence     = 1
)

var latestTimestamp = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

func (suite *LocalhostTestSuite) TestValidate() {
	testCases := []struct {
		name        string
		clientState *types.ClientState
		expPass     bool
	}{
		{
			name:        "valid client",
			clientState: types.NewClientState("chainID", 10, latestTimestamp),
			expPass:     true,
		},
		{
			name:        "invalid chain id",
			clientState: types.NewClientState(" ", 10, latestTimestamp),
			expPass:     false,
		},
		{
			name:        "invalid height",
			clientState: types.NewClientState("chainID", 0, latestTimestamp),
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
	clientState := types.NewClientState("chainID", 10, latestTimestamp)
	err := clientState.VerifyClientConsensusState(
		nil, nil, nil, 0, "", 0, nil, nil, nil,
	)
	suite.Require().Error(err)
}

func (suite *LocalhostTestSuite) TestVerifyConnectionState() {
	counterparty := connectiontypes.NewCounterparty("clientB", testConnectionID, commitmenttypes.NewMerklePrefix([]byte("ibc")))
	conn := connectiontypes.NewConnectionEnd(connectiontypes.OPEN, "clientA", counterparty, []string{"1.0.0"})

	testCases := []struct {
		name        string
		clientState *types.ClientState
		connection  connectiontypes.ConnectionEnd
		prefix      commitmenttypes.MerklePrefix
		proof       []byte
		expPass     bool
	}{
		{
			name:        "ApplyPrefix failed",
			clientState: types.NewClientState("chainID", 10, latestTimestamp),
			connection:  conn,
			prefix:      commitmenttypes.MerklePrefix{},
			expPass:     false,
		},
		{
			name:        "proof verification failed",
			clientState: types.NewClientState("chainID", 10, latestTimestamp),
			connection:  conn,
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:       []byte{},
			expPass:     false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyConnectionState(
			suite.store, suite.cdc, height, tc.prefix, tc.proof, testConnectionID, &tc.connection,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *LocalhostTestSuite) TestVerifyChannelState() {
	counterparty := channeltypes.NewCounterparty(testPortID, testChannelID)
	ch := channeltypes.NewChannel(channeltypes.OPEN, channeltypes.ORDERED, counterparty, []string{testConnectionID}, "1.0.0")

	testCases := []struct {
		name        string
		clientState *types.ClientState
		channel     channeltypes.Channel
		prefix      commitmenttypes.MerklePrefix
		proof       []byte
		expPass     bool
	}{
		{
			name:        "ApplyPrefix failed",
			clientState: types.NewClientState("chainID", 10, latestTimestamp),
			channel:     ch,
			prefix:      commitmenttypes.MerklePrefix{},
			expPass:     false,
		},
		{
			name:        "latest client height < height",
			clientState: types.NewClientState("chainID", 10, latestTimestamp),
			channel:     ch,
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass:     false,
		},
		{
			name:        "proof verification failed",
			clientState: types.NewClientState("chainID", 10, latestTimestamp),
			channel:     ch,
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:       []byte{},
			expPass:     false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyChannelState(
			suite.store, suite.cdc, height, tc.prefix, tc.proof, testPortID, testChannelID, &tc.channel,
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
		clientState *types.ClientState
		commitment  []byte
		prefix      commitmenttypes.MerklePrefix
		proof       []byte
		expPass     bool
	}{
		{
			name:        "ApplyPrefix failed",
			clientState: types.NewClientState("chainID", 10, latestTimestamp),
			commitment:  []byte{},
			prefix:      commitmenttypes.MerklePrefix{},
			expPass:     false,
		},
		{
			name:        "latest client height < height",
			clientState: types.NewClientState("chainID", 10, latestTimestamp),
			commitment:  []byte{},
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass:     false,
		},
		{
			name:        "client is frozen",
			clientState: types.NewClientState("chainID", 10, latestTimestamp),
			commitment:  []byte{},
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass:     false,
		},
		{
			name:        "proof verification failed",
			clientState: types.NewClientState("chainID", 10, latestTimestamp),
			commitment:  []byte{},
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:       []byte{},
			expPass:     false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyPacketCommitment(
			suite.store, suite.cdc, height, tc.prefix, tc.proof, testPortID, testChannelID, testSequence, tc.commitment,
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
		clientState *types.ClientState
		ack         []byte
		prefix      commitmenttypes.MerklePrefix
		proof       []byte
		expPass     bool
	}{
		{
			name:        "ApplyPrefix failed",
			clientState: types.NewClientState("chainID", 10, latestTimestamp),
			ack:         []byte{},
			prefix:      commitmenttypes.MerklePrefix{},
			expPass:     false,
		},
		{
			name:        "latest client height < height",
			clientState: types.NewClientState("chainID", 10, latestTimestamp),
			ack:         []byte{},
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass:     false,
		},
		{
			name:        "client is frozen",
			clientState: types.NewClientState("chainID", 10, latestTimestamp),
			ack:         []byte{},
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass:     false,
		},
		{
			name:        "proof verification failed",
			clientState: types.NewClientState("chainID", 10, latestTimestamp),
			ack:         []byte{},
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:       []byte{},
			expPass:     false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyPacketAcknowledgement(
			suite.store, suite.cdc, height, tc.prefix, tc.proof, testPortID, testChannelID, testSequence, tc.ack,
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
		clientState *types.ClientState
		prefix      commitmenttypes.MerklePrefix
		proof       []byte
		expPass     bool
	}{
		{
			name:        "ApplyPrefix failed",
			clientState: types.NewClientState("chainID", 10, latestTimestamp),
			prefix:      commitmenttypes.MerklePrefix{},
			expPass:     false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyPacketAcknowledgementAbsence(
			suite.store, suite.cdc, height, tc.prefix, tc.proof, testPortID, testChannelID, testSequence,
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
		clientState *types.ClientState
		prefix      commitmenttypes.MerklePrefix
		proof       []byte
		expPass     bool
	}{
		{
			name:        "ApplyPrefix failed",
			clientState: types.NewClientState("chainID", 10, latestTimestamp),
			prefix:      commitmenttypes.MerklePrefix{},
			expPass:     false,
		},
		{
			name:        "latest client height < height",
			clientState: types.NewClientState("chainID", 10, latestTimestamp),
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass:     false,
		},
		{
			name:        "client is frozen",
			clientState: types.NewClientState("chainID", 10, latestTimestamp),
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass:     false,
		},
		{
			name:        "proof verification failed",
			clientState: types.NewClientState("chainID", 10, latestTimestamp),
			prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:       []byte{},
			expPass:     false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyNextSequenceRecv(
			suite.store, suite.cdc, height, tc.prefix, tc.proof, testPortID, testChannelID, testSequence,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
