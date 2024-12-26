package cometbft

import (
	"context"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/comet"
	corecontext "cosmossdk.io/core/context"
)

func TestContextWithCometInfo(t *testing.T) {
	info := comet.Info{
		Evidence: []comet.Evidence{
			{
				Type:             comet.MisbehaviorType(1),
				Height:           100,
				Time:             time.Now(),
				TotalVotingPower: 1000,
				Validator: comet.Validator{
					Address: []byte("validator1"),
					Power:   500,
				},
			},
		},
		ValidatorsHash:  []byte("validatorshash123"),
		ProposerAddress: []byte("proposer456"),
		LastCommit: comet.CommitInfo{
			Round: 1,
			Votes: []comet.VoteInfo{},
		},
	}
	ctx := context.Background()
	newCtx := contextWithCometInfo(ctx, info)

	retrievedInfo, ok := newCtx.Value(corecontext.CometInfoKey).(comet.Info)
	require.True(t, ok)
	require.Equal(t, info, retrievedInfo)
}

func TestToCoreEvidence(t *testing.T) {
	timestamp := time.Now()
	testCases := []struct {
		name     string
		input    []abci.Misbehavior
		expected []comet.Evidence
	}{
		{
			name:     "empty evidence",
			input:    []abci.Misbehavior{},
			expected: []comet.Evidence{},
		},
		{
			name: "single evidence",
			input: []abci.Misbehavior{
				{
					Type:             1,
					Height:           100,
					Time:             timestamp,
					TotalVotingPower: 1000,
					Validator: abci.Validator{
						Address: []byte("address"),
						Power:   500,
					},
				},
			},
			expected: []comet.Evidence{
				{
					Type:             comet.MisbehaviorType(1),
					Height:           100,
					Time:             timestamp,
					TotalVotingPower: 1000,
					Validator: comet.Validator{
						Address: []byte("address"),
						Power:   500,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := toCoreEvidence(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestToCoreCommitInfo(t *testing.T) {
	input := abci.CommitInfo{
		Round: 1,
		Votes: []abci.VoteInfo{
			{
				Validator: abci.Validator{
					Address: []byte("validator1"),
					Power:   100,
				},
				BlockIdFlag: 2,
			},
		},
	}

	expected := comet.CommitInfo{
		Round: 1,
		Votes: []comet.VoteInfo{
			{
				Validator: comet.Validator{
					Address: []byte("validator1"),
					Power:   100,
				},
				BlockIDFlag: 2,
			},
		},
	}

	result := toCoreCommitInfo(input)
	require.Equal(t, expected, result)
}

func TestToCoreExtendedCommitInfo(t *testing.T) {
	input := abci.ExtendedCommitInfo{
		Round: 1,
		Votes: []abci.ExtendedVoteInfo{
			{
				Validator: abci.Validator{
					Address: []byte("validator1"),
					Power:   100,
				},
				BlockIdFlag: 2,
			},
		},
	}

	expected := comet.CommitInfo{
		Round: 1,
		Votes: []comet.VoteInfo{
			{
				Validator: comet.Validator{
					Address: []byte("validator1"),
					Power:   100,
				},
				BlockIDFlag: 2,
			},
		},
	}

	result := toCoreExtendedCommitInfo(input)
	require.Equal(t, expected, result)
}
