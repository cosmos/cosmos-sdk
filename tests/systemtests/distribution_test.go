package systemtests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestWithdrawAllRewardsCmd(t *testing.T) {
	// scenario: test distribution withdraw all rewards command
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	newAddr := cli.AddKey("newAddr")
	require.NotEmpty(t, newAddr)

	testDenom := "stake"

	var initialAmount int64 = 10000000
	initialBalance := fmt.Sprintf("%d%s", initialAmount, testDenom)
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", newAddr, initialBalance},
	)
	sut.StartChain(t)

	// query balance
	newAddrBal := cli.QueryBalance(newAddr, testDenom)
	require.Equal(t, initialAmount, newAddrBal)

	// query validator operator address
	rsp := cli.CustomQuery("q", "staking", "validators")
	validators := gjson.Get(rsp, "validators.#.operator_address").Array()
	require.GreaterOrEqual(t, len(validators), 2)
	val1Addr := validators[0].String()
	val2Addr := validators[1].String()

	var delegationAmount int64 = 100000
	delegation := fmt.Sprintf("%d%s", delegationAmount, testDenom)

	// delegate tokens to validator1
	rsp = cli.RunAndWait("tx", "staking", "delegate", val1Addr, delegation, "--from="+newAddr, "--fees=1"+testDenom)
	RequireTxSuccess(t, rsp)

	// delegate tokens to validator2
	rsp = cli.RunAndWait("tx", "staking", "delegate", val2Addr, delegation, "--from="+newAddr, "--fees=1"+testDenom)
	RequireTxSuccess(t, rsp)

	// check updated balance: newAddrBal - delegatedBal - fees
	expBal := newAddrBal - (delegationAmount * 2) - 2
	newAddrBal = cli.QueryBalance(newAddr, testDenom)
	require.Equal(t, expBal, newAddrBal)

	withdrawCmdArgs := []string{"tx", "distribution", "withdraw-all-rewards", "--from=" + newAddr, "--fees=1" + testDenom}

	// test with --max-msgs
	testCases := []struct {
		name     string
		maxMsgs  int
		expTxLen int
	}{
		{
			"--max-msgs value is 1",
			1,
			2,
		},
		{
			"--max-msgs value is 2",
			2,
			1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assertGenOnlyOutput := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
				require.Len(t, gotOutputs, 1)
				// gets output combining two objects without any space or new line
				splitOutput := strings.Split(gotOutputs[0].(string), "}{")
				require.Len(t, splitOutput, tc.expTxLen)
				return false
			}
			cmd := append(withdrawCmdArgs, fmt.Sprintf("--max-msgs=%d", tc.maxMsgs), "--generate-only")
			_ = cli.WithRunErrorMatcher(assertGenOnlyOutput).Run(cmd...)
		})
	}

	// test withdraw-all-rewards transaction
	rsp = cli.RunAndWait(withdrawCmdArgs...)
	RequireTxSuccess(t, rsp)
}
