package cli

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/testutil"
)

func TestGetBroadcastCommand_OfflineFlag(t *testing.T) {
	clientCtx := client.Context{}.WithOffline(true)
	clientCtx = clientCtx.WithTxGenerator(simappparams.MakeEncodingConfig().TxGenerator)

	cmd := GetBroadcastCommand(clientCtx)
	cmd.SetOut(ioutil.Discard)
	cmd.SetErr(ioutil.Discard)
	cmd.SetArgs([]string{fmt.Sprintf("--%s=true", flags.FlagOffline), ""})

	require.EqualError(t, cmd.Execute(), "cannot broadcast tx during offline mode")
}

func TestGetBroadcastCommand_WithoutOfflineFlag(t *testing.T) {
	clientCtx := client.Context{}
	clientCtx = clientCtx.WithTxGenerator(simappparams.MakeEncodingConfig().TxGenerator)
	cmd := GetBroadcastCommand(clientCtx)

	testDir, cleanFunc := testutil.NewTestCaseDir(t)
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
