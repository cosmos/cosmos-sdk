package types_test

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
)

func (suite *SoloMachineTestSuite) TestConsensusState() {
	consensusState := suite.solomachine.ConsensusState()

	suite.Require().Equal(exported.SoloMachine, consensusState.ClientType())
	suite.Require().Equal(uint64(0), consensusState.GetHeight().GetEpochNumber())
	suite.Require().Equal(suite.solomachine.Sequence, consensusState.GetHeight().GetEpochHeight())
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
				Sequence:    0,
				PublicKey:   suite.solomachine.ConsensusState().PublicKey,
				Timestamp:   suite.solomachine.Time,
				Diversifier: suite.solomachine.Diversifier,
			},
			false,
		},
		{
			"timestamp is zero",
			&types.ConsensusState{
				Sequence:    suite.solomachine.Sequence,
				PublicKey:   suite.solomachine.ConsensusState().PublicKey,
				Timestamp:   0,
				Diversifier: suite.solomachine.Diversifier,
			},
			false,
		},
		{
			"diversifier is blank",
			&types.ConsensusState{
				Sequence:    suite.solomachine.Sequence,
				PublicKey:   suite.solomachine.ConsensusState().PublicKey,
				Timestamp:   suite.solomachine.Time,
				Diversifier: " ",
			},
			false,
		},
		{
			"pubkey is nil",
			&types.ConsensusState{
				Sequence:    suite.solomachine.Sequence,
				Timestamp:   suite.solomachine.Time,
				Diversifier: suite.solomachine.Diversifier,
				PublicKey:   nil,
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
