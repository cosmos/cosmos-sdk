package types_test

import (
	"time"

	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
)

func (suite *TendermintTestSuite) TestConsensusStateValidateBasic() {
	testCases := []struct {
		msg            string
		consensusState *types.ConsensusState
		expectPass     bool
	}{
		{"success",
			&types.ConsensusState{
				Timestamp:          suite.now,
				Root:               commitmenttypes.NewMerkleRoot([]byte("app_hash")),
				NextValidatorsHash: suite.valsHash,
			},
			true},
		{"root is nil",
			&types.ConsensusState{
				Timestamp:          suite.now,
				Root:               commitmenttypes.MerkleRoot{},
				NextValidatorsHash: suite.valsHash,
			},
			false},
		{"root is empty",
			&types.ConsensusState{
				Timestamp:          suite.now,
				Root:               commitmenttypes.MerkleRoot{},
				NextValidatorsHash: suite.valsHash,
			},
			false},
		{"nextvalshash is invalid",
			&types.ConsensusState{
				Timestamp:          suite.now,
				Root:               commitmenttypes.NewMerkleRoot([]byte("app_hash")),
				NextValidatorsHash: []byte("hi"),
			},
			false},

		{"timestamp is zero",
			&types.ConsensusState{
				Timestamp:          time.Time{},
				Root:               commitmenttypes.NewMerkleRoot([]byte("app_hash")),
				NextValidatorsHash: suite.valsHash,
			},
			false},
	}

	for i, tc := range testCases {
		tc := tc

		suite.Require().Equal(tc.consensusState.ClientType(), types.Tendermint)
		suite.Require().Equal(tc.consensusState.GetRoot(), tc.consensusState.Root)

		if tc.expectPass {
			suite.Require().NoError(tc.consensusState.ValidateBasic(), "valid test case %d failed: %s", i, tc.msg)
		} else {
			suite.Require().Error(tc.consensusState.ValidateBasic(), "invalid test case %d passed: %s", i, tc.msg)
		}
	}
}
