package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

func (suite *SoloMachineTestSuite) TestCheckMisbehaviourAndUpdateState() {
	var (
		clientState clientexported.ClientState
		evidence    clientexported.Misbehaviour
	)

	testCases := []struct {
		name    string
		setup   func()
		expPass bool
	}{
		{
			"valid misbehaviour evidence",
			func() {
				clientState = suite.solomachine.ClientState()
				evidence = suite.solomachine.CreateEvidence()
			},
			true,
		},
		{
			"client is frozen",
			func() {
				cs := suite.solomachine.ClientState()
				cs.FrozenHeight = 1
				clientState = cs
				evidence = suite.solomachine.CreateEvidence()
			},
			false,
		},
		{
			"wrong client state type",
			func() {
				clientState = ibctmtypes.ClientState{}
				evidence = suite.solomachine.CreateEvidence()
			},
			false,
		},
		{
			"invalid evidence type",
			func() {
				clientState = suite.solomachine.ClientState()
				evidence = ibctmtypes.Evidence{}
			},
			false,
		},
		{
			"equal data in signatures",
			func() {
				clientState = suite.solomachine.ClientState()

				// store in tmp var before assigning to interface type
				ev := suite.solomachine.CreateEvidence()
				ev.SignatureOne.Data = ev.SignatureTwo.Data
				evidence = ev
			},
			false,
		},
		{
			"invalid first signature",
			func() {
				clientState = suite.solomachine.ClientState()

				// store in temp before assigning to interface type
				ev := suite.solomachine.CreateEvidence()

				msg := []byte("DATA ONE")
				data := append(sdk.Uint64ToBigEndian(suite.solomachine.Sequence+1), msg...)
				sig, err := suite.solomachine.PrivateKey.Sign(data)
				suite.Require().NoError(err)

				ev.SignatureOne.Signature = sig
				ev.SignatureOne.Data = msg
				evidence = ev
			},
			false,
		},
		{
			"invalid second signature",
			func() {
				clientState = suite.solomachine.ClientState()

				// store in temp before assigning to interface type
				ev := suite.solomachine.CreateEvidence()

				msg := []byte("DATA TWO")
				data := append(sdk.Uint64ToBigEndian(suite.solomachine.Sequence+1), msg...)
				sig, err := suite.solomachine.PrivateKey.Sign(data)
				suite.Require().NoError(err)

				ev.SignatureTwo.Signature = sig
				ev.SignatureTwo.Data = msg
				evidence = ev
			},
			false,
		},
		{
			"signatures sign over different sequence",
			func() {
				clientState = suite.solomachine.ClientState()

				// store in temp before assigning to interface type
				ev := suite.solomachine.CreateEvidence()

				// Signature One
				msg := []byte("DATA ONE")
				// sequence used is plus 1
				data := append(sdk.Uint64ToBigEndian(suite.solomachine.Sequence+1), msg...)
				sig, err := suite.solomachine.PrivateKey.Sign(data)
				suite.Require().NoError(err)

				ev.SignatureOne.Signature = sig
				ev.SignatureOne.Data = msg

				// Signature Two
				msg = []byte("DATA TWO")
				// sequence used is minus 1
				data = append(sdk.Uint64ToBigEndian(suite.solomachine.Sequence-1), msg...)
				sig, err = suite.solomachine.PrivateKey.Sign(data)
				suite.Require().NoError(err)

				ev.SignatureTwo.Signature = sig
				ev.SignatureTwo.Data = msg

				evidence = ev

			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			// setup test
			tc.setup()

			clientState, err := clientState.CheckMisbehaviourAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), suite.store, evidence)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().True(clientState.IsFrozen(), "client not frozen")
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(clientState)
			}
		})
	}
}
