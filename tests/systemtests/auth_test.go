package systemtests

import (
	"fmt"
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestAuthSignTxCmd(t *testing.T) {
	// scenario: test auth sign command
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	// get validator addresses
	val1Addr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, val1Addr)

	// get validator addresses
	val2Addr := cli.GetKeyAddr("node1")
	require.NotEmpty(t, val2Addr)

	sut.StartChain(t)

	// run bank tx send with --generate-only flag
	bankSendGenCmd := []string{"tx", "bank", "send", val1Addr, val2Addr, "1000stake", "--generate-only"}

	var opFile *os.File

	assertGenOnlyOutput := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
		require.Len(t, gotOutputs, 1)
		rsp := gotOutputs[0].(string)
		// get msg from output
		msgs := gjson.Get(rsp, "body.messages").Array()
		require.Len(t, msgs, 1)
		// check from address is equal to account1 address
		fromAddr := gjson.Get(msgs[0].String(), "from_address").String()
		require.Equal(t, val1Addr, fromAddr)
		// check to address is equal to account2 address
		toAddr := gjson.Get(msgs[0].String(), "to_address").String()
		require.Equal(t, val2Addr, toAddr)

		// write to temp file
		opFile = testutil.WriteToNewTempFile(t, rsp)
		defer opFile.Close()

		return false
	}
	_ = cli.WithRunErrorMatcher(assertGenOnlyOutput).Run(bankSendGenCmd...)

	// query node0 account details
	rsp := cli.CustomQuery("q", "auth", "account", val1Addr)
	accSeq := gjson.Get(rsp, "account.value.sequence").Int()
	accNum := 1
	signTxCmd := []string{"tx", "sign", opFile.Name(), "--from=" + val1Addr}

	testCases := []struct {
		name      string
		extraArgs []string
	}{
		{
			"valid tx sign",
			[]string{},
		},
		{
			"valid tx sign with offline mode",
			[]string{
				"--offline",
				fmt.Sprintf("--account-number=%d", accNum),
				fmt.Sprintf("--sequence=%d", accSeq),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := append(signTxCmd, tc.extraArgs...)
			assertSignTxOutput := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
				require.Len(t, gotOutputs, 1)
				rsp := gotOutputs[0].(string)

				signatures := gjson.Get(rsp, "signatures").Array()
				require.Len(t, signatures, 1)
				return false
			}

			_ = cli.WithRunErrorMatcher(assertSignTxOutput).Run(cmd...)
		})
	}
}
