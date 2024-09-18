package systemtests

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/testutil"
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
	allowedAddr := cli.AddKey("allowed")
	require.NotEqual(t, granterAddr, allowedAddr)
	notAllowedAddr := cli.AddKey("notAllowed")
	require.NotEqual(t, granterAddr, notAllowedAddr)

	denom := "stake"
	var initialAmount int64 = 10000000
	initialBalance := fmt.Sprintf("%d%s", initialAmount, denom)
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", granteeAddr, initialBalance},
		[]string{"genesis", "add-genesis-account", allowedAddr, initialBalance},
	)
	sut.StartChain(t)

	// query balancec
	granteeBal := cli.QueryBalance(granteeAddr, denom)
	require.Equal(t, initialAmount, granteeBal)
	allowedAddrBal := cli.QueryBalance(allowedAddr, denom)
	require.Equal(t, initialAmount, allowedAddrBal)

	spendLimitAmount := 1000

	// test exec send authorization

	// create send authorization grant
	rsp := cli.RunAndWait("tx", "authz", "grant", granteeAddr, "send",
		"--spend-limit="+fmt.Sprintf("%d%s", spendLimitAmount, denom),
		"--allow-list="+allowedAddr,
		"--from", granterAddr)
	RequireTxSuccess(t, rsp)

	execCmdArgs := []string{"tx", "authz", "exec"}

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
			notAllowedAddr,
			granteeAddr,
			20,
			true,
			"authorization not found",
		},
		{
			"amount greater than spend limit",
			notAllowedAddr,
			granteeAddr,
			spendLimitAmount + 5,
			true,
			"authorization not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// msg vote
			bankTx := fmt.Sprintf(`
{
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
			execMsg := testutil.WriteToNewTempFile(t, bankTx)
			defer execMsg.Close()
			fmt.Println("Exec file", execMsg.Name())
			time.Sleep(time.Minute * 10)

			cmd := append(append(execCmdArgs, execMsg.Name()), "--from="+tc.grantee)
			if tc.expectErr {
				rsp := cli.Run(cmd...)
				RequireTxFailure(t, rsp)
				require.Contains(t, rsp, tc.expErrMsg)
			} else {
				rsp := cli.RunAndWait(cmd...)
				RequireTxSuccess(t, rsp)
			}
		})
	}

}
