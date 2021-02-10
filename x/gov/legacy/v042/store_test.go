package v042_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v040gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v040"
	v042gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v042"
)

func TestStoreMigration(t *testing.T) {
	govKey := sdk.NewKVStoreKey("gov")
	ctx := testutil.DefaultContext(govKey, sdk.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(govKey)

	_, _, addr1 := testdata.KeyTestPubAddr()
	proposalID := uint64(6)
	now := time.Now()
	// Use dummy value for all keys.
	value := []byte("foo")

	testCases := []struct {
		name   string
		oldKey []byte
		newKey []byte
	}{
		{
			"ProposalKey",
			v040gov.ProposalKey(proposalID),
			v042gov.ProposalKey(proposalID),
		},
		{
			"ActiveProposalQueue",
			v040gov.ActiveProposalQueueKey(proposalID, now),
			v042gov.ActiveProposalQueueKey(proposalID, now),
		},
		{
			"InactiveProposalQueue",
			v040gov.InactiveProposalQueueKey(proposalID, now),
			v042gov.InactiveProposalQueueKey(proposalID, now),
		},
		{
			"ProposalIDKey",
			v040gov.ProposalIDKey,
			v042gov.ProposalIDKey,
		},
		{
			"DepositKey",
			v040gov.DepositKey(proposalID, addr1),
			v042gov.DepositKey(proposalID, addr1),
		},
		{
			"VotesKeyPrefix",
			v040gov.VoteKey(proposalID, addr1),
			v042gov.VoteKey(proposalID, addr1),
		},
	}

	// Set all the old keys to the store
	for _, tc := range testCases {
		store.Set(tc.oldKey, value)
	}

	// Run migrations.
	err := v042gov.MigrateStore(ctx, govKey)
	require.NoError(t, err)

	// Make sure the new keys are set and old keys are deleted.
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if bytes.Compare(tc.oldKey, tc.newKey) != 0 {
				require.Nil(t, store.Get(tc.oldKey))
			}
			require.Equal(t, value, store.Get(tc.newKey))
		})
	}
}
