package types_test

import (
	"time"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

func (suite *TendermintTestSuite) TestConsensusStateValidateBasic() {
	testCases := []struct {
		msg            string
		consensusState ibctmtypes.ConsensusState
		expectPass     bool
	}{
		{"success",
			ibctmtypes.ConsensusState{
				Timestamp:          suite.now,
				Height:             height,
				Root:               commitmenttypes.NewMerkleRoot([]byte("app_hash")),
				NextValidatorsHash: suite.valsHash,
			},
			true},
		{"root is nil",
			ibctmtypes.ConsensusState{
				Timestamp:          suite.now,
				Height:             height,
				Root:               commitmenttypes.MerkleRoot{},
				NextValidatorsHash: suite.valsHash,
			},
			false},
		{"root is empty",
			ibctmtypes.ConsensusState{
				Timestamp:          suite.now,
				Height:             height,
				Root:               commitmenttypes.MerkleRoot{},
				NextValidatorsHash: suite.valsHash,
			},
			false},
		{"nextvalshash is invalid",
			ibctmtypes.ConsensusState{
				Timestamp:          suite.now,
				Height:             height,
				Root:               commitmenttypes.NewMerkleRoot([]byte("app_hash")),
				NextValidatorsHash: []byte("hi"),
			},
			false},

		{"height is 0",
			ibctmtypes.ConsensusState{
				Timestamp:          suite.now,
				Height:             0,
				Root:               commitmenttypes.NewMerkleRoot([]byte("app_hash")),
				NextValidatorsHash: suite.valsHash,
			},
			false},
		{"timestamp is zero",
			ibctmtypes.ConsensusState{
				Timestamp:          time.Time{},
				Height:             height,
				Root:               commitmenttypes.NewMerkleRoot([]byte("app_hash")),
				NextValidatorsHash: suite.valsHash,
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
