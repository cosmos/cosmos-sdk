//go:build system_test

package systemtests

import (
	"fmt"
	"os"
	"sync"
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
	govModule          = "gov"
	authModule         = "auth"

	genesisAmount = 1000000000000
	stakeAmount   = 10000000000
	feeAmount     = 1
	depositAmount = 50000000
	poolAmount    = 100
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

func submitGovProposal(t *testing.T, validatorAddress string, propFile *os.File) {
	t.Helper()

	sut := systemtests.Sut
	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)

	args := []string{
		"tx", govModule, "submit-proposal",
		propFile.Name(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, validatorAddress),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(feeAmount))).String()),
	}

	rsp := cli.Run(args...)
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	systemtests.RequireTxSuccess(t, txResult)
}

func voteAndEnsureProposalPassed(t *testing.T, validatorAddress string, propID int) {
	t.Helper()

	sut := systemtests.Sut
	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)

	// check the proposal
	proposalsResp := cli.CustomQuery("q", govModule, "proposals")
	proposals := gjson.Get(proposalsResp, "proposals.#.id").Array()
	require.NotEmpty(t, proposals)

	rsp := cli.CustomQuery("q", govModule, "proposal", fmt.Sprintf("%d", propID))
	status := gjson.Get(rsp, "proposal.status")
	require.Equal(t, "PROPOSAL_STATUS_VOTING_PERIOD", status.String())

	// vote on the proposal
	args := []string{
		"tx", govModule, "vote", fmt.Sprintf("%d", propID), "yes",
		fmt.Sprintf("--%s=%s", flags.FlagFrom, validatorAddress),
	}
	rsp = cli.Run(args...)
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	systemtests.RequireTxSuccess(t, txResult)

	time.Sleep(11 * time.Second)
	systemtests.Sut.AwaitNextBlock(t)

	// ensure that vote has passed
	rsp = cli.CustomQuery("q", "gov", "proposal", fmt.Sprintf("%d", propID))
	status = gjson.Get(rsp, "proposal.status")
	require.Equal(t, "PROPOSAL_STATUS_PASSED", status.String())
}

func getGovAddress(t *testing.T) string {
	t.Helper()

	sut := systemtests.Sut
	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)

	// get gov module address
	resp := cli.CustomQuery("q", authModule, "module-account", "gov")
	govAddress := gjson.Get(resp, "account.value.address").String()
	_, bz, err := bech32.DecodeAndConvert(govAddress)
	assert.NoError(t, err)
	govAddress, err = bech32.ConvertAndEncode(sdk.Bech32MainPrefix, bz)
	assert.NoError(t, err)

	return govAddress
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
		[]string{"genesis", "add-genesis-account", account1Addr, fmt.Sprintf("%d%s", genesisAmount, sdk.DefaultBondDenom)},
	)

	sut.StartChain(t)

	// query validator address to delegate tokens
	rsp := cli.CustomQuery("q", stakingModule, "validators")
	valAddr := gjson.Get(rsp, "validators.#.operator_address").Array()[0].String()

	// get validator address
	valSigner := gjson.Get(cli.Keys("keys", "list"), "0.address").String()
	require.NotEmpty(t, valSigner)

	// stake tokens
	rsp = cli.Run(
		"tx",
		stakingModule,
		"delegate",
		valAddr,
		fmt.Sprintf("%d%s", stakeAmount, sdk.DefaultBondDenom),
		"--from="+account1Addr,
		fmt.Sprintf("--fees=%d%s", feeAmount, sdk.DefaultBondDenom),
	)
	systemtests.RequireTxSuccess(t, rsp)

	beforeBalance := cli.QueryBalance(account1Addr, sdk.DefaultBondDenom)
	assert.Equal(t, int64(genesisAmount-stakeAmount-feeAmount), beforeBalance)

	rsp = cli.CustomQuery("q", stakingModule, "delegation", account1Addr, valAddr)
	assert.Equal(t, int64(stakeAmount), gjson.Get(rsp, "delegation_response.balance.amount").Int(), rsp)
	assert.Equal(t, sdk.DefaultBondDenom, gjson.Get(rsp, "delegation_response.balance.denom").String(), rsp)

	t.Run("check x/distribution query does not work when using x/protocolpool", func(t *testing.T) {
		failingCli := cli.WithRunErrorMatcher(func(t assert.TestingT, err error, msgAndArgs ...interface{}) (ok bool) {
			assert.Error(t, err)
			return false
		})
		// query the community pool - should fail for x/distribution
		_ = failingCli.CustomQuery("q", distributionModule, "community-pool")
	})

	t.Run("check x/protocolpool community pool query", func(t *testing.T) {
		// query will work for x/protocolpool
		rsp = cli.CustomQuery("q", protocolPoolModule, "community-pool")
		poolAmount := gjson.Get(rsp, "pool.0.amount").Int()
		assert.True(t, poolAmount > 0, rsp)
		assert.Equal(t, sdk.DefaultBondDenom, gjson.Get(rsp, "pool.0.denom").String(), rsp)

		t.Log("block height", sut.CurrentHeight(), "\n")
		sut.AwaitNBlocks(t, 1)
		t.Log("block height", sut.CurrentHeight(), "\n")

		rsp = cli.CustomQuery("q", protocolPoolModule, "community-pool")
		newPoolAmount := gjson.Get(rsp, "pool.0.amount").Int()
		assert.Equal(t, sdk.DefaultBondDenom, gjson.Get(rsp, "pool.0.denom").String(), rsp)
		// check that staking is continually rewarded
		assert.True(t, newPoolAmount > poolAmount, rsp)
	})

	govAddress := getGovAddress(t)

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
  				"amount": "%d"
			}]
		}
	],
	"title": "My awesome title",
	"summary": "My awesome description",
	"deposit": "%s"
}`,
			govAddress,
			account1Addr,
			poolAmount,
			sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(depositAmount)),
		)
		validPropFile := systemtests.StoreTempFile(t, []byte(validProp))
		defer validPropFile.Close()

		submitGovProposal(t, valSigner, validPropFile)
	})

	balanceBefore := cli.QueryBalance(account1Addr, sdk.DefaultBondDenom)
	voteAndEnsureProposalPassed(t, valSigner, 1)

	// ensure that vote has passed
	t.Run("ensure that the vote has passed", func(t *testing.T) {
		// check that the funds were distributed
		// should be previous balance plus amount from the pool (100) plus the deposit amount (50000000)
		balanceAfter := cli.QueryBalance(account1Addr, sdk.DefaultBondDenom)
		require.Equal(t, balanceBefore+poolAmount+depositAmount-feeAmount, balanceAfter)
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
		[]string{"genesis", "add-genesis-account", account1Addr, fmt.Sprintf("%d%s", genesisAmount, sdk.DefaultBondDenom)},
	)

	systemtests.Sut.StartChain(t)

	govAddress := getGovAddress(t)
	duration := 30 * time.Second
	// wait long enough that it will be expired
	buffer := 11 * time.Second
	expiry := time.Now().Add(duration).UTC()
	var balanceBefore int64
	wg := new(sync.WaitGroup)
	wg.Add(2)
	time.AfterFunc(duration+buffer, func() {
		wg.Done()
	})
	go func() {
		defer wg.Done()
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
				sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(depositAmount)),
			)
			validPropFile := systemtests.StoreTempFile(t, []byte(validProp))
			defer validPropFile.Close()

			submitGovProposal(t, valAddr, validPropFile)
		})

		// get balance before any distribution
		balanceBefore = cli.QueryBalance(account1Addr, sdk.DefaultBondDenom)
		voteAndEnsureProposalPassed(t, valAddr, 1)

		// ensure that vote has passed
		t.Run("ensure that the vote has passed", func(t *testing.T) {
			// check that the fund exists
			rsp := cli.CustomQuery("q", protocolPoolModule, "continuous-fund", account1Addr)
			gotExpiry := gjson.Get(rsp, "continuous_fund.expiry").Time()
			require.Equal(t, expiry.Truncate(time.Second), gotExpiry.Truncate(time.Second))
			recipient := gjson.Get(rsp, "continuous_fund.recipient").String()
			require.Equal(t, account1Addr, recipient)
		})
	}()

	wg.Wait()
	systemtests.Sut.AwaitNextBlock(t)

	t.Run("check balance and that the fund is expired", func(t *testing.T) {
		failingCli := cli.WithRunErrorMatcher(func(t assert.TestingT, err error, msgAndArgs ...interface{}) (ok bool) {
			assert.Error(t, err)
			return false
		})
		// query the continuous fund - should be expired
		_ = failingCli.CustomQuery("q", protocolPoolModule, "continuous-fund", account1Addr)

		// check that there is nothing in the store
		rsp := cli.CustomQuery("q", protocolPoolModule, "continuous-funds")
		require.Equal(t, "{}", rsp)

		balanceAfter := cli.QueryBalance(account1Addr, sdk.DefaultBondDenom)

		// check that our balance has increased due to fund accrual
		require.True(t, balanceBefore < balanceAfter)
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
		[]string{"genesis", "add-genesis-account", account1Addr, fmt.Sprintf("%d%s", genesisAmount, sdk.DefaultBondDenom)},
	)

	systemtests.Sut.StartChain(t)

	govAddress := getGovAddress(t)

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
			sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(depositAmount)),
		)
		validPropFile := systemtests.StoreTempFile(t, []byte(validProp))
		defer validPropFile.Close()

		submitGovProposal(t, valAddr, validPropFile)
	})

	// get balance before any distribution
	balanceBefore := cli.QueryBalance(account1Addr, sdk.DefaultBondDenom)
	voteAndEnsureProposalPassed(t, valAddr, 1)

	// ensure that vote has passed
	t.Run("ensure that the vote has passed - create", func(t *testing.T) {
		// check that the fund exists
		rsp := cli.CustomQuery("q", protocolPoolModule, "continuous-fund", account1Addr)
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
			sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(depositAmount)),
		)
		validPropFile := systemtests.StoreTempFile(t, []byte(validProp))
		defer validPropFile.Close()

		submitGovProposal(t, valAddr, validPropFile)
	})

	voteAndEnsureProposalPassed(t, valAddr, 2)

	// ensure that vote has passed
	t.Run("ensure that the vote has passed - cancel", func(t *testing.T) {
		// check that the fund does not exist
		failingCli := cli.WithRunErrorMatcher(func(t assert.TestingT, err error, msgAndArgs ...interface{}) (ok bool) {
			assert.Error(t, err)
			return false
		})
		// query the continuous fund - should be expired
		_ = failingCli.CustomQuery("q", protocolPoolModule, "continuous-funds", account1Addr)

		// check that there is nothing in the store
		rsp := cli.CustomQuery("q", protocolPoolModule, "continuous-funds")
		require.Equal(t, "{}", rsp)

		balanceAfter := cli.QueryBalance(account1Addr, sdk.DefaultBondDenom)

		// balance should be balance greater than initial balance
		require.True(t, balanceBefore < balanceAfter)
	})
}
