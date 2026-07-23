//go:build system_test

package systemtests

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/tools/systemtests"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestValidatorKeyRotation(t *testing.T) {
	sut := systemtests.Sut
	sut.ResetChain(t)

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)

	sut.StartChain(t)

	oldPubKey := systemtests.LoadValidatorPubKeyForNode(t, sut, 0)
	oldPower := systemtests.QueryCometValidatorPower(sut.RPCClient(t), oldPubKey.Bytes())
	require.Positive(t, oldPower)

	newPrivKey := ed25519.GenPrivKey()
	newPubKey := newPrivKey.PubKey()
	newPubKeyJSON := marshalPubKeyJSON(t, newPubKey)

	// Standby full node holding the new key, online and synced before the
	// rotation. It is not yet in the validator set, so it only observes until
	// CometBFT switches node0's slot to the new key, at which point it signs for
	// node0 with no downtime.
	sut.AddFullnode(t, func(_ int, nodePath string) {
		writeCometValidatorKey(t, nodePath, newPrivKey)
	})
	sut.AwaitNBlocks(t, 2)

	rsp := cli.Run(
		"tx", "staking", "rotate-cons-pub-key", newPubKeyJSON,
		"--from=node0",
		"--fees=1stake",
		"--gas=300000",
	)
	systemtests.RequireTxSuccess(t, rsp)

	sut.AwaitNBlocks(t, stakingtypes.ConsensusUpdateDelay)

	assert.Zero(t, systemtests.QueryCometValidatorPower(sut.RPCClient(t), oldPubKey.Bytes()))
	assert.Equal(t, oldPower, systemtests.QueryCometValidatorPower(sut.RPCClient(t), newPubKey.Bytes()))

	valAddr := cli.GetKeyAddrPrefix("node0", "val")
	rsp = cli.CustomQuery("q", "staking", "validator", valAddr)
	assert.Equal(t, gjson.Get(newPubKeyJSON, "key").String(), gjson.Get(rsp, "validator.consensus_pubkey.value").String(), rsp)

	// The standby now signs node0's slot under the new key; assert it does so
	// with no downtime.
	requireContinuesSigning(t, sut, cli, newPubKeyJSON)

	// continue running the chain to ensure it remains healthy
	target := sut.CurrentHeight() + 2*stakingtypes.ConsensusUpdateDelay + 4
	sut.AwaitBlockHeight(t, target, 40*time.Second)
}

func TestValidatorKeyRotationValidatorOffline(t *testing.T) {
	// Index of the validator that will rotate and then go offline
	const rotNode = 1

	sut := systemtests.Sut
	sut.ResetChain(t)

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)
	sut.StartChain(t)

	oldPubKey := systemtests.LoadValidatorPubKeyForNode(t, sut, rotNode)
	oldPower := systemtests.QueryCometValidatorPower(sut.RPCClient(t), oldPubKey.Bytes())
	require.Positive(t, oldPower)

	newPubKey := ed25519.GenPrivKey().PubKey()
	newPubKeyJSON := marshalPubKeyJSON(t, newPubKey)
	rsp := cli.Run(
		"tx", "staking", "rotate-cons-pub-key", marshalPubKeyJSON(t, newPubKey),
		"--from=node1",
		"--fees=1stake",
		"--gas=300000",
	)
	systemtests.RequireTxSuccess(t, rsp)

	// Pause the rotating validator before the swap and keep it down across the
	// swap and the blocks after. The remaining 3 nodes keep quorum and advance.
	require.NoError(t, sut.PauseNodes(rotNode))
	t.Cleanup(func() { _ = sut.ResumeNodes(rotNode) })

	// Ensure the rotation still happened for this validator, even though its
	// offline
	sut.AwaitNBlocks(t, stakingtypes.ConsensusUpdateDelay)

	assert.Zero(t, systemtests.QueryCometValidatorPower(sut.RPCClient(t), oldPubKey.Bytes()))
	assert.Equal(t, oldPower, systemtests.QueryCometValidatorPower(sut.RPCClient(t), newPubKey.Bytes()))

	valAddr := cli.GetKeyAddrPrefix(fmt.Sprintf("%s%d", "node", rotNode), "val")
	rsp = cli.CustomQuery("q", "staking", "validator", valAddr)
	assert.Equal(t, gjson.Get(newPubKeyJSON, "key").String(), gjson.Get(rsp, "validator.consensus_pubkey.value").String(), rsp)

	// continue running the chain to ensure it remains healthy
	target := sut.CurrentHeight() + 2*stakingtypes.ConsensusUpdateDelay + 4
	sut.AwaitBlockHeight(t, target, 60*time.Second)
}

func TestValidatorKeyRotationMulti(t *testing.T) {
	const rotateCount = 3

	sut := systemtests.Sut
	sut.ResetChain(t)

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)
	sut.StartChain(t)

	oldPubKeys := make([]cryptotypes.PubKey, rotateCount)
	newPrivKeys := make([]*ed25519.PrivKey, rotateCount)
	newPubKeys := make([]cryptotypes.PubKey, rotateCount)
	for i := 0; i < rotateCount; i++ {
		oldPubKeys[i] = systemtests.LoadValidatorPubKeyForNode(t, sut, i)
		require.Positive(t, systemtests.QueryCometValidatorPower(sut.RPCClient(t), oldPubKeys[i].Bytes()))
		newPrivKeys[i] = ed25519.GenPrivKey()
		newPubKeys[i] = newPrivKeys[i].PubKey()
	}

	// Bring up a standby full node per rotating validator, each holding that
	// validator's new key. Each standby takes over its slot with no downtime once
	// the swap lands, so the set never loses quorum and this runs on the default
	// node count.
	for i := 0; i < rotateCount; i++ {
		priv := newPrivKeys[i]
		sut.AddFullnode(t, func(_ int, nodePath string) {
			writeCometValidatorKey(t, nodePath, priv)
		})
	}
	sut.AwaitNBlocks(t, 2)

	// Fire the rotations back to back so their windows overlap: at the peak every
	// rotated validator is mid-swap and CometBFT applies the updates across a few
	// adjacent blocks.
	for i := 0; i < rotateCount; i++ {
		rsp := cli.Run(
			"tx", "staking", "rotate-cons-pub-key", marshalPubKeyJSON(t, newPubKeys[i]),
			fmt.Sprintf("--from=node%d", i),
			"--fees=1stake",
			"--gas=300000",
		)
		systemtests.RequireTxSuccess(t, rsp)
	}

	// ensure we produce blocks past the last swap
	target := sut.CurrentHeight() + 2*stakingtypes.ConsensusUpdateDelay + 6
	sut.AwaitBlockHeight(t, target, 90*time.Second)

	// Every rotation applied and each new key signs its slot with no downtime.
	for i := 0; i < rotateCount; i++ {
		assert.Zero(t, systemtests.QueryCometValidatorPower(sut.RPCClient(t), oldPubKeys[i].Bytes()))
		assert.Positive(t, systemtests.QueryCometValidatorPower(sut.RPCClient(t), newPubKeys[i].Bytes()))
		requireContinuesSigning(t, sut, cli, marshalPubKeyJSON(t, newPubKeys[i]))
	}
}

func TestValidatorKeyRotationJailedInWindow(t *testing.T) {
	const rotNode = 1

	sut := systemtests.Sut
	sut.ResetChain(t)

	// Jail on the first missed block past the arming height (min_signed = 1.0).
	sut.ModifyGenesisJSON(t, func(genesis []byte) []byte {
		out, err := sjson.Set(string(genesis), "app_state.slashing.params.signed_blocks_window", "5")
		require.NoError(t, err)
		out, err = sjson.Set(out, "app_state.slashing.params.min_signed_per_window", "1.000000000000000000")
		require.NoError(t, err)
		return []byte(out)
	})

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)
	sut.StartChain(t)

	// Let the signed-blocks window arm before rotating.
	sut.AwaitBlockHeight(t, 12, 60*time.Second)

	oldPubKey := systemtests.LoadValidatorPubKeyForNode(t, sut, rotNode)
	require.Positive(t, systemtests.QueryCometValidatorPower(sut.RPCClient(t), oldPubKey.Bytes()))

	newPubKey := ed25519.GenPrivKey().PubKey()
	rsp := cli.Run(
		"tx", "staking", "rotate-cons-pub-key", marshalPubKeyJSON(t, newPubKey),
		"--from=node1",
		"--fees=1stake",
		"--gas=300000",
	)
	systemtests.RequireTxSuccess(t, rsp)

	// Take the rotating validator offline immediately so downtime jailing removes
	// it from the set while its key swap is still pending.
	require.NoError(t, sut.PauseNodes(rotNode))
	t.Cleanup(func() { _ = sut.ResumeNodes(rotNode) })

	// The remaining validators must keep producing blocks through both the key
	// swap and the jailing; a bad old@0/new@0 emission would halt here.
	target := sut.CurrentHeight() + 12
	sut.AwaitBlockHeight(t, target, 90*time.Second)

	// Confirm the jail actually fired
	valAddr := cli.GetKeyAddrPrefix("node1", "val")
	rsp = cli.CustomQuery("q", "staking", "validator", valAddr)
	require.True(t, gjson.Get(rsp, "validator.jailed").Bool(), rsp)
}

// writeCometValidatorKey overwrites a node's CometBFT priv_validator_key.json
// with the given ed25519 key so the node signs consensus messages under it.
func writeCometValidatorKey(t *testing.T, nodePath string, priv *ed25519.PrivKey) {
	t.Helper()
	pub := priv.PubKey()
	doc := fmt.Sprintf(`{
  "address": "%X",
  "pub_key": {"type": "tendermint/PubKeyEd25519", "value": "%s"},
  "priv_key": {"type": "tendermint/PrivKeyEd25519", "value": "%s"}
}`,
		pub.Address(),
		base64.StdEncoding.EncodeToString(pub.Bytes()),
		base64.StdEncoding.EncodeToString(priv.Bytes()),
	)
	path := filepath.Join(systemtests.WorkDir, nodePath, "config", "priv_validator_key.json")
	require.NoError(t, os.WriteFile(path, []byte(doc), 0o600))
}

// requireContinuesSigning asserts the validator with the given consensus pubkey
// signs every block across a short interval, proving the standby holding the new
// key took over the slot with no downtime.
func requireContinuesSigning(t *testing.T, sut *systemtests.SystemUnderTest, cli *systemtests.CLIWrapper, consPubKeyJSON string) {
	t.Helper()
	signingInfo := func() (missed, indexOffset int64) {
		rsp := cli.CustomQuery("q", "slashing", "signing-info", consPubKeyJSON)
		return gjson.Get(rsp, "val_signing_info.missed_blocks_counter").Int(),
			gjson.Get(rsp, "val_signing_info.index_offset").Int()
	}
	missedBefore, idxBefore := signingInfo()
	sut.AwaitNBlocks(t, 5)
	missedAfter, idxAfter := signingInfo()

	require.Greater(t, idxAfter, idxBefore,
		"new key shows no signing activity (signing-info index not advancing)")
	require.Equal(t, missedBefore, missedAfter,
		"validator missed blocks after rotation; new key is not signing continuously")
}

func marshalPubKeyJSON(t *testing.T, pubKey cryptotypes.PubKey) string {
	t.Helper()

	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	bz, err := cdc.MarshalInterfaceJSON(pubKey)
	require.NoError(t, err)
	return string(bz)
}
