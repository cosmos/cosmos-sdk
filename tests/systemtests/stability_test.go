//go:build system_test

package systemtests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil/systemtests"
)

func TestEntireChainNonGracefulCrashRecovery(t *testing.T) {
	// Scenario:
	// 1. Start network with multiple validators
	// 2. Send transactions (sanity check)
	// 3. Crash the entire network, non-gracefully
	// 4. Restart the nodes
	// 5. Verify chain health and transaction execution
	doCrashTest(t, false)
}

func TestEntireChainGracefulCrashRecovery(t *testing.T) {
	// Scenario:
	// 1. Start network with multiple validators
	// 2. Send transactions (sanity check)
	// 3. Crash the entire network, gracefully
	// 4. Restart the nodes
	// 5. Verify chain health and transaction execution
	doCrashTest(t, true)
}

func doCrashTest(t *testing.T, graceful bool) {
	t.Helper()
	sut := systemtests.Sut
	sut.ResetChain(t)

	require.GreaterOrEqual(t, sut.NodesCount(), 2, "chaos test requires at least 2 nodes")

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)

	sender := cli.AddKey("sender")
	receiver := cli.AddKey("receiver")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", sender, "10000000stake"},
	)

	sut.StartChain(t)

	initialHeight := sut.CurrentHeight()
	t.Logf("Chain started at height %d with %d nodes", initialHeight, sut.NodesCount())

	rsp := cli.Run("tx", "bank", "send", sender, receiver, "1000stake", "--from="+sender, "--fees=1stake")
	systemtests.RequireTxSuccess(t, rsp)
	assert.Equal(t, int64(1000), cli.QueryBalance(receiver, "stake"))
	t.Log("Initial transaction successful")

	heightBeforeCrash := sut.CurrentHeight()
	t.Logf("Height before crash: %d", heightBeforeCrash)

	// Crash the entire chain
	if graceful {
		t.Log("Crashing entire chain (graceful)...")
	} else {
		t.Log("Crashing entire chain (non-graceful)...")
	}
	sut.KillChain(graceful)

	// Verify all nodes are stopped
	assert.Empty(t, sut.RunningNodes(), "all nodes should be stopped")
	assert.False(t, sut.IsNodeRunning(0), "node 0 should be stopped")
	t.Log("All nodes crashed")

	// Wait for OS to fully release LevelDB file locks after process exit
	time.Sleep(5 * time.Second)

	// Restart all nodes
	t.Log("Restarting all nodes...")
	allNodes := make([]int, sut.NodesCount())
	for i := range allNodes {
		allNodes[i] = i
	}
	require.NoError(t, sut.StartNodes(t, allNodes...))

	// Wait for at least one node to be up and syncing
	sut.AwaitNodeUp(t, "tcp://localhost:26657")

	// Wait for nodes to sync
	t.Log("Waiting for nodes to sync...")
	sut.AwaitNodesSynced(t, allNodes...)

	// Verify chain resumed from where it left off
	sut.AwaitNBlocks(t, 2)
	heightAfterRecovery := sut.CurrentHeight()
	t.Logf("Height after recovery: %d", heightAfterRecovery)
	assert.GreaterOrEqual(t, heightAfterRecovery, heightBeforeCrash, "chain should resume from previous height")

	// Verify all nodes are running
	runningNodes := sut.RunningNodes()
	assert.Len(t, runningNodes, sut.NodesCount(), "all nodes should be running")

	// Send transaction to verify chain is fully operational
	rsp = cli.Run("tx", "bank", "send", sender, receiver, "500stake", "--from="+sender, "--fees=1stake")
	systemtests.RequireTxSuccess(t, rsp)

	expectedBalance := int64(1000 + 500)
	actualBalance := cli.QueryBalance(receiver, "stake")
	assert.Equal(t, expectedBalance, actualBalance, "receiver should have correct balance after recovery")
}

func TestNodeCrashRecovery(t *testing.T) {
	// Scenario:
	// 1. Start network with multiple validators
	// 2. Send transactions (sanity check)
	// 3. Crash one validator node (non-graceful kill)
	// 4. Verify chain continues producing blocks
	// 5. Send more transactions while node is down
	// 6. Restart the crashed node
	// 7. Wait for node to sync
	// 8. Verify chain health and transaction execution

	sut := systemtests.Sut
	sut.ResetChain(t)

	require.GreaterOrEqual(t, sut.NodesCount(), 2, "chaos test requires at least 2 nodes")

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)

	sender := cli.AddKey("sender")
	receiver := cli.AddKey("receiver")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", sender, "10000000stake"},
	)

	sut.StartChain(t)

	initialHeight := sut.CurrentHeight()
	t.Logf("Chain started at height %d with %d nodes", initialHeight, sut.NodesCount())

	rsp := cli.Run("tx", "bank", "send", sender, receiver, "1000stake", "--from="+sender, "--fees=1stake")
	systemtests.RequireTxSuccess(t, rsp)
	assert.Equal(t, int64(1000), cli.QueryBalance(receiver, "stake"))
	t.Log("Initial transaction successful")

	// Record height before crash
	heightBeforeCrash := sut.CurrentHeight()
	t.Logf("Height before crash: %d", heightBeforeCrash)

	// Crash node 1 (non-graceful - simulates power failure)
	nodeToKill := 1
	t.Logf("Crashing node %d (non-graceful)...", nodeToKill)
	require.NoError(t, sut.KillNodes(false, nodeToKill))
	assert.False(t, sut.IsNodeRunning(nodeToKill), "node should be stopped")

	// Verify remaining nodes continue
	runningNodes := sut.RunningNodes()
	t.Logf("Running nodes after crash: %v", runningNodes)
	assert.NotContains(t, runningNodes, nodeToKill)

	// Wait and verify chain continues producing blocks
	time.Sleep(2 * sut.BlockTime())
	sut.AwaitNBlocks(t, 2)
	heightDuringOutage := sut.CurrentHeight()
	t.Logf("Chain continued to height %d while node %d was down", heightDuringOutage, nodeToKill)
	assert.Greater(t, heightDuringOutage, heightBeforeCrash, "chain should continue producing blocks")

	// Send transaction while node is down
	rsp = cli.Run("tx", "bank", "send", sender, receiver, "500stake", "--from="+sender, "--fees=1stake")
	systemtests.RequireTxSuccess(t, rsp)
	t.Log("Transaction during outage successful")

	// Restart the crashed node
	t.Logf("Restarting node %d...", nodeToKill)
	require.NoError(t, sut.StartNodes(t, nodeToKill))
	assert.True(t, sut.IsNodeRunning(nodeToKill), "node should be running")

	// Wait for node to sync
	t.Logf("Waiting for node %d to sync...", nodeToKill)
	sut.AwaitNodesSynced(t, nodeToKill)

	// Verify chain health after recovery
	sut.AwaitNBlocks(t, 2)
	heightAfterRecovery := sut.CurrentHeight()
	t.Logf("Height after recovery: %d", heightAfterRecovery)

	// Send final transaction to confirm full network health
	rsp = cli.Run("tx", "bank", "send", sender, receiver, "250stake", "--from="+sender, "--fees=1stake")
	systemtests.RequireTxSuccess(t, rsp)
}

func TestNodePauseResume(t *testing.T) {
	// Scenario:
	// Test node pause (SIGSTOP) and resume (SIGCONT)
	// This simulates a frozen process (long GC, I/O stall)

	sut := systemtests.Sut
	sut.ResetChain(t)

	require.GreaterOrEqual(t, sut.NodesCount(), 2, "chaos test requires at least 2 nodes")

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)

	sender := cli.AddKey("sender")
	receiver := cli.AddKey("receiver")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", sender, "10000000stake"},
	)

	sut.StartChain(t)

	rsp := cli.Run("tx", "bank", "send", sender, receiver, "1000stake", "--from="+sender, "--fees=1stake")
	systemtests.RequireTxSuccess(t, rsp)

	nodeToPause := 1
	t.Logf("Pausing node %d...", nodeToPause)
	require.NoError(t, sut.PauseNodes(nodeToPause))

	// Node is still "running" (process exists) but frozen
	assert.True(t, sut.IsNodeRunning(nodeToPause), "paused node should still appear as running")

	// Chain should continue with remaining nodes
	heightBeforePause := sut.CurrentHeight()
	sut.AwaitNBlocks(t, 2)
	assert.Greater(t, sut.CurrentHeight(), heightBeforePause, "chain should continue with paused node")

	// Send transaction while node is paused
	rsp = cli.Run("tx", "bank", "send", sender, receiver, "500stake", "--from="+sender, "--fees=1stake")
	systemtests.RequireTxSuccess(t, rsp)

	// Resume the node
	t.Logf("Resuming node %d...", nodeToPause)
	require.NoError(t, sut.ResumeNodes(nodeToPause))

	// Give it time to catch up
	time.Sleep(2 * sut.BlockTime())
	sut.AwaitNBlocks(t, 2)

	rsp = cli.Run("tx", "bank", "send", sender, receiver, "250stake", "--from="+sender, "--fees=1stake")
	systemtests.RequireTxSuccess(t, rsp)
}

func TestMultipleNodeFailures(t *testing.T) {
	// Scenario:
	// Test network resilience with multiple node failures
	// Kill nodes one at a time and verify network stays healthy

	sut := systemtests.Sut
	sut.ResetChain(t)

	require.GreaterOrEqual(t, sut.NodesCount(), 3, "this test requires at least 3 nodes")

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)

	sender := cli.AddKey("sender")
	receiver := cli.AddKey("receiver")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", sender, "10000000stake"},
	)

	sut.StartChain(t)
	t.Logf("Started chain with %d nodes", sut.NodesCount())

	rsp := cli.Run("tx", "bank", "send", sender, receiver, "1000stake", "--from="+sender, "--fees=1stake")
	systemtests.RequireTxSuccess(t, rsp)

	runningNodesBefore := sut.RunningNodes()
	// Kill node 2 (graceful)
	t.Log("Killing node 2 (graceful)...")
	require.NoError(t, sut.KillNodes(true, 2))
	sut.AwaitNBlocks(t, 2)
	runningNodesAfter := sut.RunningNodes()
	require.Len(t, runningNodesAfter, len(runningNodesBefore)-1)

	rsp = cli.Run("tx", "bank", "send", sender, receiver, "500stake", "--from="+sender, "--fees=1stake")
	systemtests.RequireTxSuccess(t, rsp)

	// Restart node 2
	t.Log("Restarting node 2...")
	require.NoError(t, sut.StartNodes(t, 2))
	sut.AwaitNodesSynced(t, 2)

	// Kill node 1
	t.Log("Killing node 1 (crash)...")
	require.NoError(t, sut.KillNodes(false, 1))
	sut.AwaitNBlocks(t, 2)

	// Txs still work
	rsp = cli.Run("tx", "bank", "send", sender, receiver, "250stake", "--from="+sender, "--fees=1stake")
	systemtests.RequireTxSuccess(t, rsp)

	// Restart node 1
	t.Log("Restarting node 1...")
	require.NoError(t, sut.StartNodes(t, 1))
	sut.AwaitNodesSynced(t, 1)

	// All nodes should be up
	runningNodes := sut.RunningNodes()
	assert.Len(t, runningNodes, sut.NodesCount(), "all nodes should be running")

	// Sanity check on the balance
	expectedBalance := int64(1000 + 500 + 250)
	assert.Equal(t, expectedBalance, cli.QueryBalance(receiver, "stake"))
}
