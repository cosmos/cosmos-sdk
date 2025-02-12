package systemtests

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	sut            *SystemUnderTest
	verbose        bool
	execBinaryName string
)

func RunTests(m *testing.M) {
	waitTime := flag.Duration("wait-time", DefaultWaitTime, "time to wait for chain events")
	nodesCount := flag.Int("nodes-count", 4, "number of nodes in the cluster")
	blockTime := flag.Duration("block-time", 1000*time.Millisecond, "block creation time")
	execBinary := flag.String("binary", "simd", "executable binary for server/ client side")
	bech32Prefix := flag.String("bech32", "cosmos", "bech32 prefix to be used with addresses")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.Parse()

	// fail fast on most common setup issue
	requireEnoughFileHandlers(*nodesCount + 1) // +1 as tests may start another node

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	WorkDir = dir
	if verbose {
		println("Work dir: ", WorkDir)
	}
	initSDKConfig(*bech32Prefix)

	DefaultWaitTime = *waitTime
	if *execBinary == "" {
		panic("executable binary name must not be empty")
	}
	execBinaryName = *execBinary

	sut = NewSystemUnderTest(*execBinary, verbose, *nodesCount, *blockTime)
	sut.SetupChain() // setup chain and keyring

	// run tests
	exitCode := m.Run()

	// postprocess
	sut.StopChain()
	if verbose || exitCode != 0 {
		sut.PrintBuffer()
		printResultFlag(exitCode == 0)
	}

	os.Exit(exitCode)
}

func GetSystemUnderTest() *SystemUnderTest {
	return sut
}

func IsVerbose() bool {
	return verbose
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

func printResultFlag(ok bool) {
	if ok {
		fmt.Println(successFlag)
	} else {
		fmt.Println(failureFlag)
	}
}
