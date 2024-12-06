//go:build system_test

package systemtests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	systest "cosmossdk.io/systemtests"
)

func TestStakeUnstake(t *testing.T) {
	// Scenario:
	// delegate tokens to validator
	// check validator has been updated
	// undelegate some tokens

	systest.Sut.ResetChain(t)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, "10000000stake"},
	)

	systest.Sut.StartChain(t)

	// query validator address to delegate tokens
	rsp := cli.CustomQuery("q", "staking", "validators")
	valAddr := gjson.Get(rsp, "validators.#.operator_address").Array()[0].String()
	valPk := gjson.Get(rsp, "validators.#.consensus_pubkey.value").Array()[0].String()

	// stake tokens
	rsp = cli.RunAndWait("tx", "staking", "delegate", valAddr, "1000000stake", "--from="+account1Addr, "--fees=1stake")
	systest.RequireTxSuccess(t, rsp)

	t.Log(cli.QueryBalance(account1Addr, "stake"))
	assert.Equal(t, int64(8999999), cli.QueryBalance(account1Addr, "stake"))

	// check validator has been updated
	if systest.IsV2() {
		rsp = cli.CustomQuery("q", "comet", "block-results", gjson.Get(rsp, "height").String())
	} else {
		rsp = cli.CustomQuery("q", "block-results", gjson.Get(rsp, "height").String())
	}
	validatorUpdates := gjson.Get(rsp, "validator_updates").Array()
	assert.NotEmpty(t, validatorUpdates)
	vpk := gjson.Get(validatorUpdates[0].String(), "pub_key_bytes").String()
	assert.Equal(t, vpk, valPk)

	rsp = cli.CustomQuery("q", "staking", "delegation", account1Addr, valAddr)
	assert.Equal(t, "1000000", gjson.Get(rsp, "delegation_response.balance.amount").String(), rsp)
	assert.Equal(t, "stake", gjson.Get(rsp, "delegation_response.balance.denom").String(), rsp)

	// unstake tokens
	rsp = cli.RunAndWait("tx", "staking", "unbond", valAddr, "5000stake", "--from="+account1Addr, "--fees=1stake")
	systest.RequireTxSuccess(t, rsp)

	rsp = cli.CustomQuery("q", "staking", "delegation", account1Addr, valAddr)
	assert.Equal(t, "995000", gjson.Get(rsp, "delegation_response.balance.amount").String(), rsp)
	assert.Equal(t, "stake", gjson.Get(rsp, "delegation_response.balance.denom").String(), rsp)

	rsp = cli.CustomQuery("q", "staking", "unbonding-delegation", account1Addr, valAddr)
	assert.Equal(t, "5000", gjson.Get(rsp, "unbond.entries.#.balance").Array()[0].String(), rsp)
}
