package systemtests

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/cometbft/cometbft/v2/privval"
	"github.com/stretchr/testify/require"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// LoadValidatorPubKeyForNode load validator nodes consensus pub key for given node number
func LoadValidatorPubKeyForNode(t *testing.T, sut *SystemUnderTest, nodeNumber int) cryptotypes.PubKey {
	t.Helper()
	return LoadValidatorPubKey(t, filepath.Join(WorkDir, sut.nodePath(nodeNumber), "config", "priv_validator_key.json"))
}

// LoadValidatorPubKey load validator nodes consensus pub key from disk
func LoadValidatorPubKey(t *testing.T, keyFile string) cryptotypes.PubKey {
	t.Helper()
	filePV := privval.LoadFilePVEmptyState(keyFile, "")
	pubKey, err := filePV.GetPubKey()
	require.NoError(t, err)
	valPubKey, err := cryptocodec.FromCmtPubKeyInterface(pubKey)
	require.NoError(t, err)
	return valPubKey
}

// QueryCometValidatorPowerForNode returns the validator's power from tendermint RPC endpoint. 0 when not found
func QueryCometValidatorPowerForNode(t *testing.T, sut *SystemUnderTest, nodeNumber int) int64 {
	t.Helper()
	pubKebBz := LoadValidatorPubKeyForNode(t, sut, nodeNumber).Bytes()
	return QueryCometValidatorPower(sut.RPCClient(t), pubKebBz)
}

func QueryCometValidatorPower(c RPCClient, pubKebBz []byte) int64 {
	for _, v := range c.Validators() {
		if bytes.Equal(v.PubKey.Bytes(), pubKebBz) {
			return v.VotingPower
		}
	}
	return 0
}
