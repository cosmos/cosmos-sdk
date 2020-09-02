package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

func (suite *SoloMachineTestSuite) TestCheckMisbehaviourAndUpdateState() {
	var (
		clientState  exported.ClientState
		misbehaviour exported.Misbehaviour
	)

	testCases := []struct {
		name    string
		setup   func()
		expPass bool
	}{
		{
			"valid misbehaviour",
			func() {
				clientState = suite.solomachine.ClientState()
				misbehaviour = suite.solomachine.CreateMisbehaviour()
			},
			true,
		},
		{
			"client is frozen",
			func() {
				cs := suite.solomachine.ClientState()
				cs.FrozenSequence = 1
				clientState = cs
				misbehaviour = suite.solomachine.CreateMisbehaviour()
			},
			false,
		},
		{
			"wrong client state type",
			func() {
				clientState = ibctmtypes.ClientState{}
				misbehaviour = suite.solomachine.CreateMisbehaviour()
			},
			false,
		},
		{
			"invalid misbehaviour type",
			func() {
				clientState = suite.solomachine.ClientState()
				misbehaviour = ibctmtypes.Misbehaviour{}
			},
			false,
		},
		{
			"invalid first signature",
			func() {
				clientState = suite.solomachine.ClientState()

				// store in temp before assigning to interface type
				m := suite.solomachine.CreateMisbehaviour()

				msg := []byte("DATA ONE")
				data := append(sdk.Uint64ToBigEndian(suite.solomachine.Sequence+1), msg...)
				sig, err := suite.solomachine.PrivateKey.Sign(data)
				suite.Require().NoError(err)

				m.SignatureOne.Signature = sig
				m.SignatureOne.Data = msg
				misbehaviour = m
			},
			false,
		},
		{
			"invalid second signature",
			func() {
				clientState = suite.solomachine.ClientState()

				// store in temp before assigning to interface type
				m := suite.solomachine.CreateMisbehaviour()

				msg := []byte("DATA TWO")
				data := append(sdk.Uint64ToBigEndian(suite.solomachine.Sequence+1), msg...)
				sig, err := suite.solomachine.PrivateKey.Sign(data)
				suite.Require().NoError(err)

				m.SignatureTwo.Signature = sig
				m.SignatureTwo.Data = msg
				misbehaviour = m
			},
			false,
		},
		{
			"signatures sign over different sequence",
			func() {
				clientState = suite.solomachine.ClientState()

				// store in temp before assigning to interface type
				m := suite.solomachine.CreateMisbehaviour()

				// Signature One
				msg := []byte("DATA ONE")
				// sequence used is plus 1
				data := append(sdk.Uint64ToBigEndian(suite.solomachine.Sequence+1), msg...)
				sig, err := suite.solomachine.PrivateKey.Sign(data)
				suite.Require().NoError(err)

				m.SignatureOne.Signature = sig
				m.SignatureOne.Data = msg

				// Signature Two
				msg = []byte("DATA TWO")
				// sequence used is minus 1
				data = append(sdk.Uint64ToBigEndian(suite.solomachine.Sequence-1), msg...)
				sig, err = suite.solomachine.PrivateKey.Sign(data)
				suite.Require().NoError(err)

				m.SignatureTwo.Signature = sig
				m.SignatureTwo.Data = msg

				misbehaviour = m

			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			// setup test
			tc.setup()

			clientState, err := clientState.CheckMisbehaviourAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), suite.store, misbehaviour)

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
