package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	solomachinetypes "github.com/cosmos/cosmos-sdk/x/ibc/06-solomachine/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

const (
	counterpartyClientIdentifier = "chainA"
	consensusHeight              = uint64(0)
	testConnectionID             = "connectionid"
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
	proof := commitmenttypes.SignatureProof{sig}

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
	counterparty := connection.NewCounterparty("clientB", testConnectionID, prefix)
	conn := connection.NewConnectionEnd(connection.OPEN, testConnectionID, "clientA", counterparty, []string{"1.0.0"})

	path, err := commitmenttypes.ApplyPrefix(prefix, host.ConnectionPath(testConnectionID))
	suite.Require().NoError(err)

	value := append(sdk.Uint64ToBigEndian(suite.sequence), []byte(path.String())...)
	bz, err := suite.cdc.MarshalBinaryBare(conn)
	suite.Require().NoError(err)
	value = append(value, bz...)

	sig, err := suite.privKey.Sign(value)
	suite.Require().NoError(err)
	proof := commitmenttypes.SignatureProof{sig}

	testCases := []struct {
		name        string
		clientState solomachinetypes.ClientState
		connection  connection.End
		prefix      commitmentexported.Prefix
		proof       commitmentexported.Proof
		expPass     bool
	}{
		{
			"successful verification",
			suite.ClientState(),
			conn,
			prefix,
			proof,
			true,
		},
		{
			"ApplyPrefix failed",
			suite.ClientState(),
			conn,
			commitmenttypes.NewSignaturePrefix([]byte{}),
			proof,
			false,
		},
		{
			"client is frozen",
			solomachinetypes.ClientState{suite.clientID, true, suite.ConsensusState()},
			conn,
			prefix,
			proof,
			false,
		},
		{
			"invalid proof type",
			suite.ClientState(),
			conn,
			prefix,
			commitmenttypes.MerkleProof{},
			false,
		},
		{
			"proof verification failed",
			suite.ClientState(),
			conn,
			prefix,
			commitmenttypes.SignatureProof{},
			false,
		},
	}

	for i, tc := range testCases {
		err := tc.clientState.VerifyClientConnectionState(
			suite.store, suite.cdc, 0, tc.prefix, tc.proof, testConnectionID, tc.conn, suite.ConsensusState(),
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

/*
func (suite *SoloMachineTestSuite) TestVerifyChannelState() {
	testCases := []struct {
		name        string
		clientState solomachinetypes.ClientState
		connection  connection.End
		prefix      commitmentexported.Prefix
		proof       commitmentexported.Proof
		expPass     bool
	}{
		{
			"successful verification",
		},
		{
			"ApplyPrefix failed",
		},
		{
			"client is frozen",
		},
		{
			"invalid proof type",
		},
		{
			"proof verification failed",
		},
	}

	for i, tc := range testCases {
		err := tc.clientState.VerifyClientConsensusState(
			suite.store, suite.aminoCdc, 0, tc.prefix, tc.proof, nil,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			suite.Require().Equal(tc.clientState.ConsensusState.Sequence, suite.GetSequenceFromStore(), "valid test case %d passed but the sequence (%d) was not updated in the store (%d) : %s", i, tc.clientState.ConsensusState.Sequence, suite.GetSequenceFromStore(), tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyPacketCommitment() {
	testCases := []struct {
		name        string
		clientState solomachinetypes.ClientState
		connection  connection.End
		prefix      commitmentexported.Prefix
		proof       commitmentexported.Proof
		expPass     bool
	}{
		{
			"successful verification",
		},
		{
			"ApplyPrefix failed",
		},
		{
			"latest client height < height",
		},
		{
			"client is frozen",
		},
		{
			"invalid proof type",
		},
		{
			"proof verification failed",
		},
	}

	for i, tc := range testCases {
		err := tc.clientState.VerifyClientConsensusState(
			suite.store, suite.aminoCdc, 0, tc.prefix, tc.proof, nil,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			suite.Require().Equal(tc.clientState.ConsensusState.Sequence, suite.GetSequenceFromStore(), "valid test case %d passed but the sequence (%d) was not updated in the store (%d) : %s", i, tc.clientState.ConsensusState.Sequence, suite.GetSequenceFromStore(), tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyPacketAcknowledgement() {
	testCases := []struct {
		name        string
		clientState solomachinetypes.ClientState
		connection  connection.End
		prefix      commitmentexported.Prefix
		proof       commitmentexported.Proof
		expPass     bool
	}{
		{
			"successful verification",
		},
		{
			"ApplyPrefix failed",
		},
		{
			"latest client height < height",
		},
		{
			"client is frozen",
		},
		{
			"invalid proof type",
		},
		{
			"proof verification failed",
		},
	}

	for i, tc := range testCases {
		err := tc.clientState.VerifyClientConsensusState(
			suite.store, suite.aminoCdc, 0, tc.prefix, tc.proof, nil,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			suite.Require().Equal(tc.clientState.ConsensusState.Sequence, suite.GetSequenceFromStore(), "valid test case %d passed but the sequence (%d) was not updated in the store (%d) : %s", i, tc.clientState.ConsensusState.Sequence, suite.GetSequenceFromStore(), tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyPacketAcknowledgementAbsence() {
	testCases := []struct {
		name        string
		clientState solomachinetypes.ClientState
		connection  connection.End
		prefix      commitmentexported.Prefix
		proof       commitmentexported.Proof
		expPass     bool
	}{
		{
			"successful verification",
		},
		{
			"ApplyPrefix failed",
		},
		{
			"latest client height < height",
		},
		{
			"client is frozen",
		},
		{
			"invalid proof type",
		},
		{
			"proof verification failed",
		},
	}

	for i, tc := range testCases {
		err := tc.clientState.VerifyClientConsensusState(
			suite.store, suite.aminoCdc, 0, tc.prefix, tc.proof, nil,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			suite.Require().Equal(tc.clientState.ConsensusState.Sequence, suite.GetSequenceFromStore(), "valid test case %d passed but the sequence (%d) was not updated in the store (%d) : %s", i, tc.clientState.ConsensusState.Sequence, suite.GetSequenceFromStore(), tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *SoloMachineTestSuite) TestVerifyNextSeqRecv() {
	testCases := []struct {
		name        string
		clientState solomachinetypes.ClientState
		connection  connection.End
		prefix      commitmentexported.Prefix
		proof       commitmentexported.Proof
		expPass     bool
	}{
		{
			"successful verification",
		},
		{
			"ApplyPrefix failed",
		},
		{
			"latest client height < height",
		},
		{
			"client is frozen",
		},
		{
			"invalid proof type",
		},
		{
			"proof verification failed",
		},
	}

	for i, tc := range testCases {
		err := tc.clientState.VerifyClientConsensusState(
			suite.store, suite.aminoCdc, 0, tc.prefix, tc.proof, nil,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			suite.Require().Equal(tc.clientState.ConsensusState.Sequence, suite.GetSequenceFromStore(), "valid test case %d passed but the sequence (%d) was not updated in the store (%d) : %s", i, tc.clientState.ConsensusState.Sequence, suite.GetSequenceFromStore(), tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
*/
