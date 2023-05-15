package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProposalKeys(t *testing.T) {
	// key proposal
	key := ProposalKey(1)
	proposalID := SplitProposalKey(key)
	require.Equal(t, int(proposalID), 1)

	// key active proposal queue
	now := time.Now()
	key = ActiveProposalQueueKey(3, now)
	proposalID, expTime := SplitActiveProposalQueueKey(key)
	require.Equal(t, int(proposalID), 3)
	require.True(t, now.Equal(expTime))

	// key inactive proposal queue
	key = InactiveProposalQueueKey(3, now)
	proposalID, expTime = SplitInactiveProposalQueueKey(key)
	require.Equal(t, int(proposalID), 3)
	require.True(t, now.Equal(expTime))

	// invalid key
	require.Panics(t, func() { SplitProposalKey([]byte("test")) })
	require.Panics(t, func() { SplitInactiveProposalQueueKey([]byte("test")) })
}
