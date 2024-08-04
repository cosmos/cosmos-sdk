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

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
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
	nodePowerBefore := QueryCometValidatorPowerForNode(t, sut, 0)
	require.NotEmpty(t, nodePowerBefore)
	t.Logf("nodePowerBefore: %v", nodePowerBefore)

	var validatorPubKey cryptotypes.PubKey
	newNode := sut.AddFullnode(t, func(nodeNumber int, nodePath string) {
		valKeyFile := filepath.Join(WorkDir, nodePath, "config", "priv_validator_key.json")
		_ = os.Remove(valKeyFile)
		_, err := copyFile(filepath.Join(WorkDir, sut.nodePath(0), "config", "priv_validator_key.json"), valKeyFile)
		require.NoError(t, err)
		validatorPubKey = LoadValidatorPubKeyForNode(t, sut, nodeNumber)
	})
	sut.AwaitNodeUp(t, fmt.Sprintf("http://%s:%d", newNode.IP, newNode.RPCPort))

	// let's wait some blocks to have evidence and update persisted
	rpc := sut.RPCClient(t)
	pkBz := validatorPubKey.Bytes()
	for i := 0; i < 20; i++ {
		sut.AwaitNextBlock(t)
		if QueryCometValidatorPower(rpc, pkBz) == 0 {
			break
		}
	}
	sut.AwaitNextBlock(t)

	// then comet status updated
	nodePowerAfter := QueryCometValidatorPowerForNode(t, sut, 0)
	require.Empty(t, nodePowerAfter)
	t.Logf("nodePowerAfter: %v", nodePowerAfter)

	// and sdk status updated
	byzantineOperatorAddr := cli.GetKeyAddrPrefix("node0", "val")
	rsp = cli.CustomQuery("q", "staking", "validator", byzantineOperatorAddr)
	assert.True(t, gjson.Get(rsp, "validator.jailed").Bool(), rsp)

	t.Log("let's run for some blocks to confirm all good")
	for i := 0; i < 10; i++ {
		sut.AwaitNextBlock(t)
	}
}
