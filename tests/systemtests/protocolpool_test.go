//go:build system_test

package systemtests

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"

	systest "cosmossdk.io/systemtests"
)

const (
	distrTestDenom = "stake"
)

func TestQueryProtocolPool(t *testing.T) {
	// scenario: test distribution validator gsrpc gateway queries
	// given a running chain

	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)
	valOperAddr := cli.GetKeyAddrPrefix("node0", "val")
	require.NotEmpty(t, valOperAddr)

	// update commission rate of node0 validator
	// generate new gentx and copy it to genesis.json before starting network
	outFile := filepath.Join(t.TempDir(), "gentx.json")
	_ = cli.RunCommandWithArgs("genesis", "gentx", "node0", "100000000"+distrTestDenom, "--chain-id="+cli.ChainID(), "--commission-rate=0.01", "--home", systest.Sut.NodeDir(0), "--keyring-backend=test", "--output-document="+outFile)
	updatedGentxBz, err := os.ReadFile(outFile) // #nosec G304
	require.NoError(t, err)

	systest.Sut.ModifyGenesisJSON(t, func(genesis []byte) []byte {
		state, err := sjson.SetRawBytes(genesis, "app_state.genutil.gen_txs.0", updatedGentxBz)
		require.NoError(t, err)
		return state
	})

	// create new address which will be used as delegator address
	delAddr := cli.AddKey("delAddr")
	require.NotEmpty(t, delAddr)

	var initialAmount int64 = 1000000000
	initialBalance := fmt.Sprintf("%d%s", initialAmount, distrTestDenom)
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", delAddr, initialBalance},
	)

	systest.Sut.StartChain(t)

	// delegate some tokens to valOperAddr
	rsp := cli.RunAndWait("tx", "staking", "delegate", valOperAddr, "100000000"+distrTestDenom, "--from="+delAddr)
	systest.RequireTxSuccess(t, rsp)

	systest.Sut.AwaitNBlocks(t, 5, 20*time.Second)

	baseurl := systest.Sut.APIAddress()

	// test delegator rewards grpc endpoint
	delegatorRewardsURL := baseurl + `/cosmos/distribution/v1beta1/delegators/%s/rewards`
	expectedAmountOutput := `{"denom":"stake","amount":"0.121275000000000000"}`
	rewardsOutput := fmt.Sprintf(`{"rewards":[{"validator_address":"%s","reward":[%s]}],"total":[%s]}`, valOperAddr, expectedAmountOutput, expectedAmountOutput)

	delegatorRewardsTestCases := []systest.RestTestCase{
		{
			Name:    "valid rewards request with valid delegator address",
			Url:     fmt.Sprintf(delegatorRewardsURL, delAddr),
			ExpCode: http.StatusOK,
			ExpOut:  rewardsOutput,
		},
		{
			Name:    "valid request(specific validator rewards)",
			Url:     fmt.Sprintf(delegatorRewardsURL+`/%s`, delAddr, valOperAddr),
			ExpCode: http.StatusOK,
			ExpOut:  fmt.Sprintf(`{"rewards":[%s]}`, expectedAmountOutput),
		},
	}
	systest.RunRestQueriesIgnoreNumbers(t, delegatorRewardsTestCases...)

	// test delegator validators grpc endpoint
	delegatorValsURL := baseurl + `/cosmos/distribution/v1beta1/delegators/%s/validators`
	valsTestCases := []systest.RestTestCase{
		{
			Name:    "gRPC request delegator validators with valid delegator address",
			Url:     fmt.Sprintf(delegatorValsURL, delAddr),
			ExpCode: http.StatusOK,
			ExpOut:  fmt.Sprintf(`{"validators":["%s"]}`, valOperAddr),
		},
	}
	systest.RunRestQueries(t, valsTestCases...)

	// test withdraw address grpc endpoint
	withdrawAddrURL := baseurl + `/cosmos/distribution/v1beta1/delegators/%s/withdraw_address`
	withdrawAddrTestCases := []systest.RestTestCase{
		{
			Name:    "gRPC request withdraw address with valid delegator address",
			Url:     fmt.Sprintf(withdrawAddrURL, delAddr),
			ExpCode: http.StatusOK,
			ExpOut:  fmt.Sprintf(`{"withdraw_address":"%s"}`, delAddr),
		},
	}
	systest.RunRestQueries(t, withdrawAddrTestCases...)

	// attempt to check the community pool in x/distribution

	// attempt to check the community pool in x/protocolpool
}
