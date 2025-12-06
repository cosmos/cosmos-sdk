package cmtservice

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSyncingResponseFields(t *testing.T) {
	// Verify the struct has the expected fields
	resp := &GetSyncingResponse{
		Syncing:             true,
		EarliestBlockHeight: 1000,
		LatestBlockHeight:   2000,
	}

	require.True(t, resp.GetSyncing())
	require.Equal(t, int64(1000), resp.GetEarliestBlockHeight())
	require.Equal(t, int64(2000), resp.GetLatestBlockHeight())
}

func TestGetSyncingResponseDefaults(t *testing.T) {
	// Verify zero values work correctly
	resp := &GetSyncingResponse{}

	require.False(t, resp.GetSyncing())
	require.Equal(t, int64(0), resp.GetEarliestBlockHeight())
	require.Equal(t, int64(0), resp.GetLatestBlockHeight())
}

func TestGetSyncingResponseNil(t *testing.T) {
	// Verify nil receiver doesn't panic
	var resp *GetSyncingResponse

	require.False(t, resp.GetSyncing())
	require.Equal(t, int64(0), resp.GetEarliestBlockHeight())
	require.Equal(t, int64(0), resp.GetLatestBlockHeight())
}
