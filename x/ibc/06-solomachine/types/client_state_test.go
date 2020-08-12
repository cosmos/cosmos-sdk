package types_test

import (
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	types "github.com/cosmos/cosmos-sdk/x/ibc/06-solomachine/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

const (
	counterpartyClientIdentifier = "chainA"
	consensusHeight              = uint64(0)
	testConnectionID             = "connectionid"
	testChannelID                = "testchannelid"
	testPortID                   = "testportid"
	timestamp                    = uint64(10)
)

var (
	invalidProof = []byte("invalid proof bytes")
	prefix       = commitmenttypes.NewMerklePrefix([]byte("ibc"))
)

func (suite *SoloMachineTestSuite) TestClientStateValidateBasic() {
	testCases := []struct {
		name        string
		clientState *types.ClientState
		expPass     bool
	}{
		{
			"valid client state",
			suite.ClientState(),
			true,
		},
		{
			"invalid client id",
			types.NewClientState("(testClientID)", "", suite.ConsensusState()),
			false,
		},
		{
			"sequence is zero",
			types.NewClientState(suite.clientID, "", &types.ConsensusState{0, suite.pubKey, timestamp}),
			false,
		},
		{
			"timstamp is zero",
			types.NewClientState(suite.clientID, "", &types.ConsensusState{1, suite.pubKey, 0}),
			false,
		},
		{
			"pubkey is empty",
			types.NewClientState(suite.clientID, "", &types.ConsensusState{suite.sequence, nil, timestamp}),
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.Validate()

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyClientConsensusState() {
	clientPrefixedPath := "clients/" + counterpartyClientIdentifier + "/" + host.ConsensusStatePath(consensusHeight)
	path, err := commitmenttypes.ApplyPrefix(prefix, clientPrefixedPath)
	suite.Require().NoError(err)

	value, err := types.ConsensusStateSignBytes(suite.cdc, suite.sequence, suite.timestamp, path, suite.ClientState().ConsensusState)
	suite.Require().NoError(err)

	sig, err := suite.privKey.Sign(value)
	suite.Require().NoError(err)

	signatureDoc := &types.Signature{
		Signature: sig,
		Timestamp: suite.timestamp,
	}

	proof, err := suite.cdc.MarshalBinaryBare(signatureDoc)
	suite.Require().NoError(err)

	testCases := []struct {
		name        string
		clientState *types.ClientState
		prefix      commitmentexported.Prefix
		proof       []byte
		expPass     bool
	}{
		{
			"successful verification",
			suite.ClientState(),
			prefix,
			proof,
			true,
		},
		{
			"ApplyPrefix failed",
			suite.ClientState(),
			commitmenttypes.MerklePrefix{},
			proof,
			false,
		},
		{
			"client is frozen",
			&types.ClientState{suite.clientID, "", 1, suite.ConsensusState()},
			prefix,
			proof,
			false,
		},
		{
			"proof is nil",
			suite.ClientState(),
			prefix,
			nil,
			false,
		},
		{
			"proof verification failed",
			suite.ClientState(),
			prefix,
			invalidProof,
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		expSeq := tc.clientState.ConsensusState.Sequence + 1

		err := tc.clientState.VerifyClientConsensusState(
			suite.store, suite.cdc, nil, suite.sequence, counterpartyClientIdentifier, consensusHeight, tc.prefix, tc.proof, tc.clientState.ConsensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %d: %s", suite.GetSequenceFromStore(), i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyConnectionState() {
	counterparty := connectiontypes.NewCounterparty("clientB", testConnectionID, prefix)
	conn := connectiontypes.NewConnectionEnd(connectiontypes.OPEN, "clientA", counterparty, []string{"1.0.0"})

	path, err := commitmenttypes.ApplyPrefix(prefix, host.ConnectionPath(testConnectionID))
	suite.Require().NoError(err)

	value, err := types.ConnectionStateSignBytes(suite.cdc, suite.sequence, suite.timestamp, path, conn)
	suite.Require().NoError(err)

	sig, err := suite.privKey.Sign(value)
	suite.Require().NoError(err)

	signatureDoc := &types.Signature{
		Signature: sig,
		Timestamp: suite.timestamp,
	}

	proof, err := suite.cdc.MarshalBinaryBare(signatureDoc)
	suite.Require().NoError(err)

	testCases := []struct {
		name        string
		clientState *types.ClientState
		prefix      commitmentexported.Prefix
		proof       []byte
		expPass     bool
	}{
		{
			"successful verification",
			suite.ClientState(),
			prefix,
			proof,
			true,
		},
		{
			"ApplyPrefix failed",
			suite.ClientState(),
			commitmenttypes.NewMerklePrefix([]byte{}),
			proof,
			false,
		},
		{
			"client is frozen",
			&types.ClientState{suite.clientID, "", 1, suite.ConsensusState()},
			prefix,
			proof,
			false,
		},
		{
			"proof is nil",
			suite.ClientState(),
			prefix,
			nil,
			false,
		},
		{
			"proof verification failed",
			suite.ClientState(),
			prefix,
			invalidProof,
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		expSeq := tc.clientState.ConsensusState.Sequence + 1

		err := tc.clientState.VerifyConnectionState(
			suite.store, suite.cdc, suite.sequence, tc.prefix, tc.proof, testConnectionID, conn,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %d: %s", suite.GetSequenceFromStore(), i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyChannelState() {
	counterparty := channeltypes.NewCounterparty(testPortID, testChannelID)
	ch := channeltypes.NewChannel(channeltypes.OPEN, channeltypes.ORDERED, counterparty, []string{testConnectionID}, "1.0.0")

	path, err := commitmenttypes.ApplyPrefix(prefix, host.ChannelPath(testPortID, testChannelID))
	suite.Require().NoError(err)

	value, err := types.ChannelStateSignBytes(suite.cdc, suite.sequence, suite.timestamp, path, ch)
	suite.Require().NoError(err)

	sig, err := suite.privKey.Sign(value)
	suite.Require().NoError(err)

	signatureDoc := &types.Signature{
		Signature: sig,
		Timestamp: suite.timestamp,
	}

	proof, err := suite.cdc.MarshalBinaryBare(signatureDoc)
	suite.Require().NoError(err)

	testCases := []struct {
		name        string
		clientState *types.ClientState
		prefix      commitmentexported.Prefix
		proof       []byte
		expPass     bool
	}{
		{
			"successful verification",
			suite.ClientState(),
			prefix,
			proof,
			true,
		},
		{
			"ApplyPrefix failed",
			suite.ClientState(),
			commitmenttypes.NewMerklePrefix([]byte{}),
			proof,
			false,
		},
		{
			"client is frozen",
			&types.ClientState{suite.clientID, "", 1, suite.ConsensusState()},
			prefix,
			proof,
			false,
		},
		{
			"proof is nil",
			suite.ClientState(),
			prefix,
			nil,
			false,
		},
		{
			"proof verification failed",
			suite.ClientState(),
			prefix,
			invalidProof,
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		expSeq := tc.clientState.ConsensusState.Sequence + 1

		err := tc.clientState.VerifyChannelState(
			suite.store, suite.cdc, suite.sequence, tc.prefix, tc.proof, testPortID, testChannelID, ch,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %d: %s", suite.GetSequenceFromStore(), i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyPacketCommitment() {
	commitmentBytes := []byte("COMMITMENT BYTES")
	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketCommitmentPath(testPortID, testChannelID, suite.sequence))
	suite.Require().NoError(err)

	value := types.PacketCommitmentSignBytes(suite.sequence, suite.timestamp, path, commitmentBytes)

	sig, err := suite.privKey.Sign(value)
	suite.Require().NoError(err)

	signatureDoc := &types.Signature{
		Signature: sig,
		Timestamp: suite.timestamp,
	}

	proof, err := suite.cdc.MarshalBinaryBare(signatureDoc)
	suite.Require().NoError(err)

	testCases := []struct {
		name        string
		clientState *types.ClientState
		prefix      commitmentexported.Prefix
		proof       []byte
		expPass     bool
	}{
		{
			"successful verification",
			suite.ClientState(),
			prefix,
			proof,
			true,
		},
		{
			"ApplyPrefix failed",
			suite.ClientState(),
			commitmenttypes.NewMerklePrefix([]byte{}),
			proof,
			false,
		},
		{
			"client is frozen",
			&types.ClientState{suite.clientID, "", 1, suite.ConsensusState()},
			prefix,
			proof,
			false,
		},
		{
			"proof is nil",
			suite.ClientState(),
			prefix,
			nil,
			false,
		},
		{
			"proof verification failed",
			suite.ClientState(),
			prefix,
			invalidProof,
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		expSeq := tc.clientState.ConsensusState.Sequence + 1

		err := tc.clientState.VerifyPacketCommitment(
			suite.store, suite.cdc, suite.sequence, tc.prefix, tc.proof, testPortID, testChannelID, suite.sequence, commitmentBytes,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %d: %s", suite.GetSequenceFromStore(), i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyPacketAcknowledgement() {
	ack := []byte("ACK")
	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketAcknowledgementPath(testPortID, testChannelID, suite.sequence))
	suite.Require().NoError(err)

	value := types.PacketAcknowledgementSignBytes(suite.sequence, suite.timestamp, path, ack)

	sig, err := suite.privKey.Sign(value)
	suite.Require().NoError(err)

	signatureDoc := &types.Signature{
		Signature: sig,
		Timestamp: suite.timestamp,
	}

	proof, err := suite.cdc.MarshalBinaryBare(signatureDoc)
	suite.Require().NoError(err)

	testCases := []struct {
		name        string
		clientState *types.ClientState
		prefix      commitmentexported.Prefix
		proof       []byte
		expPass     bool
	}{
		{
			"successful verification",
			suite.ClientState(),
			prefix,
			proof,
			true,
		},
		{
			"ApplyPrefix failed",
			suite.ClientState(),
			commitmenttypes.NewMerklePrefix([]byte{}),
			proof,
			false,
		},
		{
			"client is frozen",
			&types.ClientState{suite.clientID, "", 1, suite.ConsensusState()},
			prefix,
			proof,
			false,
		},
		{
			"proof is nil",
			suite.ClientState(),
			prefix,
			nil,
			false,
		},
		{
			"proof verification failed",
			suite.ClientState(),
			prefix,
			invalidProof,
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		expSeq := tc.clientState.ConsensusState.Sequence + 1

		err := tc.clientState.VerifyPacketAcknowledgement(
			suite.store, suite.cdc, suite.sequence, tc.prefix, tc.proof, testPortID, testChannelID, suite.sequence, ack,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %d: %s", suite.GetSequenceFromStore(), i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyPacketAcknowledgementAbsence() {
	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketAcknowledgementPath(testPortID, testChannelID, suite.sequence))
	suite.Require().NoError(err)

	value := types.PacketAcknowledgementAbsenceSignBytes(suite.sequence, suite.timestamp, path)

	sig, err := suite.privKey.Sign(value)
	suite.Require().NoError(err)

	signatureDoc := &types.Signature{
		Signature: sig,
		Timestamp: suite.timestamp,
	}

	proof, err := suite.cdc.MarshalBinaryBare(signatureDoc)
	suite.Require().NoError(err)

	testCases := []struct {
		name        string
		clientState *types.ClientState
		prefix      commitmentexported.Prefix
		proof       []byte
		expPass     bool
	}{
		{
			"successful verification",
			suite.ClientState(),
			prefix,
			proof,
			true,
		},
		{
			"ApplyPrefix failed",
			suite.ClientState(),
			commitmenttypes.NewMerklePrefix([]byte{}),
			proof,
			false,
		},
		{
			"client is frozen",
			&types.ClientState{suite.clientID, "", 1, suite.ConsensusState()},
			prefix,
			proof,
			false,
		},
		{
			"proof is nil",
			suite.ClientState(),
			prefix,
			nil,
			false,
		},
		{
			"proof verification failed",
			suite.ClientState(),
			prefix,
			invalidProof,
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		expSeq := tc.clientState.ConsensusState.Sequence + 1

		err := tc.clientState.VerifyPacketAcknowledgementAbsence(
			suite.store, suite.cdc, suite.sequence, tc.prefix, tc.proof, testPortID, testChannelID, suite.sequence,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %d: %s", suite.GetSequenceFromStore(), i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyNextSeqRecv() {
	nextSeqRecv := suite.sequence + 1
	path, err := commitmenttypes.ApplyPrefix(prefix, host.NextSequenceRecvPath(testPortID, testChannelID))
	suite.Require().NoError(err)

	value := types.NextSequenceRecvSignBytes(suite.sequence, suite.timestamp, path, nextSeqRecv)

	sig, err := suite.privKey.Sign(value)
	suite.Require().NoError(err)

	signatureDoc := &types.Signature{
		Signature: sig,
		Timestamp: suite.timestamp,
	}

	proof, err := suite.cdc.MarshalBinaryBare(signatureDoc)
	suite.Require().NoError(err)

	testCases := []struct {
		name        string
		clientState *types.ClientState
		prefix      commitmentexported.Prefix
		proof       []byte
		expPass     bool
	}{
		{
			"successful verification",
			suite.ClientState(),
			prefix,
			proof,
			true,
		},
		{
			"ApplyPrefix failed",
			suite.ClientState(),
			commitmenttypes.NewMerklePrefix([]byte{}),
			proof,
			false,
		},
		{
			"client is frozen",
			&types.ClientState{suite.clientID, "", 1, suite.ConsensusState()},
			prefix,
			proof,
			false,
		},
		{
			"proof is nil",
			suite.ClientState(),
			prefix,
			nil,
			false,
		},
		{
			"proof verification failed",
			suite.ClientState(),
			prefix,
			invalidProof,
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		expSeq := tc.clientState.ConsensusState.Sequence + 1

		err := tc.clientState.VerifyNextSequenceRecv(
			suite.store, suite.cdc, suite.sequence, tc.prefix, tc.proof, testPortID, testChannelID, nextSeqRecv,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %d: %s", suite.GetSequenceFromStore(), i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
