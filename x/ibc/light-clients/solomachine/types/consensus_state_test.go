package types_test

import (
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
)

func (suite *SoloMachineTestSuite) TestConsensusState() {
	consensusState := suite.ConsensusState()

	suite.Require().Equal(clientexported.SoloMachine, consensusState.ClientType())
	suite.Require().Equal(suite.sequence, consensusState.GetHeight())
	suite.Require().Equal(timestamp, consensusState.GetTimestamp())
	suite.Require().Nil(consensusState.GetRoot())
}

func (suite *SoloMachineTestSuite) TestConsensusStateValidateBasic() {
	testCases := []struct {
		name           string
		consensusState *solomachinetypes.ConsensusState
		expPass        bool
	}{
		{
			"valid consensus state",
			suite.ConsensusState(),
			true,
		},
		{
			"sequence is zero",
			&solomachinetypes.ConsensusState{
				Sequence: 0,
				PubKey:   suite.pubKey,
			},
			false,
		},
		{
			"pubkey is nil",
			&solomachinetypes.ConsensusState{
				Sequence: suite.sequence,
				PubKey:   nil,
			},
			false,
		},
	}

	for i, tc := range testCases {
		err := tc.consensusState.ValidateBasic()

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
