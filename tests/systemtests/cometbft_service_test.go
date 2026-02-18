//go:build system_test

package systemtests

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/creachadair/tomledit"
	"github.com/creachadair/tomledit/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/cosmos/cosmos-sdk/testutil/systemtests"
)

// TestCometBFTGetSyncingGRPC tests the GetSyncing gRPC endpoint
// to verify it returns the expected earliest_block_height and latest_block_height fields.
// This test validates the feature added in PR #25647.
func TestCometBFTGetSyncingGRPC(t *testing.T) {
	sut := systemtests.Sut
	sut.ResetChain(t)

	sut.StartChain(t)

	// Wait for a few blocks to be produced
	sut.AwaitNBlocks(t, 3)

	// Connect to gRPC endpoint
	grpcAddr := fmt.Sprintf("localhost:%d", 9090) // DefaultGrpcPort
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	queryClient := cmtservice.NewServiceClient(conn)

	// Test that the GetSyncing gRPC endpoint returns all expected fields
	t.Run("gRPC GetSyncing returns block heights", func(t *testing.T) {
		resp, err := queryClient.GetSyncing(context.Background(), &cmtservice.GetSyncingRequest{})
		require.NoError(t, err)
		require.NotNil(t, resp)

		t.Logf("gRPC GetSyncing response: syncing=%v, earliest=%d, latest=%d",
			resp.Syncing, resp.EarliestBlockHeight, resp.LatestBlockHeight)

		// Verify earliest_block_height is a valid height (>= 1)
		assert.GreaterOrEqual(t, resp.EarliestBlockHeight, int64(1),
			"earliest_block_height should be >= 1, got %d", resp.EarliestBlockHeight)

		// Verify latest_block_height is a valid height (>= 1)
		assert.GreaterOrEqual(t, resp.LatestBlockHeight, int64(1),
			"latest_block_height should be >= 1, got %d", resp.LatestBlockHeight)

		// Verify latest >= earliest (invariant)
		assert.GreaterOrEqual(t, resp.LatestBlockHeight, resp.EarliestBlockHeight,
			"latest_block_height (%d) should be >= earliest_block_height (%d)",
			resp.LatestBlockHeight, resp.EarliestBlockHeight)
	})

	// Test that latest_block_height increases as chain progresses
	t.Run("gRPC latest height increases over time", func(t *testing.T) {
		// Get initial height
		resp1, err := queryClient.GetSyncing(context.Background(), &cmtservice.GetSyncingRequest{})
		require.NoError(t, err)
		initialLatest := resp1.LatestBlockHeight
		t.Logf("Initial latest_block_height: %d", initialLatest)

		// Wait for more blocks
		sut.AwaitNBlocks(t, 2)

		// Get new height
		resp2, err := queryClient.GetSyncing(context.Background(), &cmtservice.GetSyncingRequest{})
		require.NoError(t, err)
		newLatest := resp2.LatestBlockHeight
		t.Logf("New latest_block_height: %d", newLatest)

		assert.Greater(t, newLatest, initialLatest,
			"latest_block_height should increase over time (was %d, now %d)",
			initialLatest, newLatest)
	})

	// Test that earliest_block_height remains stable on an unpruned chain
	t.Run("gRPC earliest height stable on unpruned chain", func(t *testing.T) {
		// Get initial earliest height
		resp1, err := queryClient.GetSyncing(context.Background(), &cmtservice.GetSyncingRequest{})
		require.NoError(t, err)
		initialEarliest := resp1.EarliestBlockHeight
		t.Logf("Initial earliest_block_height: %d", initialEarliest)

		// Wait for more blocks
		sut.AwaitNBlocks(t, 2)

		// Get earliest height again
		resp2, err := queryClient.GetSyncing(context.Background(), &cmtservice.GetSyncingRequest{})
		require.NoError(t, err)
		currentEarliest := resp2.EarliestBlockHeight
		t.Logf("Current earliest_block_height: %d", currentEarliest)

		// On an unpruned chain, earliest should remain at 1 or the initial value
		assert.Equal(t, initialEarliest, currentEarliest,
			"earliest_block_height should remain stable on unpruned chain (was %d, now %d)",
			initialEarliest, currentEarliest)
	})
}

// TestCometBFTGetSyncingWithBlockRetention tests that earliest_block_height
// increases when block retention (min-retain-blocks) is configured.
// This test configures aggressive pruning settings to verify the feature works.
func TestCometBFTGetSyncingWithBlockRetention(t *testing.T) {
	const minRetainBlocks = 5

	sut := systemtests.Sut

	// Modify genesis to set a low max_age_num_blocks for evidence
	// Block pruning retention height = min(commitHeight-minRetainBlocks, commitHeight-maxAgeNumBlocks, ...)
	// Default max_age_num_blocks is 100000, which would prevent pruning in tests
	sut.ModifyGenesisJSON(t, func(genesis []byte) []byte {
		state, err := sjson.Set(string(genesis), "consensus.params.evidence.max_age_num_blocks", strconv.Itoa(minRetainBlocks))
		require.NoError(t, err)
		return []byte(state)
	})

	// Configure min-retain-blocks and pruning in app.toml before starting the chain
	configureBlockPruning(t, sut, minRetainBlocks)

	sut.StartChain(t)

	// Connect to gRPC endpoint
	grpcAddr := fmt.Sprintf("localhost:%d", 9090)
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	queryClient := cmtservice.NewServiceClient(conn)

	// Get initial state
	resp, err := queryClient.GetSyncing(context.Background(), &cmtservice.GetSyncingRequest{})
	require.NoError(t, err)
	initialEarliest := resp.EarliestBlockHeight
	initialLatest := resp.LatestBlockHeight
	t.Logf("Initial state: earliest=%d, latest=%d", initialEarliest, initialLatest)

	// Verify the response structure is correct
	require.GreaterOrEqual(t, initialEarliest, int64(1), "earliest_block_height should be >= 1")
	require.GreaterOrEqual(t, initialLatest, int64(1), "latest_block_height should be >= 1")
	require.GreaterOrEqual(t, initialLatest, initialEarliest, "latest >= earliest invariant")

	// Wait for enough blocks that pruning should definitely occur
	// Need to wait for more than minRetainBlocks to trigger pruning
	blocksToWait := minRetainBlocks * 3
	t.Logf("Waiting for %d blocks to trigger block pruning (minRetainBlocks=%d)...", blocksToWait, minRetainBlocks)
	sut.AwaitNBlocks(t, int64(blocksToWait))

	// Check state after waiting - pruning should have occurred
	resp, err = queryClient.GetSyncing(context.Background(), &cmtservice.GetSyncingRequest{})
	require.NoError(t, err)
	newEarliest := resp.EarliestBlockHeight
	newLatest := resp.LatestBlockHeight
	t.Logf("After %d blocks: earliest=%d, latest=%d", blocksToWait, newEarliest, newLatest)

	// Verify latest has increased (this should always be true)
	assert.Greater(t, newLatest, initialLatest,
		"latest_block_height should have increased")

	// Verify earliest has increased due to block pruning
	// With minRetainBlocks=5 and maxAgeNumBlocks=5, pruning should occur after ~5 blocks
	assert.Greater(t, newEarliest, initialEarliest,
		"earliest_block_height should increase when blocks are pruned (minRetainBlocks=%d, was %d, now %d)",
		minRetainBlocks, initialEarliest, newEarliest)

	// Verify the block range is bounded by our retention settings
	blockDiff := newLatest - newEarliest
	t.Logf("Block range: %d (latest %d - earliest %d)", blockDiff, newLatest, newEarliest)

	// The block range should be close to minRetainBlocks (with some tolerance)
	// Allow extra tolerance since pruning timing isn't exact
	maxExpectedRange := int64(minRetainBlocks + 5)
	assert.LessOrEqual(t, blockDiff, maxExpectedRange,
		"block range should be close to minRetainBlocks (%d), got %d",
		minRetainBlocks, blockDiff)

	// Verify invariant still holds
	assert.GreaterOrEqual(t, newLatest, newEarliest,
		"latest >= earliest invariant should hold after pruning")
}

// configureBlockPruning edits the app.toml for all nodes to enable block pruning
// This requires setting min-retain-blocks, state pruning, and disabling state sync snapshots
func configureBlockPruning(t *testing.T, sut *systemtests.SystemUnderTest, minRetainBlocks int) {
	t.Helper()

	// Edit app.toml for each node
	for i := 0; i < sut.NodesCount(); i++ {
		// NodeDir already includes WorkDir, so just append config/app.toml
		appTomlPath := filepath.Join(sut.NodeDir(i), "config", "app.toml")
		systemtests.EditToml(appTomlPath, func(doc *tomledit.Document) {
			// Set min-retain-blocks for CometBFT block pruning
			setInt(doc, minRetainBlocks, "min-retain-blocks")
			// Use "everything" pruning which is the most aggressive and doesn't require
			// custom interval settings (pruning-interval minimum is 10)
			setString(doc, "everything", "pruning")
			// Disable state sync snapshots (0 = disabled) - required for "everything" pruning
			setInt(doc, 0, "state-sync", "snapshot-interval")
		})
		t.Logf("Configured block pruning (min-retain-blocks=%d, pruning=everything) in %s", minRetainBlocks, appTomlPath)
	}
}

// setInt sets an integer value in a toml document
func setInt(doc *tomledit.Document, newVal int, xpath ...string) {
	e := doc.First(xpath...)
	if e == nil {
		panic(fmt.Sprintf("not found: %v", xpath))
	}
	e.Value = parser.MustValue(strconv.Itoa(newVal))
}

// setString sets a quoted string value in a toml document
func setString(doc *tomledit.Document, newVal string, xpath ...string) {
	e := doc.First(xpath...)
	if e == nil {
		panic(fmt.Sprintf("not found: %v", xpath))
	}
	e.Value = parser.MustValue(fmt.Sprintf("%q", newVal))
}
