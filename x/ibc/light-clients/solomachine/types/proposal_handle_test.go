package types_test

import (
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
)

func (suite *SoloMachineTestSuite) TestCheckProposedHeaderAndUpdateState() {
	var header exported.Header

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"valid header", func() {
				header = suite.solomachine.CreateHeader()
			}, true,
		},
		{
			"nil header", func() {
				header = &ibctmtypes.Header{}
			}, false,
		},
		{
			"header does not update public key", func() {
				header = &types.Header{
					Sequence:     1,
					NewPublicKey: suite.solomachine.ConsensusState().PublicKey,
				}
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			clientState := suite.solomachine.ClientState()

			tc.malleate()

			clientStore := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), suite.solomachine.ClientID)

			// all cases should always fail if the client has 'AllowUpdateAfterProposal' set to false
			clientState.AllowUpdateAfterProposal = false
			cs, consState, err := clientState.CheckProposedHeaderAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), clientStore, header)
			suite.Require().Error(err)
			suite.Require().Nil(cs)
			suite.Require().Nil(consState)

			clientState.AllowUpdateAfterProposal = true
			cs, consState, err = clientState.CheckProposedHeaderAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), clientStore, header)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(header.(*types.Header).GetPubKey(), consState.(*types.ConsensusState).GetPubKey())
				suite.Require().Equal(cs.(*types.ClientState).ConsensusState, consState)
				suite.Require().Equal(header.GetHeight().GetEpochHeight(), cs.(*types.ClientState).Sequence)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(cs)
				suite.Require().Nil(consState)
			}
		})
	}

}
