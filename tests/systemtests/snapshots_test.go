//go:build system_test

package systemtests

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	systest "cosmossdk.io/systemtests"
)

const disabledLog = "--log_level=disabled"

func TestSnapshots(t *testing.T) {
	t.Skip("Not persisting properly on CI")

	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	systest.Sut.StartChain(t)

	// Wait for chain produce some blocks
	systest.Sut.AwaitNBlocks(t, 6)
	// Stop all nodes
	systest.Sut.StopChain()

	var (
		command         string
		restoreableDirs []string
	)
	node0Dir := systest.Sut.NodeDir(0)
	if systest.IsV2() {
		command = "store"
		restoreableDirs = []string{fmt.Sprintf("%s/data/application.db", node0Dir), fmt.Sprintf("%s/data/ss", node0Dir)}
	} else {
		command = "snapshots"
		restoreableDirs = []string{fmt.Sprintf("%s/data/application.db", node0Dir)}
	}

	// export snapshot at height 5
	res := cli.RunCommandWithArgs(command, "export", "--height=5", fmt.Sprintf("--home=%s", node0Dir), disabledLog)
	require.Contains(t, res, "Snapshot created at height 5")
	require.DirExists(t, fmt.Sprintf("%s/data/snapshots/5/3", node0Dir))

	// Check snapshots list
	res = cli.RunCommandWithArgs(command, "list", fmt.Sprintf("--home=%s", node0Dir), disabledLog)
	require.Contains(t, res, "height: 5")

	// Dump snapshot
	res = cli.RunCommandWithArgs(command, "dump", "5", "3", fmt.Sprintf("--home=%s", node0Dir), fmt.Sprintf("--output=%s/5-3.tar.gz", node0Dir), disabledLog)
	// Check if output file exist
	require.FileExists(t, fmt.Sprintf("%s/5-3.tar.gz", node0Dir))

	// Delete snapshots
	res = cli.RunCommandWithArgs(command, "delete", "5", "3", fmt.Sprintf("--home=%s", node0Dir), disabledLog)
	require.NoDirExists(t, fmt.Sprintf("%s/data/snapshots/5/3", node0Dir))

	// Load snapshot from file
	res = cli.RunCommandWithArgs(command, "load", fmt.Sprintf("%s/5-3.tar.gz", node0Dir), fmt.Sprintf("--home=%s", node0Dir), disabledLog)
	require.DirExists(t, fmt.Sprintf("%s/data/snapshots/5/3", node0Dir))

	// Restore from snapshots
	for _, dir := range restoreableDirs {
		require.NoError(t, os.RemoveAll(dir))
	}
	// Remove database
	err := os.RemoveAll(fmt.Sprintf("%s/data/application.db", node0Dir))
	require.NoError(t, err)
	if systest.IsV2() {
		require.NoError(t, os.RemoveAll(fmt.Sprintf("%s/data/ss", node0Dir)))
	}

	res = cli.RunCommandWithArgs(command, "restore", "5", "3", fmt.Sprintf("--home=%s", node0Dir), disabledLog)
	for _, dir := range restoreableDirs {
		require.DirExists(t, dir)
	}
}

func TestPrune(t *testing.T) {
	t.Skip("Not persisting properly on CI")

	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	systest.Sut.StartChain(t)

	// Wait for chain produce some blocks
	systest.Sut.AwaitNBlocks(t, 6)

	// Stop all nodes
	systest.Sut.StopChain()

	node0Dir := systest.Sut.NodeDir(0)

	// prune
	var command []string
	if systest.IsV2() {
		command = []string{"store", "prune", "--store.keep-recent=1"}
	} else {
		command = []string{"prune", "everything"}
	}
	res := cli.RunCommandWithArgs(append(command, fmt.Sprintf("--home=%s", node0Dir), disabledLog)...)
	require.Contains(t, res, "successfully pruned the application root multi stores")
}
