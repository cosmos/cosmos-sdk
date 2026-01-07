package systemtests

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	Sut            *SystemUnderTest
	Verbose        bool
	execBinaryName string
	// Store original configuration for ResetSut
	originalNodesCount int
	originalBlockTime  time.Duration
)

func RunTests(m *testing.M) {
	chainID := flag.String("chain-id", "testing", "chainID")
	waitTime := flag.Duration("wait-time", DefaultWaitTime, "time to wait for chain events")
	nodesCount := flag.Int("nodes-count", 4, "number of nodes in the cluster")
	blockTime := flag.Duration("block-time", 3000*time.Millisecond, "block creation time")
	execBinary := flag.String("binary", "simd", "executable binary for server/ client side")
	bech32Prefix := flag.String("bech32", "cosmos", "bech32 prefix to be used with addresses")
	flag.BoolVar(&Verbose, "verbose", false, "verbose output")
	flag.Parse()

	// fail fast on most common setup issue
	requireEnoughFileHandlers(*nodesCount + 1) // +1 as tests may start another node

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	WorkDir = dir
	if Verbose {
		println("Work dir: ", WorkDir)
	}
	initSDKConfig(*bech32Prefix)

	DefaultWaitTime = *waitTime
	if *execBinary == "" {
		panic("executable binary name must not be empty")
	}
	execBinaryName = *execBinary

	// store original configuration for ResetSut, we do this with global variables for now since Sut is global
	// and we want the same initial configuration when we do a complete reset
	originalNodesCount = *nodesCount
	originalBlockTime = *blockTime

	Sut = NewSystemUnderTest(*execBinary, Verbose, *nodesCount, *blockTime, *chainID)
	Sut.SetupChain() // setup chain and keyring

	// run tests
	exitCode := m.Run()

	// postprocess
	Sut.StopChain()
	if Verbose || exitCode != 0 {
		Sut.PrintBuffer()
		printResultFlag(exitCode == 0)
	}

	os.Exit(exitCode)
}

func GetSystemUnderTest() *SystemUnderTest {
	return Sut
}

func IsVerbose() bool {
	return Verbose
}

func GetExecutableName() string {
	return execBinaryName
}

// requireEnoughFileHandlers uses `ulimit`
func requireEnoughFileHandlers(nodesCount int) {
	ulimit, err := exec.LookPath("ulimit")
	if err != nil || ulimit == "" { // skip when not available
		return
	}

	cmd := exec.Command(ulimit, "-n")
	cmd.Dir = WorkDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("unexpected error :%#+v, output: %s", err, string(out)))
	}
	fileDescrCount, err := strconv.Atoi(strings.Trim(string(out), " \t\n"))
	if err != nil {
		panic(fmt.Sprintf("unexpected error :%#+v, output: %s", err, string(out)))
	}
	expFH := nodesCount * 260 // random number that worked on my box
	if fileDescrCount < expFH {
		panic(fmt.Sprintf("Fail fast. Insufficient setup. Run 'ulimit -n %d'", expFH))
	}
}

func initSDKConfig(bech32Prefix string) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(bech32Prefix, bech32Prefix+sdk.PrefixPublic)
	config.SetBech32PrefixForValidator(bech32Prefix+sdk.PrefixValidator+sdk.PrefixOperator, bech32Prefix+sdk.PrefixValidator+sdk.PrefixOperator+sdk.PrefixPublic)
	config.SetBech32PrefixForConsensusNode(bech32Prefix+sdk.PrefixValidator+sdk.PrefixConsensus, bech32Prefix+sdk.PrefixValidator+sdk.PrefixConsensus+sdk.PrefixPublic)
}

const (
	successFlag = `
 ___ _   _  ___ ___ ___  ___ ___ 
/ __| | | |/ __/ __/ _ \/ __/ __|
\__ \ |_| | (_| (_|  __/\__ \__ \
|___/\__,_|\___\___\___||___/___/`
	failureFlag = `
  __      _ _          _ 
 / _|    (_) |        | |
| |_ __ _ _| | ___  __| |
|  _/ _| | | |/ _ \/ _| |
| || (_| | | |  __/ (_| |
|_| \__,_|_|_|\___|\__,_|`
)

// ResetSut completely resets Sut by deleting all state and creating a fresh instance of Sut.
func ResetSut(t *testing.T) {
	t.Helper()
	// stop current instance if it exists
	if Sut != nil {
		Sut.StopChain()
	}

	// delete entire testnet directory to remove all state
	err := os.RemoveAll(filepath.Join(WorkDir, "testnet"))
	if err != nil {
		t.Fatalf("failed to remove testnet directory: %v", err)
	}

	// create fresh Sut instance with original configuration
	Sut = NewSystemUnderTest(execBinaryName, Verbose, originalNodesCount, originalBlockTime)
	Sut.SetupChain()
}

func printResultFlag(ok bool) {
	if ok {
		fmt.Println(successFlag)
	} else {
		fmt.Println(failureFlag)
	}
}
