package types_test

import (
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
)

func (suite *SoloMachineTestSuite) TestMisbehaviour() {
	misbehaviour := suite.solomachine.CreateMisbehaviour()

	suite.Require().Equal(clientexported.SoloMachine, misbehaviour.ClientType())
	suite.Require().Equal(suite.solomachine.ClientID, misbehaviour.GetClientID())
	suite.Require().Equal("client", misbehaviour.Route())
	suite.Require().Equal("client_misbehaviour", misbehaviour.Type())
	suite.Require().Equal(tmbytes.HexBytes(tmhash.Sum(types.SubModuleCdc.MustMarshalBinaryBare(&misbehaviour))), misbehaviour.Hash())
	suite.Require().Equal(int64(suite.solomachine.Sequence), misbehaviour.GetHeight())
}

func (suite *SoloMachineTestSuite) TestMisbehaviourValidateBasic() {
	testCases := []struct {
		name             string
		malleateMisbehaviour func(misbehaviour *types.Misbehaviour)
		expPass          bool
	}{
		{
			"valid misbehaviour",
			func(*types.Misbehaviour) {},
			true,
		},
		{
			"invalid client ID",
			func(misbehaviour *types.Misbehaviour) {
				misbehaviour.ClientId = "(badclientid)"
			},
			false,
		},
		{
			"sequence is zero",
			func(misbehaviour *types.Misbehaviour) {
				misbehaviour.Sequence = 0
			},
			false,
		},
		{
			"signature one sig is empty",
			func(misbehaviour *types.Misbehaviour) {
				misbehaviour.SignatureOne.Signature = []byte{}
			},
			false,
		},
		{
			"signature two sig is empty",
			func(misbehaviour *types.Misbehaviour) {
				misbehaviour.SignatureTwo.Signature = []byte{}
			},
			false,
		},
		{
			"signature one data is empty",
			func(misbehaviour *types.Misbehaviour) {
				misbehaviour.SignatureOne.Data = nil
			},
			false,
		},
		{
			"signature two data is empty",
			func(misbehaviour *types.Misbehaviour) {
				misbehaviour.SignatureTwo.Data = []byte{}
			},
			false,
		},
		{
			"signatures are identical",
			func(misbehaviour *types.Misbehaviour) {
				misbehaviour.SignatureTwo.Signature = misbehaviour.SignatureOne.Signature
			},
			false,
		},
		{
			"data signed is identical",
			func(misbehaviour *types.Misbehaviour) {
				misbehaviour.SignatureTwo.Data = misbehaviour.SignatureOne.Data
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {

			misbehaviour := suite.solomachine.CreateMisbehaviour()
			tc.malleateEvidence(&misbehaviour)

			err := misbehaviour.ValidateBasic()

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
