package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	solomachinetypes "github.com/cosmos/cosmos-sdk/x/ibc/06-solomachine/types"
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
)

var (
	prefix = commitmenttypes.NewSignaturePrefix([]byte("ibc"))
)

func (suite *SoloMachineTestSuite) TestClientStateValidateBasic() {
	testCases := []struct {
		name        string
		clientState solomachinetypes.ClientState
		expPass     bool
	}{
		{
			"valid client state",
			suite.ClientState(),
			true,
		},
		{
			"invalid client id",
			solomachinetypes.NewClientState("(testClientID)", suite.ConsensusState()),
			false,
		},
		{
			"sequence is zero",
			solomachinetypes.NewClientState(suite.clientID, solomachinetypes.ConsensusState{0, suite.privKey.PubKey()}),
			false,
		},
		{
			"pubkey is empty",
			solomachinetypes.NewClientState(suite.clientID, solomachinetypes.ConsensusState{suite.sequence, nil}),
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

	value := append(sdk.Uint64ToBigEndian(suite.sequence), []byte(path.String())...)
	bz, err := suite.aminoCdc.MarshalBinaryBare(suite.ClientState().ConsensusState)
	suite.Require().NoError(err)
	value = append(value, bz...)

	sig, err := suite.privKey.Sign(value)
	suite.Require().NoError(err)
	proof := commitmenttypes.SignatureProof{Signature: sig}

	testCases := []struct {
		name        string
		clientState solomachinetypes.ClientState
		prefix      commitmentexported.Prefix
		proof       commitmentexported.Proof
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
			commitmenttypes.SignaturePrefix{},
			proof,
			false,
		},
		{
			"client is frozen",
			solomachinetypes.ClientState{suite.clientID, true, suite.ConsensusState()},
			prefix,
			proof,
			false,
		},
		{
			"invalid proof type",
			suite.ClientState(),
			prefix,
			commitmenttypes.MerkleProof{},
			false,
		},
		{
			"proof verification failed",
			suite.ClientState(),
			prefix,
			commitmenttypes.SignatureProof{},
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyClientConsensusState(
			suite.store, suite.aminoCdc, nil, 0, counterpartyClientIdentifier, consensusHeight, tc.prefix, tc.proof, tc.clientState.ConsensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			expSeq := tc.clientState.ConsensusState.Sequence + 1
			suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %d: %s", suite.GetSequenceFromStore(), i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyConnectionState() {
	counterparty, err := connection.NewCounterparty("clientB", testConnectionID, prefix)
	suite.Require().NoError(err)
	conn := connection.NewConnectionEnd(connection.OPEN, testConnectionID, "clientA", counterparty, []string{"1.0.0"})

	path, err := commitmenttypes.ApplyPrefix(prefix, host.ConnectionPath(testConnectionID))
	suite.Require().NoError(err)

	value := append(sdk.Uint64ToBigEndian(suite.sequence), []byte(path.String())...)
	bz, err := suite.cdc.MarshalBinaryBare(&conn)
	suite.Require().NoError(err)
	value = append(value, bz...)

	sig, err := suite.privKey.Sign(value)
	suite.Require().NoError(err)
	proof := commitmenttypes.SignatureProof{Signature: sig}

	testCases := []struct {
		name        string
		clientState solomachinetypes.ClientState
		prefix      commitmentexported.Prefix
		proof       commitmentexported.Proof
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
			commitmenttypes.NewSignaturePrefix([]byte{}),
			proof,
			false,
		},
		{
			"client is frozen",
			solomachinetypes.ClientState{suite.clientID, true, suite.ConsensusState()},
			prefix,
			proof,
			false,
		},
		{
			"invalid proof type",
			suite.ClientState(),
			prefix,
			commitmenttypes.MerkleProof{},
			false,
		},
		{
			"proof verification failed",
			suite.ClientState(),
			prefix,
			commitmenttypes.SignatureProof{},
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyConnectionState(
			suite.store, suite.cdc, 0, tc.prefix, tc.proof, testConnectionID, conn, tc.clientState.ConsensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)

			expSeq := tc.clientState.ConsensusState.Sequence + 1
			suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %d: %s", suite.GetSequenceFromStore(), i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyChannelState() {
	counterparty := channel.NewCounterparty(testPortID, testChannelID)
	ch := channel.NewChannel(channel.OPEN, channel.ORDERED, counterparty, []string{testConnectionID}, "1.0.0")

	path, err := commitmenttypes.ApplyPrefix(prefix, host.ChannelPath(testPortID, testChannelID))
	suite.Require().NoError(err)

	value := append(sdk.Uint64ToBigEndian(suite.sequence), []byte(path.String())...)
	bz, err := suite.cdc.MarshalBinaryBare(&ch)
	suite.Require().NoError(err)
	value = append(value, bz...)

	sig, err := suite.privKey.Sign(value)
	suite.Require().NoError(err)
	proof := commitmenttypes.SignatureProof{Signature: sig}

	testCases := []struct {
		name        string
		clientState solomachinetypes.ClientState
		prefix      commitmentexported.Prefix
		proof       commitmentexported.Proof
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
			commitmenttypes.NewSignaturePrefix([]byte{}),
			proof,
			false,
		},
		{
			"client is frozen",
			solomachinetypes.ClientState{suite.clientID, true, suite.ConsensusState()},
			prefix,
			proof,
			false,
		},
		{
			"invalid proof type",
			suite.ClientState(),
			prefix,
			commitmenttypes.MerkleProof{},
			false,
		},
		{
			"proof verification failed",
			suite.ClientState(),
			prefix,
			commitmenttypes.SignatureProof{},
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyChannelState(
			suite.store, suite.cdc, 0, tc.prefix, tc.proof, testPortID, testChannelID, ch, tc.clientState.ConsensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)

			expSeq := tc.clientState.ConsensusState.Sequence + 1
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

	value := append(sdk.Uint64ToBigEndian(suite.sequence), []byte(path.String())...)
	value = append(value, commitmentBytes...)

	sig, err := suite.privKey.Sign(value)
	suite.Require().NoError(err)
	proof := commitmenttypes.SignatureProof{Signature: sig}

	testCases := []struct {
		name        string
		clientState solomachinetypes.ClientState
		prefix      commitmentexported.Prefix
		proof       commitmentexported.Proof
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
			commitmenttypes.NewSignaturePrefix([]byte{}),
			proof,
			false,
		},
		{
			"client is frozen",
			solomachinetypes.ClientState{suite.clientID, true, suite.ConsensusState()},
			prefix,
			proof,
			false,
		},
		{
			"invalid proof type",
			suite.ClientState(),
			prefix,
			commitmenttypes.MerkleProof{},
			false,
		},
		{
			"proof verification failed",
			suite.ClientState(),
			prefix,
			commitmenttypes.SignatureProof{},
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyPacketCommitment(
			suite.store, 0, tc.prefix, tc.proof, testPortID, testChannelID, suite.sequence, commitmentBytes, tc.clientState.ConsensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)

			expSeq := tc.clientState.ConsensusState.Sequence + 1
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

	value := append(sdk.Uint64ToBigEndian(suite.sequence), []byte(path.String())...)
	value = append(value, ack...)

	sig, err := suite.privKey.Sign(value)
	suite.Require().NoError(err)
	proof := commitmenttypes.SignatureProof{Signature: sig}

	testCases := []struct {
		name        string
		clientState solomachinetypes.ClientState
		prefix      commitmentexported.Prefix
		proof       commitmentexported.Proof
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
			commitmenttypes.NewSignaturePrefix([]byte{}),
			proof,
			false,
		},
		{
			"client is frozen",
			solomachinetypes.ClientState{suite.clientID, true, suite.ConsensusState()},
			prefix,
			proof,
			false,
		},
		{
			"invalid proof type",
			suite.ClientState(),
			prefix,
			commitmenttypes.MerkleProof{},
			false,
		},
		{
			"proof verification failed",
			suite.ClientState(),
			prefix,
			commitmenttypes.SignatureProof{},
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyPacketAcknowledgement(
			suite.store, 0, tc.prefix, tc.proof, testPortID, testChannelID, suite.sequence, ack, tc.clientState.ConsensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)

			expSeq := tc.clientState.ConsensusState.Sequence + 1
			suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %d: %s", suite.GetSequenceFromStore(), i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyPacketAcknowledgementAbsence() {
	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketAcknowledgementPath(testPortID, testChannelID, suite.sequence))
	suite.Require().NoError(err)

	value := append(sdk.Uint64ToBigEndian(suite.sequence), []byte(path.String())...)

	sig, err := suite.privKey.Sign(value)
	suite.Require().NoError(err)
	proof := commitmenttypes.SignatureProof{Signature: sig}

	testCases := []struct {
		name        string
		clientState solomachinetypes.ClientState
		prefix      commitmentexported.Prefix
		proof       commitmentexported.Proof
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
			commitmenttypes.NewSignaturePrefix([]byte{}),
			proof,
			false,
		},
		{
			"client is frozen",
			solomachinetypes.ClientState{suite.clientID, true, suite.ConsensusState()},
			prefix,
			proof,
			false,
		},
		{
			"invalid proof type",
			suite.ClientState(),
			prefix,
			commitmenttypes.MerkleProof{},
			false,
		},
		{
			"proof verification failed",
			suite.ClientState(),
			prefix,
			commitmenttypes.SignatureProof{},
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyPacketAcknowledgementAbsence(
			suite.store, 0, tc.prefix, tc.proof, testPortID, testChannelID, suite.sequence, tc.clientState.ConsensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)

			expSeq := tc.clientState.ConsensusState.Sequence + 1
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

	value := append(sdk.Uint64ToBigEndian(suite.sequence), []byte(path.String())...)
	value = append(value, sdk.Uint64ToBigEndian(nextSeqRecv)...)

	sig, err := suite.privKey.Sign(value)
	suite.Require().NoError(err)
	proof := commitmenttypes.SignatureProof{Signature: sig}

	testCases := []struct {
		name        string
		clientState solomachinetypes.ClientState
		prefix      commitmentexported.Prefix
		proof       commitmentexported.Proof
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
			commitmenttypes.NewSignaturePrefix([]byte{}),
			proof,
			false,
		},
		{
			"client is frozen",
			solomachinetypes.ClientState{suite.clientID, true, suite.ConsensusState()},
			prefix,
			proof,
			false,
		},
		{
			"invalid proof type",
			suite.ClientState(),
			prefix,
			commitmenttypes.MerkleProof{},
			false,
		},
		{
			"proof verification failed",
			suite.ClientState(),
			prefix,
			commitmenttypes.SignatureProof{},
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyNextSequenceRecv(
			suite.store, 0, tc.prefix, tc.proof, testPortID, testChannelID, nextSeqRecv, tc.clientState.ConsensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)

			expSeq := tc.clientState.ConsensusState.Sequence + 1
			suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %d: %s", suite.GetSequenceFromStore(), i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
