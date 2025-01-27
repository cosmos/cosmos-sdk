//go:build system_test

package systemtests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	systest "cosmossdk.io/systemtests"
)

const (
	validMetadata = "metadata"
)

func TestGroupCommands(t *testing.T) {
	// scenario: test group commands
	// given a running chain

	systest.Sut.ResetChain(t)
	require.GreaterOrEqual(t, systest.Sut.NodesCount(), 2)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	systest.Sut.StartChain(t)

	baseurl := systest.Sut.APIAddress()

	// test create group
	memberWeight := "5"
	validMembers := fmt.Sprintf(`
	{
		"members": [
			{
				"address": "%s",
				"weight": "%s",
				"metadata": "%s"
			}
		]
	}`, valAddr, memberWeight, validMetadata)
	validMembersFile := systest.StoreTempFile(t, []byte(validMembers))
	createGroupCmd := []string{"tx", "group", "create-group", valAddr, validMetadata, validMembersFile.Name(), "--from=" + valAddr}
	rsp := cli.RunAndWait(createGroupCmd...)
	systest.RequireTxSuccess(t, rsp)

	// query groups by admin to confirm group creation
	rsp = cli.CustomQuery("q", "group", "groups-by-admin", valAddr)
	require.Len(t, gjson.Get(rsp, "groups").Array(), 1)
	groupId := gjson.Get(rsp, "groups.0.id").String()

	// test create group policies
	for i := 0; i < 5; i++ {
		threshold := i + 1
		policyFile := systest.StoreTempFile(t, []byte(fmt.Sprintf(`{"@type":"/cosmos.group.v1.ThresholdDecisionPolicy", "threshold":"%d", "windows":{"voting_period":"30000s"}}`, threshold)))
		policyCmd := []string{"tx", "group", "create-group-policy", valAddr, groupId, validMetadata, policyFile.Name(), "--from=" + valAddr}
		rsp = cli.RunAndWait(policyCmd...)
		systest.RequireTxSuccess(t, rsp)

		groupPoliciesResp := string(systest.GetRequest(t, fmt.Sprintf("%s/cosmos/group/v1/group_policies_by_group/%s", baseurl, groupId)))
		policyAddrQuery := fmt.Sprintf("group_policies.#(decision_policy.threshold==%d).address", threshold)

		require.Equal(t, gjson.Get(groupPoliciesResp, "pagination.total").Int(), int64(threshold))
		policyAddr := gjson.Get(groupPoliciesResp, policyAddrQuery).String()
		require.NotEmpty(t, policyAddr)

		rsp = cli.RunCommandWithArgs(cli.WithTXFlags("tx", "bank", "send", valAddr, policyAddr, "1000stake", "--generate-only")...)
		require.Equal(t, policyAddr, gjson.Get(rsp, "body.messages.0.to_address").String())
	}

	// test create group policy with percentage decision policy
	percentagePolicyType := "/cosmos.group.v1.PercentageDecisionPolicy"
	policyFile := systest.StoreTempFile(t, []byte(fmt.Sprintf(`{"@type":"%s", "percentage":"%f", "windows":{"voting_period":"30000s"}}`, percentagePolicyType, 0.5)))
	policyCmd := []string{"tx", "group", "create-group-policy", valAddr, groupId, validMetadata, policyFile.Name(), "--from=" + valAddr}
	rsp = cli.RunAndWait(policyCmd...)
	systest.RequireTxSuccess(t, rsp)

	groupPoliciesResp := cli.CustomQuery("q", "group", "group-policies-by-admin", valAddr)
	require.Equal(t, gjson.Get(groupPoliciesResp, "pagination.total").Int(), int64(6))
	policyAddr := gjson.Get(groupPoliciesResp, fmt.Sprintf("group_policies.#(decision_policy.type==%s).address", percentagePolicyType)).String()
	require.NotEmpty(t, policyAddr)

	// test create proposal
	proposalJSON := fmt.Sprintf(`{
		"group_policy_address": "%s",
		"messages":[
		{
		"@type": "/cosmos.bank.v1beta1.MsgSend",
		"from_address": "%s",
		"to_address": "%s",
		"amount": [{"denom": "stake","amount": "100"}]
		}
		],
		"metadata": "%s",
		"title": "My Proposal",
		"summary": "Summary",
		"proposers": ["%s"]
	}`, policyAddr, policyAddr, valAddr, validMetadata, valAddr)
	proposalFile := systest.StoreTempFile(t, []byte(proposalJSON))
	rsp = cli.RunAndWait("tx", "group", "submit-proposal", proposalFile.Name())
	systest.RequireTxSuccess(t, rsp)

	// query proposals
	rsp = cli.CustomQuery("q", "group", "proposals-by-group-policy", policyAddr)
	require.Len(t, gjson.Get(rsp, "proposals").Array(), 1)
	proposalId := gjson.Get(rsp, "proposals.0.id").String()

	// test vote proposal
	rsp = cli.RunAndWait("tx", "group", "vote", proposalId, valAddr, "yes", validMetadata)
	systest.RequireTxSuccess(t, rsp)

	// query votes
	voteResp := string(systest.GetRequest(t, fmt.Sprintf("%s/cosmos/group/v1/vote_by_proposal_voter/%s/%s", baseurl, proposalId, valAddr)))

	require.Equal(t, "VOTE_OPTION_YES", gjson.Get(voteResp, "vote.option").String())
}
