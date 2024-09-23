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

func TestSnapshotsV2(t *testing.T) {
	if !isV2() {
		t.Skip()
	}

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)
	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, "10000000stake"},
	)

	sut.StartChain(t)
	nodeDir := filepath.Join(WorkDir, "testnet", "node0", "simdv2")

	// Wait for chain produce some blocks
	time.Sleep(time.Second * 10)
	// Stop 1 node
	err := sut.StopSingleNode()
	require.NoError(t, err)
	time.Sleep(time.Second * 5)

	// export snapshot at height 5
	res := cli.RunCommandWithArgs("store", "export", "--height=5", fmt.Sprintf("--home=%s", nodeDir))
	require.Contains(t, res, "Snapshot created at height 5")
	require.DirExists(t, fmt.Sprintf("%s/data/snapshots/5/3", nodeDir))

	// Check snapshots list
	res = cli.RunCommandWithArgs("store", "list", fmt.Sprintf("--home=%s", nodeDir))
	require.Contains(t, res, "height: 5")

	// Dump snapshot
	res = cli.RunCommandWithArgs("store", "dump", "5", "3", fmt.Sprintf("--home=%s", nodeDir), fmt.Sprintf("--output=%s/5-3.tar.gz", nodeDir))
	// Check if output file exist
	require.FileExists(t, fmt.Sprintf("%s/5-3.tar.gz", nodeDir))

	// Delete snapshots
	res = cli.RunCommandWithArgs("store", "delete", "5", "3", fmt.Sprintf("--home=%s", nodeDir))
	require.NoDirExists(t, fmt.Sprintf("%s/data/snapshots/5/3", nodeDir))

	// Load snapshot from file
	res = cli.RunCommandWithArgs("store", "load", fmt.Sprintf("%s/5-3.tar.gz", nodeDir), fmt.Sprintf("--home=%s", nodeDir))
	require.DirExists(t, fmt.Sprintf("%s/data/snapshots/5/3", nodeDir))

	// Restore from snapshots

	// Remove database
	err = os.RemoveAll(fmt.Sprintf("%s/data/application.db", nodeDir))
	require.NoError(t, err)
	err = os.RemoveAll(fmt.Sprintf("%s/data/ss", nodeDir))
	require.NoError(t, err)

	res = cli.RunCommandWithArgs("store", "restore", "5", "3", fmt.Sprintf("--home=%s", nodeDir))
	require.DirExists(t, fmt.Sprintf("%s/data/application.db", nodeDir))
	require.DirExists(t, fmt.Sprintf("%s/data/ss", nodeDir))
}

func TestPruneV2(t *testing.T) {
	if !isV2() {
		t.Skip()
	}
	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)
	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, "10000000stake"},
	)

	sut.StartChain(t)
	nodeDir := filepath.Join(WorkDir, "testnet", "node0", "simdv2")

	// Wait for chain produce some blocks
	time.Sleep(time.Second * 10)
	// Stop 1 node
	err := sut.StopSingleNode()
	require.NoError(t, err)
	time.Sleep(time.Second)

	// prune
	res := cli.RunCommandWithArgs("store", "prune", "--keep-recent=1", fmt.Sprintf("--home=%s", nodeDir))
	require.Contains(t, res, "successfully pruned the application root multi stores")
}
