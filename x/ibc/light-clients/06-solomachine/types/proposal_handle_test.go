package types_test

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/06-solomachine/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

func (suite *SoloMachineTestSuite) TestCheckSubstituteAndUpdateState() {
	var (
		subjectClientState    *types.ClientState
		substituteClientState exported.ClientState
	)

	// test singlesig and multisig public keys
	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {

		testCases := []struct {
			name     string
			malleate func()
			expPass  bool
		}{
			{
				"valid substitute", func() {
					subjectClientState.AllowUpdateAfterProposal = true
				}, true,
			},
			{
				"subject not allowed to be updated", func() {
					subjectClientState.AllowUpdateAfterProposal = false
				}, false,
			},
			{
				"substitute is not the solo machine", func() {
					substituteClientState = &ibctmtypes.ClientState{}
				}, false,
			},
			{
				"subject public key is nil", func() {
					subjectClientState.ConsensusState.PublicKey = nil
				}, false,
			},

			{
				"substitute public key is nil", func() {
					substituteClientState.(*types.ClientState).ConsensusState.PublicKey = nil
				}, false,
			},
			{
				"subject and substitute use the same public key", func() {
					substituteClientState.(*types.ClientState).ConsensusState.PublicKey = subjectClientState.ConsensusState.PublicKey
				}, false,
			},
		}

		for _, tc := range testCases {
			tc := tc

			suite.Run(tc.name, func() {
				suite.SetupTest()

				subjectClientState = solomachine.ClientState()
				subjectClientState.AllowUpdateAfterProposal = true
				substitute := ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "substitute", "testing", 5)
				substituteClientState = substitute.ClientState()

				tc.malleate()

				subjectClientStore := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), solomachine.ClientID)
				substituteClientStore := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), substitute.ClientID)

				updatedClient, err := subjectClientState.CheckSubstituteAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), subjectClientStore, substituteClientStore, substituteClientState, nil)

				if tc.expPass {
					suite.Require().NoError(err)

					suite.Require().Equal(substituteClientState.(*types.ClientState).ConsensusState, updatedClient.(*types.ClientState).ConsensusState)
					suite.Require().Equal(substituteClientState.(*types.ClientState).Sequence, updatedClient.(*types.ClientState).Sequence)
					suite.Require().Equal(uint64(0), updatedClient.(*types.ClientState).FrozenSequence)
				} else {
					suite.Require().Error(err)
					suite.Require().Nil(updatedClient)
				}
			})
		}
	}
}
