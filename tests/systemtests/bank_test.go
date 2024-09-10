package systemtests

import (
	"fmt"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestBankSendTxCmd(t *testing.T) {
	// scenario: test bank send command
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)
	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	account2Addr := cli.AddKey("account2")
	require.NotEqual(t, account1Addr, account2Addr)
	denom := "stake"
	initialAmount := 10000000
	initialBalance := fmt.Sprintf("%d%s", initialAmount, denom)
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, initialBalance},
		[]string{"genesis", "add-genesis-account", account2Addr, initialBalance},
	)
	sut.StartChain(t)

	// query accounts balances
	balance1 := cli.QueryBalance(account1Addr, "stake")
	assert.Equal(t, int64(initialAmount), balance1)
	balance2 := cli.QueryBalance(account2Addr, "stake")
	assert.Equal(t, int64(initialAmount), balance2)

	bankSendCmdArgs := []string{"tx", "bank", "send", account1Addr, account2Addr, "1000stake"}

	testCases := []struct {
		name         string
		extraArgs    []string
		expectErr    bool
		expectedCode uint32
	}{
		{
			"valid transaction",
			[]string{"--fees=1stake"},
			false,
			0,
		},
		{
			"not enough fees",
			[]string{"--fees=0stake"},
			true,
			sdkerrors.ErrInsufficientFee.ABCICode(),
		},
		{
			"not enough gas",
			[]string{"--fees=1stake", "--gas=10"},
			true,
			sdkerrors.ErrOutOfGas.ABCICode(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			cmdArgs := append(bankSendCmdArgs, tc.extraArgs...)

			if tc.expectErr {
				assertErr := func(xt assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
					assert.Len(t, gotOutputs, 1)
					code := gjson.Get(gotOutputs[0].(string), "code")
					assert.True(t, code.Exists())
					assert.Equal(t, int64(tc.expectedCode), code.Int())
					return false // always abort
				}
				rsp := cli.WithRunErrorMatcher(assertErr).Run(cmdArgs...)
				RequireTxFailure(t, rsp)
			} else {
				rsp := cli.Run(cmdArgs...)
				txResult, found := cli.AwaitTxCommitted(rsp)
				assert.True(t, found)
				RequireTxSuccess(t, txResult)
			}
		})
	}

	// test tx bank send with insufficient funds
	insufficientCmdArgs := bankSendCmdArgs[0 : len(bankSendCmdArgs)-1]
	insufficientCmdArgs = append(insufficientCmdArgs, initialBalance, "--fees=10stake")
	rsp := cli.Run(insufficientCmdArgs...)
	RequireTxFailure(t, rsp)
	assert.Contains(t, rsp, sdkerrors.ErrInsufficientFunds.Error())

	// test tx bank send with unauthorized signature
	assertUnauthorizedErr := func(t assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
		assert.Len(t, gotOutputs, 1)
		code := gjson.Get(gotOutputs[0].(string), "code")
		assert.True(t, code.Exists())
		assert.Equal(t, int64(sdkerrors.ErrUnauthorized.ABCICode()), code.Int())
		return false
	}
	invalidCli := cli
	invalidCli.chainID = cli.chainID + "a" // set invalid chain-id
	rsp = invalidCli.WithRunErrorMatcher(assertUnauthorizedErr).Run(bankSendCmdArgs...)
	RequireTxFailure(t, rsp)

	// test tx bank send generate only
	assertGenOnlyOutput := func(t assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
		assert.Len(t, gotOutputs, 1)
		rsp := gotOutputs[0].(string)
		// get msg from output
		msgs := gjson.Get(rsp, "body.messages").Array()
		assert.Len(t, msgs, 1)
		// check from address is equal to account1 address
		fromAddr := gjson.Get(msgs[0].String(), "from_address").String()
		assert.Equal(t, account1Addr, fromAddr)
		// check to address is equal to account2 address
		toAddr := gjson.Get(msgs[0].String(), "to_address").String()
		assert.Equal(t, account2Addr, toAddr)
		return false
	}
	genCmdArgs := append(bankSendCmdArgs, "--generate-only")
	_ = cli.WithRunErrorMatcher(assertGenOnlyOutput).Run(genCmdArgs...)

	// test tx bank send with dry-run flag
	assertDryRunOutput := func(t assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
		assert.Len(t, gotOutputs, 1)
		rsp := gotOutputs[0].(string)
		// check gas estimate value found in output
		assert.Contains(t, rsp, "gas estimate")
		return false
	}
	dryRunCmdArgs := append(bankSendCmdArgs, "--dry-run")
	_ = cli.WithRunErrorMatcher(assertDryRunOutput).Run(dryRunCmdArgs...)
}

func TestBankMultiSendTxCmd(t *testing.T) {
	// scenario: test bank multi-send command
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)
	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	account2Addr := cli.AddKey("account2")
	account3Addr := cli.AddKey("account3")
	require.NotEqual(t, account1Addr, account2Addr, account3Addr)
	denom := "stake"
	initialAmount := 10000000
	initialBalance := fmt.Sprintf("%d%s", initialAmount, denom)
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, initialBalance},
		[]string{"genesis", "add-genesis-account", account2Addr, initialBalance},
		[]string{"genesis", "add-genesis-account", account3Addr, initialBalance},
	)
	sut.StartChain(t)

	multiSendCmdArgs := []string{"tx", "bank", "multi-send", account1Addr, account2Addr, account3Addr, "1000stake", "--from=" + account1Addr}

	testCases := []struct {
		name         string
		cmdArgs      []string
		expectErr    bool
		expectedCode uint32
		expErrMsg    string
	}{
		{
			"valid transaction",
			append(multiSendCmdArgs, "--fees=1stake"),
			false,
			0,
			"",
		},
		{
			"not enough arguments",
			[]string{"tx", "bank", "multi-send", account1Addr, account2Addr, "1000stake", "--from=" + account1Addr},
			true,
			0,
			"only received 3",
		},
		{
			"not enough fees",
			append(multiSendCmdArgs, "--fees=0stake"),
			true,
			sdkerrors.ErrInsufficientFee.ABCICode(),
			"insufficient fee",
		},
		{
			"not enough gas",
			append(multiSendCmdArgs, "--fees=1stake", "--gas=10"),
			true,
			sdkerrors.ErrOutOfGas.ABCICode(),
			"out of gas",
		},
		{
			"chain-id shouldn't be used with offline and generate-only flags",
			append(multiSendCmdArgs, "--generate-only", "--offline", "-a=0", "-s=4"),
			true,
			0,
			"chain ID cannot be used",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {

			if tc.expectErr {
				assertErr := func(xt assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
					assert.Len(t, gotOutputs, 1)
					output := gotOutputs[0].(string)
					assert.Contains(t, output, tc.expErrMsg)
					if tc.expectedCode != 0 {
						code := gjson.Get(output, "code")
						assert.True(t, code.Exists())
						assert.Equal(t, int64(tc.expectedCode), code.Int())
					}
					return false // always abort
				}
				_ = cli.WithRunErrorMatcher(assertErr).Run(tc.cmdArgs...)
			} else {
				rsp := cli.Run(tc.cmdArgs...)
				txResult, found := cli.AwaitTxCommitted(rsp)
				assert.True(t, found)
				RequireTxSuccess(t, txResult)
			}
		})
	}
}

func TestBankGRPCQueries(t *testing.T) {
	// scenario: test bank grpc gateway queries
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)
	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	account2Addr := cli.AddKey("account2")
	require.NotEqual(t, account1Addr, account2Addr)
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, "10000000stake"},
	)

	sut.StartChain(t)

}
