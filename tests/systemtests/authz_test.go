//go:build system_test

package systemtests

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	systest "cosmossdk.io/systemtests"
)

const (
	msgSendTypeURL       = `/cosmos.bank.v1beta1.MsgSend`
	msgDelegateTypeURL   = `/cosmos.staking.v1beta1.MsgDelegate`
	msgVoteTypeURL       = `/cosmos.gov.v1.MsgVote`
	msgUndelegateTypeURL = `/cosmos.staking.v1beta1.MsgUndelegate`
	msgRedelegateTypeURL = `/cosmos.staking.v1beta1.MsgBeginRedelegate`
	sendAuthzTypeURL     = `/cosmos.bank.v1beta1.SendAuthorization`
	genericAuthzTypeURL  = `/cosmos.authz.v1beta1.GenericAuthorization`
	testDenom            = "stake"
)

func TestAuthzGrantTxCmd(t *testing.T) {
	// scenario: test authz grant command
	// given a running chain

	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	// get validator address which will be used as granter
	granterAddr := cli.GetKeyAddr("node0")
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

	systest.Sut.StartChain(t)

	// query validator operator address
	rsp := cli.CustomQuery("q", "staking", "validators")
	valOperAddr := gjson.Get(rsp, "validators.#.operator_address").Array()[0].String()

	grantCmdArgs := []string{"tx", "authz", "grant", "--from", granterAddr}
	expirationTime := time.Now().Add(time.Hour).Unix()

	// test grant command
	testCases := []struct {
		name      string
		grantee   string
		cmdArgs   []string
		expErrMsg string
		queryTx   bool
	}{
		{
			"invalid authorization type",
			grantee1Addr,
			[]string{"spend"},
			"invalid authorization type",
			false,
		},
		{
			"send authorization without spend-limit",
			grantee1Addr,
			[]string{"send"},
			"spend-limit should be greater than zero",
			false,
		},
		{
			"generic authorization without msg type",
			grantee1Addr,
			[]string{"generic"},
			"msg type cannot be empty",
			true,
		},
		{
			"delegate authorization without allow or deny list",
			grantee1Addr,
			[]string{"delegate"},
			"both allowed & deny list cannot be empty",
			false,
		},
		{
			"delegate authorization with invalid allowed validator address",
			grantee1Addr,
			[]string{"delegate", "--allowed-validators=invalid"},
			"decoding bech32 failed",
			false,
		},
		{
			"delegate authorization with invalid deny validator address",
			grantee1Addr,
			[]string{"delegate", "--deny-validators=invalid"},
			"decoding bech32 failed",
			false,
		},
		{
			"unbond authorization without allow or deny list",
			grantee1Addr,
			[]string{"unbond"},
			"both allowed & deny list cannot be empty",
			false,
		},
		{
			"unbond authorization with invalid allowed validator address",
			grantee1Addr,
			[]string{"unbond", "--allowed-validators=invalid"},
			"decoding bech32 failed",
			false,
		},
		{
			"unbond authorization with invalid deny validator address",
			grantee1Addr,
			[]string{"unbond", "--deny-validators=invalid"},
			"decoding bech32 failed",
			false,
		},
		{
			"redelegate authorization without allow or deny list",
			grantee1Addr,
			[]string{"redelegate"},
			"both allowed & deny list cannot be empty",
			false,
		},
		{
			"redelegate authorization with invalid allowed validator address",
			grantee1Addr,
			[]string{"redelegate", "--allowed-validators=invalid"},
			"decoding bech32 failed",
			false,
		},
		{
			"redelegate authorization with invalid deny validator address",
			grantee1Addr,
			[]string{"redelegate", "--deny-validators=invalid"},
			"decoding bech32 failed",
			false,
		},
		{
			"valid send authorization",
			grantee1Addr,
			[]string{"send", "--spend-limit=1000" + testDenom},
			"",
			false,
		},
		{
			"valid send authorization with expiration",
			grantee2Addr,
			[]string{"send", "--spend-limit=1000" + testDenom, fmt.Sprintf("--expiration=%d", expirationTime)},
			"",
			false,
		},
		{
			"valid generic authorization",
			grantee3Addr,
			[]string{"generic", "--msg-type=" + msgVoteTypeURL},
			"",
			false,
		},
		{
			"valid delegate authorization",
			grantee4Addr,
			[]string{"delegate", "--allowed-validators=" + valOperAddr},
			"",
			false,
		},
		{
			"valid unbond authorization",
			grantee5Addr,
			[]string{"unbond", "--deny-validators=" + valOperAddr},
			"",
			false,
		},
		{
			"valid redelegate authorization",
			grantee6Addr,
			[]string{"redelegate", "--allowed-validators=" + valOperAddr},
			"",
			false,
		},
	}

	grantsCount := 0

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := append(append(grantCmdArgs, tc.grantee), tc.cmdArgs...)
			if tc.expErrMsg != "" {
				if tc.queryTx {
					rsp := cli.Run(cmd...)
					systest.RequireTxFailure(t, rsp)
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
				return
			}
			rsp := cli.RunAndWait(cmd...)
			systest.RequireTxSuccess(t, rsp)

			// query granter-grantee grants
			resp := cli.CustomQuery("q", "authz", "grants", granterAddr, tc.grantee)
			grants := gjson.Get(resp, "grants").Array()
			// check grants length equal to 1 to confirm grant created successfully
			require.Len(t, grants, 1)
			grantsCount++
		})
	}

	// query grants-by-granter
	resp := cli.CustomQuery("q", "authz", "grants-by-granter", granterAddr)
	grants := gjson.Get(resp, "grants").Array()
	require.Len(t, grants, grantsCount)
}

func TestAuthzExecSendAuthorization(t *testing.T) {
	// scenario: test authz exec send authorization
	// given a running chain

	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	// get validator address which will be used as granter
	granterAddr := cli.GetKeyAddr("node0")
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

	var initialAmount int64 = 10000000
	initialBalance := fmt.Sprintf("%d%s", initialAmount, testDenom)
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", granteeAddr, initialBalance},
		[]string{"genesis", "add-genesis-account", allowedAddr, initialBalance},
		[]string{"genesis", "add-genesis-account", newAccount, initialBalance},
	)
	systest.Sut.StartChain(t)

	// query balances
	granterBal := cli.QueryBalance(granterAddr, testDenom)
	granteeBal := cli.QueryBalance(granteeAddr, testDenom)
	require.Equal(t, initialAmount, granteeBal)
	allowedAddrBal := cli.QueryBalance(allowedAddr, testDenom)
	require.Equal(t, initialAmount, allowedAddrBal)

	var spendLimitAmount int64 = 1000
	expirationTime := time.Now().Add(systest.Sut.BlockTime() * 10)

	// test exec send authorization

	// create send authorization grant
	rsp := cli.RunAndWait("tx", "authz", "grant", granteeAddr, "send",
		"--spend-limit="+fmt.Sprintf("%d%s", spendLimitAmount, testDenom),
		"--allow-list="+allowedAddr,
		"--expiration="+fmt.Sprintf("%d", expirationTime.Unix()),
		"--fees=1"+testDenom,
		"--from", granterAddr)
	systest.RequireTxSuccess(t, rsp)
	// reduce fees of above tx from granter balance
	granterBal--

	testCases := []struct {
		name      string
		grantee   string
		toAddr    string
		amount    int64
		expErrMsg string
	}{
		{
			"valid exec transaction",
			granteeAddr,
			allowedAddr,
			20,
			"",
		},
		{
			"send to not allowed address",
			granteeAddr,
			notAllowedAddr,
			10,
			"cannot send to",
		},
		{
			"amount greater than spend limit",
			granteeAddr,
			allowedAddr,
			spendLimitAmount + 5,
			"requested amount is more than spend limit",
		},
		{
			"no grant found",
			newAccount,
			granteeAddr,
			20,
			"authorization not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// msg send
			cmd := msgSendExec(t, granterAddr, tc.grantee, tc.toAddr, testDenom, tc.amount)
			if tc.expErrMsg != "" {
				rsp := cli.Run(cmd...)
				systest.RequireTxFailure(t, rsp)
				require.Contains(t, rsp, tc.expErrMsg)
				return
			}
			rsp := cli.RunAndWait(cmd...)
			systest.RequireTxSuccess(t, rsp)

			// check granter balance equals to granterBal - transferredAmount
			expGranterBal := granterBal - tc.amount
			require.Equal(t, expGranterBal, cli.QueryBalance(granterAddr, testDenom))
			granterBal = expGranterBal

			// check allowed addr balance equals to allowedAddrBal + transferredAmount
			expAllowAddrBal := allowedAddrBal + tc.amount
			require.Equal(t, expAllowAddrBal, cli.QueryBalance(allowedAddr, testDenom))
			allowedAddrBal = expAllowAddrBal
		})
	}

	// test grant expiry
	require.Eventually(t, func() bool {
		resp := cli.CustomQuery("q", "authz", "grants", granterAddr, granteeAddr)
		grants := gjson.Get(resp, "grants").Array()
		return len(grants) == 0
	}, 10*systest.Sut.BlockTime(), 200*time.Millisecond)

	execSendCmd := msgSendExec(t, granterAddr, granteeAddr, allowedAddr, testDenom, 10)
	rsp = cli.Run(execSendCmd...)
	systest.RequireTxFailure(t, rsp)
	require.Contains(t, rsp, "authorization not found")
}

func TestAuthzExecGenericAuthorization(t *testing.T) {
	// scenario: test authz exec generic authorization
	// given a running chain

	cli, granterAddr, granteeAddr := setupChain(t)

	allowedAddr := cli.AddKey("allowedAddr")
	require.NotEqual(t, granterAddr, allowedAddr)

	// query balances
	granterBal := cli.QueryBalance(granterAddr, testDenom)

	expirationTime := time.Now().Add(systest.Sut.BlockTime() * 5)

	// create generic authorization grant
	_ = cli.RunAndWait("tx", "authz", "grant", granteeAddr, "generic",
		"--msg-type="+msgSendTypeURL,
		"--expiration="+fmt.Sprintf("%d", expirationTime.Unix()),
		"--fees=1"+testDenom,
		"--from", granterAddr)
	granterBal--

	execSendCmd := msgSendExec(t, granterAddr, granteeAddr, allowedAddr, testDenom, 10)
	_ = cli.RunAndWait(execSendCmd...)
	// check granter balance equals to granterBal - transferredAmount
	expGranterBal := granterBal - 10
	require.Equal(t, expGranterBal, cli.QueryBalance(granterAddr, testDenom))

	// wait until block after expired has passed
	maxWait := max(expirationTime.Sub(time.Now()), time.Nanosecond) + 5*systest.Sut.BlockTime()
	require.Eventually(t, func() bool {
		resp := cli.CustomQuery("q", "authz", "grants", granterAddr, granteeAddr)
		grants := gjson.Get(resp, "grants").Array()
		if len(grants) != 0 {
			t.Log(time.Now().Format(time.RFC3339))
			t.Log(resp)
		}
		return len(grants) == 0
	}, maxWait, 200*time.Millisecond)
}

func TestAuthzExecDelegateAuthorization(t *testing.T) {
	// scenario: test authz exec delegate authorization
	// given a running chain

	cli, granterAddr, granteeAddr := setupChain(t)

	// query balances
	granterBal := cli.QueryBalance(granterAddr, testDenom)

	// query validator operator address
	rsp := cli.CustomQuery("q", "staking", "validators")
	validators := gjson.Get(rsp, "validators.#.operator_address").Array()
	require.GreaterOrEqual(t, len(validators), 2)
	val1Addr := validators[0].String()
	val2Addr := validators[1].String()

	var spendLimitAmount int64 = 1000
	require.Greater(t, granterBal, spendLimitAmount)

	rsp = cli.RunAndWait("tx", "authz", "grant", granteeAddr, "delegate",
		"--spend-limit="+fmt.Sprintf("%d%s", spendLimitAmount, testDenom),
		"--allowed-validators="+val1Addr,
		"--fees=1"+testDenom,
		"--from", granterAddr)
	systest.RequireTxSuccess(t, rsp)
	// reduce fees of above tx from granter balance
	granterBal--

	delegateTestCases := []struct {
		name      string
		grantee   string
		valAddr   string
		amount    int64
		expErrMsg string
	}{
		{
			"valid txn: (delegate half tokens)",
			granteeAddr,
			val1Addr,
			spendLimitAmount / 2,
			"",
		},
		{
			"amount greater than spend limit",
			granteeAddr,
			val1Addr,
			spendLimitAmount + 5,
			"negative coin amount",
		},
		{
			"delegate to not allowed address",
			granteeAddr,
			val2Addr,
			10,
			"cannot delegate",
		},
		{
			"valid txn: (delegate remaining half tokens)",
			granteeAddr,
			val1Addr,
			spendLimitAmount / 2,
			"",
		},
		{
			"no authorization found as grant prunes",
			granteeAddr,
			val1Addr,
			spendLimitAmount / 2,
			"authorization not found",
		},
	}

	execCmdArgs := []string{"tx", "authz", "exec"}

	for _, tc := range delegateTestCases {
		t.Run(tc.name, func(t *testing.T) {
			delegateTx := fmt.Sprintf(`{"@type":"%s","delegator_address":"%s","validator_address":"%s","amount":{"denom":"%s","amount":"%d"}}`,
				msgDelegateTypeURL, granterAddr, tc.valAddr, testDenom, tc.amount)
			execMsg := WriteToTempJSONFile(t, delegateTx)

			cmd := append(append(execCmdArgs, execMsg.Name()), "--from="+tc.grantee)
			if tc.expErrMsg != "" {
				rsp := cli.Run(cmd...)
				systest.RequireTxFailure(t, rsp)
				require.Contains(t, rsp, tc.expErrMsg)
				return
			}
			rsp := cli.RunAndWait(cmd...)
			systest.RequireTxSuccess(t, rsp)

			// check granter balance equals to granterBal - transferredAmount
			expGranterBal := granterBal - tc.amount
			require.Equal(t, expGranterBal, cli.QueryBalance(granterAddr, testDenom))
			granterBal = expGranterBal
		})
	}
}

func TestAuthzExecUndelegateAuthorization(t *testing.T) {
	// scenario: test authz exec undelegate authorization
	// given a running chain

	cli, granterAddr, granteeAddr := setupChain(t)

	// query validator operator address
	rsp := cli.CustomQuery("q", "staking", "validators")
	validators := gjson.Get(rsp, "validators.#.operator_address").Array()
	require.GreaterOrEqual(t, len(validators), 2)
	val1Addr := validators[0].String()
	val2Addr := validators[1].String()

	// delegate some tokens
	rsp = cli.RunAndWait("tx", "staking", "delegate", val1Addr, "10000"+testDenom, "--from="+granterAddr)
	systest.RequireTxSuccess(t, rsp)

	// query delegated tokens count
	resp := cli.CustomQuery("q", "staking", "delegation", granterAddr, val1Addr)
	delegatedAmount := gjson.Get(resp, "delegation_response.balance.amount").Int()

	rsp = cli.RunAndWait("tx", "authz", "grant", granteeAddr, "unbond",
		"--allowed-validators="+val1Addr,
		"--fees=1"+testDenom,
		"--from", granterAddr)
	systest.RequireTxSuccess(t, rsp)

	undelegateTestCases := []struct {
		name      string
		grantee   string
		valAddr   string
		amount    int64
		expErrMsg string
	}{
		{
			"valid transaction",
			granteeAddr,
			val1Addr,
			10,
			"",
		},
		{
			"undelegate to not allowed address",
			granteeAddr,
			val2Addr,
			10,
			"cannot delegate/undelegate",
		},
	}

	for _, tc := range undelegateTestCases {
		t.Run(tc.name, func(t *testing.T) {
			undelegateTx := fmt.Sprintf(`{"@type":"%s","delegator_address":"%s","validator_address":"%s","amount":{"denom":"%s","amount":"%d"}}`,
				msgUndelegateTypeURL, granterAddr, tc.valAddr, testDenom, tc.amount)
			execMsg := WriteToTempJSONFile(t, undelegateTx)

			cmd := []string{"tx", "authz", "exec", execMsg.Name(), "--from=" + tc.grantee}
			if tc.expErrMsg != "" {
				rsp := cli.Run(cmd...)
				systest.RequireTxFailure(t, rsp)
				require.Contains(t, rsp, tc.expErrMsg)
				return
			}
			rsp := cli.RunAndWait(cmd...)
			systest.RequireTxSuccess(t, rsp)

			// query delegation and check balance reduced
			expectedAmount := delegatedAmount - tc.amount
			resp = cli.CustomQuery("q", "staking", "delegation", granterAddr, val1Addr)
			delegatedAmount = gjson.Get(resp, "delegation_response.balance.amount").Int()
			require.Equal(t, expectedAmount, delegatedAmount)
		})
	}

	// revoke existing grant
	rsp = cli.RunAndWait("tx", "authz", "revoke", granteeAddr, msgUndelegateTypeURL, "--from", granterAddr)
	systest.RequireTxSuccess(t, rsp)

	// check grants between granter and grantee after revoking
	resp = cli.CustomQuery("q", "authz", "grants", granterAddr, granteeAddr)
	grants := gjson.Get(resp, "grants").Array()
	require.Len(t, grants, 0)
}

func TestAuthzExecRedelegateAuthorization(t *testing.T) {
	// scenario: test authz exec redelegate authorization
	// given a running chain

	cli, granterAddr, granteeAddr := setupChain(t)

	// query validator operator address
	rsp := cli.CustomQuery("q", "staking", "validators")
	validators := gjson.Get(rsp, "validators.#.operator_address").Array()
	require.GreaterOrEqual(t, len(validators), 2)
	val1Addr := validators[0].String()
	val2Addr := validators[1].String()

	// delegate some tokens
	rsp = cli.RunAndWait("tx", "staking", "delegate", val1Addr, "10000"+testDenom, "--from="+granterAddr)
	systest.RequireTxSuccess(t, rsp)

	// test exec redelegate authorization
	rsp = cli.RunAndWait("tx", "authz", "grant", granteeAddr, "redelegate",
		fmt.Sprintf("--allowed-validators=%s,%s", val1Addr, val2Addr),
		"--fees=1"+testDenom,
		"--from", granterAddr)
	systest.RequireTxSuccess(t, rsp)

	var redelegationAmount int64 = 10

	redelegateTx := fmt.Sprintf(`{"@type":"%s","delegator_address":"%s","validator_src_address":"%s","validator_dst_address":"%s","amount":{"denom":"%s","amount":"%d"}}`,
		msgRedelegateTypeURL, granterAddr, val1Addr, val2Addr, testDenom, redelegationAmount)
	execMsg := WriteToTempJSONFile(t, redelegateTx)

	redelegateCmd := []string{"tx", "authz", "exec", execMsg.Name(), "--from=" + granteeAddr, "--gas=500000", "--fees=10stake"}
	rsp = cli.RunAndWait(redelegateCmd...)
	systest.RequireTxSuccess(t, rsp)

	// query new delegation and check balance increased
	resp := cli.CustomQuery("q", "staking", "delegation", granterAddr, val2Addr)
	delegatedAmount := gjson.Get(resp, "delegation_response.balance.amount").Int()
	require.GreaterOrEqual(t, delegatedAmount, redelegationAmount)

	// revoke all existing grants
	rsp = cli.RunAndWait("tx", "authz", "revoke-all", "--from", granterAddr)
	systest.RequireTxSuccess(t, rsp)

	// check grants after revoking
	resp = cli.CustomQuery("q", "authz", "grants-by-granter", granterAddr)
	grants := gjson.Get(resp, "grants").Array()
	require.Len(t, grants, 0)
}

func TestAuthzGRPCQueries(t *testing.T) {
	// scenario: test authz grpc gateway queries
	// given a running chain

	cli, granterAddr, grantee1Addr := setupChain(t)

	grantee2Addr := cli.AddKey("grantee2")
	require.NotEqual(t, granterAddr, grantee2Addr)
	require.NotEqual(t, grantee1Addr, grantee2Addr)

	// create few grants
	rsp := cli.RunAndWait("tx", "authz", "grant", grantee1Addr, "send",
		"--spend-limit=10000"+testDenom,
		"--from", granterAddr)
	systest.RequireTxSuccess(t, rsp)
	grant1 := fmt.Sprintf(`"authorization":{"@type":"%s","spend_limit":[{"denom":"%s","amount":"10000"}],"allow_list":[]},"expiration":null`, sendAuthzTypeURL, testDenom)

	rsp = cli.RunAndWait("tx", "authz", "grant", grantee2Addr, "send",
		"--spend-limit=1000"+testDenom,
		"--from", granterAddr)
	systest.RequireTxSuccess(t, rsp)
	grant2 := fmt.Sprintf(`"authorization":{"@type":"%s","spend_limit":[{"denom":"%s","amount":"1000"}],"allow_list":[]},"expiration":null`, sendAuthzTypeURL, testDenom)

	rsp = cli.RunAndWait("tx", "authz", "grant", grantee2Addr, "generic",
		"--msg-type="+msgVoteTypeURL,
		"--from", granterAddr)
	systest.RequireTxSuccess(t, rsp)
	grant3 := fmt.Sprintf(`"authorization":{"@type":"%s","msg":"%s"},"expiration":null`, genericAuthzTypeURL, msgVoteTypeURL)

	rsp = cli.RunAndWait("tx", "authz", "grant", grantee2Addr, "generic",
		"--msg-type="+msgDelegateTypeURL,
		"--from", grantee1Addr)
	systest.RequireTxSuccess(t, rsp)
	grant4 := fmt.Sprintf(`"authorization":{"@type":"%s","msg":"%s"},"expiration":null`, genericAuthzTypeURL, msgDelegateTypeURL)

	baseurl := systest.Sut.APIAddress()

	// test query grant grpc endpoint
	grantURL := baseurl + "/cosmos/authz/v1beta1/grants?granter=%s&grantee=%s&msg_type_url=%s"

	bech32FailOutput := `{"code":2, "message":"decoding bech32 failed: invalid separator index -1", "details":[]}`
	emptyStrOutput := `{"code":2, "message":"empty address string is not allowed", "details":[]}`
	invalidMsgTypeOutput := `{"code":2, "message":"codespace authz code 2: authorization not found: authorization not found for invalidMsg type", "details":[]}`
	expGrantOutput := fmt.Sprintf(`{"grants":[{%s}],"pagination":null}`, grant1)

	grantTestCases := []systest.RestTestCase{
		{
			Name:       "invalid granter address",
			Url:        fmt.Sprintf(grantURL, "invalid_granter", grantee1Addr, msgSendTypeURL),
			ExpCodeGTE: http.StatusBadRequest,
			ExpOut:     bech32FailOutput,
		},
		{
			Name:       "invalid grantee address",
			Url:        fmt.Sprintf(grantURL, granterAddr, "invalid_grantee", msgSendTypeURL),
			ExpCodeGTE: http.StatusBadRequest,
			ExpOut:     bech32FailOutput,
		},
		{
			Name:       "with empty granter",
			Url:        fmt.Sprintf(grantURL, "", grantee1Addr, msgSendTypeURL),
			ExpCodeGTE: http.StatusBadRequest,
			ExpOut:     emptyStrOutput,
		},
		{
			Name:       "with empty grantee",
			Url:        fmt.Sprintf(grantURL, granterAddr, "", msgSendTypeURL),
			ExpCodeGTE: http.StatusBadRequest,
			ExpOut:     emptyStrOutput,
		},
		{
			Name:    "invalid msg-type",
			Url:     fmt.Sprintf(grantURL, granterAddr, grantee1Addr, "invalidMsg"),
			ExpCode: http.StatusInternalServerError,
			ExpOut:  invalidMsgTypeOutput,
		},
		{
			Name:    "valid grant query",
			Url:     fmt.Sprintf(grantURL, granterAddr, grantee1Addr, msgSendTypeURL),
			ExpCode: http.StatusOK,
			ExpOut:  expGrantOutput,
		},
	}

	systest.RunRestQueriesIgnoreNumbers(t, grantTestCases...)

	// test query grants grpc endpoint
	grantsURL := baseurl + "/cosmos/authz/v1beta1/grants?granter=%s&grantee=%s"

	grantsTestCases := []systest.RestTestCase{
		{
			Name:    "expect single grant",
			Url:     fmt.Sprintf(grantsURL, granterAddr, grantee1Addr),
			ExpCode: http.StatusOK,
			ExpOut:  fmt.Sprintf(`{"grants":[{%s}],"pagination":{"next_key":null,"total":"1"}}`, grant1),
		},
		{
			Name:    "expect two grants",
			Url:     fmt.Sprintf(grantsURL, granterAddr, grantee2Addr),
			ExpCode: http.StatusOK,
			ExpOut:  fmt.Sprintf(`{"grants":[{%s},{%s}],"pagination":{"next_key":null,"total":"2"}}`, grant2, grant3),
		},
		{
			Name:    "expect single grant with pagination",
			Url:     fmt.Sprintf(grantsURL+"&pagination.limit=1", granterAddr, grantee2Addr),
			ExpCode: http.StatusOK,
			ExpOut:  fmt.Sprintf(`{"grants":[{%s}],"pagination":{"next_key":"L2Nvc21vcy5nb3YudjEuTXNnVm90ZQ==","total":"0"}}`, grant2),
		},
		{
			Name:    "expect single grant with pagination limit and offset",
			Url:     fmt.Sprintf(grantsURL+"&pagination.limit=1&pagination.offset=1", granterAddr, grantee2Addr),
			ExpCode: http.StatusOK,
			ExpOut:  fmt.Sprintf(`{"grants":[{%s}],"pagination":{"next_key":null,"total":"0"}}`, grant3),
		},
		{
			Name:    "expect two grants with pagination",
			Url:     fmt.Sprintf(grantsURL+"&pagination.limit=2", granterAddr, grantee2Addr),
			ExpCode: http.StatusOK,
			ExpOut:  fmt.Sprintf(`{"grants":[{%s},{%s}],"pagination":{"next_key":null,"total":"0"}}`, grant2, grant3),
		},
	}

	systest.RunRestQueries(t, grantsTestCases...)

	// test query grants by granter grpc endpoint
	grantsByGranterURL := baseurl + "/cosmos/authz/v1beta1/grants/granter/%s"
	decodingFailedOutput := `{"code":2, "message":"decoding bech32 failed: invalid character in string: ' '", "details":[]}`
	noAuthorizationsOutput := `{"grants":[],"pagination":{"next_key":null,"total":"0"}}`
	granterQueryOutput := fmt.Sprintf(`{"grants":[{"granter":"%s","grantee":"%s",%s}],"pagination":{"next_key":null,"total":"1"}}`,
		grantee1Addr, grantee2Addr, grant4)

	granterTestCases := []systest.RestTestCase{
		{
			Name:       "invalid granter account address",
			Url:        fmt.Sprintf(grantsByGranterURL, "invalid address"),
			ExpCodeGTE: http.StatusBadRequest,
			ExpOut:     decodingFailedOutput,
		},
		{
			Name:    "no authorizations found from granter",
			Url:     fmt.Sprintf(grantsByGranterURL, grantee2Addr),
			ExpCode: http.StatusOK,
			ExpOut:  noAuthorizationsOutput,
		},
		{
			Name:    "valid granter query",
			Url:     fmt.Sprintf(grantsByGranterURL, grantee1Addr),
			ExpCode: http.StatusOK,
			ExpOut:  granterQueryOutput,
		},
	}

	systest.RunRestQueriesIgnoreNumbers(t, granterTestCases...)

	// test query grants by grantee grpc endpoint
	grantsByGranteeURL := baseurl + "/cosmos/authz/v1beta1/grants/grantee/%s"
	grantee1GrantsOutput := fmt.Sprintf(`{"grants":[{"granter":"%s","grantee":"%s",%s}],"pagination":{"next_key":null,"total":"1"}}`, granterAddr, grantee1Addr, grant1)

	granteeTestCases := []systest.RestTestCase{
		{
			Name:       "invalid grantee account address",
			Url:        fmt.Sprintf(grantsByGranteeURL, "invalid address"),
			ExpCodeGTE: http.StatusBadRequest,
			ExpOut:     decodingFailedOutput,
		},
		{
			Name:    "no authorizations found from grantee",
			Url:     fmt.Sprintf(grantsByGranteeURL, granterAddr),
			ExpCode: http.StatusOK,
			ExpOut:  noAuthorizationsOutput,
		},
		{
			Name:    "valid grantee query",
			Url:     fmt.Sprintf(grantsByGranteeURL, grantee1Addr),
			ExpCode: http.StatusOK,
			ExpOut:  grantee1GrantsOutput,
		},
	}

	systest.RunRestQueriesIgnoreNumbers(t, granteeTestCases...)
}

func setupChain(t *testing.T) (*systest.CLIWrapper, string, string) {
	t.Helper()

	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	require.GreaterOrEqual(t, systest.Sut.NodesCount(), 2)

	// get validators' address which will be used as granter and grantee
	granterAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, granterAddr)
	granteeAddr := cli.GetKeyAddr("node1")
	require.NotEmpty(t, granteeAddr)

	systest.Sut.StartChain(t)

	return cli, granterAddr, granteeAddr
}

func msgSendExec(t *testing.T, granter, grantee, toAddr, denom string, amount int64) []string {
	t.Helper()

	bankTx := fmt.Sprintf(`{
		"@type": "%s",
		"from_address": "%s",
		"to_address": "%s",
		"amount": [
			{
				"denom": "%s",
				"amount": "%d"
			}
		]
	}`, msgSendTypeURL, granter, toAddr, denom, amount)
	execMsg := WriteToTempJSONFile(t, bankTx)

	execSendCmd := []string{"tx", "authz", "exec", execMsg.Name(), "--from=" + grantee}
	return execSendCmd
}

// Write the given string to a new temporary json file.
// Returns an file for the test to use.
func WriteToTempJSONFile(tb testing.TB, s string) *os.File {
	tb.Helper()

	tmpFile, err := os.CreateTemp(tb.TempDir(), "test-*.json")
	require.NoError(tb, err)

	// Write to the temporary file
	_, err = tmpFile.WriteString(s)
	require.NoError(tb, err)

	// Close the file after writing
	err = tmpFile.Close()
	require.NoError(tb, err)

	return tmpFile
}
