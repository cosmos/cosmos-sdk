package types_test

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/06-solomachine/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

func (suite *SoloMachineTestSuite) TestCheckProposedHeaderAndUpdateState() {
	var header exported.Header

	// test singlesig and multisig public keys
	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {

		testCases := []struct {
			name     string
			malleate func()
			expPass  bool
		}{
			{
				"valid header", func() {
					header = solomachine.CreateHeader()
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
						NewPublicKey: solomachine.ConsensusState().PublicKey,
					}
				}, false,
			},
		}

		for _, tc := range testCases {
			tc := tc

			suite.Run(tc.name, func() {
				suite.SetupTest()

				clientState := solomachine.ClientState()

				tc.malleate()

				clientStore := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), solomachine.ClientID)

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

					smConsState, ok := consState.(*types.ConsensusState)
					suite.Require().True(ok)
					smHeader, ok := header.(*types.Header)
					suite.Require().True(ok)

					suite.Require().Equal(cs.(*types.ClientState).ConsensusState, consState)

					headerPubKey, err := smHeader.GetPubKey()
					suite.Require().NoError(err)

					consStatePubKey, err := smConsState.GetPubKey()
					suite.Require().NoError(err)

					suite.Require().Equal(headerPubKey, consStatePubKey)
					suite.Require().Equal(smHeader.NewDiversifier, smConsState.Diversifier)
					suite.Require().Equal(smHeader.Timestamp, smConsState.Timestamp)
					suite.Require().Equal(smHeader.GetHeight().GetRevisionHeight(), cs.(*types.ClientState).Sequence)
				} else {
					suite.Require().Error(err)
					suite.Require().Nil(cs)
					suite.Require().Nil(consState)
				}
			})
		}
	}
}
