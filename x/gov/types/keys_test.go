package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var addr = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

func TestProposalKeys(t *testing.T) {
	// key proposal
	key := KeyProposal(1)
	proposalID := SplitKeyProposal(key)
	require.Equal(t, int(proposalID), 1)

	// key active proposal queue
	now := time.Now()
	key = KeyActiveProposalQueue(3, now)
	proposalID, expTime := SplitKeyActiveProposalQueue(key)
	require.Equal(t, int(proposalID), 3)
	require.True(t, now.Equal(expTime))

	// key inactive proposal queue
	key = KeyInactiveProposalQueue(3, now)
	proposalID, expTime = SplitKeyInactiveProposalQueue(key)
	require.Equal(t, int(proposalID), 3)
	require.True(t, now.Equal(expTime))

	// invalid key
	require.Panics(t, func() { SplitKeyProposal([]byte("test")) })
	require.Panics(t, func() { SplitKeyInactiveProposalQueue([]byte("test")) })
}

func TestDepositKeys(t *testing.T) {

	key := KeyProposalDeposits(2)
	proposalID := SplitKeyProposal(key)
	require.Equal(t, int(proposalID), 2)

	key = KeyProposalDeposit(2, addr)
	proposalID, depositorAddr := SplitKeyDeposit(key)
	require.Equal(t, int(proposalID), 2)
	require.Equal(t, addr, depositorAddr)

	// invalid key
	addr2 := sdk.AccAddress("test1")
	key = KeyProposalDeposit(5, addr2)
	require.Panics(t, func() { SplitKeyDeposit(key) })
}

func TestVoteKeys(t *testing.T) {

	key := KeyProposalVotes(2)
	proposalID := SplitKeyProposal(key)
	require.Equal(t, int(proposalID), 2)

	key = KeyProposalVote(2, addr)
	proposalID, voterAddr := SplitKeyDeposit(key)
	require.Equal(t, int(proposalID), 2)
	require.Equal(t, addr, voterAddr)

	// invalid key
	addr2 := sdk.AccAddress("test1")
	key = KeyProposalVote(5, addr2)
	require.Panics(t, func() { SplitKeyVote(key) })
}
