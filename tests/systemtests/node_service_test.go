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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"cosmossdk.io/systemtests"

	"github.com/cosmos/cosmos-sdk/client/grpc/node"
)

// TestNodeStatusGRPC tests the Status gRPC endpoint to verify earliest_store_height.
func TestNodeStatusGRPC(t *testing.T) {
	sut := systemtests.Sut
	sut.ResetChain(t)
	sut.StartChain(t)
	sut.AwaitNBlocks(t, 3)

	grpcAddr := fmt.Sprintf("localhost:%d", 9090)
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	queryClient := node.NewServiceClient(conn)

	t.Run("returns valid store heights", func(t *testing.T) {
		resp, err := queryClient.Status(context.Background(), &node.StatusRequest{})
		require.NoError(t, err)
		t.Logf("Status response: earliest_store_height=%d, height=%d", resp.EarliestStoreHeight, resp.Height)

		assert.GreaterOrEqual(t, resp.EarliestStoreHeight, uint64(1))
		assert.GreaterOrEqual(t, resp.Height, uint64(1))
		assert.GreaterOrEqual(t, resp.Height, resp.EarliestStoreHeight)
	})

	t.Run("earliest stable on unpruned chain", func(t *testing.T) {
		resp1, err := queryClient.Status(context.Background(), &node.StatusRequest{})
		require.NoError(t, err)
		initial := resp1.EarliestStoreHeight

		sut.AwaitNBlocks(t, 2)

		resp2, err := queryClient.Status(context.Background(), &node.StatusRequest{})
		require.NoError(t, err)
		assert.Equal(t, initial, resp2.EarliestStoreHeight)
	})
}

// TestNodeStatusWithStatePruning tests earliest_store_height increases with state pruning.
func TestNodeStatusWithStatePruning(t *testing.T) {
	const pruningKeepRecent = 5
	const pruningInterval = 10

	sut := systemtests.Sut
	sut.ResetChain(t)

	// Configure state pruning
	for i := 0; i < sut.NodesCount(); i++ {
		appTomlPath := filepath.Join(sut.NodeDir(i), "config", "app.toml")
		systemtests.EditToml(appTomlPath, func(doc *tomledit.Document) {
			setNodeString(doc, "custom", "pruning")
			setNodeString(doc, strconv.Itoa(pruningKeepRecent), "pruning-keep-recent")
			setNodeString(doc, strconv.Itoa(pruningInterval), "pruning-interval")
		})
	}

	sut.StartChain(t)

	grpcAddr := fmt.Sprintf("localhost:%d", 9090)
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	queryClient := node.NewServiceClient(conn)

	resp, err := queryClient.Status(context.Background(), &node.StatusRequest{})
	require.NoError(t, err)
	initialEarliest := resp.EarliestStoreHeight
	t.Logf("Initial: earliest_store_height=%d, height=%d", initialEarliest, resp.Height)

	// Wait for pruning to occur
	blocksToWait := pruningInterval + pruningKeepRecent + 5
	t.Logf("Waiting %d blocks for state pruning...", blocksToWait)
	sut.AwaitNBlocks(t, int64(blocksToWait))

	resp, err = queryClient.Status(context.Background(), &node.StatusRequest{})
	require.NoError(t, err)
	t.Logf("After %d blocks: earliest_store_height=%d, height=%d", blocksToWait, resp.EarliestStoreHeight, resp.Height)

	assert.Greater(t, resp.EarliestStoreHeight, initialEarliest,
		"earliest_store_height should increase after pruning")
	assert.GreaterOrEqual(t, resp.Height, resp.EarliestStoreHeight)
}

func setNodeString(doc *tomledit.Document, val string, xpath ...string) {
	e := doc.First(xpath...)
	if e == nil {
		panic(fmt.Sprintf("not found: %v", xpath))
	}
	e.Value = parser.MustValue(fmt.Sprintf("%q", val))
}
