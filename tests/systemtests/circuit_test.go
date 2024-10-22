package systemtests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestCircuitCmds(t *testing.T) {
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

	accountAddr := cli.AddKey("account")
	require.NotEmpty(t, accountAddr)

	// fund tokens to new created addresses
	var amount int64 = 100000
	denom := "stake"
	rsp := cli.FundAddress(allMsgsAcc, fmt.Sprintf("%d%s", amount, denom))
	RequireTxSuccess(t, rsp)
	require.Equal(t, amount, cli.QueryBalance(allMsgsAcc, denom))

	rsp = cli.FundAddress(someMsgsAcc, fmt.Sprintf("%d%s", amount, denom))
	RequireTxSuccess(t, rsp)
	require.Equal(t, amount, cli.QueryBalance(someMsgsAcc, denom))

	rsp = cli.FundAddress(accountAddr, fmt.Sprintf("%d%s", amount, denom))
	RequireTxSuccess(t, rsp)
	require.Equal(t, amount, cli.QueryBalance(accountAddr, denom))

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
		limtTypeURLS  string
		expPermission string
	}{
		{
			"set new super admin",
			superAdmin2,
			3,
			"",
			"LEVEL_SUPER_ADMIN",
		},
		{
			"set all msgs level to address",
			allMsgsAcc,
			2,
			"",
			"LEVEL_ALL_MSGS",
		},
		{
			"set some msgs level to address",
			someMsgsAcc,
			1,
			"/cosmos.bank.v1beta1.MsgSend, /cosmos.bank.v1beta1.MsgMultiSend",
			"LEVEL_SOME_MSGS",
		},
	}

	for _, tc := range authorizeTestCases {
		t.Run(tc.name, func(t *testing.T) {
			permissionJSON := fmt.Sprintf(`{"level":%d,"limit_type_urls":[]}`, tc.level)
			if tc.limtTypeURLS != "" {
				permissionJSON = fmt.Sprintf(`{"level":%d,"limit_type_urls":["%s"]}`, tc.level, tc.limtTypeURLS)
			}
			rsp = cli.RunAndWait("tx", "circuit", "authorize", tc.address, permissionJSON, "--from="+superAdmin)
			RequireTxSuccess(t, rsp)

			// query account permissions
			rsp = cli.CustomQuery("q", "circuit", "account", tc.address)
			require.Equal(t, tc.expPermission, gjson.Get(rsp, "permission.level").String())
			if tc.limtTypeURLS != "" {
				require.Equal(t, tc.limtTypeURLS, gjson.Get(rsp, "permission.limit_type_urls.0").String())
			}
		})
	}
}
