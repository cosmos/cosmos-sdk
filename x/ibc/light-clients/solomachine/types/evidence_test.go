package types_test

import (
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
)

func (suite *SoloMachineTestSuite) TestEvidence() {
	ev := suite.Evidence()

	suite.Require().Equal(clientexported.SoloMachine, ev.ClientType())
	suite.Require().Equal(suite.clientID, ev.GetClientID())
	suite.Require().Equal("client", ev.Route())
	suite.Require().Equal("client_misbehaviour", ev.Type())
	suite.Require().Equal(tmbytes.HexBytes(tmhash.Sum(solomachinetypes.SubModuleCdc.MustMarshalBinaryBare(&ev))), ev.Hash())
	suite.Require().Equal(int64(suite.sequence), ev.GetHeight())
}

func (suite *SoloMachineTestSuite) TestEvidenceValidateBasic() {
	testCases := []struct {
		name             string
		malleateEvidence func(ev *solomachinetypes.Evidence)
		expPass          bool
	}{
		{
			"valid evidence",
			func(*solomachinetypes.Evidence) {},
			true,
		},
		{
			"invalid client ID",
			func(ev *solomachinetypes.Evidence) {
				ev.ClientID = "(badclientid)"
			},
			false,
		},
		{
			"sequence is zero",
			func(ev *solomachinetypes.Evidence) {
				ev.Sequence = 0
			},
			false,
		},
		{
			"signature one sig is empty",
			func(ev *solomachinetypes.Evidence) {
				ev.SignatureOne.Signature = []byte{}
			},
			false,
		},
		{
			"signature two sig is empty",
			func(ev *solomachinetypes.Evidence) {
				ev.SignatureTwo.Signature = []byte{}
			},
			false,
		},
		{
			"signature one data is empty",
			func(ev *solomachinetypes.Evidence) {
				ev.SignatureOne.Data = nil
			},
			false,
		},
		{
			"signature two data is empty",
			func(ev *solomachinetypes.Evidence) {
				ev.SignatureTwo.Data = []byte{}
			},
			false,
		},
		{
			"signatures are identical",
			func(ev *solomachinetypes.Evidence) {
				ev.SignatureTwo.Signature = ev.SignatureOne.Signature
			},
			false,
		},
		{
			"data signed is identical",
			func(ev *solomachinetypes.Evidence) {
				ev.SignatureTwo.Data = ev.SignatureOne.Data
			},
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		ev := suite.Evidence()
		tc.malleateEvidence(&ev)

		err := ev.ValidateBasic()

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
