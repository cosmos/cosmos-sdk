package cli

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestGetCommandEncode(t *testing.T) {
	clientCtx := client.Context{}
	clientCtx = clientCtx.WithTxGenerator(simappparams.MakeEncodingConfig().TxGenerator)

	cmd := GetEncodeCommand(clientCtx)

	viper.Set(flags.FlagOffline, false)

	testDir, cleanFunc := tests.NewTestCaseDir(t)
	t.Cleanup(cleanFunc)
	txContents := []byte("{\"type\":\"cosmos-sdk/StdTx\",\"value\":{\"msg\":[{\"type\":\"cosmos-sdk/MsgSend\",\"value\":{\"from_address\":\"cosmos1cxlt8kznps92fwu3j6npahx4mjfutydyene2qw\",\"to_address\":\"cosmos1wc8mpr8m3sy3ap3j7fsgqfzx36um05pystems4\",\"amount\":[{\"denom\":\"stake\",\"amount\":\"10000\"}]}}],\"fee\":{\"amount\":[],\"gas\":\"200000\"},\"signatures\":null,\"memo\":\"\"}}")
	txFileName := filepath.Join(testDir, "tx.json")
	err := ioutil.WriteFile(txFileName, txContents, 0644)
	require.NoError(t, err)

	err = cmd.RunE(cmd, []string{txFileName})

	require.NoError(t, err)
}
