package systemtests

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const (
	msgVoteTypeURL = `/cosmos.gov.v1.MsgVote`
)

func TestAuthzGrantTxCmd(t *testing.T) {
	// scenario: test authz grant command
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	// get validator address which will be used as granter
	granterAddr := gjson.Get(cli.Keys("keys", "list"), "1.address").String()
	require.NotEmpty(t, granterAddr)

	// add grantee keys which will be used for each valid transaction
	grantee1Addr := cli.AddKey("grantee1")
	grantee2Addr := cli.AddKey("grantee2")
	require.NotEqual(t, granterAddr, grantee2Addr)
	grantee3Addr := cli.AddKey("grantee3")
	require.NotEqual(t, granterAddr, grantee3Addr)
	grantee4Addr := cli.AddKey("grantee4")
	require.NotEqual(t, granterAddr, grantee4Addr)
	grantee5Addr := cli.AddKey("grantee5")
	require.NotEqual(t, granterAddr, grantee5Addr)
	grantee6Addr := cli.AddKey("grantee6")
	require.NotEqual(t, granterAddr, grantee6Addr)

	sut.StartChain(t)

	// query validator operator address
	rsp := cli.CustomQuery("q", "staking", "validators")
	valOperAddr := gjson.Get(rsp, "validators.#.operator_address").Array()[0].String()

	grantCmdArgs := []string{"tx", "authz", "grant", "--from", granterAddr}
	expirationTime := time.Now().Add(time.Minute * time.Duration(10)).Unix()

	// test grant command
	testCases := []struct {
		name      string
		grantee   string
		cmdArgs   []string
		expectErr bool
		expErrMsg string
		queryTx   bool
	}{
		{
			"not enough arguments",
			grantee1Addr,
			[]string{},
			true,
			"accepts 2 arg(s), received 1",
			false,
		},
		{
			"invalid authorization type",
			grantee1Addr,
			[]string{"spend"},
			true,
			"invalid authorization type",
			false,
		},
		{
			"send authorization without spend-limit",
			grantee1Addr,
			[]string{"send"},
			true,
			"spend-limit should be greater than zero",
			false,
		},
		{
			"generic authorization without msg type",
			grantee1Addr,
			[]string{"generic"},
			true,
			"msg type cannot be empty",
			true,
		},
		{
			"delegate authorization without allow or deny list",
			grantee1Addr,
			[]string{"delegate"},
			true,
			"both allowed & deny list cannot be empty",
			false,
		},
		{
			"delegate authorization with invalid allowed validator address",
			grantee1Addr,
			[]string{"delegate", "--allowed-validators=invalid"},
			true,
			"decoding bech32 failed",
			false,
		},
		{
			"delegate authorization with invalid deny validator address",
			grantee1Addr,
			[]string{"delegate", "--deny-validators=invalid"},
			true,
			"decoding bech32 failed",
			false,
		},
		{
			"unbond authorization without allow or deny list",
			grantee1Addr,
			[]string{"unbond"},
			true,
			"both allowed & deny list cannot be empty",
			false,
		},
		{
			"unbond authorization with invalid allowed validator address",
			grantee1Addr,
			[]string{"unbond", "--allowed-validators=invalid"},
			true,
			"decoding bech32 failed",
			false,
		},
		{
			"unbond authorization with invalid deny validator address",
			grantee1Addr,
			[]string{"unbond", "--deny-validators=invalid"},
			true,
			"decoding bech32 failed",
			false,
		},
		{
			"redelegate authorization without allow or deny list",
			grantee1Addr,
			[]string{"redelegate"},
			true,
			"both allowed & deny list cannot be empty",
			false,
		},
		{
			"redelegate authorization with invalid allowed validator address",
			grantee1Addr,
			[]string{"redelegate", "--allowed-validators=invalid"},
			true,
			"decoding bech32 failed",
			false,
		},
		{
			"redelegate authorization with invalid deny validator address",
			grantee1Addr,
			[]string{"redelegate", "--deny-validators=invalid"},
			true,
			"decoding bech32 failed",
			false,
		},
		{
			"valid send authorization",
			grantee1Addr,
			[]string{"send", "--spend-limit=1000stake"},
			false,
			"",
			false,
		},
		{
			"valid send authorization with expiration",
			grantee2Addr,
			[]string{"send", "--spend-limit=1000stake", fmt.Sprintf("--expiration=%d", expirationTime)},
			false,
			"",
			false,
		},
		{
			"valid generic authorization",
			grantee3Addr,
			[]string{"generic", "--msg-type=" + msgVoteTypeURL},
			false,
			"",
			false,
		},
		{
			"valid delegate authorization",
			grantee4Addr,
			[]string{"delegate", "--allowed-validators=" + valOperAddr},
			false,
			"",
			false,
		},
		{
			"valid unbond authorization",
			grantee5Addr,
			[]string{"unbond", "--deny-validators=" + valOperAddr},
			false,
			"",
			false,
		},
		{
			"valid redelegate authorization",
			grantee6Addr,
			[]string{"redelegate", "--allowed-validators=" + valOperAddr},
			false,
			"",
			false,
		},
	}

	grantsCount := 0

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := append(append(grantCmdArgs, tc.grantee), tc.cmdArgs...)
			if tc.expectErr {
				if tc.queryTx {
					rsp := cli.Run(cmd...)
					RequireTxFailure(t, rsp)
					require.Contains(t, rsp, tc.expErrMsg)
				} else {
					assertErr := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
						require.Len(t, gotOutputs, 1)
						output := gotOutputs[0].(string)
						require.Contains(t, output, tc.expErrMsg)
						return false
					}
					_ = cli.WithRunErrorMatcher(assertErr).Run(cmd...)
				}
			} else {
				rsp := cli.RunAndWait(cmd...)
				RequireTxSuccess(t, rsp)

				// query granter-grantee grants
				resp := cli.CustomQuery("q", "authz", "grants", granterAddr, tc.grantee)
				grants := gjson.Get(resp, "grants").Array()
				// check grants length equal to 1 to confirm grant created successfully
				require.Len(t, grants, 1)
				grantsCount++
			}
		})
	}

	// query grants-by-granter
	resp := cli.CustomQuery("q", "authz", "grants-by-granter", granterAddr)
	grants := gjson.Get(resp, "grants").Array()
	require.Len(t, grants, grantsCount)
}

func TestAuthzExecTxCmd(t *testing.T) {
	// scenario: test authz grant command
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	// get validator address which will be used as granter
	granterAddr := gjson.Get(cli.Keys("keys", "list"), "1.address").String()
	require.NotEmpty(t, granterAddr)

	// add grantee keys which will be used for each valid transaction
	granteeAddr := cli.AddKey("grantee")
	require.NotEqual(t, granterAddr, granteeAddr)
	allowedAddr := cli.AddKey("allowed")
	require.NotEqual(t, granteeAddr, allowedAddr)
	notAllowedAddr := cli.AddKey("notAllowed")
	require.NotEqual(t, granteeAddr, notAllowedAddr)
	newAccount := cli.AddKey("newAccount")
	require.NotEqual(t, granteeAddr, newAccount)

	denom := "stake"
	var initialAmount int64 = 10000000
	initialBalance := fmt.Sprintf("%d%s", initialAmount, denom)
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", granteeAddr, initialBalance},
		[]string{"genesis", "add-genesis-account", allowedAddr, initialBalance},
		[]string{"genesis", "add-genesis-account", newAccount, initialBalance},
	)
	sut.StartChain(t)

	// query balances
	granterBal := cli.QueryBalance(granterAddr, denom)
	granteeBal := cli.QueryBalance(granteeAddr, denom)
	require.Equal(t, initialAmount, granteeBal)
	allowedAddrBal := cli.QueryBalance(allowedAddr, denom)
	require.Equal(t, initialAmount, allowedAddrBal)

	spendLimitAmount := 1000
	expirationTime := time.Now().Add(time.Second * 10).Unix()

	execCmdArgs := []string{"tx", "authz", "exec"}

	// test exec command basic checks
	execErrTestCases := []struct {
		name      string
		cmdArgs   []string
		expErrMsg string
	}{
		{
			"not enough arguments",
			[]string{"--from=" + granteeAddr},
			"accepts 1 arg(s), received 0",
		},
		{
			"invalid json path",
			[]string{"/invalid/file.txt", "--from=" + granteeAddr},
			"invalid argument",
		},
	}

	for _, tc := range execErrTestCases {
		cmd := append(execCmdArgs, tc.cmdArgs...)
		t.Run(tc.name, func(t *testing.T) {
			assertErr := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
				require.Len(t, gotOutputs, 1)
				output := gotOutputs[0].(string)
				require.Contains(t, output, tc.expErrMsg)
				return false
			}
			_ = cli.WithRunErrorMatcher(assertErr).Run(cmd...)

		})
	}

	// test exec send authorization

	// create send authorization grant
	rsp := cli.RunAndWait("tx", "authz", "grant", granteeAddr, "send",
		"--spend-limit="+fmt.Sprintf("%d%s", spendLimitAmount, denom),
		"--allow-list="+allowedAddr,
		"--expiration="+fmt.Sprintf("%d", expirationTime),
		"--fees=1stake",
		"--from", granterAddr)
	RequireTxSuccess(t, rsp)
	// reduce fees of above tx from granter balance
	granterBal = granterBal - 1

	testCases := []struct {
		name      string
		grantee   string
		toAddr    string
		amount    int
		expectErr bool
		expErrMsg string
	}{
		{
			"valid exec transaction",
			granteeAddr,
			allowedAddr,
			20,
			false,
			"",
		},
		{
			"no grant found",
			newAccount,
			granteeAddr,
			20,
			true,
			"authorization not found",
		},
		{
			"amount greater than spend limit",
			granteeAddr,
			allowedAddr,
			spendLimitAmount + 5,
			true,
			"requested amount is more than spend limit",
		},
		{
			"send to not allowed address",
			granteeAddr,
			notAllowedAddr,
			10,
			true,
			"cannot send to",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// msg vote
			bankTx := fmt.Sprintf(`{
    "@type": "/cosmos.bank.v1beta1.MsgSend",
    "from_address": "%s",
    "to_address": "%s",
    "amount": [
        {
            "denom": "%s",
            "amount": "%d"
        }
    ]
}`, granterAddr, tc.toAddr, denom, tc.amount)
			execMsg := WriteToTempJSONFile(t, bankTx)
			defer execMsg.Close()

			cmd := append(append(execCmdArgs, execMsg.Name()), "--from="+tc.grantee)
			if tc.expectErr {
				rsp := cli.Run(cmd...)
				RequireTxFailure(t, rsp)
				require.Contains(t, rsp, tc.expErrMsg)
			} else {
				rsp := cli.RunAndWait(cmd...)
				RequireTxSuccess(t, rsp)

				// check granter balance equals to granterBal - transferredAmount
				expGranterBal := granterBal - int64(tc.amount)
				require.Equal(t, expGranterBal, cli.QueryBalance(granterAddr, denom))
				granterBal = expGranterBal

				// check allowed addr balance equals to allowedAddrBal + transferredAmount
				fmt.Println("Allowed.....", allowedAddrBal, tc.amount)
				expAllowAddrBal := allowedAddrBal + int64(tc.amount)
				require.Equal(t, expAllowAddrBal, cli.QueryBalance(allowedAddr, denom))
				allowedAddrBal = expAllowAddrBal
			}
		})
	}

	// test grant expiry
	time.Sleep(time.Second * 10)
	bankTx := fmt.Sprintf(`{
		"@type": "/cosmos.bank.v1beta1.MsgSend",
		"from_address": "%s",
		"to_address": "%s",
		"amount": [
			{
				"denom": "%s",
				"amount": "%d"
			}
		]
	}`, granterAddr, allowedAddr, denom, 10)
	execMsg := WriteToTempJSONFile(t, bankTx)
	defer execMsg.Close()

	execSendCmd := append(append(execCmdArgs, execMsg.Name()), "--from="+granteeAddr)
	rsp = cli.Run(execSendCmd...)
	RequireTxFailure(t, rsp)
	require.Contains(t, rsp, "authorization not found")
}

// Write the given string to a new temporary json file.
// Returns an file for the test to use.
func WriteToTempJSONFile(t testing.TB, s string) *os.File {
	t.Helper()

	tmpFile, err := os.CreateTemp(t.TempDir(), "test-*.json")
	require.Nil(t, err)
	defer tmpFile.Close()

	// Write to the temporary file
	_, err = tmpFile.WriteString(s)
	require.Nil(t, err)

	return tmpFile
}
