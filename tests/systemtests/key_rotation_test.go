//go:build system_test

package systemtests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

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

	newPubKey := ed25519.GenPrivKey().PubKey()
	newPubKeyJSON := marshalPubKeyJSON(t, newPubKey)

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
