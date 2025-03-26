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

func TestQueryProtocolPool(t *testing.T) {
	// Scenario:
	// delegate tokens to validator
	// check distribution

	sut := systemtests.Sut
	sut.ResetChain(t)

	// set up gov params so we can pass props quickly
	modifyGovParams(t)

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

	// get validator address
	valSigner := gjson.Get(cli.Keys("keys", "list"), "0.address").String()
	require.NotEmpty(t, valSigner)

	// stake tokens
	rsp = cli.Run("tx", stakingModule, "delegate", valAddr, "10000000000stake", "--from="+account1Addr, "--fees=1stake")
	systemtests.RequireTxSuccess(t, rsp)

	beforeBalance := cli.QueryBalance(account1Addr, stakingToken)
	assert.Equal(t, int64(989999999999), beforeBalance)

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

	rsp = cli.CustomQuery("q", protocolPoolModule, "community-pool")
	poolAmount := gjson.Get(rsp, "pool.0.amount").Int()
	assert.True(t, poolAmount > 0, rsp)
	assert.Equal(t, stakingToken, gjson.Get(rsp, "pool.0.denom").String(), rsp)

	// fund the community pool and query
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
			"@type": "/cosmos.protocolpool.v1.MsgCommunityPoolSpend",
			"authority": "%s",
			"recipient": "%s",
			"amount": [{
				"denom": "stake",
  				"amount": "100"
			}]
		}
	],
	"title": "My awesome title",
	"summary": "My awesome description",
	"deposit": "%s"
}`,
			govAddress,
			account1Addr,
			sdk.NewCoin(stakingToken, math.NewInt(50000000)),
		)
		validPropFile := systemtests.StoreTempFile(t, []byte(validProp))
		defer validPropFile.Close()

		args := []string{
			"tx", "gov", "submit-proposal",
			validPropFile.Name(),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, valSigner),
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
			fmt.Sprintf("--%s=%s", flags.FlagFrom, valSigner),
		}
		rsp = cli.Run(args...)
		txResult, found := cli.AwaitTxCommitted(rsp)
		require.True(t, found)
		systemtests.RequireTxSuccess(t, txResult)
	})

	balanceBefore := cli.QueryBalance(account1Addr, stakingToken)

	time.Sleep(11 * time.Second)
	systemtests.Sut.AwaitNextBlock(t)

	// ensure that vote has passed
	t.Run("ensure that the vote has passed", func(t *testing.T) {
		rsp := cli.CustomQuery("q", "gov", "proposal", "1")
		status := gjson.Get(rsp, "proposal.status")
		require.Equal(t, "PROPOSAL_STATUS_PASSED", status.String())

		// check that the funds were distributed
		// should be previous balance plus amount from the pool (100) plus the deposit amount (50000000)
		balanceAfter := cli.QueryBalance(account1Addr, stakingToken)
		require.Equal(t, balanceBefore+100+50000000, balanceAfter)
	})
}

// Create a continuous fund
// - submit prop and vote until passed
// Check that funds are distributed and continuous fund is cleaned up once expired
func TestContinuousFunds(t *testing.T) {
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

	expiry := time.Now().Add(20 * time.Second).UTC()

	t.Run("valid proposal", func(t *testing.T) {
		// Create a valid new proposal JSON.
		validProp := fmt.Sprintf(`
{
	"messages": [
		{
			"@type": "/cosmos.protocolpool.v1.MsgCreateContinuousFund",
			"authority": "%s",
			"recipient": "%s",
			"percentage": "0.5",
			"expiry": "%s"
		}
	],
	"title": "My awesome title",
	"summary": "My awesome description",
	"deposit": "%s"
}`,
			govAddress,
			account1Addr,
			expiry.Format(time.RFC3339),
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

	// get balance before any distribution
	balanceBefore := cli.QueryBalance(account1Addr, stakingToken)
	time.Sleep(11 * time.Second)
	systemtests.Sut.AwaitNextBlock(t)

	// ensure that vote has passed
	t.Run("ensure that the vote has passed", func(t *testing.T) {
		rsp := cli.CustomQuery("q", "gov", "proposal", "1")
		status := gjson.Get(rsp, "proposal.status")
		require.Equal(t, "PROPOSAL_STATUS_PASSED", status.String())

		// check that the fund exists
		rsp = cli.CustomQuery("q", "protocolpool", "continuous-fund", account1Addr)
		gotExpiry := gjson.Get(rsp, "continuous_fund.expiry").Time()
		require.Equal(t, expiry.Truncate(time.Second), gotExpiry.Truncate(time.Second))
		recipient := gjson.Get(rsp, "continuous_fund.recipient").String()
		require.Equal(t, account1Addr, recipient)
	})

	// wait long enough that it will be expired
	time.Sleep(11 * time.Second)
	systemtests.Sut.AwaitNextBlock(t)

	t.Run("check balance and that the fund is expired", func(t *testing.T) {
		failingCli := cli.WithRunErrorMatcher(func(t assert.TestingT, err error, msgAndArgs ...interface{}) (ok bool) {
			assert.Error(t, err)
			return false
		})
		// query the continuous fund - should be expired
		_ = failingCli.CustomQuery("q", "protocolpool", "continuous-fund", account1Addr)

		// check that there is nothing in the store
		rsp := cli.CustomQuery("q", "protocolpool", "continuous-funds")
		require.Equal(t, "{}", rsp)

		balanceAfter := cli.QueryBalance(account1Addr, stakingToken)

		// balance should be balance before + 412 (community pool value added * 0.5)
		require.Equal(t, balanceBefore+412, balanceAfter)
	})
}

// Create a continuous fund
// - submit prop and vote until passed (no expiry)
// Create a cancellation prop
//   - submit prop and vote until passed
//
// Check that some funds have been distributed and that the fund is canceled.
func TestCancelContinuousFunds(t *testing.T) {
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

	t.Run("valid proposal - create", func(t *testing.T) {
		// Create a valid new proposal JSON.
		validProp := fmt.Sprintf(`
{
	"messages": [
		{
			"@type": "/cosmos.protocolpool.v1.MsgCreateContinuousFund",
			"authority": "%s",
			"recipient": "%s",
			"percentage": "0.5"
		}
	],
	"title": "My awesome title",
	"summary": "My awesome description",
	"deposit": "%s"
}`,
			govAddress,
			account1Addr,
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

	t.Run("vote on proposal - create", func(t *testing.T) {
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

	// get balance before any distribution
	balanceBefore := cli.QueryBalance(account1Addr, stakingToken)
	time.Sleep(11 * time.Second)
	systemtests.Sut.AwaitNextBlock(t)

	// ensure that vote has passed
	t.Run("ensure that the vote has passed - create", func(t *testing.T) {
		rsp := cli.CustomQuery("q", "gov", "proposal", "1")
		status := gjson.Get(rsp, "proposal.status")
		require.Equal(t, "PROPOSAL_STATUS_PASSED", status.String())

		// check that the fund exists
		rsp = cli.CustomQuery("q", "protocolpool", "continuous-fund", account1Addr)
		recipient := gjson.Get(rsp, "continuous_fund.recipient").String()
		require.Equal(t, account1Addr, recipient)
	})

	t.Run("valid proposal - cancel", func(t *testing.T) {
		// Create a valid new proposal JSON.
		validProp := fmt.Sprintf(`
{
	"messages": [
		{
			"@type": "/cosmos.protocolpool.v1.MsgCancelContinuousFund",
			"authority": "%s",
			"recipient": "%s"
		}
	],
	"title": "My awesome title",
	"summary": "My awesome description",
	"deposit": "%s"
}`,
			govAddress,
			account1Addr,
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

	t.Run("vote on proposal - cancel", func(t *testing.T) {
		// check the proposal
		proposalsResp := cli.CustomQuery("q", "gov", "proposals")
		proposals := gjson.Get(proposalsResp, "proposals.#.id").Array()
		require.NotEmpty(t, proposals)

		rsp := cli.CustomQuery("q", "gov", "proposal", "2")
		status := gjson.Get(rsp, "proposal.status")
		require.Equal(t, "PROPOSAL_STATUS_VOTING_PERIOD", status.String())

		// vote on the proposal
		args := []string{
			"tx", "gov", "vote", "2", "yes",
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
	t.Run("ensure that the vote has passed - cancel", func(t *testing.T) {
		rsp := cli.CustomQuery("q", "gov", "proposal", "2")
		status := gjson.Get(rsp, "proposal.status")
		require.Equal(t, "PROPOSAL_STATUS_PASSED", status.String())

		// check that the fund does not exist
		failingCli := cli.WithRunErrorMatcher(func(t assert.TestingT, err error, msgAndArgs ...interface{}) (ok bool) {
			assert.Error(t, err)
			return false
		})
		// query the continuous fund - should be expired
		_ = failingCli.CustomQuery("q", "protocolpool", "continuous-fund", account1Addr)

		// check that there is nothing in the store
		rsp = cli.CustomQuery("q", "protocolpool", "continuous-funds")
		require.Equal(t, "{}", rsp)

		balanceAfter := cli.QueryBalance(account1Addr, stakingToken)

		// balance should be balance before + 1442 (community pool value added * 0.5)
		require.Equal(t, balanceBefore+1442, balanceAfter)
	})
}
