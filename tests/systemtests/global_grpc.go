//go:build system_test

package systemtests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	systest "cosmossdk.io/systemtests"
)

func TestGRPCReflection(t *testing.T) {
	// scenario: test grpc reflection
	// given a running chain

	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	// get validator address
	valAddr := cli.GetKeyAddr("node0")

	// add new key
	receiverAddr := cli.AddKey("account1")
	denom := "stake"
	systest.Sut.StartChain(t)

	// query validator balance and make sure it has enough balance
	var transferAmount int64 = 1000
	valBalance := cli.QueryBalance(valAddr, denom)
	require.Greater(t, valBalance, transferAmount, "not enough balance found with validator")

	bankSendCmdArgs := []string{"tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom)}

	// test valid transaction
	rsp := cli.Run(append(bankSendCmdArgs, "--fees=1stake")...)
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	systest.RequireTxSuccess(t, txResult)
	// check valaddr balance equals to valBalance-(transferedAmount+feeAmount)
	require.Equal(t, valBalance-(transferAmount+1), cli.QueryBalance(valAddr, denom))
	// check receiver balance equals to transferAmount
	require.Equal(t, transferAmount, cli.QueryBalance(receiverAddr, denom))

	// test tx bank send with insufficient funds
	insufficientCmdArgs := bankSendCmdArgs[0 : len(bankSendCmdArgs)-1]
	insufficientCmdArgs = append(insufficientCmdArgs, fmt.Sprintf("%d%s", valBalance, denom), "--fees=10stake")
	rsp = cli.Run(insufficientCmdArgs...)
	systest.RequireTxFailure(t, rsp)
	require.Contains(t, rsp, "insufficient funds")

	// test tx bank send with unauthorized signature
	assertUnauthorizedErr := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
		var hasString bool
		for _, output := range gotOutputs {
			if strings.Contains(output.(string), "signature verification failed; please verify account number") {
				hasString = true
				break
			}
		}
		require.True(t, hasString, "outputs doesn't contain: signature verification failed; please verify account number")
		return false
	}
	invalidCli := cli.WithChainID(cli.ChainID() + "a") // set invalid chain-id
	rsp = invalidCli.WithRunErrorMatcher(assertUnauthorizedErr).Run(bankSendCmdArgs...)
	systest.RequireTxFailure(t, rsp)

	// test tx bank send generate only
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
		require.Equal(t, receiverAddr, toAddr)
		return false
	}
	genCmdArgs := append(bankSendCmdArgs, "--generate-only")
	_ = cli.WithRunErrorMatcher(assertGenOnlyOutput).Run(genCmdArgs...)

	// test tx bank send with dry-run flag
	assertDryRunOutput := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
		require.Len(t, gotOutputs, 1)
		rsp := gotOutputs[0].(string)
		// check gas estimate value found in output
		require.Contains(t, rsp, "gas estimate")
		return false
	}
	dryRunCmdArgs := append(bankSendCmdArgs, "--dry-run")
	_ = cli.WithRunErrorMatcher(assertDryRunOutput).Run(dryRunCmdArgs...)
}
