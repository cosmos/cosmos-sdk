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

func TestGetBlockResultsResponseFields(t *testing.T) {
	// Verify the struct has the expected fields
	resp := &GetBlockResultsResponse{
		Height:  1000,
		AppHash: []byte{0x01, 0x02, 0x03},
	}

	require.Equal(t, int64(1000), resp.GetHeight())
	require.Equal(t, []byte{0x01, 0x02, 0x03}, resp.GetAppHash())
	require.Nil(t, resp.GetTxsResults())
	require.Nil(t, resp.GetFinalizeBlockEvents())
	require.Nil(t, resp.GetValidatorUpdates())
	require.Nil(t, resp.GetConsensusParamUpdates())
}

func TestGetBlockResultsResponseDefaults(t *testing.T) {
	// Verify zero values work correctly
	resp := &GetBlockResultsResponse{}

	require.Equal(t, int64(0), resp.GetHeight())
	require.Nil(t, resp.GetAppHash())
	require.Nil(t, resp.GetTxsResults())
	require.Nil(t, resp.GetFinalizeBlockEvents())
}

func TestGetBlockResultsResponseNil(t *testing.T) {
	// Verify nil receiver doesn't panic
	var resp *GetBlockResultsResponse

	require.Equal(t, int64(0), resp.GetHeight())
	require.Nil(t, resp.GetAppHash())
	require.Nil(t, resp.GetTxsResults())
	require.Nil(t, resp.GetFinalizeBlockEvents())
	require.Nil(t, resp.GetValidatorUpdates())
	require.Nil(t, resp.GetConsensusParamUpdates())
}

func TestGetBlockResultsRequestFields(t *testing.T) {
	req := &GetBlockResultsRequest{
		Height: 500,
	}

	require.Equal(t, int64(500), req.GetHeight())
}

func TestGetBlockResultsRequestDefaults(t *testing.T) {
	req := &GetBlockResultsRequest{}
	require.Equal(t, int64(0), req.GetHeight())
}
