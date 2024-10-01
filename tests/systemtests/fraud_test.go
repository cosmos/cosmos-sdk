//go:build system_test

package systemtests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestValidatorDoubleSign(t *testing.T) {
	// Scenario:
	//   given: a running chain
	//   when: a second instance with the same val key signs a block
	//   then: the validator is removed from the active set and jailed forever
	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)
	sut.StartChain(t)

	// Check the validator is in the active set
	rsp := cli.CustomQuery("q", "staking", "validators")
	t.Log(rsp)
	validatorPubKey := LoadValidatorPubKeyForNode(t, sut, 0)
	rpc, pkBz := sut.RPCClient(t), validatorPubKey.Bytes()

	nodePowerBefore := QueryCometValidatorPower(rpc, pkBz)
	require.NotEmpty(t, nodePowerBefore)
	t.Logf("nodePowerBefore: %v", nodePowerBefore)

	newNode := sut.AddFullnode(t, func(nodeNumber int, nodePath string) {
		valKeyFile := filepath.Join(WorkDir, nodePath, "config", "priv_validator_key.json")
		_ = os.Remove(valKeyFile)
		_ = MustCopyFile(filepath.Join(WorkDir, sut.nodePath(0), "config", "priv_validator_key.json"), valKeyFile)
	})
	sut.AwaitNodeUp(t, fmt.Sprintf("http://%s:%d", newNode.IP, newNode.RPCPort))

	// let's wait some blocks to have evidence and update persisted
	var nodePowerAfter int64 = -1
	for i := 0; i < 30; i++ {
		sut.AwaitNextBlock(t)
		if nodePowerAfter = QueryCometValidatorPower(rpc, pkBz); nodePowerAfter == 0 {
			break
		}
		t.Logf("wait %d", sut.CurrentHeight())
	}
	// then comet status updated
	require.Empty(t, nodePowerAfter)

	// and sdk status updated
	byzantineOperatorAddr := cli.GetKeyAddrPrefix("node0", "val")
	rsp = cli.CustomQuery("q", "staking", "validator", byzantineOperatorAddr)
	assert.True(t, gjson.Get(rsp, "validator.jailed").Bool(), rsp)

	// let's run for some blocks to confirm all good
	sut.AwaitNBlocks(t, 5)
}
