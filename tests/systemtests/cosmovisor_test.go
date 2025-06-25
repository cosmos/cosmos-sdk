//go:build system_test

package systemtests

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	systest "cosmossdk.io/systemtests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

func TestCosmovisorUpgrade(t *testing.T) {
	const (
		upgrade1Height       = 25
		upgrade1Name         = "v053-to-v054" // must match UpgradeName in simapp/upgrades.go
		upgrade2Height int64 = 30
		upgrade2Name         = "manual1"
	)
	// Scenario:
	// start a legacy chain with some state
	// when a chain upgrade proposal is executed
	// then the chain upgrades successfully
	systest.Sut.StopChain()

	currentBranchBinary := systest.Sut.ExecBinary()

	legacyBinary := systest.WorkDir + "/binaries/v0.53/simd"
	systest.Sut.SetExecBinary(legacyBinary)
	systest.Sut.SetupChain()

	votingPeriod := 5 * time.Second // enough time to vote
	systest.Sut.ModifyGenesisJSON(t, systest.SetGovVotingPeriod(t, votingPeriod))

	systest.Sut.StartChainWithCosmovisor(t)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	govAddr := sdk.AccAddress(address.Module("gov")).String()
	// submit upgrade proposal
	proposal := fmt.Sprintf(`
	{
	"messages": [
	 {
	  "@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
	  "authority": %q,
	  "plan": {
	   "name": %q,
	   "height": "%d"
	  }
	 }
	],
	"metadata": "ipfs://CID",
	"deposit": "100000000stake",
	"title": "my upgrade",
	"summary": "testing"
	}`, govAddr, upgrade1Name, upgrade1Height)
	proposalID := cli.SubmitAndVoteGovProposal(proposal)

	// add binary for gov upgrade
	systest.Sut.ExecCosmovisor(
		t,
		true,
		"add-upgrade",
		upgrade1Name,
		currentBranchBinary,
	)

	systest.Sut.AwaitBlockHeight(t, 21, 60*time.Second)

	t.Logf("current_height: %d\n", systest.Sut.CurrentHeight())
	raw := cli.CustomQuery("q", "gov", "proposal", proposalID)
	proposalStatus := gjson.Get(raw, "proposal.status").String()
	require.Equal(t, "PROPOSAL_STATUS_PASSED", proposalStatus, raw)

	wrapperTxt := fmt.Sprintf(`#!/usr/bin/env bash
set -e
TEST_MANUAL_UPGRADE_HEIGHT="%d" exec %s "$@"`, upgrade2Height, currentBranchBinary)
	wrapperPath := filepath.Join(systest.WorkDir, "testnet", fmt.Sprintf("%s.sh", upgrade2Name))
	wrapperPath, err := filepath.Abs(wrapperPath)
	require.NoError(t, err, "failed to get absolute path for manual upgrade script")
	err = os.WriteFile(wrapperPath, []byte(wrapperTxt), 0o755)
	require.NoError(t, err, "failed to write manual upgrade script")

	// add manual upgrade
	systest.Sut.ExecCosmovisor(
		t,
		true,
		"add-upgrade",
		upgrade2Name,
		wrapperPath,
		fmt.Sprintf("--upgrade-height=%d", upgrade2Height),
	)

	systest.Sut.AwaitBlockHeight(t, upgrade1Height+1)

	regex, err := regexp.Compile(fmt.Sprintf(`UPGRADE %q NEEDED at height: %d:  module=x/upgrade`,
		upgrade1Name, upgrade1Height))
	require.NoError(t, err)
	require.Equal(t, systest.Sut.NodesCount(), systest.Sut.FindLogMessage(regex))
	// TODO check current binary

	systest.Sut.AwaitBlockHeight(t, upgrade2Height+1)
	regex, err = regexp.Compile(fmt.Sprintf(`halt per configuration height %d`,
		upgrade2Height))
	require.NoError(t, err)
	require.Equal(t, systest.Sut.NodesCount(), systest.Sut.FindLogMessage(regex))
	// TODO check current binary

	// smoke test that new version runs
	cli = systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	got := cli.Run("tx", "protocolpool", "fund-community-pool", "100stake", "--from=node0")
	systest.RequireTxSuccess(t, got)
}
