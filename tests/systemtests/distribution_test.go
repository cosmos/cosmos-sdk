//go:build system_test

package systemtests

import (
	"fmt"
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

func TestDistrValidatorGRPCQueries(t *testing.T) {
	// scenario: test distribution validator gsrpc gateway queries
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)
	valOperAddr := cli.GetKeyAddrPrefix("node0", "val")
	require.NotEmpty(t, valOperAddr)

	denom := "stake"
	sut.StartChain(t)

	sut.AwaitNBlocks(t, 3)

	baseurl := sut.APIAddress()
	expectedAmountOutput := fmt.Sprintf(`{"denom":"%s","amount":"203.105000000000000000"}`, denom)

	// test params grpc endpoint
	paramsURL := baseurl + "/cosmos/distribution/v1beta1/params"

	paramsTestCases := []GRPCTestCase{
		{
			"gRPC request params",
			paramsURL,
			`{"params":{"community_tax":"0.020000000000000000","base_proposer_reward":"0.000000000000000000","bonus_proposer_reward":"0.000000000000000000","withdraw_addr_enabled":true}}`,
		},
	}
	RunGRPCQueries(t, paramsTestCases)

	// test validator distribution info grpc endpoint
	validatorsURL := baseurl + `/cosmos/distribution/v1beta1/validators/%s`
	decodingFailedOutput := `{"code":2, "message":"decoding bech32 failed: invalid separator index -1", "details":[]}`
	validatorsOutput := fmt.Sprintf(`{"operator_address":"%s","self_bond_rewards":[],"commission":[%s]}`, valAddr, expectedAmountOutput)

	validatorsTestCases := []GRPCTestCase{
		{
			"invalid validator gRPC request with wrong validator address",
			fmt.Sprintf(validatorsURL, "wrongaddress"),
			decodingFailedOutput,
		},
		{
			"gRPC request validator with valid validator address",
			fmt.Sprintf(validatorsURL, valOperAddr),
			validatorsOutput,
		},
	}
	TestGRPCQueryIgnoreNumbers(t, validatorsTestCases)

	// test outstanding rewards grpc endpoint
	outstandingRewardsURL := baseurl + `/cosmos/distribution/v1beta1/validators/%s/outstanding_rewards`

	rewardsTestCases := []GRPCTestCase{
		{
			"invalid outstanding rewards gRPC request with wrong validator address",
			fmt.Sprintf(outstandingRewardsURL, "wrongaddress"),
			decodingFailedOutput,
		},
		{
			"gRPC request outstanding rewards with valid validator address",
			fmt.Sprintf(outstandingRewardsURL, valOperAddr),
			fmt.Sprintf(`{"rewards":{"rewards":[%s]}}`, expectedAmountOutput),
		},
	}
	TestGRPCQueryIgnoreNumbers(t, rewardsTestCases)

	// test validator commission grpc endpoint
	commissionURL := baseurl + `/cosmos/distribution/v1beta1/validators/%s/commission`

	commissionTestCases := []GRPCTestCase{
		{
			"invalid commission gRPC request with wrong validator address",
			fmt.Sprintf(commissionURL, "wrongaddress"),
			decodingFailedOutput,
		},
		{
			"gRPC request commission with valid validator address",
			fmt.Sprintf(commissionURL, valOperAddr),
			fmt.Sprintf(`{"commission":{"commission":[%s]}}`, expectedAmountOutput),
		},
	}
	TestGRPCQueryIgnoreNumbers(t, commissionTestCases)

	// test validator slashes grpc endpoint
	slashURL := baseurl + `/cosmos/distribution/v1beta1/validators/%s/slashes`
	invalidHeightOutput := `{"code":3, "message":"strconv.ParseUint: parsing \"-3\": invalid syntax", "details":[]}`

	slashTestCases := []GRPCTestCase{
		{
			"invalid slashes gRPC request with wrong validator address",
			fmt.Sprintf(slashURL, "wrongaddress"),
			`{"code":3, "message":"invalid validator address", "details":[]}`,
		},
		{
			"invalid start height",
			fmt.Sprintf(slashURL+`?starting_height=%s&ending_height=%s`, valOperAddr, "-3", "3"),
			invalidHeightOutput,
		},
		{
			"invalid end height",
			fmt.Sprintf(slashURL+`?starting_height=%s&ending_height=%s`, valOperAddr, "1", "-3"),
			invalidHeightOutput,
		},
		{
			"valid request get slashes",
			fmt.Sprintf(slashURL+`?starting_height=%s&ending_height=%s`, valOperAddr, "1", "3"),
			`{"slashes":[],"pagination":{"next_key":null,"total":"0"}}`,
		},
	}
	RunGRPCQueries(t, slashTestCases)
}

func TestDistrDelegatorGRPCQueries(t *testing.T) {
	// scenario: test distribution validator gsrpc gateway queries
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)
	denom := "stake"

	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)
	valOperAddr := cli.GetKeyAddrPrefix("node0", "val")
	require.NotEmpty(t, valOperAddr)

	// update commission rate of node0 validator
	// generate new gentx and copy it to genesis.json before starting network
	rsp := cli.RunCommandWithArgs("genesis", "gentx", "node0", "100000000"+denom, "--chain-id="+cli.chainID, "--commission-rate=0.01", "--home", sut.nodePath(0), "--keyring-backend=test")
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
	initialBalance := fmt.Sprintf("%d%s", initialAmount, denom)
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", delAddr, initialBalance},
	)

	sut.StartChain(t)

	// delegate some tokens to valOperAddr
	rsp = cli.RunAndWait("tx", "staking", "delegate", valOperAddr, "100000000"+denom, "--from="+delAddr)
	RequireTxSuccess(t, rsp)

	sut.AwaitNBlocks(t, 5)

	baseurl := sut.APIAddress()
	decodingFailedOutput := `{"code":2, "message":"decoding bech32 failed: invalid separator index -1", "details":[]}`

	// test delegator rewards grpc endpoint
	delegatorRewardsURL := baseurl + `/cosmos/distribution/v1beta1/delegators/%s/rewards`
	expectedAmountOutput := `{"denom":"stake","amount":"0.121275000000000000"}`
	rewardsOutput := fmt.Sprintf(`{"rewards":[{"validator_address":"%s","reward":[%s]}],"total":[%s]}`, valOperAddr, expectedAmountOutput, expectedAmountOutput)

	delegatorRewardsTestCases := []GRPCTestCase{
		{
			"wrong delegator address",
			fmt.Sprintf(delegatorRewardsURL, "wrongdeladdress"),
			decodingFailedOutput,
		},
		{
			"valid rewards request with valid delegator address",
			fmt.Sprintf(delegatorRewardsURL, delAddr),
			rewardsOutput,
		},
		{
			"wrong validator address (specific validator rewards)",
			fmt.Sprintf(delegatorRewardsURL+`/%s`, delAddr, "wrongvaladdress"),
			decodingFailedOutput,
		},
		{
			"valid request(specific validator rewards)",
			fmt.Sprintf(delegatorRewardsURL+`/%s`, delAddr, valOperAddr),
			fmt.Sprintf(`{"rewards":[%s]}`, expectedAmountOutput),
		},
	}
	TestGRPCQueryIgnoreNumbers(t, delegatorRewardsTestCases)

	// test delegator validators grpc endpoint
	delegatorValsURL := baseurl + `/cosmos/distribution/v1beta1/delegators/%s/validators`
	valsTestCases := []GRPCTestCase{
		{
			"invalid delegator validators gRPC request with wrong delegator address",
			fmt.Sprintf(delegatorValsURL, "wrongaddress"),
			decodingFailedOutput,
		},
		{
			"gRPC request delegator validators with valid delegator address",
			fmt.Sprintf(delegatorValsURL, delAddr),
			fmt.Sprintf(`{"validators":["%s"]}`, valOperAddr),
		},
	}
	RunGRPCQueries(t, valsTestCases)

	// test withdraw address grpc endpoint
	withdrawAddrURL := baseurl + `/cosmos/distribution/v1beta1/delegators/%s/withdraw_address`
	withdrawAddrTestCases := []GRPCTestCase{
		{
			"invalid withdraw address gRPC request with wrong delegator address",
			fmt.Sprintf(withdrawAddrURL, "wrongaddress"),
			decodingFailedOutput,
		},
		{
			"gRPC request withdraw address with valid delegator address",
			fmt.Sprintf(withdrawAddrURL, delAddr),
			fmt.Sprintf(`{"withdraw_address":"%s"}`, delAddr),
		},
	}
	RunGRPCQueries(t, withdrawAddrTestCases)
}
