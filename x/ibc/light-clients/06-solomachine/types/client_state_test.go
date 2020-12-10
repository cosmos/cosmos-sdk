package types_test

import (
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/06-solomachine/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

const (
	counterpartyClientIdentifier = "chainA"
	testConnectionID             = "connectionid"
	testChannelID                = "testchannelid"
	testPortID                   = "testportid"
)

var (
	prefix          = commitmenttypes.NewMerklePrefix([]byte("ibc"))
	consensusHeight = clienttypes.ZeroHeight()
)

func (suite *SoloMachineTestSuite) TestClientStateValidateBasic() {
	// test singlesig and multisig public keys
	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {

		testCases := []struct {
			name        string
			clientState *types.ClientState
			expPass     bool
		}{
			{
				"valid client state",
				solomachine.ClientState(),
				true,
			},
			{
				"empty ClientState",
				&types.ClientState{},
				false,
			},
			{
				"sequence is zero",
				types.NewClientState(0, &types.ConsensusState{solomachine.ConsensusState().PublicKey, solomachine.Diversifier, solomachine.Time}, false),
				false,
			},
			{
				"timestamp is zero",
				types.NewClientState(1, &types.ConsensusState{solomachine.ConsensusState().PublicKey, solomachine.Diversifier, 0}, false),
				false,
			},
			{
				"diversifier is blank",
				types.NewClientState(1, &types.ConsensusState{solomachine.ConsensusState().PublicKey, "  ", 1}, false),
				false,
			},
			{
				"pubkey is empty",
				types.NewClientState(1, &types.ConsensusState{nil, solomachine.Diversifier, solomachine.Time}, false),
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
}

func (suite *SoloMachineTestSuite) TestInitialize() {
	// test singlesig and multisig public keys
	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {
		malleatedConsensus := solomachine.ClientState().ConsensusState
		malleatedConsensus.Timestamp = malleatedConsensus.Timestamp + 10

		testCases := []struct {
			name      string
			consState exported.ConsensusState
			expPass   bool
		}{
			{
				"valid consensus state",
				solomachine.ConsensusState(),
				true,
			},
			{
				"nil consensus state",
				nil,
				false,
			},
			{
				"invalid consensus state: Tendermint consensus state",
				&ibctmtypes.ConsensusState{},
				false,
			},
			{
				"invalid consensus state: consensus state does not match consensus state in client",
				malleatedConsensus,
				false,
			},
		}

		for _, tc := range testCases {
			err := solomachine.ClientState().Initialize(
				suite.chainA.GetContext(), suite.chainA.Codec,
				suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), "solomachine"),
				tc.consState,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid testcase: %s failed", tc.name)
			} else {
				suite.Require().Error(err, "invalid testcase: %s passed", tc.name)
			}
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyClientState() {
	// create client for tendermint so we can use client state for verification
	clientA, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
	clientState := suite.chainA.GetClientState(clientA)
	path := suite.solomachine.GetClientStatePath(counterpartyClientIdentifier)

	// test singlesig and multisig public keys
	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {

		value, err := types.ClientStateSignBytes(suite.chainA.Codec, solomachine.Sequence, solomachine.Time, solomachine.Diversifier, path, clientState)
		suite.Require().NoError(err)

		sig := solomachine.GenerateSignature(value)

		signatureDoc := &types.TimestampedSignatureData{
			SignatureData: sig,
			Timestamp:     solomachine.Time,
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
				solomachine.ClientState(),
				prefix,
				proof,
				true,
			},
			{
				"ApplyPrefix failed",
				solomachine.ClientState(),
				nil,
				proof,
				false,
			},
			{
				"client is frozen",
				&types.ClientState{
					Sequence:                 1,
					FrozenSequence:           1,
					ConsensusState:           solomachine.ConsensusState(),
					AllowUpdateAfterProposal: false,
				},
				prefix,
				proof,
				false,
			},
			{
				"consensus state in client state is nil",
				types.NewClientState(1, nil, false),
				prefix,
				proof,
				false,
			},
			{
				"client state latest height is less than sequence",
				types.NewClientState(solomachine.Sequence-1,
					&types.ConsensusState{
						Timestamp: solomachine.Time,
						PublicKey: solomachine.ConsensusState().PublicKey,
					}, false),
				prefix,
				proof,
				false,
			},
			{
				"consensus state timestamp is greater than signature",
				types.NewClientState(solomachine.Sequence,
					&types.ConsensusState{
						Timestamp: solomachine.Time + 1,
						PublicKey: solomachine.ConsensusState().PublicKey,
					}, false),
				prefix,
				proof,
				false,
			},

			{
				"proof is nil",
				solomachine.ClientState(),
				prefix,
				nil,
				false,
			},
			{
				"proof verification failed",
				solomachine.ClientState(),
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
					expSeq = tc.clientState.Sequence + 1
				}

				err := tc.clientState.VerifyClientState(
					suite.store, suite.chainA.Codec, solomachine.GetHeight(), tc.prefix, counterpartyClientIdentifier, tc.proof, clientState,
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
}

func (suite *SoloMachineTestSuite) TestVerifyClientConsensusState() {
	// create client for tendermint so we can use consensus state for verification
	clientA, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
	clientState := suite.chainA.GetClientState(clientA)
	consensusState, found := suite.chainA.GetConsensusState(clientA, clientState.GetLatestHeight())
	suite.Require().True(found)

	path := suite.solomachine.GetConsensusStatePath(counterpartyClientIdentifier, consensusHeight)

	// test singlesig and multisig public keys
	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {

		value, err := types.ConsensusStateSignBytes(suite.chainA.Codec, solomachine.Sequence, solomachine.Time, solomachine.Diversifier, path, consensusState)
		suite.Require().NoError(err)

		sig := solomachine.GenerateSignature(value)
		signatureDoc := &types.TimestampedSignatureData{
			SignatureData: sig,
			Timestamp:     solomachine.Time,
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
				solomachine.ClientState(),
				prefix,
				proof,
				true,
			},
			{
				"ApplyPrefix failed",
				solomachine.ClientState(),
				nil,
				proof,
				false,
			},
			{
				"client is frozen",
				&types.ClientState{
					Sequence:                 1,
					FrozenSequence:           1,
					ConsensusState:           solomachine.ConsensusState(),
					AllowUpdateAfterProposal: false,
				},
				prefix,
				proof,
				false,
			},
			{
				"consensus state in client state is nil",
				types.NewClientState(1, nil, false),
				prefix,
				proof,
				false,
			},
			{
				"client state latest height is less than sequence",
				types.NewClientState(solomachine.Sequence-1,
					&types.ConsensusState{
						Timestamp: solomachine.Time,
						PublicKey: solomachine.ConsensusState().PublicKey,
					}, false),
				prefix,
				proof,
				false,
			},
			{
				"consensus state timestamp is greater than signature",
				types.NewClientState(solomachine.Sequence,
					&types.ConsensusState{
						Timestamp: solomachine.Time + 1,
						PublicKey: solomachine.ConsensusState().PublicKey,
					}, false),
				prefix,
				proof,
				false,
			},

			{
				"proof is nil",
				solomachine.ClientState(),
				prefix,
				nil,
				false,
			},
			{
				"proof verification failed",
				solomachine.ClientState(),
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
					expSeq = tc.clientState.Sequence + 1
				}

				err := tc.clientState.VerifyClientConsensusState(
					suite.store, suite.chainA.Codec, solomachine.GetHeight(), counterpartyClientIdentifier, consensusHeight, tc.prefix, tc.proof, consensusState,
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
}

func (suite *SoloMachineTestSuite) TestVerifyConnectionState() {
	counterparty := connectiontypes.NewCounterparty("clientB", testConnectionID, prefix)
	conn := connectiontypes.NewConnectionEnd(connectiontypes.OPEN, "clientA", counterparty, connectiontypes.ExportedVersionsToProto(connectiontypes.GetCompatibleVersions()), 0)

	path := suite.solomachine.GetConnectionStatePath(testConnectionID)

	// test singlesig and multisig public keys
	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {

		value, err := types.ConnectionStateSignBytes(suite.chainA.Codec, solomachine.Sequence, solomachine.Time, solomachine.Diversifier, path, conn)
		suite.Require().NoError(err)

		sig := solomachine.GenerateSignature(value)
		signatureDoc := &types.TimestampedSignatureData{
			SignatureData: sig,
			Timestamp:     solomachine.Time,
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
				solomachine.ClientState(),
				prefix,
				proof,
				true,
			},
			{
				"ApplyPrefix failed",
				solomachine.ClientState(),
				commitmenttypes.NewMerklePrefix([]byte{}),
				proof,
				false,
			},
			{
				"client is frozen",
				&types.ClientState{
					Sequence:                 1,
					FrozenSequence:           1,
					ConsensusState:           solomachine.ConsensusState(),
					AllowUpdateAfterProposal: false,
				},
				prefix,
				proof,
				false,
			},
			{
				"proof is nil",
				solomachine.ClientState(),
				prefix,
				nil,
				false,
			},
			{
				"proof verification failed",
				solomachine.ClientState(),
				prefix,
				suite.GetInvalidProof(),
				false,
			},
		}

		for i, tc := range testCases {
			tc := tc

			expSeq := tc.clientState.Sequence + 1

			err := tc.clientState.VerifyConnectionState(
				suite.store, suite.chainA.Codec, solomachine.GetHeight(), tc.prefix, tc.proof, testConnectionID, conn,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
				suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %d: %s", suite.GetSequenceFromStore(), i, tc.name)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			}
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyChannelState() {
	counterparty := channeltypes.NewCounterparty(testPortID, testChannelID)
	ch := channeltypes.NewChannel(channeltypes.OPEN, channeltypes.ORDERED, counterparty, []string{testConnectionID}, "1.0.0")

	path := suite.solomachine.GetChannelStatePath(testPortID, testChannelID)

	// test singlesig and multisig public keys
	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {

		value, err := types.ChannelStateSignBytes(suite.chainA.Codec, solomachine.Sequence, solomachine.Time, solomachine.Diversifier, path, ch)
		suite.Require().NoError(err)

		sig := solomachine.GenerateSignature(value)
		signatureDoc := &types.TimestampedSignatureData{
			SignatureData: sig,
			Timestamp:     solomachine.Time,
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
				solomachine.ClientState(),
				prefix,
				proof,
				true,
			},
			{
				"ApplyPrefix failed",
				solomachine.ClientState(),
				nil,
				proof,
				false,
			},
			{
				"client is frozen",
				&types.ClientState{
					Sequence:                 1,
					FrozenSequence:           1,
					ConsensusState:           solomachine.ConsensusState(),
					AllowUpdateAfterProposal: false,
				},
				prefix,
				proof,
				false,
			},
			{
				"proof is nil",
				solomachine.ClientState(),
				prefix,
				nil,
				false,
			},
			{
				"proof verification failed",
				solomachine.ClientState(),
				prefix,
				suite.GetInvalidProof(),
				false,
			},
		}

		for i, tc := range testCases {
			tc := tc

			expSeq := tc.clientState.Sequence + 1

			err := tc.clientState.VerifyChannelState(
				suite.store, suite.chainA.Codec, solomachine.GetHeight(), tc.prefix, tc.proof, testPortID, testChannelID, ch,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
				suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %d: %s", suite.GetSequenceFromStore(), i, tc.name)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			}
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyPacketCommitment() {
	commitmentBytes := []byte("COMMITMENT BYTES")

	// test singlesig and multisig public keys
	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {

		path := solomachine.GetPacketCommitmentPath(testPortID, testChannelID)

		value, err := types.PacketCommitmentSignBytes(suite.chainA.Codec, solomachine.Sequence, solomachine.Time, solomachine.Diversifier, path, commitmentBytes)
		suite.Require().NoError(err)

		sig := solomachine.GenerateSignature(value)
		signatureDoc := &types.TimestampedSignatureData{
			SignatureData: sig,
			Timestamp:     solomachine.Time,
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
				solomachine.ClientState(),
				prefix,
				proof,
				true,
			},
			{
				"ApplyPrefix failed",
				solomachine.ClientState(),
				commitmenttypes.NewMerklePrefix([]byte{}),
				proof,
				false,
			},
			{
				"client is frozen",
				&types.ClientState{
					Sequence:                 1,
					FrozenSequence:           1,
					ConsensusState:           solomachine.ConsensusState(),
					AllowUpdateAfterProposal: false,
				},
				prefix,
				proof,
				false,
			},
			{
				"proof is nil",
				solomachine.ClientState(),
				prefix,
				nil,
				false,
			},
			{
				"proof verification failed",
				solomachine.ClientState(),
				prefix,
				suite.GetInvalidProof(),
				false,
			},
		}

		for i, tc := range testCases {
			tc := tc

			expSeq := tc.clientState.Sequence + 1

			err := tc.clientState.VerifyPacketCommitment(
				suite.store, suite.chainA.Codec, solomachine.GetHeight(), 0, 0, tc.prefix, tc.proof, testPortID, testChannelID, solomachine.Sequence, commitmentBytes,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
				suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %d: %s", suite.GetSequenceFromStore(), i, tc.name)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			}
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyPacketAcknowledgement() {
	ack := []byte("ACK")
	// test singlesig and multisig public keys
	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {

		path := solomachine.GetPacketAcknowledgementPath(testPortID, testChannelID)

		value, err := types.PacketAcknowledgementSignBytes(suite.chainA.Codec, solomachine.Sequence, solomachine.Time, solomachine.Diversifier, path, ack)
		suite.Require().NoError(err)

		sig := solomachine.GenerateSignature(value)
		signatureDoc := &types.TimestampedSignatureData{
			SignatureData: sig,
			Timestamp:     solomachine.Time,
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
				solomachine.ClientState(),
				prefix,
				proof,
				true,
			},
			{
				"ApplyPrefix failed",
				solomachine.ClientState(),
				commitmenttypes.NewMerklePrefix([]byte{}),
				proof,
				false,
			},
			{
				"client is frozen",
				&types.ClientState{
					Sequence:                 1,
					FrozenSequence:           1,
					ConsensusState:           solomachine.ConsensusState(),
					AllowUpdateAfterProposal: false,
				},
				prefix,
				proof,
				false,
			},
			{
				"proof is nil",
				solomachine.ClientState(),
				prefix,
				nil,
				false,
			},
			{
				"proof verification failed",
				solomachine.ClientState(),
				prefix,
				suite.GetInvalidProof(),
				false,
			},
		}

		for i, tc := range testCases {
			tc := tc

			expSeq := tc.clientState.Sequence + 1

			err := tc.clientState.VerifyPacketAcknowledgement(
				suite.store, suite.chainA.Codec, solomachine.GetHeight(), 0, 0, tc.prefix, tc.proof, testPortID, testChannelID, solomachine.Sequence, ack,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
				suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %d: %s", suite.GetSequenceFromStore(), i, tc.name)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			}
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyPacketReceiptAbsence() {
	// test singlesig and multisig public keys
	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {

		// absence uses receipt path as well
		path := solomachine.GetPacketReceiptPath(testPortID, testChannelID)

		value, err := types.PacketReceiptAbsenceSignBytes(suite.chainA.Codec, solomachine.Sequence, solomachine.Time, solomachine.Diversifier, path)
		suite.Require().NoError(err)

		sig := solomachine.GenerateSignature(value)
		signatureDoc := &types.TimestampedSignatureData{
			SignatureData: sig,
			Timestamp:     solomachine.Time,
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
				solomachine.ClientState(),
				prefix,
				proof,
				true,
			},
			{
				"ApplyPrefix failed",
				solomachine.ClientState(),
				commitmenttypes.NewMerklePrefix([]byte{}),
				proof,
				false,
			},
			{
				"client is frozen",
				&types.ClientState{
					Sequence:                 1,
					FrozenSequence:           1,
					ConsensusState:           solomachine.ConsensusState(),
					AllowUpdateAfterProposal: false,
				},
				prefix,
				proof,
				false,
			},
			{
				"proof is nil",
				solomachine.ClientState(),
				prefix,
				nil,
				false,
			},
			{
				"proof verification failed",
				solomachine.ClientState(),
				prefix,
				suite.GetInvalidProof(),
				false,
			},
		}

		for i, tc := range testCases {
			tc := tc

			expSeq := tc.clientState.Sequence + 1

			err := tc.clientState.VerifyPacketReceiptAbsence(
				suite.store, suite.chainA.Codec, solomachine.GetHeight(), 0, 0, tc.prefix, tc.proof, testPortID, testChannelID, solomachine.Sequence,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
				suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %d: %s", suite.GetSequenceFromStore(), i, tc.name)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			}
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyNextSeqRecv() {
	// test singlesig and multisig public keys
	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {

		nextSeqRecv := solomachine.Sequence + 1
		path := solomachine.GetNextSequenceRecvPath(testPortID, testChannelID)

		value, err := types.NextSequenceRecvSignBytes(suite.chainA.Codec, solomachine.Sequence, solomachine.Time, solomachine.Diversifier, path, nextSeqRecv)
		suite.Require().NoError(err)

		sig := solomachine.GenerateSignature(value)
		signatureDoc := &types.TimestampedSignatureData{
			SignatureData: sig,
			Timestamp:     solomachine.Time,
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
				solomachine.ClientState(),
				prefix,
				proof,
				true,
			},
			{
				"ApplyPrefix failed",
				solomachine.ClientState(),
				commitmenttypes.NewMerklePrefix([]byte{}),
				proof,
				false,
			},
			{
				"client is frozen",
				&types.ClientState{
					Sequence:                 1,
					FrozenSequence:           1,
					ConsensusState:           solomachine.ConsensusState(),
					AllowUpdateAfterProposal: false,
				},
				prefix,
				proof,
				false,
			},
			{
				"proof is nil",
				solomachine.ClientState(),
				prefix,
				nil,
				false,
			},
			{
				"proof verification failed",
				solomachine.ClientState(),
				prefix,
				suite.GetInvalidProof(),
				false,
			},
		}

		for i, tc := range testCases {
			tc := tc

			expSeq := tc.clientState.Sequence + 1

			err := tc.clientState.VerifyNextSequenceRecv(
				suite.store, suite.chainA.Codec, solomachine.GetHeight(), 0, 0, tc.prefix, tc.proof, testPortID, testChannelID, nextSeqRecv,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
				suite.Require().Equal(expSeq, suite.GetSequenceFromStore(), "sequence not updated in the store (%d) on valid test case %d: %s", suite.GetSequenceFromStore(), i, tc.name)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			}
		}
	}
}
