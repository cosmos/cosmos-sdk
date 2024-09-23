package systemtests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const (
	msgSendTypeURL       = `/cosmos.bank.v1beta1.MsgSend`
	msgDelegateTypeURL   = `/cosmos.staking.v1beta1.MsgDelegate`
	msgVoteTypeURL       = `/cosmos.gov.v1.MsgVote`
	msgUndelegateTypeURL = `/cosmos.staking.v1beta1.MsgUndelegate`
	msgRedelegateTypeURL = `/cosmos.staking.v1beta1.MsgBeginRedelegate`
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
	// scenario: test authz exec command
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

	var spendLimitAmount int64 = 1000
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
	granterBal--

	testCases := []struct {
		name      string
		grantee   string
		toAddr    string
		amount    int64
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
			"send to not allowed address",
			granteeAddr,
			notAllowedAddr,
			10,
			true,
			"cannot send to",
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
			"no grant found",
			newAccount,
			granteeAddr,
			20,
			true,
			"authorization not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// msg send
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
}`, msgSendTypeURL, granterAddr, tc.toAddr, denom, tc.amount)
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
				expGranterBal := granterBal - tc.amount
				require.Equal(t, expGranterBal, cli.QueryBalance(granterAddr, denom))
				granterBal = expGranterBal

				// check allowed addr balance equals to allowedAddrBal + transferredAmount
				expAllowAddrBal := allowedAddrBal + tc.amount
				require.Equal(t, expAllowAddrBal, cli.QueryBalance(allowedAddr, denom))
				allowedAddrBal = expAllowAddrBal
			}
		})
	}

	// test grant expiry
	time.Sleep(time.Second * 10)
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
	}`, msgSendTypeURL, granterAddr, allowedAddr, denom, 10)
	execMsg := WriteToTempJSONFile(t, bankTx)
	defer execMsg.Close()

	execSendCmd := append(append(execCmdArgs, execMsg.Name()), "--from="+granteeAddr)
	rsp = cli.Run(execSendCmd...)
	RequireTxFailure(t, rsp)
	require.Contains(t, rsp, "authorization not found")

	// test exec generic authorization

	expirationTime = time.Now().Add(time.Second * 5).Unix()

	// create generic authorization grant
	rsp = cli.RunAndWait("tx", "authz", "grant", granteeAddr, "generic",
		"--msg-type="+msgSendTypeURL,
		"--expiration="+fmt.Sprintf("%d", expirationTime),
		"--fees=1stake",
		"--from", granterAddr)
	RequireTxSuccess(t, rsp)
	granterBal--

	rsp = cli.RunAndWait(execSendCmd...)
	RequireTxSuccess(t, rsp)
	// check granter balance equals to granterBal - transferredAmount
	expGranterBal := granterBal - 10
	require.Equal(t, expGranterBal, cli.QueryBalance(granterAddr, denom))
	granterBal = expGranterBal

	time.Sleep(time.Second * 5)

	// check grants after expiration
	resp := cli.CustomQuery("q", "authz", "grants", granterAddr, granteeAddr)
	grants := gjson.Get(resp, "grants").Array()
	require.Len(t, grants, 0)

	// test exec delegate authorization

	// query validator operator address
	rsp = cli.CustomQuery("q", "staking", "validators")
	validators := gjson.Get(rsp, "validators.#.operator_address").Array()
	require.GreaterOrEqual(t, len(validators), 2)
	val1Addr := validators[0].String()
	val2Addr := validators[1].String()

	require.Greater(t, granterBal, spendLimitAmount)

	rsp = cli.RunAndWait("tx", "authz", "grant", granteeAddr, "delegate",
		"--spend-limit="+fmt.Sprintf("%d%s", spendLimitAmount, denom),
		"--allowed-validators="+val1Addr,
		"--fees=1stake",
		"--from", granterAddr)
	RequireTxSuccess(t, rsp)
	// reduce fees of above tx from granter balance
	granterBal--

	delegateTestCases := []struct {
		name      string
		grantee   string
		valAddr   string
		amount    int64
		expectErr bool
		expErrMsg string
	}{
		{
			"valid txn: (delegate half tokens)",
			granteeAddr,
			val1Addr,
			spendLimitAmount / 2,
			false,
			"",
		},
		{
			"amount greater than spend limit",
			granteeAddr,
			val1Addr,
			spendLimitAmount + 5,
			true,
			"negative coin amount",
		},
		{
			"delegate to not allowed address",
			granteeAddr,
			val2Addr,
			10,
			true,
			"cannot delegate",
		},
		{
			"valid txn: (delegate remaining half tokens)",
			granteeAddr,
			val1Addr,
			spendLimitAmount / 2,
			false,
			"",
		},
		{
			"no authorization found as grant prunes",
			granteeAddr,
			val1Addr,
			spendLimitAmount / 2,
			true,
			"authorization not found",
		},
	}

	for _, tc := range delegateTestCases {
		t.Run(tc.name, func(t *testing.T) {
			delegateTx := fmt.Sprintf(`{"@type":"%s","delegator_address":"%s","validator_address":"%s","amount":{"denom":"%s","amount":"%d"}}`,
				msgDelegateTypeURL, granterAddr, tc.valAddr, denom, tc.amount)
			execMsg := WriteToTempJSONFile(t, delegateTx)
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
				expGranterBal := granterBal - tc.amount
				require.Equal(t, expGranterBal, cli.QueryBalance(granterAddr, denom))
				granterBal = expGranterBal
			}
		})
	}

	// test exec undelegate authorization

	// query delegated tokens count
	resp = cli.CustomQuery("q", "staking", "delegation", granterAddr, val1Addr)
	delegatedAmount := gjson.Get(resp, "delegation_response.balance.amount").Int()

	rsp = cli.RunAndWait("tx", "authz", "grant", granteeAddr, "unbond",
		"--allowed-validators="+val1Addr,
		"--fees=1stake",
		"--from", granterAddr)
	RequireTxSuccess(t, rsp)

	undelegateTestCases := []struct {
		name      string
		grantee   string
		valAddr   string
		amount    int64
		expectErr bool
		expErrMsg string
	}{
		{
			"valid transaction",
			granteeAddr,
			val1Addr,
			10,
			false,
			"",
		},
		{
			"undelegate to not allowed address",
			granteeAddr,
			val2Addr,
			10,
			true,
			"cannot delegate/undelegate",
		},
	}

	for _, tc := range undelegateTestCases {
		t.Run(tc.name, func(t *testing.T) {
			undelegateTx := fmt.Sprintf(`{"@type":"%s","delegator_address":"%s","validator_address":"%s","amount":{"denom":"%s","amount":"%d"}}`,
				msgUndelegateTypeURL, granterAddr, tc.valAddr, denom, tc.amount)
			execMsg := WriteToTempJSONFile(t, undelegateTx)
			defer execMsg.Close()

			cmd := append(append(execCmdArgs, execMsg.Name()), "--from="+tc.grantee)
			if tc.expectErr {
				rsp := cli.Run(cmd...)
				RequireTxFailure(t, rsp)
				require.Contains(t, rsp, tc.expErrMsg)
			} else {
				rsp := cli.RunAndWait(cmd...)
				RequireTxSuccess(t, rsp)

				// query delegation and check balance reduced
				expectedAmount := delegatedAmount - tc.amount
				resp = cli.CustomQuery("q", "staking", "delegation", granterAddr, val1Addr)
				delegatedAmount = gjson.Get(resp, "delegation_response.balance.amount").Int()
				require.Equal(t, expectedAmount, delegatedAmount)
			}
		})
	}

	// revoke existing grant
	rsp = cli.RunAndWait("tx", "authz", "revoke", granteeAddr, msgUndelegateTypeURL, "--from", granterAddr)
	RequireTxSuccess(t, rsp)

	// check grants between granter and grantee after revoking
	resp = cli.CustomQuery("q", "authz", "grants", granterAddr, granteeAddr)
	grants = gjson.Get(resp, "grants").Array()
	require.Len(t, grants, 0)

	// test exec redelegate authorization
	rsp = cli.RunAndWait("tx", "authz", "grant", granteeAddr, "redelegate",
		fmt.Sprintf("--allowed-validators=%s,%s", val1Addr, val2Addr),
		"--fees=1stake",
		"--from", granterAddr)
	RequireTxSuccess(t, rsp)

	var redelegationAmount int64 = 10

	redelegateTx := fmt.Sprintf(`{"@type":"%s","delegator_address":"%s","validator_src_address":"%s","validator_dst_address":"%s","amount":{"denom":"%s","amount":"%d"}}`,
		msgRedelegateTypeURL, granterAddr, val1Addr, val2Addr, denom, redelegationAmount)
	execMsg = WriteToTempJSONFile(t, redelegateTx)
	defer execMsg.Close()

	redelegateCmd := append(append(execCmdArgs, execMsg.Name()), "--from="+granteeAddr, "--gas=auto")
	rsp = cli.RunAndWait(redelegateCmd...)
	RequireTxSuccess(t, rsp)

	// query new delegation and check balance increased
	resp = cli.CustomQuery("q", "staking", "delegation", granterAddr, val2Addr)
	delegatedAmount = gjson.Get(resp, "delegation_response.balance.amount").Int()
	require.GreaterOrEqual(t, delegatedAmount, redelegationAmount)

	// revoke all existing grants
	rsp = cli.RunAndWait("tx", "authz", "revoke-all", "--from", granterAddr)
	RequireTxSuccess(t, rsp)

	// check grants after revoking
	resp = cli.CustomQuery("q", "authz", "grants-by-granter", granterAddr)
	grants = gjson.Get(resp, "grants").Array()
	require.Len(t, grants, 0)
}

func TestAuthzGRPCQueries(t *testing.T) {
	// scenario: test authz grpc gateway queries
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	// get validator address which will be used as granter
	granterAddr := gjson.Get(cli.Keys("keys", "list"), "1.address").String()
	require.NotEmpty(t, granterAddr)

	// add grantee keys
	grantee1Addr := cli.AddKey("grantee1")
	require.NotEqual(t, granterAddr, grantee1Addr)

	grantee2Addr := cli.AddKey("grantee2")
	require.NotEqual(t, granterAddr, grantee2Addr)

	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", grantee1Addr, "10000000stake"},
	)

	// start chain
	sut.StartChain(t)

	// create few grants
	rsp := cli.RunAndWait("tx", "authz", "grant", grantee1Addr, "send",
		"--spend-limit=10000stake",
		"--fees=1stake",
		"--from", granterAddr)
	RequireTxSuccess(t, rsp)
	grant1 := `{"authorization":{"@type":"/cosmos.bank.v1beta1.SendAuthorization","spend_limit":[{"denom":"stake","amount":"10000"}],"allow_list":[]},"expiration":null}`

	rsp = cli.RunAndWait("tx", "authz", "grant", grantee2Addr, "send",
		"--spend-limit=10000stake",
		"--fees=1stake",
		"--from", granterAddr)
	RequireTxSuccess(t, rsp)
	grant2 := `{"authorization":{"@type":"/cosmos.bank.v1beta1.SendAuthorization","spend_limit":[{"denom":"stake","amount":"10000"}],"allow_list":[]},"expiration":null}`

	rsp = cli.RunAndWait("tx", "authz", "grant", grantee2Addr, "generic",
		"--msg-type="+msgVoteTypeURL,
		"--fees=1stake",
		"--from", granterAddr)
	RequireTxSuccess(t, rsp)
	grant3 := `{"authorization":{"@type":"/cosmos.authz.v1beta1.GenericAuthorization","msg":"/cosmos.gov.v1.MsgVote"},"expiration":null}`

	baseurl := fmt.Sprintf("http://localhost:%d", apiPortStart)

	// test query grant grpc endpoint
	grantURL := baseurl + "/cosmos/authz/v1beta1/grants?granter=%s&grantee=%s&msg_type_url=%s"

	grantTestCases := []GRPCTestCase{
		{
			"invalid granter address",
			fmt.Sprintf(grantURL, "invalid_granter", grantee1Addr, msgSendTypeURL),
			"decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid grantee address",
			fmt.Sprintf(grantURL, granterAddr, "invalid_grantee", msgSendTypeURL),
			"decoding bech32 failed: invalid separator index -1",
		},
		{
			"with empty granter",
			fmt.Sprintf(grantURL, "", grantee1Addr, msgSendTypeURL),
			"empty address string is not allowed",
		},
		{
			"with empty grantee",
			fmt.Sprintf(grantURL, granterAddr, "", msgSendTypeURL),
			"empty address string is not allowed",
		},
		{
			"invalid msg-type",
			fmt.Sprintf(grantURL, granterAddr, grantee1Addr, "invalidMsg"),
			"authorization not found for invalidMsg type",
		},
		{
			"valid query",
			fmt.Sprintf(grantURL, granterAddr, grantee1Addr, msgSendTypeURL),
			grant1,
		},
	}

	RunGRPCQueries(t, grantTestCases)

	// test query grants grpc endpoint
	grantsURL := baseurl + "/cosmos/authz/v1beta1/grants?granter=%s&grantee=%s"
	grantsTestCases := []GRPCTestCase{
		{
			"expect single grant",
			fmt.Sprintf(grantsURL, granterAddr, grantee1Addr),
			grant1,
		},
		{
			"expect two grants",
			fmt.Sprintf(grantsURL, granterAddr, grantee2Addr),
			fmt.Sprintf("[%s,%s]", grant2, grant3),
		},
		{
			"expect single grant with pagination",
			fmt.Sprintf(grantsURL+"&pagination.limit=1", granterAddr, grantee2Addr),
			grant2,
		},
		{
			"expect single grant with pagination limit and offset",
			fmt.Sprintf(grantsURL+"&pagination.limit=1&pagination.offset=1", granterAddr, grantee2Addr),
			grant3,
		},
		{
			"expect two grants with pagination",
			fmt.Sprintf(grantsURL+"&pagination.limit=2", granterAddr, grantee2Addr),
			fmt.Sprintf("[%s,%s]", grant2, grant3),
		},
	}

	RunGRPCQueries(t, grantsTestCases)

	// test query grants by granter grpc endpoint
	grantsByGranterURL := baseurl + "/cosmos/authz/v1beta1/grants/granter/%s"
	granterTestCases := []GRPCTestCase{
		{
			"invalid account address",
			fmt.Sprintf(grantsByGranterURL, "invalid address"),
			"decoding bech32 failed",
		},
		{
			"no authorizations found",
			fmt.Sprintf(grantsByGranterURL, grantee1Addr),
			`"grants":[]`,
		},
		{
			"valid query",
			fmt.Sprintf(grantsByGranterURL, granterAddr),
			`"total":"3"`,
		},
	}

	RunGRPCQueries(t, granterTestCases)

	// test query grants by grantee grpc endpoint
	grantsByGranteeURL := baseurl + "/cosmos/authz/v1beta1/grants/grantee/%s"
	granteeTestCases := []GRPCTestCase{
		{
			"invalid account address",
			fmt.Sprintf(grantsByGranteeURL, "invalid address"),
			"decoding bech32 failed",
		},
		{
			"no authorizations found",
			fmt.Sprintf(grantsByGranteeURL, granterAddr),
			`"grants":[]`,
		},
		{
			"valid query",
			fmt.Sprintf(grantsByGranteeURL, grantee1Addr),
			`"total":"1"`,
		},
		{
			"valid query with more grants",
			fmt.Sprintf(grantsByGranteeURL, grantee2Addr),
			`"total":"2"`,
		},
	}

	RunGRPCQueries(t, granteeTestCases)
}
