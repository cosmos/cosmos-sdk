//go:build system_test

package systemtests

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	systest "cosmossdk.io/systemtests"
)

func TestUnorderedTXDuplicate(t *testing.T) {
	// scenario: test unordered tx duplicate
	// given a running chain with a tx in the unordered tx pool
	// when a new tx with the same hash is broadcasted
	// then the new tx should be rejected

	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	account2Addr := cli.AddKey("account2")
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, "10000000stake"},
	)

	systest.Sut.StartChain(t)

	timeoutTimestamp := time.Now().Add(time.Minute)
	// send tokens
	cmd := []string{"tx", "bank", "send", account1Addr, account2Addr, "5000stake", "--from=" + account1Addr, "--fees=1stake", fmt.Sprintf("--timeout-timestamp=%v", timeoutTimestamp.Unix()), "--unordered", "--sequence=1", "--note=1"}
	rsp1 := cli.Run(cmd...)
	systest.RequireTxSuccess(t, rsp1)

	assertDuplicateErr := func(xt assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
		require.Len(t, gotOutputs, 1)
		output := gotOutputs[0].(string)
		code := gjson.Get(output, "code")
		require.True(t, code.Exists())
		require.Equal(t, int64(19), code.Int()) // 19 == already in mempool.
		return false                            // always abort
	}
	rsp2 := cli.WithRunErrorMatcher(assertDuplicateErr).Run(cmd...)
	systest.RequireTxFailure(t, rsp2)

	require.Eventually(t, func() bool {
		return cli.QueryBalance(account2Addr, "stake") == 5000
	}, 10*systest.Sut.BlockTime(), 200*time.Millisecond, "TX was not executed before timeout")
}

func TestTxBackwardsCompatability(t *testing.T) {
	// Scenario:
	// A transaction generated from a v0.53 chain without unordered and timeout_timestamp flags set should succeed.
	// Conversely, a transaction generated from a v0.53 chain with unordered and timeout_timestamp flags set should fail.
	var (
		denom                = "stake"
		transferAmount int64 = 1000
	)
	systest.Sut.ResetChain(t)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")

	systest.Sut.StartChain(t)

	// create unsigned tx
	bankSendCmdArgs := []string{"tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--chain-id=" + cli.ChainID(), "--fees=10stake", "--sign-mode=direct", "--generate-only"}
	res := cli.RunCommandWithArgs(bankSendCmdArgs...)
	txFile := systest.StoreTempFile(t, []byte(res))
	res = cli.RunCommandWithArgs("tx", "sign", txFile.Name(), "--from="+valAddr, "--chain-id="+cli.ChainID(), "--keyring-backend=test", "--home="+systest.Sut.NodeDir(0))
	signedTxFile := systest.StoreTempFile(t, []byte(res))
	// encodedTx := cli.RunCommandWithArgs("tx", "encode", signedTxFile.Name())

	systest.Sut.StopChain()

	// Now we're going to switch to a v.50 chain.
	legacyBinary := systest.WorkDir + "/binaries/simd_v50"
	systest.Sut.SetExecBinary(legacyBinary)
	systest.Sut.SetTestnetInitializer(systest.LegacyInitializerWithBinary(legacyBinary, systest.Sut))
	systest.Sut.SetupChain()
	systest.Sut.StartChain(t)

	cli = systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	res = cli.Run("tx", "broadcast", signedTxFile.Name())
	systest.RequireTxSuccess(t, res)

	//res = cli.CustomQuery("tx", "decode", fmt.Sprintf(`%s`, encodedTx))
	//res, err := fixJSONIntegerResponse(res)
	//require.NoError(t, err)
	//tx := &txtypes.Tx{}
	//err = json.Unmarshal([]byte(res), tx)
	//require.NoError(t, err)
}

func fixJSONIntegerResponse(input string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	processed := doReplace(data, []string{})

	result, err := json.Marshal(processed)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(result), nil
}

// replaces integers in data only if we're not in "fee.amount"
func doReplace(data interface{}, path []string) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, val := range v {
			newPath := append(path, key)
			v[key] = doReplace(val, newPath)
		}
	case []interface{}:
		for i, val := range v {
			v[i] = doReplace(val, path)
		}
	case string:
		// we don't want to change it if its fee.amount
		if len(path) >= 3 &&
			path[len(path)-3] == "fee" &&
			path[len(path)-2] == "amount" {
			return v
		}
		if num, err := strconv.Atoi(v); err == nil {
			return num
		}
	}
	return data
}
