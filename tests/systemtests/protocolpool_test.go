//go:build system_test

package systemtests

import (
	"cosmossdk.io/systemtests"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"testing"
)

const (
	stakingModule      = "staking"
	distributionModule = "distribution"
	protocolPoolModule = "protocolpool"

	stakingToken = "stake"
)

func TestQueryProtocolPool(t *testing.T) {
	// Scenario:
	// delegate tokens to validator
	// check distribution

	sut := systemtests.Sut
	sut.ResetChain(t)

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)

	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, "1000000000000stake"},
	)

	sut.StartChain(t)

	// query validator address to delegate tokens
	rsp := cli.CustomQuery("q", stakingModule, "validators")
	valAddr := gjson.Get(rsp, "validators.#.operator_address").Array()[0].String()

	// stake tokens
	rsp = cli.Run("tx", stakingModule, "delegate", valAddr, "10000000000stake", "--from="+account1Addr, "--fees=1stake")
	systemtests.RequireTxSuccess(t, rsp)

	t.Log(cli.QueryBalance(account1Addr, stakingToken))
	assert.Equal(t, int64(989999999999), cli.QueryBalance(account1Addr, stakingToken))

	rsp = cli.CustomQuery("q", stakingModule, "delegation", account1Addr, valAddr)
	assert.Equal(t, "10000000000", gjson.Get(rsp, "delegation_response.balance.amount").String(), rsp)
	assert.Equal(t, stakingToken, gjson.Get(rsp, "delegation_response.balance.denom").String(), rsp)

	failingCli := cli.WithRunErrorMatcher(func(t assert.TestingT, err error, msgAndArgs ...interface{}) (ok bool) {
		assert.Error(t, err)
		return false
	})
	// query the community pool - should fail for x/distribution
	rsp = failingCli.CustomQuery("q", distributionModule, "community-pool")

	// query will work for x/protocolpool
	rsp = cli.CustomQuery("q", protocolPoolModule, "community-pool")
	assert.True(t, gjson.Get(rsp, "pool.0.amount").Int() > 0, rsp)
	assert.Equal(t, stakingToken, gjson.Get(rsp, "pool.0.denom").String(), rsp)

	t.Log("block height", sut.CurrentHeight(), "\n")

	sut.AwaitNBlocks(t, 5)
	t.Log("block height", sut.CurrentHeight(), "\n")

}
