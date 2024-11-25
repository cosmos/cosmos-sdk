//go:build system_test

package systemtests

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	distrTestDenom = "stake"
)

func TestWithdrawAllRewardsCmd(t *testing.T) {
	// scenario: test distribution withdraw all rewards command
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	newAddr := cli.AddKey("newAddr")
	require.NotEmpty(t, newAddr)

	var initialAmount int64 = 10000000
	initialBalance := fmt.Sprintf("%d%s", initialAmount, distrTestDenom)
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", newAddr, initialBalance},
	)
	sut.StartChain(t)

	// query balance
	newAddrBal := cli.QueryBalance(newAddr, distrTestDenom)
	require.Equal(t, initialAmount, newAddrBal)

	// query validator operator address
	rsp := cli.CustomQuery("q", "staking", "validators")
	validators := gjson.Get(rsp, "validators.#.operator_address").Array()
	require.GreaterOrEqual(t, len(validators), 2)
	val1Addr := validators[0].String()
	val2Addr := validators[1].String()

	var delegationAmount int64 = 100000
	delegation := fmt.Sprintf("%d%s", delegationAmount, distrTestDenom)

	// delegate tokens to validator1
	rsp = cli.RunAndWait("tx", "staking", "delegate", val1Addr, delegation, "--from="+newAddr, "--fees=1"+distrTestDenom)
	RequireTxSuccess(t, rsp)

	// delegate tokens to validator2
	rsp = cli.RunAndWait("tx", "staking", "delegate", val2Addr, delegation, "--from="+newAddr, "--fees=1"+distrTestDenom)
	RequireTxSuccess(t, rsp)

	// check updated balance: newAddrBal - delegatedBal - fees
	expBal := newAddrBal - (delegationAmount * 2) - 2
	newAddrBal = cli.QueryBalance(newAddr, distrTestDenom)
	require.Equal(t, expBal, newAddrBal)

	withdrawCmdArgs := []string{"tx", "distribution", "withdraw-all-rewards", "--from=" + newAddr, "--fees=1" + distrTestDenom}

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
				// split txs with new line
				splitOutput := strings.Split(strings.Trim(gotOutputs[0].(string), "\n"), "\n")
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

func TestDistrValidatorGRPCQueries(t *testing.T) {
	// scenario: test distribution validator grpc gateway queries
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)
	valOperAddr := cli.GetKeyAddrPrefix("node0", "val")
	require.NotEmpty(t, valOperAddr)

	sut.StartChain(t)

	sut.AwaitNBlocks(t, 3)

	baseurl := sut.APIAddress()
	expectedAmountOutput := fmt.Sprintf(`{"denom":"%s","amount":"203.105000000000000000"}`, distrTestDenom)

	// test params grpc endpoint
	paramsURL := baseurl + "/cosmos/distribution/v1beta1/params"

	paramsTestCases := []RestTestCase{
		{
			"gRPC request params",
			paramsURL,
			http.StatusOK,
			`{"params":{"community_tax":"0.020000000000000000","base_proposer_reward":"0.000000000000000000","bonus_proposer_reward":"0.000000000000000000","withdraw_addr_enabled":true}}`,
		},
	}
	RunRestQueries(t, paramsTestCases)

	// test validator distribution info grpc endpoint
	validatorsURL := baseurl + `/cosmos/distribution/v1beta1/validators/%s`
	validatorsOutput := fmt.Sprintf(`{"operator_address":"%s","self_bond_rewards":[],"commission":[%s]}`, valAddr, expectedAmountOutput)

	validatorsTestCases := []RestTestCase{
		{
			"gRPC request validator with valid validator address",
			fmt.Sprintf(validatorsURL, valOperAddr),
			http.StatusOK,
			validatorsOutput,
		},
	}
	TestRestQueryIgnoreNumbers(t, validatorsTestCases)

	// test outstanding rewards grpc endpoint
	outstandingRewardsURL := baseurl + `/cosmos/distribution/v1beta1/validators/%s/outstanding_rewards`

	rewardsTestCases := []RestTestCase{
		{
			"gRPC request outstanding rewards with valid validator address",
			fmt.Sprintf(outstandingRewardsURL, valOperAddr),
			http.StatusOK,
			fmt.Sprintf(`{"rewards":{"rewards":[%s]}}`, expectedAmountOutput),
		},
	}
	TestRestQueryIgnoreNumbers(t, rewardsTestCases)

	// test validator commission grpc endpoint
	commissionURL := baseurl + `/cosmos/distribution/v1beta1/validators/%s/commission`

	commissionTestCases := []RestTestCase{
		{
			"gRPC request commission with valid validator address",
			fmt.Sprintf(commissionURL, valOperAddr),
			http.StatusOK,
			fmt.Sprintf(`{"commission":{"commission":[%s]}}`, expectedAmountOutput),
		},
	}
	TestRestQueryIgnoreNumbers(t, commissionTestCases)

	// test validator slashes grpc endpoint
	slashURL := baseurl + `/cosmos/distribution/v1beta1/validators/%s/slashes`
	invalidHeightOutput := `{"code":3, "message":"strconv.ParseUint: parsing \"-3\": invalid syntax", "details":[]}`

	slashTestCases := []RestTestCase{
		{
			"invalid start height",
			fmt.Sprintf(slashURL+`?starting_height=%s&ending_height=%s`, valOperAddr, "-3", "3"),
			http.StatusBadRequest,
			invalidHeightOutput,
		},
		{
			"invalid end height",
			fmt.Sprintf(slashURL+`?starting_height=%s&ending_height=%s`, valOperAddr, "1", "-3"),
			http.StatusBadRequest,
			invalidHeightOutput,
		},
		{
			"valid request get slashes",
			fmt.Sprintf(slashURL+`?starting_height=%s&ending_height=%s`, valOperAddr, "1", "3"),
			http.StatusOK,
			`{"slashes":[],"pagination":{"next_key":null,"total":"0"}}`,
		},
	}
	RunRestQueries(t, slashTestCases)
}

func TestDistrDelegatorGRPCQueries(t *testing.T) {
	// scenario: test distribution validator gsrpc gateway queries
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)
	valOperAddr := cli.GetKeyAddrPrefix("node0", "val")
	require.NotEmpty(t, valOperAddr)

	// update commission rate of node0 validator
	// generate new gentx and copy it to genesis.json before starting network
	rsp := cli.RunCommandWithArgs("genesis", "gentx", "node0", "100000000"+distrTestDenom, "--chain-id="+cli.chainID, "--commission-rate=0.01", "--home", sut.nodePath(0), "--keyring-backend=test")
	// extract gentx path from above command output
	re := regexp.MustCompile(`"(.*?\.json)"`)
	match := re.FindStringSubmatch(rsp)
	require.GreaterOrEqual(t, len(match), 1)

	updatedGentx := filepath.Join(WorkDir, match[1])
	updatedGentxBz, err := os.ReadFile(updatedGentx) // #nosec G304
	require.NoError(t, err)

	sut.ModifyGenesisJSON(t, func(genesis []byte) []byte {
		state, err := sjson.SetRawBytes(genesis, "app_state.genutil.gen_txs.0", updatedGentxBz)
		require.NoError(t, err)
		return state
	})

	// create new address which will be used as delegator address
	delAddr := cli.AddKey("delAddr")
	require.NotEmpty(t, delAddr)

	var initialAmount int64 = 1000000000
	initialBalance := fmt.Sprintf("%d%s", initialAmount, distrTestDenom)
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", delAddr, initialBalance},
	)

	sut.StartChain(t)

	// delegate some tokens to valOperAddr
	rsp = cli.RunAndWait("tx", "staking", "delegate", valOperAddr, "100000000"+distrTestDenom, "--from="+delAddr)
	RequireTxSuccess(t, rsp)

	sut.AwaitNBlocks(t, 5)

	baseurl := sut.APIAddress()

	// test delegator rewards grpc endpoint
	delegatorRewardsURL := baseurl + `/cosmos/distribution/v1beta1/delegators/%s/rewards`
	expectedAmountOutput := `{"denom":"stake","amount":"0.121275000000000000"}`
	rewardsOutput := fmt.Sprintf(`{"rewards":[{"validator_address":"%s","reward":[%s]}],"total":[%s]}`, valOperAddr, expectedAmountOutput, expectedAmountOutput)

	delegatorRewardsTestCases := []RestTestCase{
		{
			"valid rewards request with valid delegator address",
			fmt.Sprintf(delegatorRewardsURL, delAddr),
			http.StatusOK,
			rewardsOutput,
		},
		{
			"valid request(specific validator rewards)",
			fmt.Sprintf(delegatorRewardsURL+`/%s`, delAddr, valOperAddr),
			http.StatusOK,
			fmt.Sprintf(`{"rewards":[%s]}`, expectedAmountOutput),
		},
	}
	TestRestQueryIgnoreNumbers(t, delegatorRewardsTestCases)

	// test delegator validators grpc endpoint
	delegatorValsURL := baseurl + `/cosmos/distribution/v1beta1/delegators/%s/validators`
	valsTestCases := []RestTestCase{
		{
			"gRPC request delegator validators with valid delegator address",
			fmt.Sprintf(delegatorValsURL, delAddr),
			http.StatusOK,
			fmt.Sprintf(`{"validators":["%s"]}`, valOperAddr),
		},
	}
	RunRestQueries(t, valsTestCases)

	// test withdraw address grpc endpoint
	withdrawAddrURL := baseurl + `/cosmos/distribution/v1beta1/delegators/%s/withdraw_address`
	withdrawAddrTestCases := []RestTestCase{
		{
			"gRPC request withdraw address with valid delegator address",
			fmt.Sprintf(withdrawAddrURL, delAddr),
			http.StatusOK,
			fmt.Sprintf(`{"withdraw_address":"%s"}`, delAddr),
		},
	}
	RunRestQueries(t, withdrawAddrTestCases)
}
