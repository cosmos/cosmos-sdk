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
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

//go:build system_test

package systemtests

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/cosmos/cosmos-sdk/tools/systemtests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	upgradeHeight int64 = 50
	upgradeName         = "pos-to-poa" // must match simapp/upgrades.go
)

// TestPOStoPoaUpgrade runs the full POS → POA chain upgrade and verifies:
//   - POA validators match pre-upgrade bonded set
//   - Delegator tokens returned (including third-party and in-flight unbondings)
//   - Active governance proposal failed and deposit refunded
//   - Total supply preserved
//   - POA governance, fee distribution, and bank transfers work post-upgrade
func TestPOStoPoaUpgrade(t *testing.T) {
	sut := systemtests.Sut
	sut.StopChain()

	poaBinary := sut.ExecBinary()
	poaInitializer := sut.TestnetInitializer()

	posBinary := systemtests.WorkDir + "/binaries/pos/simd"
	sut.SetExecBinary(posBinary)
	sut.SetTestnetInitializer(systemtests.InitializerWithBinary(posBinary, sut))
	sut.SetupChain()

	votingPeriod := 10 * time.Second
	sut.ModifyGenesisJSON(t, systemtests.SetGovVotingPeriod(t, votingPeriod))

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)
	delegatorAddr := cli.AddKey("delegator")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", delegatorAddr, "50000000stake"},
	)

	sut.StartChain(t, fmt.Sprintf("--halt-height=%d", upgradeHeight+1))

	cli = systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)
	govAddr := sdk.AccAddress(address.Module("gov")).String()

	// --- Pre-upgrade: build interesting state ---

	stakingValsRaw := cli.CustomQuery("q", "staking", "validators")
	stakingVals := gjson.Get(stakingValsRaw, "validators").Array()
	require.NotEmpty(t, stakingVals)
	preUpgradeValCount := len(stakingVals)
	valAddr := gjson.Get(stakingVals[0].Raw, "operator_address").String()

	// Third-party delegation.
	rsp := cli.Run("tx", "staking", "delegate", valAddr, "10000000stake",
		"--from="+delegatorAddr, "--fees=1stake")
	systemtests.RequireTxSuccess(t, rsp)

	delegatorBalancePreUnbond := cli.QueryBalance(delegatorAddr, "stake")

	// In-flight unbonding.
	rsp = cli.Run("tx", "staking", "unbond", valAddr, "5000000stake",
		"--from="+delegatorAddr, "--fees=1stake")
	systemtests.RequireTxSuccess(t, rsp)

	rsp = cli.CustomQuery("q", "staking", "unbonding-delegation", delegatorAddr, valAddr)
	require.NotEmpty(t, gjson.Get(rsp, "unbond.entries").Array())

	// Active governance proposal (not voted on).
	activeProposal := fmt.Sprintf(`{
		"messages": [],
		"metadata": "ipfs://CID",
		"deposit": "10000000stake",
		"title": "Should be failed by upgrade",
		"summary": "This proposal should be in voting period when the upgrade fires"
	}`)
	rsp = cli.SubmitGovProposal(activeProposal, "--from=node0")
	systemtests.RequireTxSuccess(t, rsp)

	raw := cli.CustomQuery("q", "gov", "proposals", "--depositor", cli.GetKeyAddr("node0"))
	proposals := gjson.Get(raw, "proposals.#.id").Array()
	require.NotEmpty(t, proposals)
	activeProposalID := proposals[len(proposals)-1].String()

	node0Addr := cli.GetKeyAddr("node0")
	node0BalancePre := cli.QueryBalance(node0Addr, "stake")
	preUpgradeSupply := cli.CustomQuery("q", "bank", "total")

	_ = delegatorBalancePreUnbond // recorded for debugging

	// --- Upgrade: submit proposal, vote, wait ---

	upgradeProposal := fmt.Sprintf(`{
		"messages": [{
			"@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
			"authority": %q,
			"plan": {
				"name": %q,
				"height": "%d"
			}
		}],
		"metadata": "ipfs://CID",
		"deposit": "100000000stake",
		"title": "POS to POA migration",
		"summary": "Migrate chain from Proof-of-Stake to Proof-of-Authority"
	}`, govAddr, upgradeName, upgradeHeight)

	upgradeProposalID := cli.SubmitAndVoteGovProposal(upgradeProposal)

	sut.AwaitBlockHeight(t, upgradeHeight-1, 180*time.Second)
	raw = cli.CustomQuery("q", "gov", "proposal", upgradeProposalID)
	require.Equal(t, "PROPOSAL_STATUS_PASSED", gjson.Get(raw, "proposal.status").String())

	sut.AwaitUpgradeInfo(t)
	sut.StopChain()

	// --- Swap binary and restart ---

	sut.SetExecBinary(poaBinary)
	sut.SetTestnetInitializer(poaInitializer)
	sut.StartChain(t)

	require.GreaterOrEqual(t, sut.CurrentHeight(), upgradeHeight+1)
	cli = systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)

	// --- Post-upgrade verification ---

	t.Run("poa validators match pre-upgrade set", func(t *testing.T) {
		rsp := cli.CustomQuery("q", "poa", "validators")
		poaVals := gjson.Get(rsp, "validators").Array()
		require.NotEmpty(t, poaVals)
		assert.Equal(t, preUpgradeValCount, len(poaVals))

		for _, v := range poaVals {
			assert.Greater(t, gjson.Get(v.Raw, "power").Int(), int64(0))
			assert.NotEmpty(t, gjson.Get(v.Raw, "metadata.moniker").String())
		}
	})

	t.Run("poa total power is positive", func(t *testing.T) {
		rsp := cli.CustomQuery("q", "poa", "total-power")
		require.Greater(t, gjson.Get(rsp, "total_power").Int(), int64(0))
	})

	t.Run("poa admin is governance module", func(t *testing.T) {
		rsp := cli.CustomQuery("q", "poa", "params")
		assert.Equal(t, govAddr, gjson.Get(rsp, "params.admin").String())
	})

	t.Run("delegator tokens returned", func(t *testing.T) {
		balance := cli.QueryBalance(delegatorAddr, "stake")
		assert.Greater(t, balance, int64(49000000),
			"delegator should have most of their original 50M back")
	})

	t.Run("active proposal failed and deposit refunded", func(t *testing.T) {
		raw := cli.CustomQuery("q", "gov", "proposal", activeProposalID)
		status := gjson.Get(raw, "proposal.status").String()
		assert.NotEqual(t, "PROPOSAL_STATUS_VOTING_PERIOD", status)
		assert.NotEqual(t, "PROPOSAL_STATUS_PASSED", status)

		node0BalancePost := cli.QueryBalance(node0Addr, "stake")
		assert.Greater(t, node0BalancePost, node0BalancePre,
			"node0 should have more tokens after upgrade (unbonded delegation + deposit refund)")
	})

	t.Run("total supply preserved", func(t *testing.T) {
		postUpgradeSupply := cli.CustomQuery("q", "bank", "total")
		preDenom := gjson.Get(preUpgradeSupply, `supply.#(denom=="stake").amount`).Int()
		postDenom := gjson.Get(postUpgradeSupply, `supply.#(denom=="stake").amount`).Int()
		require.Greater(t, preDenom, int64(0))
		require.Greater(t, postDenom, int64(0))
		assert.InDelta(t, preDenom, postDenom, float64(preDenom/1000))
	})

	t.Run("chain produces blocks", func(t *testing.T) {
		startHeight := sut.CurrentHeight()
		sut.AwaitNBlocks(t, 3, 30*time.Second)
		assert.Greater(t, sut.CurrentHeight(), startHeight)
	})

	t.Run("bank transfers work", func(t *testing.T) {
		recipientAddr := cli.AddKey("recipient")
		rsp := cli.Run("tx", "bank", "send", "node0", recipientAddr, "1000stake", "--fees=1stake")
		systemtests.RequireTxSuccess(t, rsp)
		assert.Equal(t, int64(1000), cli.QueryBalance(recipientAddr, "stake"))
	})

	t.Run("poa governance works", func(t *testing.T) {
		rsp := cli.CustomQuery("q", "poa", "validators")
		poaVals := gjson.Get(rsp, "validators").Array()
		require.NotEmpty(t, poaVals)

		// Find a local key that is a POA validator.
		var voterKeyName string
		for _, v := range poaVals {
			opAddr := gjson.Get(v.Raw, "metadata.operator_address").String()
			if name := getKeyNameForAddress(t, cli, opAddr); name != "" {
				voterKeyName = name
				break
			}
		}
		require.NotEmpty(t, voterKeyName)

		textProposal := fmt.Sprintf(`{
			"messages": [],
			"metadata": "ipfs://CID",
			"deposit": "10000000stake",
			"title": "Post-upgrade POA governance test",
			"summary": "Verify governance works after POS to POA migration"
		}`)
		propFile := systemtests.StoreTempFile(t, []byte(textProposal))
		rsp = cli.Run("tx", "gov", "submit-proposal", propFile.Name(),
			"--from="+voterKeyName, "--fees=1stake", "--gas=auto")
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNextBlock(t)

		rsp = cli.CustomQuery("q", "gov", "proposals")
		allProposals := gjson.Get(rsp, "proposals").Array()
		require.NotEmpty(t, allProposals)
		govTestProposalID := gjson.Get(allProposals[len(allProposals)-1].Raw, "id").String()

		// All validators vote yes in parallel.
		var wg sync.WaitGroup
		for i := 0; i < sut.NodesCount(); i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				cli.Run("tx", "gov", "vote", govTestProposalID, "yes",
					"--from="+fmt.Sprintf("node%d", i), "--fees=1stake")
			}(i)
		}
		wg.Wait()
		sut.AwaitNextBlock(t)
		sut.AwaitNextBlock(t)

		time.Sleep(votingPeriod + 3*time.Second)
		sut.AwaitNextBlock(t)

		rsp = cli.CustomQuery("q", "gov", "proposal", govTestProposalID)
		assert.Equal(t, "PROPOSAL_STATUS_PASSED", gjson.Get(rsp, "proposal.status").String())
	})

	t.Run("fees route to poa validators", func(t *testing.T) {
		recipientAddr := cli.GetKeyAddr("recipient")
		for i := 0; i < 5; i++ {
			rsp := cli.Run("tx", "bank", "send", "node0", recipientAddr,
				"100stake", "--fees=10000stake")
			systemtests.RequireTxSuccess(t, rsp)
		}
		sut.AwaitNBlocks(t, 2, 15*time.Second)

		rsp := cli.CustomQuery("q", "poa", "validators")
		poaVals := gjson.Get(rsp, "validators").Array()
		require.NotEmpty(t, poaVals)

		opAddr := gjson.Get(poaVals[0].Raw, "metadata.operator_address").String()
		feesRsp := cli.CustomQuery("q", "poa", "withdrawable-fees", opAddr)

		fees := gjson.Get(feesRsp, "fees.fees").Array()
		require.NotEmpty(t, fees, "validator should have withdrawable fees")
		assert.NotEqual(t, "0", gjson.Get(fees[0].Raw, "amount").String())
	})
}
