package types_test

import (
	"time"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func (suite *TendermintTestSuite) TestConsensusStateValidateBasic() {
	testCases := []struct {
		msg            string
		consensusState ibctmtypes.ConsensusState
		expectPass     bool
	}{
		{"success",
			ibctmtypes.ConsensusState{
				Timestamp:    suite.now,
				Root:         commitment.NewRoot([]byte("app_hash")),
				ValidatorSet: suite.valSet,
			},
			true},
		{"root is nil",
			ibctmtypes.ConsensusState{
				Timestamp:    suite.now,
				Root:         nil,
				ValidatorSet: suite.valSet,
			},
			false},
		{"root is empty",
			ibctmtypes.ConsensusState{
				Timestamp:    suite.now,
				Root:         commitment.Root{},
				ValidatorSet: suite.valSet,
			},
			false},
		{"valset is nil",
			ibctmtypes.ConsensusState{
				Timestamp:    suite.now,
				Root:         commitment.NewRoot([]byte("app_hash")),
				ValidatorSet: nil,
			},
			false},
		{"timestamp is zero",
			ibctmtypes.ConsensusState{
				Timestamp:    time.Time{},
				Root:         commitment.NewRoot([]byte("app_hash")),
				ValidatorSet: suite.valSet,
			},
			false},
	}

	for i, tc := range testCases {
		tc := tc

		suite.Require().Equal(tc.consensusState.ClientType(), clientexported.Tendermint)
		suite.Require().Equal(tc.consensusState.GetRoot(), tc.consensusState.Root)

		if tc.expectPass {
			suite.Require().NoError(tc.consensusState.ValidateBasic(), "valid test case %d failed: %s", i, tc.msg)
		} else {
			suite.Require().Error(tc.consensusState.ValidateBasic(), "invalid test case %d passed: %s", i, tc.msg)
		}
	}
}
