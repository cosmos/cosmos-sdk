package systemtests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestAccountCreation(t *testing.T) {
	// scenario: test account creation
	// given a running chain
	// when accountA is sending funds to accountB,
	// AccountB should not be created
	// when accountB is sending funds to accountA,
	// AccountB should be created

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

	// query account1
	rsp := cli.CustomQuery("q", "auth", "account", account1Addr)
	assert.Equal(t, account1Addr, gjson.Get(rsp, "account.value.address").String(), rsp)

	rsp1 := cli.RunAndWait("tx", "bank", "send", account1Addr, account2Addr, "5000stake", "--from="+account1Addr, "--fees=1stake")
	RequireTxSuccess(t, rsp1)

	// query account2
	assertNotFound := func(t assert.TestingT, err error, msgAndArgs ...interface{}) (ok bool) {
		return strings.Contains(err.Error(), "not found: key not found")
	}
	_ = cli.WithRunErrorMatcher(assertNotFound).CustomQuery("q", "auth", "account", account2Addr)

	rsp3 := cli.RunAndWait("tx", "bank", "send", account2Addr, account1Addr, "1000stake", "--from="+account2Addr, "--fees=1stake")
	RequireTxSuccess(t, rsp3)

	// query account2 to make sure its created
	rsp4 := cli.CustomQuery("q", "auth", "account", account2Addr)
	assert.Equal(t, "1", gjson.Get(rsp4, "account.value.sequence").String(), rsp4)
	rsp5 := cli.CustomQuery("q", "auth", "account", account1Addr)
	assert.Equal(t, "1", gjson.Get(rsp5, "account.value.sequence").String(), rsp5)
}
