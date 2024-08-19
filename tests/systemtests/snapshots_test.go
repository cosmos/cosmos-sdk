//go:build system_test

package systemtests

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const nodeDir = "./testnet/node0/simdv2"

func TestSnapshots(t *testing.T) {
	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)
	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, "10000000stake"},
	)

	sut.StartChain(t)

	// Wait for chain produce some blocks
	time.Sleep(time.Second * 10)
	// Stop 1 node
	err := sut.StopSingleNode()
	require.NoError(t, err)
	time.Sleep(time.Second)	

	// export snapshot at height 5
	res := cli.RunWithArgs("store", "export", "--height=5", fmt.Sprintf("--home=%s", nodeDir))
	require.Contains(t, res, "Snapshot created at height 5")
	require.DirExists(t, fmt.Sprintf("%s/data/snapshots/5/3", nodeDir))
	
	// Check snapshots list
	res = cli.RunWithArgs("store", "list", fmt.Sprintf("--home=%s", nodeDir))
	require.Contains(t, res, "height: 5")

	// Dump snapshot
	res = cli.RunWithArgs("store", "dump", "5", "3", fmt.Sprintf("--home=%s", nodeDir), "--output=./testnet/node0/simdv2/5-3.tar.gz")
	// Check if output file exist
	require.FileExists(t, fmt.Sprintf("%s/5-3.tar.gz", nodeDir))

	// Delete snapshots
	res = cli.RunWithArgs("store", "delete", "5", "3", fmt.Sprintf("--home=%s", nodeDir))
	require.NoDirExists(t, fmt.Sprintf("%s/data/snapshots/5/3", nodeDir))

	// Load snapshot from file
	res = cli.RunWithArgs("store", "load", fmt.Sprintf("%s/5-3.tar.gz", nodeDir), fmt.Sprintf("--home=%s", nodeDir))
	require.DirExists(t, fmt.Sprintf("%s/data/snapshots/5/3", nodeDir))

	// Restore from snapshots

	// Remove sc & ss database
	err = os.RemoveAll(fmt.Sprintf("%s/data/application.db", nodeDir))
	require.NoError(t, err)
	err = os.RemoveAll(fmt.Sprintf("%s/data/ss", nodeDir))
	require.NoError(t, err)

	res = cli.RunWithArgs("store", "restore", "5", "3", fmt.Sprintf("--home=%s", nodeDir))
	require.DirExists(t, fmt.Sprintf("%s/data/application.db", nodeDir))
	require.DirExists(t, fmt.Sprintf("%s/data/ss", nodeDir))

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

	// Wait for chain produce some blocks
	time.Sleep(time.Second * 10)
	// Stop 1 node
	err := sut.StopSingleNode()
	require.NoError(t, err)
	time.Sleep(time.Second)	

	// prune
	res := cli.RunWithArgs("store", "prune", "--keep-recent=1", fmt.Sprintf("--home=%s", nodeDir))
	require.Contains(t, res, "successfully pruned the application root multi stores")

	// Start the node
	sut.StartSingleNode(t, nodeDir)
}	
