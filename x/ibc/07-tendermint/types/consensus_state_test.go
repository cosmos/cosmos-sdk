package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func TestConsensusStateValidateBasic(t *testing.T) {
	testCases := []struct {
		msg            string
		consensusState tendermint.ConsensusState
		expectPass     bool
	}{
		{"success",
			tendermint.ConsensusState{
				Root:             commitment.NewRoot([]byte("app_hash")),
				ValidatorSetHash: []byte("valset_hash"),
			},
			true},
		{"root is nil",
			tendermint.ConsensusState{
				Root:             nil,
				ValidatorSetHash: []byte("valset_hash"),
			},
			false},
		{"root is empty",
			tendermint.ConsensusState{
				Root:             commitment.Root{},
				ValidatorSetHash: []byte("valset_hash"),
			},
			false},
		{"invalid client type",
			tendermint.ConsensusState{
				Root:             commitment.NewRoot([]byte("app_hash")),
				ValidatorSetHash: []byte{},
			},
			false},
	}

	for i, tc := range testCases {
		tc := tc

		require.Equal(t, tc.consensusState.ClientType(), clientexported.Tendermint)
		require.Equal(t, tc.consensusState.GetRoot(), tc.consensusState.Root)

		if tc.expectPass {
			require.NoError(t, tc.consensusState.ValidateBasic(), "valid test case %d failed: %s", i, tc.msg)
		} else {
			require.Error(t, tc.consensusState.ValidateBasic(), "invalid test case %d passed: %s", i, tc.msg)
		}
	}
}
