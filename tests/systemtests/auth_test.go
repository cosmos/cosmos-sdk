package systemtests

import (
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

	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	newAccount := cli.AddKey("newAccount")

	sut.StartChain(t)

	// run bank tx send with --generate-only flag
	bankSendGenCmd := []string{"tx", "bank", "send", valAddr, newAccount, "1000stake", "--generate-only"}

	var opFile *os.File

	assertGenOnlyOutput := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
		require.Len(t, gotOutputs, 1)
		rsp := gotOutputs[0].(string)
		// get msg from output
		msgs := gjson.Get(rsp, "body.messages").Array()
		require.Len(t, msgs, 1)
		// check from address is equal to account1 address
		fromAddr := gjson.Get(msgs[0].String(), "from_address").String()
		require.Equal(t, valAddr, fromAddr)
		// check to address is equal to account2 address
		toAddr := gjson.Get(msgs[0].String(), "to_address").String()
		require.Equal(t, newAccount, toAddr)

		// write to temp file
		opFile = testutil.WriteToNewTempFile(t, rsp)
		defer opFile.Close()

		return false
	}
	_ = cli.WithRunErrorMatcher(assertGenOnlyOutput).Run(bankSendGenCmd...)
}
