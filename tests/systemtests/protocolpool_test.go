//go:build system_test

package systemtests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"cosmossdk.io/math"
	"cosmossdk.io/systemtests"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
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

	t.Run("check x/distribution query does not work when using x/protocolpool", func(t *testing.T) {
		failingCli := cli.WithRunErrorMatcher(func(t assert.TestingT, err error, msgAndArgs ...interface{}) (ok bool) {
			assert.Error(t, err)
			return false
		})
		// query the community pool - should fail for x/distribution
		rsp = failingCli.CustomQuery("q", distributionModule, "community-pool")
	})

	t.Run("check x/protocolpool community pool query", func(t *testing.T) {
		// query will work for x/protocolpool
		rsp = cli.CustomQuery("q", protocolPoolModule, "community-pool")
		poolAmount := gjson.Get(rsp, "pool.0.amount").Int()
		assert.True(t, poolAmount > 0, rsp)
		assert.Equal(t, stakingToken, gjson.Get(rsp, "pool.0.denom").String(), rsp)

		t.Log("block height", sut.CurrentHeight(), "\n")
		sut.AwaitNBlocks(t, 1)
		t.Log("block height", sut.CurrentHeight(), "\n")

		rsp = cli.CustomQuery("q", protocolPoolModule, "community-pool")
		newPoolAmount := gjson.Get(rsp, "pool.0.amount").Int()
		assert.Equal(t, stakingToken, gjson.Get(rsp, "pool.0.denom").String(), rsp)
		// check that staking is continually rewarded
		assert.True(t, newPoolAmount > poolAmount, rsp)
	})

	// fund the community pool and query
}

func modifyGovParams(t *testing.T) {
	t.Helper()
	// set up params so that we should just auto pass
	systemtests.Sut.ModifyGenesisJSON(t,
		func(genesis []byte) []byte {
			state, err := sjson.Set(string(genesis), "app_state.gov.params.max_deposit_period", (1 * time.Second).String())
			require.NoError(t, err)
			return []byte(state)
		},
		func(genesis []byte) []byte {
			state, err := sjson.Set(string(genesis), "app_state.gov.params.voting_period", (11 * time.Second).String())
			require.NoError(t, err)
			return []byte(state)
		},
		func(genesis []byte) []byte {
			state, err := sjson.Set(string(genesis), "app_state.gov.params.veto_threshold", "0.000001")
			require.NoError(t, err)
			return []byte(state)
		},
		func(genesis []byte) []byte {
			state, err := sjson.Set(string(genesis), "app_state.gov.params.threshold", "0.0000001")
			require.NoError(t, err)
			return []byte(state)
		},
		func(genesis []byte) []byte {
			state, err := sjson.Set(string(genesis), "app_state.gov.params.quorum", "0.0000001")
			require.NoError(t, err)
			return []byte(state)
		},
	)
}

func TestBudget(t *testing.T) {
	// given a running chain

	systemtests.Sut.ResetChain(t)
	cli := systemtests.NewCLIWrapper(t, systemtests.Sut, systemtests.Verbose)

	// set up gov params so we can pass props quickly
	modifyGovParams(t)

	// get validator address
	valAddr := gjson.Get(cli.Keys("keys", "list"), "0.address").String()
	require.NotEmpty(t, valAddr)

	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	systemtests.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, "1000000000000stake"},
	)

	systemtests.Sut.StartChain(t)

	// get gov module address
	resp := cli.CustomQuery("q", "auth", "module-account", "gov")
	govAddress := gjson.Get(resp, "account.value.address").String()
	_, bz, err := bech32.DecodeAndConvert(govAddress)
	assert.NoError(t, err)
	govAddress, err = bech32.ConvertAndEncode(sdk.Bech32MainPrefix, bz)
	assert.NoError(t, err)

	t.Run("valid proposal", func(t *testing.T) {
		// Create a valid new proposal JSON.
		validProp := fmt.Sprintf(`
{
	"messages": [
		{
			"@type": "/cosmos.protocolpool.v1.MsgCreateBudget",
			"authority": "%s",
			"recipient_address": "%s",
			"budget_per_tranche": {
  				"denom": "stake",
  				"amount": "10"
			},
			"tranches": 10,
			"period": "%s"
		}
	],
	"title": "My awesome title",
	"summary": "My awesome description",
	"deposit": "%s"
}`,
			govAddress,
			account1Addr,
			1*time.Second,
			sdk.NewCoin(stakingToken, math.NewInt(50000000)),
		)
		validPropFile := systemtests.StoreTempFile(t, []byte(validProp))
		defer validPropFile.Close()

		args := []string{
			"tx", "gov", "submit-proposal",
			validPropFile.Name(),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, valAddr),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(stakingToken, math.NewInt(10))).String()),
		}

		rsp := cli.Run(args...)
		txResult, found := cli.AwaitTxCommitted(rsp)
		require.True(t, found)
		systemtests.RequireTxSuccess(t, txResult)
	})

	t.Run("vote on proposal", func(t *testing.T) {
		// check the proposal
		proposalsResp := cli.CustomQuery("q", "gov", "proposals")
		proposals := gjson.Get(proposalsResp, "proposals.#.id").Array()
		require.NotEmpty(t, proposals)

		rsp := cli.CustomQuery("q", "gov", "proposal", "1")
		status := gjson.Get(rsp, "proposal.status")
		require.Equal(t, "PROPOSAL_STATUS_VOTING_PERIOD", status.String())

		// vote on the proposal
		args := []string{
			"tx", "gov", "vote", "1", "yes",
			fmt.Sprintf("--%s=%s", flags.FlagFrom, valAddr),
		}
		rsp = cli.Run(args...)
		txResult, found := cli.AwaitTxCommitted(rsp)
		require.True(t, found)
		systemtests.RequireTxSuccess(t, txResult)
	})

	time.Sleep(11 * time.Second)
	systemtests.Sut.AwaitNextBlock(t)

	// ensure that vote has passed
	t.Run("ensure that the vote has passed", func(t *testing.T) {
		rsp := cli.CustomQuery("q", "gov", "proposal", "1")
		status := gjson.Get(rsp, "proposal.status")
		require.Equal(t, "PROPOSAL_STATUS_PASSED", status.String())

		// check that the budget exists
		rsp = cli.CustomQuery("q", "protocolpool", "unclaimed-budget", account1Addr)
		tranchesLeft := gjson.Get(rsp, "tranches_left").Int()
		require.Equal(t, int64(10), tranchesLeft)
	})

	time.Sleep(11 * time.Second)
	systemtests.Sut.AwaitNextBlock(t)

	t.Run("claim the budget (wrong address will fail)", func(t *testing.T) {
		// claim the budget (right address fails)
		args := []string{
			"tx", "protocolpool", "claim-budget", valAddr,
			fmt.Sprintf("--%s=%s", flags.FlagFrom, valAddr),
		}
		rsp := cli.Run(args...)
		require.Contains(t, rsp, "no budget found for recipient")
	})

	t.Run("claim the budget (right address passes)", func(t *testing.T) {
		balanceBefore := cli.QueryBalance(account1Addr, stakingToken)

		// claim the budget (right address passes)
		args := []string{
			"tx", "protocolpool", "claim-budget", account1Addr,
			fmt.Sprintf("--%s=%s", flags.FlagFrom, account1Addr),
		}
		rsp := cli.Run(args...)
		txResult, found := cli.AwaitTxCommitted(rsp)
		require.True(t, found)
		systemtests.RequireTxSuccess(t, txResult)

		// check budget is updated (trances should be expired)
		rsp = cli.CustomQuery("q", "protocolpool", "unclaimed-budget", account1Addr)
		tranchesLeft := gjson.Get(rsp, "tranches_left").Int()
		require.Equal(t, int64(0), tranchesLeft)
		claimed := gjson.Get(rsp, "claimed_amount.amount").Int()
		require.Equal(t, int64(100), claimed)

		balanceAfter := cli.QueryBalance(account1Addr, stakingToken)

		// balance should be equal to balanceBefore + claim - fee
		expectedBalance := balanceBefore + claimed - cli.GetFeeAmount(t)[0].Amount.Int64()
		require.Equal(t, expectedBalance, balanceAfter)
	})
}
