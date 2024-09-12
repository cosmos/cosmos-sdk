package systemtests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestExecAuthorizationCmd(t *testing.T) {
	// scenario: test authz exec command
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	// get validator address
	valAddr := gjson.Get(cli.Keys("keys", "list"), "1.address").String()
	require.NotEmpty(t, valAddr)

	// add few more keys
	account1Addr := cli.AddKey("account1")
	account2Addr := cli.AddKey("account2")
	require.NotEqual(t, account1Addr, account2Addr)

	denom := "stake"
	var initialAmount int64 = 10000000
	initialBalance := fmt.Sprintf("%d%s", initialAmount, denom)
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, initialBalance},
	)
	sut.StartChain(t)

	// query accounts balances
	account1Bal := cli.QueryBalance(account1Addr, denom)
	require.Equal(t, initialAmount, account1Bal)
}
