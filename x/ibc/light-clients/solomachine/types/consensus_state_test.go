package types_test

import (
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
)

func (suite *SoloMachineTestSuite) TestConsensusState() {
	consensusState := suite.solomachine.ConsensusState()

	suite.Require().Equal(clientexported.SoloMachine, consensusState.ClientType())
	suite.Require().Equal(suite.solomachine.Sequence, consensusState.GetHeight())
	suite.Require().Equal(suite.solomachine.Time, consensusState.GetTimestamp())
	suite.Require().Nil(consensusState.GetRoot())
}

func (suite *SoloMachineTestSuite) TestConsensusStateValidateBasic() {
	testCases := []struct {
		name           string
		consensusState *types.ConsensusState
		expPass        bool
	}{
		{
			"valid consensus state",
			suite.solomachine.ConsensusState(),
			true,
		},
		{
			"sequence is zero",
			&types.ConsensusState{
				Sequence:  0,
				PublicKey: suite.solomachine.ConsensusState().PublicKey,
			},
			false,
		},
		{
			"pubkey is nil",
			&types.ConsensusState{
				Sequence:  suite.solomachine.Sequence,
				PublicKey: nil,
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {

			err := tc.consensusState.ValidateBasic()

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
