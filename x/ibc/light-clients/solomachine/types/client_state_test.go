package types_test

import (
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
)

const (
	counterpartyClientIdentifier = "chainA"
	testConnectionID             = "connectionid"
	testChannelID                = "testchannelid"
	testPortID                   = "testportid"
)

var (
	prefix          = commitmenttypes.NewMerklePrefix([]byte("ibc"))
	consensusHeight = clienttypes.Height{}
)

func (suite *SoloMachineTestSuite) TestClientStateValidateBasic() {
	testCases := []struct {
		name        string
		clientState *types.ClientState
		expPass     bool
	}{
		{
			"valid client state",
			suite.solomachine.ClientState(),
			true,
		},
		{
			"empty ClientState",
			&types.ClientState{},
			false,
		},
		{
			"sequence is zero",
			types.NewClientState(&types.ConsensusState{0, suite.solomachine.ConsensusState().PublicKey, suite.solomachine.Time}, false),
			false,
		},
		{
			"timstamp is zero",
			types.NewClientState(&types.ConsensusState{1, suite.solomachine.ConsensusState().PublicKey, 0}, false),
			false,
		},
		{
			"pubkey is empty",
			types.NewClientState(&types.ConsensusState{suite.solomachine.Sequence, nil, suite.solomachine.Time}, false),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {

			err := tc.clientState.Validate()

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *SoloMachineTestSuite) TestVerifyClientState() {
	// create client for tendermint so we can use client state for verification
	clientA, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
	clientState := suite.chainA.GetClientState(clientA)

	clientPrefixedPath := "clients/" + counterpartyClientIdentifier + "/" + host.ClientStatePath()
	path, err := commitmenttypes.ApplyPrefix(prefix, clientPrefixedPath)
	suite.Require().NoError(err)

	value, err := types.ClientStateSignBytes(suite.chainA.Codec, suite.solomachine.Sequence, suite.solomachine.Time, path, clientState)
	suite.Require().NoError(err)

	sig, err := suite.solomachine.PrivateKey.Sign(value)
	suite.Require().NoError(err)

	signatureDoc := &types.TimestampedSignature{
		Signature: sig,
		Timestamp: suite.solomachine.Time,
	}

	proof, err := suite.chainA.Codec.MarshalBinaryBare(signatureDoc)
	suite.Require().NoError(err)

	testCases := []struct {
		name        string
		clientState *types.ClientState
		prefix      exported.Prefix
		proof       []byte
		expPass     bool
	}{
		{
			"successful verification",
			suite.solomachine.ClientState(),
			prefix,
			proof,
			true,
		},
		{
			"ApplyPrefix failed",
			suite.solomachine.ClientState(),
			nil,
			proof,
			false,
		},
		{
			"client is frozen",
			&types.ClientState{1, suite.solomachine.ConsensusState(), false},
			prefix,
			proof,
			false,
		},
		{
			"consensus state in client state is nil",
			types.NewClientState(nil, false),
			prefix,
			proof,
			false,
		},
		{
			"client state latest height is less than sequence",
			types.NewClientState(
				&types.ConsensusState{
					Sequence:  suite.solomachine.Sequence - 1,
					Timestamp: suite.solomachine.Time,
					PublicKey: suite.solomachine.ConsensusState().PublicKey,
				}, false),
			prefix,
			proof,
			false,
		},
		{
			"consensus state timestamp is greater than signature",
			types.NewClientState(
				&types.ConsensusState{
					Sequence:  suite.solomachine.Sequence,
					Timestamp: suite.solomachine.Time + 1,
					PublicKey: suite.solomachine.ConsensusState().PublicKey,
				}, false),
			prefix,
			proof,
			false,
		},

		{
			"proof is nil",
			suite.solomachine.ClientState(),
			prefix,
			nil,
			false,
		},
		{
			"proof verification failed",
			suite.solomachine.ClientState(),
			prefix,
			suite.GetInvalidProof(),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {

			var expSeq uint64
			if tc.clientState.ConsensusState != nil {
				expSeq = tc.clientState.ConsensusState.Sequence + 1
			}

			err := tc.clientState.VerifyClientState(
				suite.store, suite.chainA.Codec, nil, suite.solomachine.GetHeight(), tc.prefix, counterpartyClientIdentifier, tc.proof, clientState,
			)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %s", suite.GetSequenceFromStore(), tc.name)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *SoloMachineTestSuite) TestVerifyClientConsensusState() {
	// create client for tendermint so we can use consensus state for verification
	clientA, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
	clientState := suite.chainA.GetClientState(clientA)
	consensusState, found := suite.chainA.GetConsensusState(clientA, clientState.GetLatestHeight())
	suite.Require().True(found)

	clientPrefixedPath := "clients/" + counterpartyClientIdentifier + "/" + host.ConsensusStatePath(consensusHeight)
	path, err := commitmenttypes.ApplyPrefix(prefix, clientPrefixedPath)
	suite.Require().NoError(err)

	value, err := types.ConsensusStateSignBytes(suite.chainA.Codec, suite.solomachine.Sequence, suite.solomachine.Time, path, consensusState)
	suite.Require().NoError(err)

	sig, err := suite.solomachine.PrivateKey.Sign(value)
	suite.Require().NoError(err)

	signatureDoc := &types.TimestampedSignature{
		Signature: sig,
		Timestamp: suite.solomachine.Time,
	}

	proof, err := suite.chainA.Codec.MarshalBinaryBare(signatureDoc)
	suite.Require().NoError(err)

	testCases := []struct {
		name        string
		clientState *types.ClientState
		prefix      exported.Prefix
		proof       []byte
		expPass     bool
	}{
		{
			"successful verification",
			suite.solomachine.ClientState(),
			prefix,
			proof,
			true,
		},
		{
			"ApplyPrefix failed",
			suite.solomachine.ClientState(),
			nil,
			proof,
			false,
		},
		{
			"client is frozen",
			&types.ClientState{1, suite.solomachine.ConsensusState(), false},
			prefix,
			proof,
			false,
		},
		{
			"consensus state in client state is nil",
			types.NewClientState(nil, false),
			prefix,
			proof,
			false,
		},
		{
			"client state latest height is less than sequence",
			types.NewClientState(
				&types.ConsensusState{
					Sequence:  suite.solomachine.Sequence - 1,
					Timestamp: suite.solomachine.Time,
					PublicKey: suite.solomachine.ConsensusState().PublicKey,
				}, false),
			prefix,
			proof,
			false,
		},
		{
			"consensus state timestamp is greater than signature",
			types.NewClientState(
				&types.ConsensusState{
					Sequence:  suite.solomachine.Sequence,
					Timestamp: suite.solomachine.Time + 1,
					PublicKey: suite.solomachine.ConsensusState().PublicKey,
				}, false),
			prefix,
			proof,
			false,
		},

		{
			"proof is nil",
			suite.solomachine.ClientState(),
			prefix,
			nil,
			false,
		},
		{
			"proof verification failed",
			suite.solomachine.ClientState(),
			prefix,
			suite.GetInvalidProof(),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {

			var expSeq uint64
			if tc.clientState.ConsensusState != nil {
				expSeq = tc.clientState.ConsensusState.Sequence + 1
			}

			err := tc.clientState.VerifyClientConsensusState(
				suite.store, suite.chainA.Codec, nil, suite.solomachine.GetHeight(), counterpartyClientIdentifier, consensusHeight, tc.prefix, tc.proof, consensusState,
			)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %s", suite.GetSequenceFromStore(), tc.name)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *SoloMachineTestSuite) TestVerifyConnectionState() {
	counterparty := connectiontypes.NewCounterparty("clientB", testConnectionID, prefix)
	conn := connectiontypes.NewConnectionEnd(connectiontypes.OPEN, "clientA", counterparty, []string{"1.0.0"})

	path, err := commitmenttypes.ApplyPrefix(prefix, host.ConnectionPath(testConnectionID))
	suite.Require().NoError(err)

	value, err := types.ConnectionStateSignBytes(suite.chainA.Codec, suite.solomachine.Sequence, suite.solomachine.Time, path, conn)
	suite.Require().NoError(err)

	sig, err := suite.solomachine.PrivateKey.Sign(value)
	suite.Require().NoError(err)

	signatureDoc := &types.TimestampedSignature{
		Signature: sig,
		Timestamp: suite.solomachine.Time,
	}

	proof, err := suite.chainA.Codec.MarshalBinaryBare(signatureDoc)
	suite.Require().NoError(err)

	testCases := []struct {
		name        string
		clientState *types.ClientState
		prefix      exported.Prefix
		proof       []byte
		expPass     bool
	}{
		{
			"successful verification",
			suite.solomachine.ClientState(),
			prefix,
			proof,
			true,
		},
		{
			"ApplyPrefix failed",
			suite.solomachine.ClientState(),
			commitmenttypes.NewMerklePrefix([]byte{}),
			proof,
			false,
		},
		{
			"client is frozen",
			&types.ClientState{1, suite.solomachine.ConsensusState(), false},
			prefix,
			proof,
			false,
		},
		{
			"proof is nil",
			suite.solomachine.ClientState(),
			prefix,
			nil,
			false,
		},
		{
			"proof verification failed",
			suite.solomachine.ClientState(),
			prefix,
			suite.GetInvalidProof(),
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		expSeq := tc.clientState.ConsensusState.Sequence + 1

		err := tc.clientState.VerifyConnectionState(
			suite.store, suite.chainA.Codec, suite.solomachine.GetHeight(), tc.prefix, tc.proof, testConnectionID, conn,
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

	value, err := types.ChannelStateSignBytes(suite.chainA.Codec, suite.solomachine.Sequence, suite.solomachine.Time, path, ch)
	suite.Require().NoError(err)

	sig, err := suite.solomachine.PrivateKey.Sign(value)
	suite.Require().NoError(err)

	signatureDoc := &types.TimestampedSignature{
		Signature: sig,
		Timestamp: suite.solomachine.Time,
	}

	proof, err := suite.chainA.Codec.MarshalBinaryBare(signatureDoc)
	suite.Require().NoError(err)

	testCases := []struct {
		name        string
		clientState *types.ClientState
		prefix      exported.Prefix
		proof       []byte
		expPass     bool
	}{
		{
			"successful verification",
			suite.solomachine.ClientState(),
			prefix,
			proof,
			true,
		},
		{
			"ApplyPrefix failed",
			suite.solomachine.ClientState(),
			nil,
			proof,
			false,
		},
		{
			"client is frozen",
			&types.ClientState{1, suite.solomachine.ConsensusState(), false},
			prefix,
			proof,
			false,
		},
		{
			"proof is nil",
			suite.solomachine.ClientState(),
			prefix,
			nil,
			false,
		},
		{
			"proof verification failed",
			suite.solomachine.ClientState(),
			prefix,
			suite.GetInvalidProof(),
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		expSeq := tc.clientState.ConsensusState.Sequence + 1

		err := tc.clientState.VerifyChannelState(
			suite.store, suite.chainA.Codec, suite.solomachine.GetHeight(), tc.prefix, tc.proof, testPortID, testChannelID, ch,
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
	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketCommitmentPath(testPortID, testChannelID, suite.solomachine.Sequence))
	suite.Require().NoError(err)

	value := types.PacketCommitmentSignBytes(suite.solomachine.Sequence, suite.solomachine.Time, path, commitmentBytes)

	sig, err := suite.solomachine.PrivateKey.Sign(value)
	suite.Require().NoError(err)

	signatureDoc := &types.TimestampedSignature{
		Signature: sig,
		Timestamp: suite.solomachine.Time,
	}

	proof, err := suite.chainA.Codec.MarshalBinaryBare(signatureDoc)
	suite.Require().NoError(err)

	testCases := []struct {
		name        string
		clientState *types.ClientState
		prefix      exported.Prefix
		proof       []byte
		expPass     bool
	}{
		{
			"successful verification",
			suite.solomachine.ClientState(),
			prefix,
			proof,
			true,
		},
		{
			"ApplyPrefix failed",
			suite.solomachine.ClientState(),
			commitmenttypes.NewMerklePrefix([]byte{}),
			proof,
			false,
		},
		{
			"client is frozen",
			&types.ClientState{1, suite.solomachine.ConsensusState(), false},
			prefix,
			proof,
			false,
		},
		{
			"proof is nil",
			suite.solomachine.ClientState(),
			prefix,
			nil,
			false,
		},
		{
			"proof verification failed",
			suite.solomachine.ClientState(),
			prefix,
			suite.GetInvalidProof(),
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		expSeq := tc.clientState.ConsensusState.Sequence + 1

		err := tc.clientState.VerifyPacketCommitment(
			suite.store, suite.chainA.Codec, suite.solomachine.GetHeight(), tc.prefix, tc.proof, testPortID, testChannelID, suite.solomachine.Sequence, commitmentBytes,
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
	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketAcknowledgementPath(testPortID, testChannelID, suite.solomachine.Sequence))
	suite.Require().NoError(err)

	value := types.PacketAcknowledgementSignBytes(suite.solomachine.Sequence, suite.solomachine.Time, path, ack)

	sig, err := suite.solomachine.PrivateKey.Sign(value)
	suite.Require().NoError(err)

	signatureDoc := &types.TimestampedSignature{
		Signature: sig,
		Timestamp: suite.solomachine.Time,
	}

	proof, err := suite.chainA.Codec.MarshalBinaryBare(signatureDoc)
	suite.Require().NoError(err)

	testCases := []struct {
		name        string
		clientState *types.ClientState
		prefix      exported.Prefix
		proof       []byte
		expPass     bool
	}{
		{
			"successful verification",
			suite.solomachine.ClientState(),
			prefix,
			proof,
			true,
		},
		{
			"ApplyPrefix failed",
			suite.solomachine.ClientState(),
			commitmenttypes.NewMerklePrefix([]byte{}),
			proof,
			false,
		},
		{
			"client is frozen",
			&types.ClientState{1, suite.solomachine.ConsensusState(), false},
			prefix,
			proof,
			false,
		},
		{
			"proof is nil",
			suite.solomachine.ClientState(),
			prefix,
			nil,
			false,
		},
		{
			"proof verification failed",
			suite.solomachine.ClientState(),
			prefix,
			suite.GetInvalidProof(),
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		expSeq := tc.clientState.ConsensusState.Sequence + 1

		err := tc.clientState.VerifyPacketAcknowledgement(
			suite.store, suite.chainA.Codec, suite.solomachine.GetHeight(), tc.prefix, tc.proof, testPortID, testChannelID, suite.solomachine.Sequence, ack,
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
	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketAcknowledgementPath(testPortID, testChannelID, suite.solomachine.Sequence))
	suite.Require().NoError(err)

	value := types.PacketAcknowledgementAbsenceSignBytes(suite.solomachine.Sequence, suite.solomachine.Time, path)

	sig, err := suite.solomachine.PrivateKey.Sign(value)
	suite.Require().NoError(err)

	signatureDoc := &types.TimestampedSignature{
		Signature: sig,
		Timestamp: suite.solomachine.Time,
	}

	proof, err := suite.chainA.Codec.MarshalBinaryBare(signatureDoc)
	suite.Require().NoError(err)

	testCases := []struct {
		name        string
		clientState *types.ClientState
		prefix      exported.Prefix
		proof       []byte
		expPass     bool
	}{
		{
			"successful verification",
			suite.solomachine.ClientState(),
			prefix,
			proof,
			true,
		},
		{
			"ApplyPrefix failed",
			suite.solomachine.ClientState(),
			commitmenttypes.NewMerklePrefix([]byte{}),
			proof,
			false,
		},
		{
			"client is frozen",
			&types.ClientState{1, suite.solomachine.ConsensusState(), false},
			prefix,
			proof,
			false,
		},
		{
			"proof is nil",
			suite.solomachine.ClientState(),
			prefix,
			nil,
			false,
		},
		{
			"proof verification failed",
			suite.solomachine.ClientState(),
			prefix,
			suite.GetInvalidProof(),
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		expSeq := tc.clientState.ConsensusState.Sequence + 1

		err := tc.clientState.VerifyPacketAcknowledgementAbsence(
			suite.store, suite.chainA.Codec, suite.solomachine.GetHeight(), tc.prefix, tc.proof, testPortID, testChannelID, suite.solomachine.Sequence,
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
	nextSeqRecv := suite.solomachine.Sequence + 1
	path, err := commitmenttypes.ApplyPrefix(prefix, host.NextSequenceRecvPath(testPortID, testChannelID))
	suite.Require().NoError(err)

	value := types.NextSequenceRecvSignBytes(suite.solomachine.Sequence, suite.solomachine.Time, path, nextSeqRecv)

	sig, err := suite.solomachine.PrivateKey.Sign(value)
	suite.Require().NoError(err)

	signatureDoc := &types.TimestampedSignature{
		Signature: sig,
		Timestamp: suite.solomachine.Time,
	}

	proof, err := suite.chainA.Codec.MarshalBinaryBare(signatureDoc)
	suite.Require().NoError(err)

	testCases := []struct {
		name        string
		clientState *types.ClientState
		prefix      exported.Prefix
		proof       []byte
		expPass     bool
	}{
		{
			"successful verification",
			suite.solomachine.ClientState(),
			prefix,
			proof,
			true,
		},
		{
			"ApplyPrefix failed",
			suite.solomachine.ClientState(),
			commitmenttypes.NewMerklePrefix([]byte{}),
			proof,
			false,
		},
		{
			"client is frozen",
			&types.ClientState{1, suite.solomachine.ConsensusState(), false},
			prefix,
			proof,
			false,
		},
		{
			"proof is nil",
			suite.solomachine.ClientState(),
			prefix,
			nil,
			false,
		},
		{
			"proof verification failed",
			suite.solomachine.ClientState(),
			prefix,
			suite.GetInvalidProof(),
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		expSeq := tc.clientState.ConsensusState.Sequence + 1

		err := tc.clientState.VerifyNextSequenceRecv(
			suite.store, suite.chainA.Codec, suite.solomachine.GetHeight(), tc.prefix, tc.proof, testPortID, testChannelID, nextSeqRecv,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %d: %s", suite.GetSequenceFromStore(), i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
