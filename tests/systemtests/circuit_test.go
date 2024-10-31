//go:build system_test

package systemtests

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

var someMsgs = []string{"/cosmos.bank.v1beta1.MsgSend", "/cosmos.bank.v1beta1.MsgMultiSend"}

func TestCircuitCommands(t *testing.T) {
	// scenario: test circuit commands
	// given a running chain

	sut.ResetChain(t)
	require.GreaterOrEqual(t, sut.NodesCount(), 2)

	cli := NewCLIWrapper(t, sut, verbose)

	// get validator addresses
	superAdmin := cli.GetKeyAddr("node0")
	require.NotEmpty(t, superAdmin)

	superAdmin2 := cli.GetKeyAddr("node1")
	require.NotEmpty(t, superAdmin2)

	// short voting period
	// update expedited voting period to avoid validation error
	sut.ModifyGenesisJSON(
		t,
		SetGovVotingPeriod(t, time.Second*8),
		SetGovExpeditedVotingPeriod(t, time.Second*7),
	)

	sut.StartChain(t)

	allMsgsAcc := cli.AddKey("allMsgsAcc")
	require.NotEmpty(t, allMsgsAcc)

	someMsgsAcc := cli.AddKey("someMsgsAcc")
	require.NotEmpty(t, someMsgsAcc)

	// fund tokens to new created addresses
	var amount int64 = 100000
	denom := "stake"
	rsp := cli.FundAddress(allMsgsAcc, fmt.Sprintf("%d%s", amount, denom))
	RequireTxSuccess(t, rsp)
	require.Equal(t, amount, cli.QueryBalance(allMsgsAcc, denom))

	rsp = cli.FundAddress(someMsgsAcc, fmt.Sprintf("%d%s", amount, denom))
	RequireTxSuccess(t, rsp)
	require.Equal(t, amount, cli.QueryBalance(someMsgsAcc, denom))

	// query gov module account address
	rsp = cli.CustomQuery("q", "auth", "module-account", "gov")
	govModAddr := gjson.Get(rsp, "account.value.address")

	// create a proposal to add super admin
	validProposal := fmt.Sprintf(`
	{
		"messages": [
			{
			"@type": "/cosmos.circuit.v1.MsgAuthorizeCircuitBreaker",
			"granter": "%s",
			"grantee": "%s",
			"permissions": {"level": 3, "limit_type_urls": []}
			}
		],
 	 	"title": "Params update proposal",
  		"deposit": "10000000stake",
  		"summary": "A short summary of my proposal"
	}`, govModAddr, superAdmin)
	proposalFile := StoreTempFile(t, []byte(validProposal))

	rsp = cli.RunAndWait("tx", "gov", "submit-proposal", proposalFile.Name(), "--from="+superAdmin)
	RequireTxSuccess(t, rsp)

	// vote to proposal from two validators
	rsp = cli.RunAndWait("tx", "gov", "vote", "1", "yes", "--from="+superAdmin)
	RequireTxSuccess(t, rsp)
	rsp = cli.RunAndWait("tx", "gov", "vote", "1", "yes", "--from="+superAdmin2)
	RequireTxSuccess(t, rsp)

	// wait for proposal to pass
	time.Sleep(time.Second * 8)

	rsp = cli.CustomQuery("q", "circuit", "accounts")

	level := gjson.Get(rsp, fmt.Sprintf("accounts.#(address==%s).permissions.level", superAdmin)).String()
	require.Equal(t, "LEVEL_SUPER_ADMIN", level)

	authorizeTestCases := []struct {
		name          string
		address       string
		level         int
		limtTypeURLs  []string
		expPermission string
	}{
		{
			"set new super admin",
			superAdmin2,
			3,
			[]string{},
			"LEVEL_SUPER_ADMIN",
		},
		{
			"set all msgs level to address",
			allMsgsAcc,
			2,
			[]string{},
			"LEVEL_ALL_MSGS",
		},
		{
			"set some msgs level to address",
			someMsgsAcc,
			1,
			someMsgs,
			"LEVEL_SOME_MSGS",
		},
	}

	for _, tc := range authorizeTestCases {
		t.Run(tc.name, func(t *testing.T) {
			permissionJSON := fmt.Sprintf(`{"level":%d,"limit_type_urls":[]}`, tc.level)
			if len(tc.limtTypeURLs) != 0 {
				permissionJSON = fmt.Sprintf(`{"level":%d,"limit_type_urls":["%s"]}`, tc.level, strings.Join(tc.limtTypeURLs[:], `","`))
			}
			rsp = cli.RunAndWait("tx", "circuit", "authorize", tc.address, permissionJSON, "--from="+superAdmin)
			RequireTxSuccess(t, rsp)

			// query account permissions
			rsp = cli.CustomQuery("q", "circuit", "account", tc.address)
			require.Equal(t, tc.expPermission, gjson.Get(rsp, "permission.level").String())
			if len(tc.limtTypeURLs) != 0 {
				listStr := gjson.Get(rsp, "permission.limit_type_urls").String()

				// convert string to array
				var msgsList []string
				require.NoError(t, json.Unmarshal([]byte(listStr), &msgsList))

				require.EqualValues(t, tc.limtTypeURLs, msgsList)
			}
		})
	}

	// test disable tx command
	testCircuitTxCommand(t, cli, "disable", superAdmin, superAdmin2, allMsgsAcc, someMsgsAcc)

	// test reset tx command
	testCircuitTxCommand(t, cli, "reset", superAdmin, superAdmin2, allMsgsAcc, someMsgsAcc)
}

func testCircuitTxCommand(t *testing.T, cli *CLIWrapper, txType, superAdmin, superAdmin2, allMsgsAcc, someMsgsAcc string) {
	t.Helper()

	disableTestCases := []struct {
		name        string
		fromAddr    string
		disableMsgs []string
		executeTxs  [][]string
	}{
		{
			txType + " msgs with super admin",
			superAdmin,
			[]string{"/cosmos.gov.v1.MsgVote"},
			[][]string{
				{
					"tx", "gov", "vote", "3", "yes", "--from=" + superAdmin,
				},
			},
		},
		{
			txType + " msgs with all msgs level address",
			allMsgsAcc,
			[]string{"/cosmos.gov.v1.MsgDeposit"},
			[][]string{
				{
					"tx", "gov", "deposit", "3", "1000stake", "--from=" + allMsgsAcc,
				},
			},
		},
		{
			txType + " msgs with some msgs level address",
			someMsgsAcc,
			someMsgs,
			[][]string{
				{
					"tx", "bank", "send", superAdmin, someMsgsAcc, "10000stake",
				},
				{
					"tx", "bank", "multi-send", superAdmin, someMsgsAcc, superAdmin2, "10000stake", "--from=" + superAdmin,
				},
			},
		},
	}

	for _, tc := range disableTestCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := []string{"tx", "circuit", txType, "--from=" + tc.fromAddr}
			cmd = append(cmd, tc.disableMsgs...)
			rsp := cli.RunAndWait(cmd...)
			RequireTxSuccess(t, rsp)

			// execute given type transaction
			rsp = cli.CustomQuery("q", "circuit", "disabled-list")
			var list []string
			if gjson.Get(rsp, "disabled_list").Exists() {
				listJSON := gjson.Get(rsp, "disabled_list").Raw

				// convert string to array
				require.NoError(t, json.Unmarshal([]byte(listJSON), &list))
			}
			for _, msg := range tc.disableMsgs {
				if txType == "disable" {
					require.Contains(t, list, msg)
				} else {
					require.NotContains(t, list, msg)
				}
			}

			// test given msg transaction to confirm
			for _, tx := range tc.executeTxs {
				tx = append(tx, "--fees=2stake")
				rsp = cli.RunCommandWithArgs(cli.withTXFlags(tx...)...)
				if txType == "disable" {
					RequireTxFailure(t, rsp)
					require.Contains(t, gjson.Get(rsp, "raw_log").String(), "tx type not allowed")
				} else {
					RequireTxSuccess(t, rsp)
				}
				// wait for sometime to avoid sequence error
				time.Sleep(time.Second * 2)
			}
		})
	}
}
