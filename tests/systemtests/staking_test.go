//go:build system_test

package systemtests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestStakeUnstake(t *testing.T) {
	// Scenario:
	// delegate tokens to validator
	// undelegate some tokens

	sut.ResetChain(t)

	cli := NewCLIWrapper(t, sut, verbose)

	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, "10000000stake"},
	)

	sut.StartChain(t)

	// query validator address to delegate tokens
	rsp := cli.CustomQuery("q", "staking", "validators")
	valAddr := gjson.Get(rsp, "validators.#.operator_address").Array()[0].String()

	// stake tokens
	rsp = cli.RunAndWait("tx", "staking", "delegate", valAddr, "10000stake", "--from="+account1Addr, "--fees=1stake")
	RequireTxSuccess(t, rsp)

	t.Log(cli.QueryBalance(account1Addr, "stake"))
	assert.Equal(t, int64(9989999), cli.QueryBalance(account1Addr, "stake"))

	rsp = cli.CustomQuery("q", "staking", "delegation", account1Addr, valAddr)
	assert.Equal(t, "10000", gjson.Get(rsp, "delegation_response.balance.amount").String(), rsp)
	assert.Equal(t, "stake", gjson.Get(rsp, "delegation_response.balance.denom").String(), rsp)

	// unstake tokens
	rsp = cli.RunAndWait("tx", "staking", "unbond", valAddr, "5000stake", "--from="+account1Addr, "--fees=1stake")
	RequireTxSuccess(t, rsp)

	rsp = cli.CustomQuery("q", "staking", "delegation", account1Addr, valAddr)
	assert.Equal(t, "5000", gjson.Get(rsp, "delegation_response.balance.amount").String(), rsp)
	assert.Equal(t, "stake", gjson.Get(rsp, "delegation_response.balance.denom").String(), rsp)

	rsp = cli.CustomQuery("q", "staking", "unbonding-delegation", account1Addr, valAddr)
	assert.Equal(t, "5000", gjson.Get(rsp, "unbond.entries.#.balance").Array()[0].String(), rsp)
}
