//go:build system_test

package systemtests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	systest "github.com/cosmos/cosmos-sdk/testutil/systemtests"
)

const (
	loadTestEnv         = "COSMOS_RUN_HEAVY_LOAD_TEST"
	loadTestSenderCount = 100
	loadTestTxsPerBatch = 200
	loadTestBatches     = 200
	loadTestFundAmount  = "10000000stake"
	loadTestSendAmount  = "1stake"
	loadTestBlockWait   = 30 * time.Second

	loadTestLightSenderCount = 50
	loadTestLightTxCount     = 10000
	loadTestLightWorkers     = 50 // limit concurrent simd processes to avoid thrashing

	// Mini load test: runs in short mode on PRs (~1-2 min)
	loadTestMiniSenderCount = 5
	loadTestMiniTxCount    = 1000
	loadTestMiniWorkers    = 10

	loadTestReceiverCount = 50

	loadTestSenderPrefix      = "sender"
	loadTestReceiverPrefix    = "receiver"
	loadTestInitialBalance    = "10000000000stake"
	loadTestInitialBalanceInt = 10000000000
)

// txHashWithTime holds a tx hash and its broadcast time for inclusion stats.
type txHashWithTime struct {
	hash  string
	bcast time.Time
}

// nodeEndpoint holds the RPC and gRPC addresses for a node.
type nodeEndpoint struct {
	RPC  string
	GRPC string
}

// loadTestSetup holds the common chain setup for load tests.
type loadTestSetup struct {
	cli           *systest.CLIWrapper
	senderNames   []string
	senderAddrs   []string
	receiverAddrs []string
	nodeEndpoints []nodeEndpoint
}

// setupLoadTestChain resets the chain, creates senders and receivers, funds them in genesis, starts the chain, and returns the setup.
func setupLoadTestChain(t *testing.T, senderCount, receiverCount int, fundAmount string) (*systest.SystemUnderTest, *loadTestSetup) {
	t.Helper()
	sut := systest.Sut
	sut.ResetChain(t)
	require.GreaterOrEqual(t, sut.NodesCount(), 2, "load test requires at least 2 nodes")

	cli := systest.NewCLIWrapper(t, sut, systest.Verbose)
	senderNames := make([]string, senderCount)
	senderAddrs := make([]string, senderCount)
	for i := 0; i < senderCount; i++ {
		name := fmt.Sprintf("%s%d", loadTestSenderPrefix, i)
		senderNames[i] = name
		senderAddrs[i] = cli.AddKey(name)
	}
	receiverAddrs := make([]string, receiverCount)
	for i := 0; i < receiverCount; i++ {
		name := fmt.Sprintf("%s%d", loadTestReceiverPrefix, i)
		receiverAddrs[i] = cli.AddKey(name)
	}

	genesisCmds := make([][]string, 0, senderCount+receiverCount)
	for _, addr := range receiverAddrs {
		genesisCmds = append(genesisCmds, []string{"genesis", "add-genesis-account", addr, fundAmount})
	}
	for _, addr := range senderAddrs {
		genesisCmds = append(genesisCmds, []string{"genesis", "add-genesis-account", addr, fundAmount})
	}
	sut.ModifyGenesisCLI(t, genesisCmds...)
	sut.StartChain(t)
	sut.AwaitNBlocks(t, 2)

	nodeEndpoints := make([]nodeEndpoint, sut.NodesCount())
	for i := 0; i < sut.NodesCount(); i++ {
		nodeEndpoints[i] = nodeEndpoint{
			RPC:  fmt.Sprintf("tcp://127.0.0.1:%d", systest.DefaultRpcPort+i),
			GRPC: fmt.Sprintf("127.0.0.1:%d", systest.DefaultGrpcPort+i),
		}
	}

	return sut, &loadTestSetup{cli, senderNames, senderAddrs, receiverAddrs, nodeEndpoints}
}

// gatherLoadTestStats scans blocks for the active window (first-to-last with txs),
// then reports committed count, TPS, block utilization, and avg time to block inclusion.
func gatherLoadTestStats(t *testing.T, sut *systest.SystemUnderTest, committed, heightBeforeBroadcast int64, txHashesWithTime []txHashWithTime) {
	t.Helper()
	heightAfterWait := sut.CurrentHeight()
	blockRpc := sut.RPCClient(t)
	type blockInfo struct {
		height int64
		txCnt  int
		time   time.Time
	}
	var withTxs []blockInfo
	for h := heightBeforeBroadcast + 1; h <= heightAfterWait; h++ {
		blk, err := blockRpc.Block(context.Background(), &h)
		if err != nil {
			continue
		}
		n := len(blk.Block.Txs)
		if n > 0 {
			withTxs = append(withTxs, blockInfo{h, n, blk.Block.Time})
		}
	}

	if len(withTxs) == 0 {
		t.Logf("load stats: no blocks with txs in range")
		return
	}
	first, last := withTxs[0], withTxs[len(withTxs)-1]
	activeDuration := last.time.Sub(first.time)
	var withTxsStr []string
	for _, b := range withTxs {
		withTxsStr = append(withTxsStr, fmt.Sprintf("h%d(%d)", b.height, b.txCnt))
	}
	if activeDuration > 0 {
		tps := float64(committed) / activeDuration.Seconds()
		t.Logf("load stats (active window h%d–h%d): committed=%d, avg TPS=%.1f (over %s)", first.height, last.height, committed, tps, activeDuration.Round(time.Millisecond))
	} else {
		t.Logf("load stats: all txs in a single block (h%d); TPS undefined", first.height)
	}
	t.Logf("block utilization: %d blocks %v", len(withTxs), withTxsStr)

	const maxSample = 50
	sample := txHashesWithTime
	if len(sample) > maxSample {
		sample = sample[:maxSample]
	}
	if len(sample) > 0 {
		var totalInclusion time.Duration
		var included int
		for _, e := range sample {
			txRes, err := blockRpc.TxByHash(context.Background(), e.hash)
			if err != nil {
				continue
			}
			if txRes.Height < first.height || txRes.Height > last.height {
				continue
			}
			blk, err := blockRpc.Block(context.Background(), &txRes.Height)
			if err != nil {
				continue
			}
			blockTime := blk.Block.Time
			inclusion := blockTime.Sub(e.bcast)
			if inclusion >= 0 {
				totalInclusion += inclusion
				included++
			}
		}
		if included > 0 {
			avgInclusion := totalInclusion / time.Duration(included)
			t.Logf("load stats: avg time to block inclusion=%s (sampled %d/%d txs in active window)", avgInclusion.Round(time.Millisecond), included, len(sample))
		}
	}
}

// TestHeavyLoad introduces sustained transaction load across multiple senders, receivers, and RPC nodes.
// It is gated by COSMOS_RUN_HEAVY_LOAD_TEST=1 to avoid slowing CI.
//
// Pattern (adapted from evm's live_repro.go):
// - Multiple senders and receivers; sends distributed round-robin to reduce account contention
// - Round-robin broadcast to RPC nodes (not validator) to maximize mempool contention
// - Verifies chain health (apphash consistency) after each batch
func TestHeavyLoad(t *testing.T) {
	if os.Getenv(loadTestEnv) != "1" {
		t.Skipf("set %s=1 to run the heavy load test", loadTestEnv)
	}

	sut, setup := setupLoadTestChain(t, loadTestSenderCount, loadTestReceiverCount, loadTestFundAmount)
	senderNames, receiverAddrs, nodeEndpoints := setup.senderNames, setup.receiverAddrs, setup.nodeEndpoints

	// Wait for chain to stabilize
	time.Sleep(2 * sut.BlockTime())

	loadStart := time.Now()

	// Use unordered txs to avoid sequence conflicts when many txs from same sender are in flight.
	// Distribute sends across multiple receivers to reduce account contention.
	// Each tx needs a unique timeout; use batch+inner index to stagger creation time.
	var totalSent, totalSkipped atomic.Int64
	for batch := 0; batch < loadTestBatches; batch++ {
		var wg sync.WaitGroup
		txIdx := 0
		for i := 0; i < loadTestTxsPerBatch; i++ {
			for si, senderName := range senderNames {
				idx := txIdx
				txIdx++
				toAddr := receiverAddrs[(si+idx)%len(receiverAddrs)]
				wg.Add(1)
				go func(fromKey, toAddr, nodeAddr string) {
					defer wg.Done()
					c := systest.NewCLIWrapper(t, sut, systest.Verbose).
						WithNodeAddress(nodeAddr).
						WithRunErrorsIgnored()
					rsp, _ := c.RunOnly("tx", "bank", "send", fromKey, toAddr, loadTestSendAmount, "--from="+fromKey, "--fees=1stake", "--unordered", "--timeout-duration=5m")
					if gjson.Get(rsp, "code").Int() == 0 {
						totalSent.Add(1)
					} else {
						totalSkipped.Add(1)
					}
				}(senderName, toAddr, nodeEndpoints[si%len(nodeEndpoints)].RPC)
			}
		}
		wg.Wait()

		// Wait for next block and verify apphash consistency
		sut.AwaitNBlocks(t, 1, loadTestBlockWait)

		// Check for consensus failure in logs
		for i := 0; i < sut.NodesCount(); i++ {
			logPath := filepath.Join(systest.WorkDir, "testnet", fmt.Sprintf("node%d.out", i))
			data, err := os.ReadFile(logPath)
			if err == nil && strings.Contains(string(data), "CONSENSUS FAILURE") {
				t.Fatalf("CONSENSUS FAILURE detected on node%d at batch=%d", i, batch)
			}
		}

		if batch%10 == 0 {
			t.Logf("batch=%d sent=%d skipped=%d", batch, totalSent.Load(), totalSkipped.Load())
		}
	}

	elapsed := time.Since(loadStart)
	sent := totalSent.Load()
	t.Logf("completed heavy load test: %d txs sent, %d skipped", sent, totalSkipped.Load())
	if elapsed > 0 && sent > 0 {
		t.Logf("load stats: avg TPS=%.1f (%.0f txs over %s)", float64(sent)/elapsed.Seconds(), float64(sent), elapsed.Round(time.Millisecond))
	}
}

// TestHeavyLoadMini runs in short mode on PRs. Small load (~200 txs) to validate broadcaster and chain under load.
func TestHeavyLoadMini(t *testing.T) {
	runProgrammaticLoadTest(t, loadTestMiniSenderCount, loadTestMiniTxCount, loadTestMiniWorkers, loadTestInitialBalance)
}

// TestHeavyLoadLight is a lighter variant for the extended suite. Uses 10k txs; skipped in short mode.
func TestHeavyLoadLight(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping light load test in short mode")
	}
	runProgrammaticLoadTest(t, loadTestLightSenderCount, loadTestLightTxCount, loadTestLightWorkers, loadTestInitialBalance)
}

func runProgrammaticLoadTest(t *testing.T, senderCount, txCount, workers int, fundAmount string) {
	t.Helper()
	receiverCount := loadTestReceiverCount
	if senderCount < receiverCount {
		receiverCount = senderCount
	}
	sut, setup := setupLoadTestChain(t, senderCount, receiverCount, fundAmount)
	cli := systest.NewCLIWrapper(t, sut, false)
	senderNames, receiverAddrs, nodeEndpoints := setup.senderNames, setup.receiverAddrs, setup.nodeEndpoints

	// Use programmatic broadcast (no simd spawn per tx) for speed. One broadcaster per node.
	keyringDir := systest.KeyringDir(systest.WorkDir, sut.OutputDir())
	broadcasters := make(map[string]*systest.LoadTestBroadcaster)
	for _, ep := range nodeEndpoints {
		bc, err := systest.NewLoadTestBroadcaster(keyringDir, sut.ChainID(), ep.RPC, ep.GRPC)
		if err != nil {
			t.Fatalf("load test broadcaster for %s: %v", ep.RPC, err)
		}
		broadcasters[ep.RPC] = bc
		t.Cleanup(func() { _ = bc.Close() })
	}

	var sent, failed atomic.Int64
	var txHashesMu sync.Mutex
	txHashesWithTime := make([]txHashWithTime, 0, txCount)

	heightBeforeBroadcast := sut.CurrentHeight()
	type job struct {
		idx          int
		senderName   string
		receiverAddr string
		nodeAddr     string
	}
	jobs := make(chan job, txCount)
	for i := 0; i < txCount; i++ {
		jobs <- job{
			idx:          i,
			senderName:   senderNames[i%senderCount],
			receiverAddr: receiverAddrs[i%len(receiverAddrs)],
			nodeAddr:     nodeEndpoints[i%len(nodeEndpoints)].RPC,
		}
	}
	close(jobs)

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				s, f := sent.Load(), failed.Load()
				pct := 100 * float64(s+f) / float64(txCount)
				t.Logf("load progress: %d/%d txs (%.0f%%) — %d sent, %d failed", s+f, txCount, pct, s, f)
			}
		}
	}()

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				bc := broadcasters[j.nodeAddr]
				txHash, code, err := bc.BroadcastBankSendUnordered(j.senderName, j.receiverAddr, "10stake", "1stake", j.idx)
				if err == nil && code == 0 {
					sent.Add(1)
					if txHash != "" {
						txHashesMu.Lock()
						txHashesWithTime = append(txHashesWithTime, txHashWithTime{txHash, time.Now()})
						txHashesMu.Unlock()
					}
				} else {
					failed.Add(1)
				}
			}
		}()
	}
	wg.Wait()
	close(done)

	t.Logf("broadcast complete: %d accepted, %d rejected", sent.Load(), failed.Load())
	require.Greater(t, sent.Load(), int64(0), "at least some txs should be accepted")

	// Wait for blocks to process the mempool
	sut.AwaitNBlocks(t, 10)

	var totalReceived int64
	for _, addr := range receiverAddrs {
		totalReceived += cli.QueryBalance(addr, "stake")
	}
	initialTotal := int64(loadTestInitialBalanceInt) * int64(len(receiverAddrs))
	require.Greater(t, totalReceived, initialTotal, "receivers should have received some transfers")

	txHashesMu.Lock()
	sample := txHashesWithTime
	txHashesMu.Unlock()
	gatherLoadTestStats(t, sut, (totalReceived-initialTotal)/10, heightBeforeBroadcast, sample)
}
