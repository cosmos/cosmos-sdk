//go:build system_test

package systemtests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func getBinaryNameAndPrefix(isSnapshot bool) (string, string) {
	if sut.execBinary == filepath.Join(WorkDir, "binaries", "simd") {
		if isSnapshot {
			return "simd", "snapshots"
		}
		return "simd", ""
	}
	return "simdv2", "store"
}

func TestSnapshots(t *testing.T) {
	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)
	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, "10000000stake"},
	)

	sut.StartChain(t)

	binary, snapshotPrefix := getBinaryNameAndPrefix(true)
	nodeDir := fmt.Sprintf("./testnet/node0/%s", binary)

	// Wait for chain produce some blocks
	time.Sleep(time.Second * 10)
	// Stop 1 node
	err := sut.StopSingleNode()
	require.NoError(t, err)
	time.Sleep(time.Second)

	// export snapshot at height 5
	res := cli.RunCommandWithArgs(snapshotPrefix, "export", "--height=5", fmt.Sprintf("--home=%s", nodeDir))
	require.Contains(t, res, "Snapshot created at height 5")
	require.DirExists(t, fmt.Sprintf("%s/data/snapshots/5/3", nodeDir))

	// Check snapshots list
	res = cli.RunCommandWithArgs(snapshotPrefix, "list", fmt.Sprintf("--home=%s", nodeDir))
	require.Contains(t, res, "height: 5")

	// Dump snapshot
	res = cli.RunCommandWithArgs(snapshotPrefix, "dump", "5", "3", fmt.Sprintf("--home=%s", nodeDir), fmt.Sprintf("--output=%s/5-3.tar.gz", nodeDir))
	// Check if output file exist
	require.FileExists(t, fmt.Sprintf("%s/5-3.tar.gz", nodeDir))

	// Delete snapshots
	res = cli.RunCommandWithArgs(snapshotPrefix, "delete", "5", "3", fmt.Sprintf("--home=%s", nodeDir))
	require.NoDirExists(t, fmt.Sprintf("%s/data/snapshots/5/3", nodeDir))

	// Load snapshot from file
	res = cli.RunCommandWithArgs(snapshotPrefix, "load", fmt.Sprintf("%s/5-3.tar.gz", nodeDir), fmt.Sprintf("--home=%s", nodeDir))
	require.DirExists(t, fmt.Sprintf("%s/data/snapshots/5/3", nodeDir))

	// Restore from snapshots

	// Remove database
	err = os.RemoveAll(fmt.Sprintf("%s/data/application.db", nodeDir))
	require.NoError(t, err)

	// Only v2 have ss database
	if binary == "simdv2" {
		err = os.RemoveAll(fmt.Sprintf("%s/data/ss", nodeDir))
		require.NoError(t, err)
	}

	res = cli.RunCommandWithArgs(snapshotPrefix, "restore", "5", "3", fmt.Sprintf("--home=%s", nodeDir))
	require.DirExists(t, fmt.Sprintf("%s/data/application.db", nodeDir))
	if binary == "simdv2" {
		require.DirExists(t, fmt.Sprintf("%s/data/ss", nodeDir))
	}

	// Start the node
	sut.StartSingleNode(t, nodeDir)
}

func TestPrune(t *testing.T) {
	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)
	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, "10000000stake"},
	)

	sut.StartChain(t)

	binary, prefix := getBinaryNameAndPrefix(false)
	nodeDir := fmt.Sprintf("./testnet/node0/%s", binary)

	// Wait for chain produce some blocks
	time.Sleep(time.Second * 10)
	// Stop 1 node
	err := sut.StopSingleNode()
	require.NoError(t, err)
	time.Sleep(time.Second)

	// prune
	var res string
	if binary == "simdv2" {
		res = cli.RunCommandWithArgs(prefix, "prune", "--keep-recent=1", fmt.Sprintf("--home=%s", nodeDir))
	} else {
		res = cli.RunCommandWithArgs("prune", "everything", fmt.Sprintf("--home=%s", nodeDir))
	}
	require.Contains(t, res, "successfully pruned the application root multi stores")
	// Start the node
	sut.StartSingleNode(t, nodeDir)
}
