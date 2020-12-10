package types_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/06-solomachine/types"
	solomachinetypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/06-solomachine/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

func (suite *SoloMachineTestSuite) TestVerifySignature() {
	cdc := suite.chainA.App.AppCodec()
	signBytes := []byte("sign bytes")

	singleSignature := suite.solomachine.GenerateSignature(signBytes)
	singleSigData, err := solomachinetypes.UnmarshalSignatureData(cdc, singleSignature)
	suite.Require().NoError(err)

	multiSignature := suite.solomachineMulti.GenerateSignature(signBytes)
	multiSigData, err := solomachinetypes.UnmarshalSignatureData(cdc, multiSignature)
	suite.Require().NoError(err)

	testCases := []struct {
		name      string
		publicKey cryptotypes.PubKey
		sigData   signing.SignatureData
		expPass   bool
	}{
		{
			"single signature with regular public key",
			suite.solomachine.PublicKey,
			singleSigData,
			true,
		},
		{
			"multi signature with multisig public key",
			suite.solomachineMulti.PublicKey,
			multiSigData,
			true,
		},
		{
			"single signature with multisig public key",
			suite.solomachineMulti.PublicKey,
			singleSigData,
			false,
		},
		{
			"multi signature with regular public key",
			suite.solomachine.PublicKey,
			multiSigData,
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			err := solomachinetypes.VerifySignature(tc.publicKey, signBytes, tc.sigData)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *SoloMachineTestSuite) TestClientStateSignBytes() {
	cdc := suite.chainA.App.AppCodec()

	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {
		// success
		path := solomachine.GetClientStatePath(counterpartyClientIdentifier)
		bz, err := types.ClientStateSignBytes(cdc, solomachine.Sequence, solomachine.Time, solomachine.Diversifier, path, solomachine.ClientState())
		suite.Require().NoError(err)
		suite.Require().NotNil(bz)

		// nil client state
		bz, err = types.ClientStateSignBytes(cdc, solomachine.Sequence, solomachine.Time, solomachine.Diversifier, path, nil)
		suite.Require().Error(err)
		suite.Require().Nil(bz)
	}
}

func (suite *SoloMachineTestSuite) TestConsensusStateSignBytes() {
	cdc := suite.chainA.App.AppCodec()

	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {
		// success
		path := solomachine.GetConsensusStatePath(counterpartyClientIdentifier, consensusHeight)
		bz, err := types.ConsensusStateSignBytes(cdc, solomachine.Sequence, solomachine.Time, solomachine.Diversifier, path, solomachine.ConsensusState())
		suite.Require().NoError(err)
		suite.Require().NotNil(bz)

		// nil consensus state
		bz, err = types.ConsensusStateSignBytes(cdc, solomachine.Sequence, solomachine.Time, solomachine.Diversifier, path, nil)
		suite.Require().Error(err)
		suite.Require().Nil(bz)
	}
}
