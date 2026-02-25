// IMPORTANT LICENSE NOTICE
//
// SPDX-License-Identifier: CosmosLabs-Evaluation-Only
//
// This file is NOT licensed under the Apache License 2.0.
//
// Licensed under the Cosmos Labs Source Available Evaluation License, which forbids:
// - commercial use,
// - production use, and
// - redistribution.
//
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/group/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

//go:build system_test

package systemtests

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/cosmos/cosmos-sdk/testutil/systemtests"
)

const groupModule = "group"

// thresholdPolicyJSON returns JSON for a threshold decision policy (1 vote to pass, short voting period for tests)
func thresholdPolicyJSON(votingPeriod string) string {
	return fmt.Sprintf(`{"@type":"/cosmos.group.v1.ThresholdDecisionPolicy","threshold":"1","windows":{"voting_period":"%s","min_execution_period":"0s"}}`, votingPeriod)
}

func TestGroupQueries(t *testing.T) {
	// Scenario:
	// Test group module query endpoints on a fresh chain
	// - groups (list groups - may be empty initially)
	// - group-info (requires group_id - skip if no groups)
	// - group-members (requires group_id)
	// - group-policies-by-group (requires group_id)

	sut := systemtests.GetSystemUnderTest()
	sut.ResetChain(t)

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.IsVerbose())

	sut.StartChain(t)

	t.Run("query groups", func(t *testing.T) {
		rsp := cli.CustomQuery("q", groupModule, "groups")
		require.NotEmpty(t, rsp, "response should not be empty")
		// Response has "groups" (possibly empty array) or "pagination" when no groups
		groups := gjson.Get(rsp, "groups").Array()
		t.Logf("Groups count: %d", len(groups))
	})

	t.Run("query group info with invalid id", func(t *testing.T) {
		// Query with non-existent group ID - expect error (group not found)
		rsp := cli.WithRunErrorsIgnored().CustomQuery("q", groupModule, "group-info", "999")
		require.Contains(t, rsp, "not found", "expected 'not found' for invalid group ID")
	})
}

func TestCreateGroupAndQuery(t *testing.T) {
	// Scenario:
	// - Create a group with members
	// - Query group info
	// - Create group policy
	// - Query group policies

	sut := systemtests.GetSystemUnderTest()
	sut.ResetChain(t)

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.IsVerbose())

	// Add accounts for admin and members
	adminAddr := cli.AddKey("groupadmin")
	member1Addr := cli.AddKey("member1")
	member2Addr := cli.AddKey("member2")

	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", adminAddr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", member1Addr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", member2Addr, "10000000stake"},
	)

	sut.StartChain(t)

	t.Run("create group", func(t *testing.T) {
		// Create group: admin, metadata, members-json-file
		membersJSON := fmt.Sprintf(`{"members":[{"address":"%s","weight":"1","metadata":""},{"address":"%s","weight":"1","metadata":""}]}`,
			member1Addr, member2Addr)
		membersFile := systemtests.StoreTempFile(t, []byte(membersJSON))
		defer membersFile.Close()

		rsp := cli.Run(
			"tx", groupModule, "create-group",
			adminAddr,
			"test group metadata",
			membersFile.Name(),
			"--fees=1stake",
		)
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNextBlock(t)
	})

	t.Run("query groups after create", func(t *testing.T) {
		rsp := cli.CustomQuery("q", groupModule, "groups")
		groups := gjson.Get(rsp, "groups").Array()
		require.NotEmpty(t, groups, "should have at least one group")
		groupID := gjson.Get(groups[0].Raw, "id").String()
		require.NotEmpty(t, groupID, "group should have id")
		t.Logf("Created group ID: %s", groupID)
	})

	t.Run("query group info", func(t *testing.T) {
		rsp := cli.CustomQuery("q", groupModule, "groups")
		groups := gjson.Get(rsp, "groups").Array()
		require.NotEmpty(t, groups)
		groupID := gjson.Get(groups[0].Raw, "id").String()

		rsp = cli.CustomQuery("q", groupModule, "group-info", groupID)
		info := gjson.Get(rsp, "info")
		require.True(t, info.Exists(), "group info should exist")
		require.Equal(t, adminAddr, gjson.Get(rsp, "info.admin").String())
	})

	t.Run("query group members", func(t *testing.T) {
		rsp := cli.CustomQuery("q", groupModule, "groups")
		groups := gjson.Get(rsp, "groups").Array()
		require.NotEmpty(t, groups)
		groupID := gjson.Get(groups[0].Raw, "id").String()

		rsp = cli.CustomQuery("q", groupModule, "group-members", groupID)
		members := gjson.Get(rsp, "members").Array()
		require.Len(t, members, 2, "should have 2 members")
	})

	t.Run("query groups by admin", func(t *testing.T) {
		rsp := cli.CustomQuery("q", groupModule, "groups-by-admin", adminAddr)
		groups := gjson.Get(rsp, "groups").Array()
		require.Len(t, groups, 1, "admin should have 1 group")
		require.Equal(t, adminAddr, gjson.Get(groups[0].Raw, "admin").String())
	})

	t.Run("query groups by member", func(t *testing.T) {
		rsp := cli.CustomQuery("q", groupModule, "groups-by-member", member1Addr)
		groups := gjson.Get(rsp, "groups").Array()
		require.Len(t, groups, 1, "member should be in 1 group")
	})
}

func TestCreateGroupWithPolicy(t *testing.T) {
	// Scenario: Create group with policy in one tx, then query group policies
	sut := systemtests.GetSystemUnderTest()
	sut.ResetChain(t)

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.IsVerbose())

	adminAddr := cli.AddKey("policyadmin")
	member1Addr := cli.AddKey("polmember1")
	member2Addr := cli.AddKey("polmember2")

	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", adminAddr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", member1Addr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", member2Addr, "10000000stake"},
	)

	sut.StartChain(t)

	t.Run("create group with policy", func(t *testing.T) {
		membersJSON := fmt.Sprintf(`{"members":[{"address":"%s","weight":"1","metadata":""},{"address":"%s","weight":"1","metadata":""}]}`,
			member1Addr, member2Addr)
		membersFile := systemtests.StoreTempFile(t, []byte(membersJSON))
		defer membersFile.Close()

		policyJSON := thresholdPolicyJSON("10s")
		policyFile := systemtests.StoreTempFile(t, []byte(policyJSON))
		defer policyFile.Close()

		rsp := cli.Run(
			"tx", groupModule, "create-group-with-policy",
			adminAddr,
			"group metadata",
			"policy metadata",
			membersFile.Name(),
			policyFile.Name(),
			"--fees=1stake",
		)
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNextBlock(t)
	})

	t.Run("query group policies by group", func(t *testing.T) {
		rsp := cli.CustomQuery("q", groupModule, "groups")
		groups := gjson.Get(rsp, "groups").Array()
		require.NotEmpty(t, groups)
		groupID := gjson.Get(groups[0].Raw, "id").String()

		rsp = cli.CustomQuery("q", groupModule, "group-policies-by-group", groupID)
		policies := gjson.Get(rsp, "group_policies").Array()
		require.Len(t, policies, 1, "should have 1 group policy")
		policyAddr := gjson.Get(policies[0].Raw, "address").String()
		require.NotEmpty(t, policyAddr)
		t.Logf("Group policy address: %s", policyAddr)
	})

	t.Run("query group policies by admin", func(t *testing.T) {
		rsp := cli.CustomQuery("q", groupModule, "group-policies-by-admin", adminAddr)
		policies := gjson.Get(rsp, "group_policies").Array()
		require.Len(t, policies, 1)
	})
}

func TestCreateGroupPolicyForExistingGroup(t *testing.T) {
	// Scenario: Create group, then create group policy separately
	sut := systemtests.GetSystemUnderTest()
	sut.ResetChain(t)

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.IsVerbose())

	adminAddr := cli.AddKey("admin2")
	member1Addr := cli.AddKey("mem2a")
	member2Addr := cli.AddKey("mem2b")

	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", adminAddr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", member1Addr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", member2Addr, "10000000stake"},
	)

	sut.StartChain(t)

	// Create group first
	membersJSON := fmt.Sprintf(`{"members":[{"address":"%s","weight":"1","metadata":""},{"address":"%s","weight":"1","metadata":""}]}`,
		member1Addr, member2Addr)
	membersFile := systemtests.StoreTempFile(t, []byte(membersJSON))
	defer membersFile.Close()

	rsp := cli.Run(
		"tx", groupModule, "create-group",
		adminAddr, "group meta", membersFile.Name(),
		"--fees=1stake",
	)
	systemtests.RequireTxSuccess(t, rsp)
	sut.AwaitNextBlock(t)

	// Get group ID
	rsp = cli.CustomQuery("q", groupModule, "groups")
	groups := gjson.Get(rsp, "groups").Array()
	require.NotEmpty(t, groups)
	groupID := gjson.Get(groups[0].Raw, "id").String()

	t.Run("create group policy", func(t *testing.T) {
		policyFile := systemtests.StoreTempFile(t, []byte(thresholdPolicyJSON("10s")))
		defer policyFile.Close()

		rsp := cli.Run(
			"tx", groupModule, "create-group-policy",
			adminAddr, groupID, "policy metadata", policyFile.Name(),
			"--fees=1stake",
		)
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNextBlock(t)
	})

	t.Run("verify group policies", func(t *testing.T) {
		rsp := cli.CustomQuery("q", groupModule, "group-policies-by-group", groupID)
		policies := gjson.Get(rsp, "group_policies").Array()
		require.Len(t, policies, 1)
	})
}

func TestUpdateGroupAdmin(t *testing.T) {
	// Scenario: Create group, update admin to new address, verify
	sut := systemtests.GetSystemUnderTest()
	sut.ResetChain(t)

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.IsVerbose())

	adminAddr := cli.AddKey("oldadmin")
	newAdminAddr := cli.AddKey("newadmin")
	memberAddr := cli.AddKey("updatemember")

	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", adminAddr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", newAdminAddr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", memberAddr, "10000000stake"},
	)

	sut.StartChain(t)

	membersJSON := fmt.Sprintf(`{"members":[{"address":"%s","weight":"1","metadata":""}]}`, memberAddr)
	membersFile := systemtests.StoreTempFile(t, []byte(membersJSON))
	defer membersFile.Close()

	rsp := cli.Run("tx", groupModule, "create-group", adminAddr, "meta", membersFile.Name(), "--fees=1stake")
	systemtests.RequireTxSuccess(t, rsp)
	sut.AwaitNextBlock(t)

	rsp = cli.CustomQuery("q", groupModule, "groups")
	groups := gjson.Get(rsp, "groups").Array()
	groupID := gjson.Get(groups[0].Raw, "id").String()

	t.Run("update group admin", func(t *testing.T) {
		rsp := cli.Run(
			"tx", groupModule, "update-group-admin",
			adminAddr, groupID, newAdminAddr,
			"--fees=1stake",
		)
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNextBlock(t)
	})

	t.Run("verify new admin", func(t *testing.T) {
		rsp := cli.CustomQuery("q", groupModule, "group-info", groupID)
		require.Equal(t, newAdminAddr, gjson.Get(rsp, "info.admin").String())
	})
}

func TestUpdateGroupMembers(t *testing.T) {
	// Scenario: Create group, add/remove members
	sut := systemtests.GetSystemUnderTest()
	sut.ResetChain(t)

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.IsVerbose())

	adminAddr := cli.AddKey("memadmin")
	member1Addr := cli.AddKey("m1")
	member2Addr := cli.AddKey("m2")
	member3Addr := cli.AddKey("m3")

	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", adminAddr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", member1Addr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", member2Addr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", member3Addr, "10000000stake"},
	)

	sut.StartChain(t)

	membersJSON := fmt.Sprintf(`{"members":[{"address":"%s","weight":"1","metadata":""},{"address":"%s","weight":"1","metadata":""}]}`,
		member1Addr, member2Addr)
	membersFile := systemtests.StoreTempFile(t, []byte(membersJSON))
	defer membersFile.Close()

	rsp := cli.Run("tx", groupModule, "create-group", adminAddr, "meta", membersFile.Name(), "--fees=1stake")
	systemtests.RequireTxSuccess(t, rsp)
	sut.AwaitNextBlock(t)

	rsp = cli.CustomQuery("q", groupModule, "groups")
	groups := gjson.Get(rsp, "groups").Array()
	groupID := gjson.Get(groups[0].Raw, "id").String()

	t.Run("add new member", func(t *testing.T) {
		// Update: keep m1, m2, add m3
		updatesJSON := fmt.Sprintf(`{"members":[{"address":"%s","weight":"1","metadata":""},{"address":"%s","weight":"1","metadata":""},{"address":"%s","weight":"1","metadata":""}]}`,
			member1Addr, member2Addr, member3Addr)
		updatesFile := systemtests.StoreTempFile(t, []byte(updatesJSON))
		defer updatesFile.Close()

		rsp := cli.Run(
			"tx", groupModule, "update-group-members",
			adminAddr, groupID, updatesFile.Name(),
			"--fees=1stake",
		)
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNextBlock(t)
	})

	t.Run("verify 3 members", func(t *testing.T) {
		rsp := cli.CustomQuery("q", groupModule, "group-members", groupID)
		members := gjson.Get(rsp, "members").Array()
		require.Len(t, members, 3)
	})
}

func TestSubmitProposalVoteAndExec(t *testing.T) {
	// Scenario: Create group with policy, fund policy, submit proposal (bank send), vote, exec
	sut := systemtests.GetSystemUnderTest()
	sut.ResetChain(t)

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.IsVerbose())

	adminAddr := cli.AddKey("proposaladmin")
	member1Addr := cli.AddKey("propmember1")
	recipientAddr := cli.AddKey("recipient")

	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", adminAddr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", member1Addr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", recipientAddr, "10000000stake"},
	)

	sut.StartChain(t)

	// Create group with policy
	membersJSON := fmt.Sprintf(`{"members":[{"address":"%s","weight":"1","metadata":""}]}`, member1Addr)
	membersFile := systemtests.StoreTempFile(t, []byte(membersJSON))
	defer membersFile.Close()

	policyFile := systemtests.StoreTempFile(t, []byte(thresholdPolicyJSON("10s")))
	defer policyFile.Close()

	rsp := cli.Run(
		"tx", groupModule, "create-group-with-policy",
		adminAddr, "group meta", "policy meta",
		membersFile.Name(), policyFile.Name(),
		"--fees=1stake",
	)
	systemtests.RequireTxSuccess(t, rsp)
	sut.AwaitNextBlock(t)

	// Get group policy address
	rsp = cli.CustomQuery("q", groupModule, "group-policies-by-admin", adminAddr)
	policies := gjson.Get(rsp, "group_policies").Array()
	require.Len(t, policies, 1)
	policyAddr := gjson.Get(policies[0].Raw, "address").String()

	// Fund the group policy
	rsp = cli.Run("tx", "bank", "send", adminAddr, policyAddr, "1000stake", "--fees=1stake")
	systemtests.RequireTxSuccess(t, rsp)
	sut.AwaitNextBlock(t)

	t.Run("submit proposal", func(t *testing.T) {
		// Proposal: group policy sends 10 stake to recipient
		msg := map[string]interface{}{
			"@type":        "/cosmos.bank.v1beta1.MsgSend",
			"from_address": policyAddr,
			"to_address":   recipientAddr,
			"amount":       []map[string]string{{"denom": "stake", "amount": "10"}},
		}
		msgBz, _ := json.Marshal(msg)
		proposal := map[string]interface{}{
			"group_policy_address": policyAddr,
			"messages":             []json.RawMessage{msgBz},
			"metadata":             "",
			"title":                "Send 10 stake",
			"summary":              "Proposal to send 10 stake to recipient",
			"proposers":            []string{member1Addr},
		}
		proposalBz, _ := json.Marshal(proposal)
		proposalFile := systemtests.StoreTempFile(t, proposalBz)
		defer proposalFile.Close()

		rsp := cli.Run("tx", groupModule, "submit-proposal", proposalFile.Name(), "--fees=1stake")
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNextBlock(t)
	})

	t.Run("query proposals and vote", func(t *testing.T) {
		rsp := cli.CustomQuery("q", groupModule, "proposals-by-group-policy", policyAddr)
		proposals := gjson.Get(rsp, "proposals").Array()
		require.NotEmpty(t, proposals)
		proposalID := gjson.Get(proposals[0].Raw, "id").String()

		// Vote yes
		rsp = cli.Run(
			"tx", groupModule, "vote",
			proposalID, member1Addr, "VOTE_OPTION_YES", "",
			"--fees=1stake",
		)
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNextBlock(t)
	})

	t.Run("wait for voting period and exec", func(t *testing.T) {
		sut.AwaitNBlocks(t, 5) // Wait for voting period (10s) to end

		rsp := cli.CustomQuery("q", groupModule, "proposals-by-group-policy", policyAddr)
		proposals := gjson.Get(rsp, "proposals").Array()
		require.NotEmpty(t, proposals)
		proposalID := gjson.Get(proposals[0].Raw, "id").String()
		status := gjson.Get(proposals[0].Raw, "status").String()
		t.Logf("Proposal %s status: %s", proposalID, status)

		// Exec proposal (member1 or any voter can exec)
		rsp = cli.Run(
			"tx", groupModule, "exec",
			proposalID,
			"--from="+member1Addr,
			"--fees=1stake",
		)
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNextBlock(t)
	})

	t.Run("verify recipient received funds", func(t *testing.T) {
		bal := cli.QueryBalance(recipientAddr, "stake")
		require.GreaterOrEqual(t, bal, int64(10000010), "recipient should have received 10 stake (initial 10M + 10)")
	})
}

func TestLeaveGroup(t *testing.T) {
	// Scenario: Create group with 2 members, one leaves, verify member count
	sut := systemtests.GetSystemUnderTest()
	sut.ResetChain(t)

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.IsVerbose())

	adminAddr := cli.AddKey("leaveadmin")
	member1Addr := cli.AddKey("leavem1")
	member2Addr := cli.AddKey("leavem2")

	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", adminAddr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", member1Addr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", member2Addr, "10000000stake"},
	)

	sut.StartChain(t)

	membersJSON := fmt.Sprintf(`{"members":[{"address":"%s","weight":"1","metadata":""},{"address":"%s","weight":"1","metadata":""}]}`,
		member1Addr, member2Addr)
	membersFile := systemtests.StoreTempFile(t, []byte(membersJSON))
	defer membersFile.Close()

	rsp := cli.Run("tx", groupModule, "create-group", adminAddr, "meta", membersFile.Name(), "--fees=1stake")
	systemtests.RequireTxSuccess(t, rsp)
	sut.AwaitNextBlock(t)

	rsp = cli.CustomQuery("q", groupModule, "groups")
	groups := gjson.Get(rsp, "groups").Array()
	groupID := gjson.Get(groups[0].Raw, "id").String()

	t.Run("member leaves group", func(t *testing.T) {
		rsp := cli.Run(
			"tx", groupModule, "leave-group",
			member2Addr, groupID,
			"--fees=1stake",
		)
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNextBlock(t)
	})

	t.Run("verify one member remains", func(t *testing.T) {
		rsp := cli.CustomQuery("q", groupModule, "group-members", groupID)
		members := gjson.Get(rsp, "members").Array()
		require.Len(t, members, 1)
		require.Equal(t, member1Addr, gjson.Get(members[0].Raw, "member.address").String())
	})
}
