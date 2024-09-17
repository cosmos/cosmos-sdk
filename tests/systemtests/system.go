package systemtests

import (
	"bufio"
	"bytes"
	"container/ring"
	"context"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/cometbft/cometbft/libs/sync"
	client "github.com/cometbft/cometbft/rpc/client/http"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"

	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// WorkDir is the directory where tests are executed. Path should be relative to this dir
	WorkDir string

	// ExecBinaryUnversionedRegExp regular expression to extract the unversioned binary name
	ExecBinaryUnversionedRegExp = regexp.MustCompile(`^(\w+)-?.*$`)
)

type TestnetInitializer interface {
	Initialize()
}

// SystemUnderTest blockchain provisioning
type SystemUnderTest struct {
	execBinary         string
	blockListener      *EventListener
	currentHeight      atomic.Int64
	outputDir          string
	testnetInitializer TestnetInitializer

	// blockTime is the expected/desired block time. This is not going to be very precise
	// since Tendermint consensus does not allow specifying it directly.
	blockTime         time.Duration
	rpcAddr           string
	initialNodesCount int
	nodesCount        int
	minGasPrice       string
	cleanupFn         []CleanupFn
	outBuff           *ring.Ring
	errBuff           *ring.Ring
	out               io.Writer
	verbose           bool
	ChainStarted      bool
	projectName       string
	dirty             bool // requires full reset when marked dirty

	pidsLock sync.RWMutex
	pids     map[int]struct{}
	chainID  string
}

func NewSystemUnderTest(execBinary string, verbose bool, nodesCount int, blockTime time.Duration, initer ...TestnetInitializer) *SystemUnderTest {
	if execBinary == "" {
		panic("executable binary name must not be empty")
	}
	nameTokens := ExecBinaryUnversionedRegExp.FindAllString(execBinary, 1)
	if len(nameTokens) == 0 || nameTokens[0] == "" {
		panic("failed to parse project name from binary")
	}

	execBinary = filepath.Join(WorkDir, "binaries", execBinary)
	s := &SystemUnderTest{
		chainID:           "testing",
		execBinary:        execBinary,
		outputDir:         "./testnet",
		blockTime:         blockTime,
		rpcAddr:           "tcp://localhost:26657",
		initialNodesCount: nodesCount,
		outBuff:           ring.New(100),
		errBuff:           ring.New(100),
		out:               os.Stdout,
		verbose:           verbose,
		minGasPrice:       fmt.Sprintf("0.000001%s", sdk.DefaultBondDenom),
		projectName:       nameTokens[0],
		pids:              make(map[int]struct{}, nodesCount),
	}
	s.testnetInitializer = NewSingleHostTestnetCmdInitializer(execBinary, WorkDir, s.chainID, s.outputDir, s.initialNodesCount, s.minGasPrice, s.CommitTimeout(), s.Log)
	return s
}

func (s *SystemUnderTest) SetExecBinary(binary string) {
	s.execBinary = binary
}

func (s *SystemUnderTest) SetTestnetInitializer(testnetInitializer TestnetInitializer) {
	s.testnetInitializer = testnetInitializer
}

// CommitTimeout returns the max time to wait for a commit. Default to 90% of block time
func (s *SystemUnderTest) CommitTimeout() time.Duration {
	// The commit timeout is a lower bound for the block time. We try to set it to a level that allows us to reach the expected block time.
	return time.Duration((int64(s.blockTime) * 90) / 100) // leave 10% for all other operations
}

func (s *SystemUnderTest) SetupChain() {
	s.Logf("Setup chain: %s\n", s.outputDir)
	if err := os.RemoveAll(filepath.Join(WorkDir, s.outputDir)); err != nil {
		panic(err.Error())
	}
	s.testnetInitializer.Initialize()
	s.nodesCount = s.initialNodesCount

	// modify genesis with system test defaults
	src := filepath.Join(WorkDir, s.nodePath(0), "config", "genesis.json")
	genesisBz, err := os.ReadFile(src) // #nosec G304
	if err != nil {
		panic(fmt.Sprintf("failed to load genesis: %s", err))
	}

	genesisBz, err = sjson.SetRawBytes(genesisBz, "consensus.params.block.max_gas", []byte(fmt.Sprintf(`"%d"`, 10_000_000)))
	if err != nil {
		panic(fmt.Sprintf("failed set block max gas: %s", err))
	}
	s.withEachNodeHome(func(i int, home string) {
		if err := saveGenesis(home, genesisBz); err != nil {
			panic(err)
		}
	})

	// backup genesis
	dest := filepath.Join(WorkDir, s.nodePath(0), "config", "genesis.json.orig")
	if _, err := copyFile(src, dest); err != nil {
		panic(fmt.Sprintf("copy failed :%#+v", err))
	}
	// backup keyring
	src = filepath.Join(WorkDir, s.nodePath(0), "keyring-test")
	dest = filepath.Join(WorkDir, s.outputDir, "keyring-test")
	if err := copyFilesInDir(src, dest); err != nil {
		panic(fmt.Sprintf("copy files from dir :%#+v", err))
	}
}

func (s *SystemUnderTest) StartChain(t *testing.T, xargs ...string) {
	t.Helper()
	s.Log("Start chain\n")
	s.ChainStarted = true
	s.startNodesAsync(t, append([]string{"start", "--log_level=info", "--log_no_color"}, xargs...)...)

	s.AwaitNodeUp(t, s.rpcAddr)

	t.Log("Start new block listener")
	s.blockListener = NewEventListener(t, s.rpcAddr)
	s.cleanupFn = append(s.cleanupFn,
		s.blockListener.Subscribe("tm.event='NewBlock'", func(e ctypes.ResultEvent) (more bool) {
			newBlock, ok := e.Data.(tmtypes.EventDataNewBlock)
			require.True(t, ok, "unexpected type %T", e.Data)
			s.currentHeight.Store(newBlock.Block.Height)
			return true
		}),
	)
	s.AwaitNextBlock(t, 4e9)
}

// MarkDirty whole chain will be reset when marked dirty
func (s *SystemUnderTest) MarkDirty() {
	s.dirty = true
}

// IsDirty true when non default genesis or other state modification were applied that might create incompatibility for tests
func (s *SystemUnderTest) IsDirty() bool {
	return s.dirty
}

// watchLogs stores stdout/stderr in a file and in a ring buffer to output the last n lines on test error
func (s *SystemUnderTest) watchLogs(node int, cmd *exec.Cmd) {
	logfile, err := os.Create(filepath.Join(WorkDir, s.outputDir, fmt.Sprintf("node%d.out", node)))
	if err != nil {
		panic(fmt.Sprintf("open logfile error %#+v", err))
	}

	errReader, err := cmd.StderrPipe()
	if err != nil {
		panic(fmt.Sprintf("stderr reader error %#+v", err))
	}
	stopRingBuffer := make(chan struct{})
	go appendToBuf(io.TeeReader(errReader, logfile), s.errBuff, stopRingBuffer)

	outReader, err := cmd.StdoutPipe()
	if err != nil {
		panic(fmt.Sprintf("stdout reader error %#+v", err))
	}
	go appendToBuf(io.TeeReader(outReader, logfile), s.outBuff, stopRingBuffer)
	s.cleanupFn = append(s.cleanupFn, func() {
		close(stopRingBuffer)
		_ = logfile.Close()
	})
}

func appendToBuf(r io.Reader, b *ring.Ring, stop <-chan struct{}) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		select {
		case <-stop:
			return
		default:
		}
		text := scanner.Text()
		// filter out noise
		if isLogNoise(text) {
			continue
		}
		b.Value = text
		b = b.Next()
	}
}

func isLogNoise(text string) bool {
	for _, v := range []string{
		"\x1b[36mmodule=\x1b[0mrpc-server", // "module=rpc-server",
	} {
		if strings.Contains(text, v) {
			return true
		}
	}
	return false
}

// AwaitUpgradeInfo blocks util an upgrade info file is persisted to disk
func (s *SystemUnderTest) AwaitUpgradeInfo(t *testing.T) {
	t.Helper()
	var found bool
	for !found {
		s.withEachNodeHome(func(i int, home string) {
			_, err := os.Stat(filepath.Join(s.nodePath(0), "data", "upgrade-info.json"))
			switch {
			case err == nil:
				found = true
			case !os.IsNotExist(err):
				t.Fatal(err.Error())
			}
		})
		time.Sleep(s.blockTime / 2)
	}
}

func (s *SystemUnderTest) AwaitChainStopped() {
	for s.anyNodeRunning() {
		time.Sleep(s.blockTime)
	}
}

// AwaitNodeUp ensures the node is running
func (s *SystemUnderTest) AwaitNodeUp(t *testing.T, rpcAddr string) {
	t.Helper()
	t.Logf("Await node is up: %s", rpcAddr)
	timeout := DefaultWaitTime
	ctx, done := context.WithTimeout(context.Background(), timeout)
	defer done()

	started := make(chan struct{})
	go func() { // query for a non empty block on status page
		t.Logf("Checking node status: %s\n", rpcAddr)
		for {
			con, err := client.New(rpcAddr, "/websocket")
			if err != nil || con.Start() != nil {
				time.Sleep(time.Second)
				continue
			}
			result, err := con.Status(ctx)
			if err != nil || result.SyncInfo.LatestBlockHeight < 1 {
				_ = con.Stop()
				continue
			}
			t.Logf("Node started. Current block: %d\n", result.SyncInfo.LatestBlockHeight)
			_ = con.Stop()
			started <- struct{}{}
		}
	}()
	select {
	case <-started:
	case <-ctx.Done():
		require.NoError(t, ctx.Err())
	case <-time.NewTimer(timeout).C:
		t.Fatalf("timeout waiting for node start: %s", timeout)
	}
}

// StopChain stops the system under test and executes all registered cleanup callbacks
func (s *SystemUnderTest) StopChain() {
	s.Log("Stop chain\n")
	if !s.ChainStarted {
		return
	}

	for _, c := range s.cleanupFn {
		c()
	}
	s.cleanupFn = nil
	// send SIGTERM
	s.withEachPid(func(p *os.Process) {
		go func() {
			if err := p.Signal(syscall.SIGTERM); err != nil {
				s.Logf("failed to stop node with pid %d: %s\n", p.Pid, err)
			}
		}()
	})
	// give some final time to shut down
	s.withEachPid(func(p *os.Process) {
		time.Sleep(200 * time.Millisecond)
	})
	// goodbye
	for ; s.anyNodeRunning(); time.Sleep(100 * time.Millisecond) {
		s.withEachPid(func(p *os.Process) {
			s.Logf("killing node %d\n", p.Pid)
			if err := p.Kill(); err != nil {
				s.Logf("failed to kill node with pid %d: %s\n", p.Pid, err)
			}
		})
	}
	s.ChainStarted = false
}

func (s *SystemUnderTest) withEachPid(cb func(p *os.Process)) {
	s.pidsLock.RLock()
	pids := maps.Keys(s.pids)
	s.pidsLock.RUnlock()

	for pid := range pids {
		p, err := os.FindProcess(pid)
		if err != nil {
			continue
		}
		cb(p)
	}
}

// PrintBuffer prints the chain logs to the console
func (s *SystemUnderTest) PrintBuffer() {
	s.outBuff.Do(func(v interface{}) {
		if v != nil {
			_, _ = fmt.Fprintf(s.out, "out> %s\n", v)
		}
	})
	fmt.Fprint(s.out, "8< chain err -----------------------------------------\n")
	s.errBuff.Do(func(v interface{}) {
		if v != nil {
			_, _ = fmt.Fprintf(s.out, "err> %s\n", v)
		}
	})
}

// AwaitBlockHeight blocks until te target height is reached. An optional timeout parameter can be passed to abort early
func (s *SystemUnderTest) AwaitBlockHeight(t *testing.T, targetHeight int64, timeout ...time.Duration) {
	t.Helper()
	require.Greater(t, targetHeight, s.currentHeight.Load())
	var maxWaitTime time.Duration
	if len(timeout) != 0 {
		maxWaitTime = timeout[0]
	} else {
		maxWaitTime = time.Duration(targetHeight-s.currentHeight.Load()+3) * s.blockTime
	}
	abort := time.NewTimer(maxWaitTime).C
	for {
		select {
		case <-abort:
			t.Fatalf("Timeout - block %d not reached within %s", targetHeight, maxWaitTime)
			return
		default:
			if current := s.AwaitNextBlock(t); current >= targetHeight {
				return
			}
		}
	}
}

// AwaitNextBlock is a first class function that any caller can use to ensure a new block was minted.
// Returns the new height
func (s *SystemUnderTest) AwaitNextBlock(t *testing.T, timeout ...time.Duration) int64 {
	t.Helper()
	maxWaitTime := s.blockTime * 3
	if len(timeout) != 0 { // optional argument to overwrite default timeout
		maxWaitTime = timeout[0]
	}
	done := make(chan int64)
	go func() {
		for start, current := s.currentHeight.Load(), s.currentHeight.Load(); current == start; current = s.currentHeight.Load() {
			time.Sleep(s.blockTime)
		}
		done <- s.currentHeight.Load()
		close(done)
	}()
	select {
	case v := <-done:
		return v
	case <-time.NewTimer(maxWaitTime).C:
		t.Fatalf("Timeout - no block within %s", maxWaitTime)
		return -1
	}
}

// ResetDirtyChain reset chain when non default setup or state (dirty)
func (s *SystemUnderTest) ResetDirtyChain(t *testing.T) {
	t.Helper()
	if s.IsDirty() {
		s.ResetChain(t)
	}
}

// ResetChain stops and clears all nodes state via 'unsafe-reset-all'
func (s *SystemUnderTest) ResetChain(t *testing.T) {
	t.Helper()
	t.Log("Reset chain")
	s.StopChain()
	restoreOriginalGenesis(t, s)
	restoreOriginalKeyring(t, s)
	s.resetBuffers()

	// remove all additional nodes
	for i := s.initialNodesCount; i < s.nodesCount; i++ {
		_ = os.RemoveAll(filepath.Join(WorkDir, s.nodePath(i)))
		_ = os.Remove(filepath.Join(WorkDir, s.outputDir, fmt.Sprintf("node%d.out", i)))
	}
	s.nodesCount = s.initialNodesCount

	// reset all validator nodes
	s.ForEachNodeExecAndWait(t, []string{"comet", "unsafe-reset-all"})
	s.currentHeight.Store(0)
	s.dirty = false
}

// ModifyGenesisCLI executes the CLI commands to modify the genesis
func (s *SystemUnderTest) ModifyGenesisCLI(t *testing.T, cmds ...[]string) {
	t.Helper()
	s.ForEachNodeExecAndWait(t, cmds...)
	s.MarkDirty()
}

type GenesisMutator func([]byte) []byte

// ModifyGenesisJSON resets the chain and executes the callbacks to update the json representation
// The mutator callbacks after each other receive the genesis as raw bytes and return the updated genesis for the next.
// example:
//
//	return func(genesis []byte) []byte {
//		val, _ := json.Marshal(sdk.NewDecCoins(fees...))
//		state, _ := sjson.SetRawBytes(genesis, "app_state.globalfee.params.minimum_gas_prices", val)
//		return state
//	}
func (s *SystemUnderTest) ModifyGenesisJSON(t *testing.T, mutators ...GenesisMutator) {
	t.Helper()
	s.ResetChain(t)
	s.modifyGenesisJSON(t, mutators...)
}

// modify json without enforcing a reset
func (s *SystemUnderTest) modifyGenesisJSON(t *testing.T, mutators ...GenesisMutator) {
	t.Helper()
	require.Empty(t, s.currentHeight.Load(), "forced chain reset required")
	current, err := os.ReadFile(filepath.Join(WorkDir, s.nodePath(0), "config", "genesis.json"))
	require.NoError(t, err)
	for _, m := range mutators {
		current = m(current)
	}
	out := storeTempFile(t, current)
	defer os.Remove(out.Name())
	s.setGenesis(t, out.Name())
	s.MarkDirty()
}

// ReadGenesisJSON returns current genesis.json content as raw string
func (s *SystemUnderTest) ReadGenesisJSON(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(WorkDir, s.nodePath(0), "config", "genesis.json"))
	require.NoError(t, err)
	return string(content)
}

// setGenesis copy genesis file to all nodes
func (s *SystemUnderTest) setGenesis(t *testing.T, srcPath string) {
	t.Helper()
	in, err := os.Open(srcPath)
	require.NoError(t, err)
	defer in.Close()
	var buf bytes.Buffer

	_, err = io.Copy(&buf, in)
	require.NoError(t, err)

	s.withEachNodeHome(func(i int, home string) {
		require.NoError(t, saveGenesis(home, buf.Bytes()))
	})
}

func saveGenesis(home string, content []byte) error {
	out, err := os.Create(filepath.Join(WorkDir, home, "config", "genesis.json"))
	if err != nil {
		return fmt.Errorf("out file: %w", err)
	}
	defer out.Close()

	if _, err = io.Copy(out, bytes.NewReader(content)); err != nil {
		return fmt.Errorf("write out file: %w", err)
	}

	if err = out.Close(); err != nil {
		return fmt.Errorf("close out file: %w", err)
	}
	return nil
}

// ForEachNodeExecAndWait runs the given app executable commands for all cluster nodes synchronously
// The commands output is returned for each node.
func (s *SystemUnderTest) ForEachNodeExecAndWait(t *testing.T, cmds ...[]string) [][]string {
	t.Helper()
	result := make([][]string, s.nodesCount)
	s.withEachNodeHome(func(i int, home string) {
		result[i] = make([]string, len(cmds))
		for j, xargs := range cmds {
			xargs = append(xargs, "--home", home)
			s.Logf("Execute `%s %s`\n", s.execBinary, strings.Join(xargs, " "))
			out := MustRunShellCmd(t, s.execBinary, xargs...)
			s.Logf("Result: %s\n", out)
			result[i][j] = out
		}
	})
	return result
}

func MustRunShellCmd(t *testing.T, cmd string, args ...string) string {
	t.Helper()
	out, err := RunShellCmd(cmd, args...)
	require.NoError(t, err)
	return out
}

func RunShellCmd(cmd string, args ...string) (string, error) {
	c := exec.Command( //nolint:gosec // used by tests only
		locateExecutable(cmd),
		args...,
	)
	c.Dir = WorkDir
	out, err := c.Output()
	if err != nil {
		return string(out), fmt.Errorf("run `%s %s`: out: %s: %w", cmd, strings.Join(args, " "), string(out), err)
	}
	return string(out), nil
}

// startNodesAsync runs the given app cli command for all cluster nodes and returns without waiting
func (s *SystemUnderTest) startNodesAsync(t *testing.T, xargs ...string) {
	t.Helper()
	s.withEachNodeHome(func(i int, home string) {
		args := append(xargs, "--home="+home)
		s.Logf("Execute `%s %s`\n", s.execBinary, strings.Join(args, " "))
		cmd := exec.Command( //nolint:gosec // used by tests only
			locateExecutable(s.execBinary),
			args...,
		)
		cmd.Dir = WorkDir
		s.watchLogs(i, cmd)
		require.NoError(t, cmd.Start(), "node %d", i)
		s.Logf("Node started: %d\n", cmd.Process.Pid)

		// cleanup when stopped
		s.awaitProcessCleanup(cmd)
	})
}

func (s *SystemUnderTest) awaitProcessCleanup(cmd *exec.Cmd) {
	pid := cmd.Process.Pid
	s.pidsLock.Lock()
	s.pids[pid] = struct{}{}
	s.pidsLock.Unlock()
	go func() {
		_ = cmd.Wait() // blocks until shutdown
		s.Logf("Node stopped: %d\n", pid)
		s.pidsLock.Lock()
		delete(s.pids, pid)
		s.pidsLock.Unlock()
	}()
}

func (s *SystemUnderTest) withEachNodeHome(cb func(i int, home string)) {
	for i := 0; i < s.nodesCount; i++ {
		cb(i, s.nodePath(i))
	}
}

// nodePath returns the path of the node within the work dir. not absolute
func (s *SystemUnderTest) nodePath(i int) string {
	return NodePath(i, s.outputDir, s.projectName)
}

func NodePath(n int, outputDir, name string) string {
	return fmt.Sprintf("%s/node%d/%s", outputDir, n, name)
}

func (s *SystemUnderTest) Log(msg string) {
	if s.verbose {
		_, _ = fmt.Fprint(s.out, msg)
	}
}

func (s *SystemUnderTest) Logf(msg string, args ...interface{}) {
	s.Log(fmt.Sprintf(msg, args...))
}

func (s *SystemUnderTest) RPCClient(t *testing.T) RPCClient {
	t.Helper()
	return NewRPCClient(t, s.rpcAddr)
}

func (s *SystemUnderTest) AllPeers(t *testing.T) []string {
	t.Helper()
	result := make([]string, s.nodesCount)
	for i, n := range s.AllNodes(t) {
		result[i] = n.PeerAddr()
	}
	return result
}

func (s *SystemUnderTest) AllNodes(t *testing.T) []Node {
	t.Helper()
	return AllNodes(t, s)
}

func AllNodes(t *testing.T, s *SystemUnderTest) []Node {
	t.Helper()
	result := make([]Node, s.nodesCount)
	outs := s.ForEachNodeExecAndWait(t, []string{"comet", "show-node-id"})
	ip := "127.0.0.1"
	if false { // is there still a use case for external ip?
		var err error
		ip, err = server.ExternalIP()
		require.NoError(t, err)
	}
	for i, out := range outs {
		result[i] = Node{
			ID:      strings.TrimSpace(out[0]),
			IP:      ip,
			RPCPort: 26657 + i, // as defined in testnet command
			P2PPort: 16656 + i, // as defined in testnet command
		}
	}
	return result
}

func (s *SystemUnderTest) resetBuffers() {
	s.outBuff = ring.New(100)
	s.errBuff = ring.New(100)
}

// AddFullnode starts a new fullnode that connects to the existing chain but is not a validator.
func (s *SystemUnderTest) AddFullnode(t *testing.T, beforeStart ...func(nodeNumber int, nodePath string)) Node {
	t.Helper()
	s.MarkDirty()
	s.nodesCount++
	nodeNumber := s.nodesCount - 1
	nodePath := s.nodePath(nodeNumber)
	_ = os.RemoveAll(nodePath) // drop any legacy path, just in case

	// prepare new node
	moniker := fmt.Sprintf("node%d", nodeNumber)
	args := []string{"init", moniker, "--home=" + nodePath, "--overwrite"}
	s.Logf("Execute `%s %s`\n", s.execBinary, strings.Join(args, " "))
	cmd := exec.Command( //nolint:gosec // used by tests only
		locateExecutable(s.execBinary),
		args...,
	)
	cmd.Dir = WorkDir
	s.watchLogs(nodeNumber, cmd)
	require.NoError(t, cmd.Run(), "failed to start node with id %d", nodeNumber)
	require.NoError(t, saveGenesis(nodePath, []byte(s.ReadGenesisJSON(t))))

	configPath := filepath.Join(WorkDir, nodePath, "config")

	// quick hack: copy config and overwrite by start params
	for _, tomlFile := range []string{"config.toml", "app.toml"} {
		configFile := filepath.Join(configPath, tomlFile)
		_ = os.Remove(configFile)
		_, err := copyFile(filepath.Join(WorkDir, s.nodePath(0), "config", tomlFile), configFile)
		require.NoError(t, err)
	}
	// start node
	allNodes := s.AllNodes(t)
	node := allNodes[len(allNodes)-1]
	peers := make([]string, len(allNodes)-1)
	for i, n := range allNodes[0 : len(allNodes)-1] {
		peers[i] = n.PeerAddr()
	}
	for _, c := range beforeStart {
		c(nodeNumber, nodePath)
	}
	args = []string{
		"start",
		"--p2p.persistent_peers=" + strings.Join(peers, ","),
		fmt.Sprintf("--p2p.laddr=tcp://localhost:%d", node.P2PPort),
		fmt.Sprintf("--rpc.laddr=tcp://localhost:%d", node.RPCPort),
		fmt.Sprintf("--grpc.address=localhost:%d", 9090+nodeNumber),
		"--p2p.pex=false",
		"--moniker=" + moniker,
		"--log_level=info",
		"--log_no_color",
		"--home", nodePath,
	}
	s.Logf("Execute `%s %s`\n", s.execBinary, strings.Join(args, " "))
	cmd = exec.Command( //nolint:gosec // used by tests only
		locateExecutable(s.execBinary),
		args...,
	)
	cmd.Dir = WorkDir
	s.watchLogs(nodeNumber, cmd)
	require.NoError(t, cmd.Start(), "node %d", nodeNumber)
	s.awaitProcessCleanup(cmd)
	return node
}

// NewEventListener constructor for Eventlistener with system rpc address
func (s *SystemUnderTest) NewEventListener(t *testing.T) *EventListener {
	t.Helper()
	return NewEventListener(t, s.rpcAddr)
}

// is any process let running?
func (s *SystemUnderTest) anyNodeRunning() bool {
	s.pidsLock.RLock()
	defer s.pidsLock.RUnlock()
	return len(s.pids) != 0
}

func (s *SystemUnderTest) CurrentHeight() int64 {
	return s.currentHeight.Load()
}

type Node struct {
	ID      string
	IP      string
	RPCPort int
	P2PPort int
}

func (n Node) PeerAddr() string {
	return fmt.Sprintf("%s@%s:%d", n.ID, n.IP, n.P2PPort)
}

func (n Node) RPCAddr() string {
	return fmt.Sprintf("tcp://%s:%d", n.IP, n.RPCPort)
}

// locateExecutable looks up the binary on the OS path.
func locateExecutable(file string) string {
	if strings.TrimSpace(file) == "" {
		panic("executable binary name must not be empty")
	}
	path, err := exec.LookPath(file)
	if err != nil {
		panic(fmt.Sprintf("unexpected error with file %q: %s", file, err.Error()))
	}
	if path == "" {
		panic(fmt.Sprintf("%q not found", file))
	}
	return path
}

// EventListener watches for events on the chain
type EventListener struct {
	t      *testing.T
	client *client.HTTP
}

// NewEventListener event listener
func NewEventListener(t *testing.T, rpcAddr string) *EventListener {
	t.Helper()
	httpClient, err := client.New(rpcAddr, "/websocket")
	require.NoError(t, err)
	require.NoError(t, httpClient.Start())
	return &EventListener{client: httpClient, t: t}
}

var DefaultWaitTime = 30 * time.Second

type (
	CleanupFn     func()
	EventConsumer func(e ctypes.ResultEvent) (more bool)
)

// Subscribe to receive events for a topic. Does not block.
// For query syntax See https://docs.cosmos.network/master/core/events.html#subscribing-to-events
func (l *EventListener) Subscribe(query string, cb EventConsumer) func() {
	ctx, done := context.WithCancel(context.Background())
	l.t.Cleanup(done)
	eventsChan, err := l.client.WSEvents.Subscribe(ctx, "testing", query)
	require.NoError(l.t, err)
	cleanup := func() {
		ctx, _ := context.WithTimeout(ctx, DefaultWaitTime)     //nolint:govet // used in cleanup only
		go l.client.WSEvents.Unsubscribe(ctx, "testing", query) //nolint:errcheck // used by tests only
		done()
	}
	go func() {
		for e := range eventsChan {
			if !cb(e) {
				return
			}
		}
	}()
	return cleanup
}

// AwaitQuery blocks and waits for a single result or timeout. This can be used with `broadcast-mode=async`.
// For query syntax See https://docs.cosmos.network/master/core/events.html#subscribing-to-events
func (l *EventListener) AwaitQuery(query string, optMaxWaitTime ...time.Duration) *ctypes.ResultEvent {
	c, result := CaptureSingleEventConsumer()
	maxWaitTime := DefaultWaitTime
	if len(optMaxWaitTime) != 0 {
		maxWaitTime = optMaxWaitTime[0]
	}
	cleanupFn := l.Subscribe(query, TimeoutConsumer(l.t, maxWaitTime, c))
	l.t.Cleanup(cleanupFn)
	return result
}

// TimeoutConsumer is an event consumer decorator with a max wait time. Panics when wait time exceeded without
// a result returned
func TimeoutConsumer(t *testing.T, maxWaitTime time.Duration, next EventConsumer) EventConsumer {
	t.Helper()
	ctx, done := context.WithCancel(context.Background())
	t.Cleanup(done)
	timeout := time.NewTimer(maxWaitTime)
	timedOut := make(chan struct{}, 1)
	go func() {
		select {
		case <-ctx.Done():
		case <-timeout.C:
			timedOut <- struct{}{}
			close(timedOut)
		}
	}()
	return func(e ctypes.ResultEvent) (more bool) {
		select {
		case <-timedOut:
			t.Fatalf("Timeout waiting for new events %s", maxWaitTime)
			return false
		default:
			timeout.Reset(maxWaitTime)
			result := next(e)
			if !result {
				done()
			}
			return result
		}
	}
}

// CaptureSingleEventConsumer consumes one event. No timeout
func CaptureSingleEventConsumer() (EventConsumer, *ctypes.ResultEvent) {
	var result ctypes.ResultEvent
	return func(e ctypes.ResultEvent) (more bool) {
		return false
	}, &result
}

// CaptureAllEventsConsumer is an `EventConsumer` that captures all events until `done()` is called to stop or timeout happens.
// The consumer works async in the background and returns all the captured events when `done()` is called.
// This can be used to verify that certain events have happened.
// Example usage:
//
//		c, done := CaptureAllEventsConsumer(t)
//		query := `tm.event='Tx'`
//		cleanupFn := l.Subscribe(query, c)
//		t.Cleanup(cleanupFn)
//
//	 // do something in your test that create events
//
//		assert.Len(t, done(), 1) // then verify your assumption
func CaptureAllEventsConsumer(t *testing.T, optMaxWaitTime ...time.Duration) (c EventConsumer, done func() []ctypes.ResultEvent) {
	t.Helper()
	maxWaitTime := DefaultWaitTime
	if len(optMaxWaitTime) != 0 {
		maxWaitTime = optMaxWaitTime[0]
	}
	var (
		mu             sync.Mutex
		capturedEvents []ctypes.ResultEvent
		exit           bool
	)
	collectEventsConsumer := func(e ctypes.ResultEvent) (more bool) {
		mu.Lock()
		defer mu.Unlock()
		if exit {
			return false
		}
		capturedEvents = append(capturedEvents, e)
		return true
	}

	return TimeoutConsumer(t, maxWaitTime, collectEventsConsumer), func() []ctypes.ResultEvent {
		mu.Lock()
		defer mu.Unlock()
		exit = true
		return capturedEvents
	}
}

// restoreOriginalGenesis replace nodes genesis by the one created on setup
func restoreOriginalGenesis(t *testing.T, s *SystemUnderTest) {
	t.Helper()
	src := filepath.Join(WorkDir, s.nodePath(0), "config", "genesis.json.orig")
	s.setGenesis(t, src)
}

// restoreOriginalKeyring replaces test keyring with original
func restoreOriginalKeyring(t *testing.T, s *SystemUnderTest) {
	t.Helper()
	dest := filepath.Join(WorkDir, s.outputDir, "keyring-test")
	require.NoError(t, os.RemoveAll(dest))
	for i := 0; i < s.initialNodesCount; i++ {
		src := filepath.Join(WorkDir, s.nodePath(i), "keyring-test")
		require.NoError(t, copyFilesInDir(src, dest))
	}
}

// copyFile copy source file to dest file path
func copyFile(src, dest string) (*os.File, error) {
	in, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return out, err
}

// copyFilesInDir copy files in src dir to dest path
func copyFilesInDir(src, dest string) error {
	err := os.MkdirAll(dest, 0o750)
	if err != nil {
		return fmt.Errorf("mkdirs: %w", err)
	}
	fs, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}
	for _, f := range fs {
		if f.IsDir() {
			continue
		}
		if _, err := copyFile(filepath.Join(src, f.Name()), filepath.Join(dest, f.Name())); err != nil {
			return fmt.Errorf("copy file: %q: %w", f.Name(), err)
		}
	}
	return nil
}

func storeTempFile(t *testing.T, content []byte) *os.File {
	t.Helper()
	out, err := os.CreateTemp(t.TempDir(), "genesis")
	require.NoError(t, err)
	_, err = io.Copy(out, bytes.NewReader(content))
	require.NoError(t, err)
	require.NoError(t, out.Close())
	return out
}
