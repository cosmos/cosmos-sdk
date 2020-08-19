package solomachine_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	solomachine "github.com/cosmos/cosmos-sdk/x/ibc/06-solomachine"
	solomachinetypes "github.com/cosmos/cosmos-sdk/x/ibc/06-solomachine/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

func (suite *SoloMachineTestSuite) TestCheckValidity() {
	var (
		clientState clientexported.ClientState
		header      clientexported.Header
	)

	testCases := []struct {
		name    string
		setup   func()
		expPass bool
	}{
		{
			"successful update",
			func() {
				clientState = suite.ClientState()
				header = suite.CreateHeader()
			},
			true,
		},
		{
			"wrong client state type",
			func() {
				clientState = ibctmtypes.ClientState{}
				header = suite.CreateHeader()
			},
			false,
		},
		{
			"invalid header type",
			func() {
				clientState = suite.ClientState()
				header = ibctmtypes.Header{}
			},
			false,
		},
		{
			"wrong sequence in header",
			func() {
				clientState = suite.ClientState()
				// store in temp before assigning to interface type
				h := suite.CreateHeader()
				h.Sequence++
				header = h
			},
			false,
		},
		{
			"signature uses wrong sequence",
			func() {
				clientState = suite.ClientState()
				suite.sequence++
				header = suite.CreateHeader()
			},
			false,
		},
		{
			"signature uses new pubkey to sign",
			func() {
				// store in temp before assinging to interface type
				cs := suite.ClientState()
				h := suite.CreateHeader()

				// generate invalid signature
				data := append(sdk.Uint64ToBigEndian(cs.ConsensusState.Sequence), h.NewPubKey.Bytes()...)
				sig, err := suite.privKey.Sign(data)
				suite.Require().NoError(err)
				h.Signature = sig

				clientState = cs
				header = h

			},
			false,
		},
		{
			"signature signs over old pubkey",
			func() {
				// store in temp before assinging to interface type
				cs := suite.ClientState()
				oldPrivKey := suite.privKey
				h := suite.CreateHeader()

				// generate invalid signature
				data := append(sdk.Uint64ToBigEndian(cs.ConsensusState.Sequence), cs.ConsensusState.PubKey.Bytes()...)
				sig, err := oldPrivKey.Sign(data)
				suite.Require().NoError(err)
				h.Signature = sig

				clientState = cs
				header = h
			},
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc
		// setup test
		tc.setup()

		clientState, consensusState, err := solomachine.CheckValidityAndUpdateState(clientState, header)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			suite.Require().Equal(header.(solomachinetypes.Header).NewPubKey, clientState.(solomachinetypes.ClientState).ConsensusState.PubKey, "valid test case %d failed with wrong updated pubkey: %s", i, tc.name)
			suite.Require().False(clientState.(solomachinetypes.ClientState).Frozen, "valid test case %d failed with frozen client: %s", i, tc.name)
			suite.Require().Equal(header.(solomachinetypes.Header).Sequence+1, clientState.(solomachinetypes.ClientState).ConsensusState.Sequence, "valid test case %d failed with wrong updated sequence: %s", i, tc.name)
			suite.Require().Equal(consensusState, clientState.(solomachinetypes.ClientState).ConsensusState, "valid test case %d failed with non-matching consensus state relative to client state: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			suite.Require().Nil(clientState, "invalid test case %d passed: %s", i, tc.name)
			suite.Require().Nil(consensusState, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
