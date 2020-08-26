package types_test

import (
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
)

func (suite *SoloMachineTestSuite) TestEvidence() {
	ev := suite.solomachine.CreateEvidence()

	suite.Require().Equal(clientexported.SoloMachine, ev.ClientType())
	suite.Require().Equal(suite.solomachine.ClientID, ev.GetClientID())
	suite.Require().Equal("client", ev.Route())
	suite.Require().Equal("client_misbehaviour", ev.Type())
	suite.Require().Equal(tmbytes.HexBytes(tmhash.Sum(types.SubModuleCdc.MustMarshalBinaryBare(ev))), ev.Hash())
	suite.Require().Equal(int64(suite.solomachine.Sequence), ev.GetHeight())
}

func (suite *SoloMachineTestSuite) TestEvidenceValidateBasic() {
	testCases := []struct {
		name             string
		malleateEvidence func(ev *types.Evidence)
		expPass          bool
	}{
		{
			"valid evidence",
			func(*types.Evidence) {},
			true,
		},
		{
			"invalid client ID",
			func(ev *types.Evidence) {
				ev.ClientId = "(badclientid)"
			},
			false,
		},
		{
			"sequence is zero",
			func(ev *types.Evidence) {
				ev.Sequence = 0
			},
			false,
		},
		{
			"signature one sig is empty",
			func(ev *types.Evidence) {
				ev.SignatureOne.Signature = []byte{}
			},
			false,
		},
		{
			"signature two sig is empty",
			func(ev *types.Evidence) {
				ev.SignatureTwo.Signature = []byte{}
			},
			false,
		},
		{
			"signature one data is empty",
			func(ev *types.Evidence) {
				ev.SignatureOne.Data = nil
			},
			false,
		},
		{
			"signature two data is empty",
			func(ev *types.Evidence) {
				ev.SignatureTwo.Data = []byte{}
			},
			false,
		},
		{
			"signatures are identical",
			func(ev *types.Evidence) {
				ev.SignatureTwo.Signature = ev.SignatureOne.Signature
			},
			false,
		},
		{
			"data signed is identical",
			func(ev *types.Evidence) {
				ev.SignatureTwo.Data = ev.SignatureOne.Data
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {

			ev := suite.solomachine.CreateEvidence()
			tc.malleateEvidence(ev)

			err := ev.ValidateBasic()

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
