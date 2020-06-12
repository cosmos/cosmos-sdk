package cli

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestGetCommandDecode(t *testing.T) {
	clientCtx := client.Context{}
	clientCtx = clientCtx.WithTxGenerator(simappparams.MakeEncodingConfig().TxGenerator)

	cmd := GetDecodeCommand(clientCtx)

	viper.Set("hex", false)

	txContents := []byte("{\"type\":\"cosmos-sdk/StdTx\",\"value\":{\"msg\":[{\"type\":\"cosmos-sdk/MsgSend\",\"value\":{\"from_address\":\"cosmos1cxlt8kznps92fwu3j6npahx4mjfutydyene2qw\",\"to_address\":\"cosmos1wc8mpr8m3sy3ap3j7fsgqfzx36um05pystems4\",\"amount\":[{\"denom\":\"stake\",\"amount\":\"10000\"}]}}],\"fee\":{\"amount\":[],\"gas\":\"200000\"},\"signatures\":null,\"memo\":\"\"}}")
	hexStr := hex.EncodeToString(txContents)

	txBytesBase64 := base64.StdEncoding.EncodeToString(txContents)

	fmt.Println("err", hexStr)
	fmt.Println("hex", string(txBytesBase64))

	err := cmd.RunE(cmd, []string{hexStr})
	fmt.Println("err", err)

	require.NoError(t, err)
}
