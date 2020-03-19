package cli

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/tests"
)

func TestGetBroadcastCommand_OfflineFlag(t *testing.T) {
	codec := amino.NewCodec()
	cmd := GetBroadcastCommand(codec)

	viper.Set(flags.FlagOffline, true)

	err := cmd.RunE(nil, []string{})
	require.EqualError(t, err, "cannot broadcast tx during offline mode")
}

func TestGetBroadcastCommand_WithoutOfflineFlag(t *testing.T) {
	codec := amino.NewCodec()
	cmd := GetBroadcastCommand(codec)

	viper.Set(flags.FlagOffline, false)

	testDir, cleanFunc := tests.NewTestCaseDir(t)
	t.Cleanup(cleanFunc)

	// Create new file with tx
	txContents := []byte("{\"type\":\"cosmos-sdk/StdTx\",\"value\":{\"msg\":[{\"type\":\"cosmos-sdk/MsgSend\",\"value\":{\"from_address\":\"cosmos1cxlt8kznps92fwu3j6npahx4mjfutydyene2qw\",\"to_address\":\"cosmos1wc8mpr8m3sy3ap3j7fsgqfzx36um05pystems4\",\"amount\":[{\"denom\":\"stake\",\"amount\":\"10000\"}]}}],\"fee\":{\"amount\":[],\"gas\":\"200000\"},\"signatures\":null,\"memo\":\"\"}}")
	txFileName := filepath.Join(testDir, "tx.json")
	err := ioutil.WriteFile(txFileName, txContents, 0644)
	require.NoError(t, err)

	err = cmd.RunE(cmd, []string{txFileName})

	// We test it tries to broadcast but we set unsupported tx to get the error.
	require.EqualError(t, err, "unsupported return type ; supported types: sync, async, block")
}
