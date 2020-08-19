package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine"
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
				clientState = suite.ClientState()
				evidence = suite.Evidence()
			},
			true,
		},
		{
			"wrong client state type",
			func() {
				clientState = ibctmtypes.ClientState{}
				evidence = suite.Evidence()
			},
			false,
		},
		{
			"invalid evidence type",
			func() {
				clientState = suite.ClientState()
				evidence = ibctmtypes.Evidence{}
			},
			false,
		},
		{
			"equal data in signatures",
			func() {
				clientState = suite.ClientState()

				// store in tmp var before assigning to interface type
				ev := suite.Evidence()
				ev.SignatureOne.Data = ev.SignatureTwo.Data
				evidence = ev
			},
			false,
		},
		{
			"invalid first signature",
			func() {
				clientState = suite.ClientState()

				// store in temp before assigning to interface type
				ev := suite.Evidence()

				msg := []byte("DATA ONE")
				data := append(sdk.Uint64ToBigEndian(suite.sequence+1), msg...)
				sig, err := suite.privKey.Sign(data)
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
				clientState = suite.ClientState()

				// store in temp before assigning to interface type
				ev := suite.Evidence()

				msg := []byte("DATA TWO")
				data := append(sdk.Uint64ToBigEndian(suite.sequence+1), msg...)
				sig, err := suite.privKey.Sign(data)
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
				clientState = suite.ClientState()

				// store in temp before assigning to interface type
				ev := suite.Evidence()

				// Signature One
				msg := []byte("DATA ONE")
				// sequence used is plus 1
				data := append(sdk.Uint64ToBigEndian(suite.sequence+1), msg...)
				sig, err := suite.privKey.Sign(data)
				suite.Require().NoError(err)

				ev.SignatureOne.Signature = sig
				ev.SignatureOne.Data = msg

				// Signature Two
				msg = []byte("DATA TWO")
				// sequence used is minus 1
				data = append(sdk.Uint64ToBigEndian(suite.sequence-1), msg...)
				sig, err = suite.privKey.Sign(data)
				suite.Require().NoError(err)

				ev.SignatureTwo.Signature = sig
				ev.SignatureTwo.Data = msg

				evidence = ev

			},
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc
		// setup test
		tc.setup()

		clientState, err := solomachine.CheckMisbehaviourAndUpdateState(clientState, suite.ConsensusState(), evidence)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			suite.Require().True(clientState.IsFrozen(), "valid test case %d did not freeze the client: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			suite.Require().Nil(clientState, "invalid test case %d returned non-nil client state: %s", i, tc.name)
		}
	}
}
