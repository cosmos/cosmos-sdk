//go:build system_test

package systemtests

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/jsonpb"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	systest "cosmossdk.io/systemtests"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func TestCosmovisorUpgrade(t *testing.T) {
	t.Run("gov upgrade, then manual upgrade", func(t *testing.T) {
		// this test
		// 1. starts a legacy v0.53 chain with Cosmovisor
		// 2. submits a gov upgrade proposal to switch to v0.54
		// 3. adds a binary for the gov upgrade
		// 4. waits for the upgrade to be applied and checks the symlink
		// 5. adds a manual upgrade which simapp has configured to make a small state breaking update
		// 6. waits for the manual upgrade to be applied and checks the symlink
		const (
			upgrade1Height       = 25
			upgrade1Name         = "v053-to-v054" // must match UpgradeName in simapp/upgrades.go
			upgrade2Height int64 = 30
			upgrade2Name         = "manual1"
		)

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

		requireCurrentPointsTo(t, "genesis")

		systest.Sut.AwaitBlockHeight(t, 21, 60*time.Second)

		t.Logf("current_height: %d\n", systest.Sut.CurrentHeight())
		raw := cli.CustomQuery("q", "gov", "proposal", proposalID)
		proposalStatus := gjson.Get(raw, "proposal.status").String()
		require.Equal(t, "PROPOSAL_STATUS_PASSED", proposalStatus, raw)

		// we create a wrapper for the current branch binary which sets up the manual upgrade
		wrapperPath := createWrapper(t, upgrade2Name, upgrade2Height, currentBranchBinary)

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

		requireCurrentPointsTo(t, fmt.Sprintf("upgrades/%s", upgrade1Name))
		// make sure a gov upgrade was triggered
		regex, err := regexp.Compile(fmt.Sprintf(`UPGRADE %q NEEDED at height: %d:  module=x/upgrade`,
			upgrade1Name, upgrade1Height))
		require.NoError(t, err)
		require.Equal(t, systest.Sut.NodesCount(), systest.Sut.FindLogMessage(regex))
		// make sure the upgrade-info.json was readable by nodes when they restarted
		regex, err = regexp.Compile("read upgrade info from disk")
		require.NoError(t, err)
		require.Equal(t, systest.Sut.NodesCount(), systest.Sut.FindLogMessage(regex))

		systest.Sut.AwaitBlockHeight(t, upgrade2Height+1)

		requireCurrentPointsTo(t, fmt.Sprintf("upgrades/%s", upgrade2Name))

		// smoke test that new version runs
		cli = systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
		got := cli.Run("tx", "protocolpool", "fund-community-pool", "100stake", "--from=node0")
		systest.RequireTxSuccess(t, got)
	})

	t.Run("manual upgrade", func(t *testing.T) {
		// this test:
		// 1. starts a legacy v0.53 chain with Cosmovisor
		// 2. adds a manual upgrade to v0.54 which has an environment variable set to manually perform the migration
		// 3. waits for the manual upgrade to be applied and checks the symlink
		const (
			upgradeHeight = 10
			upgradeName   = "v053-to-v054" // must match UpgradeName in simapp/upgrades.go
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

		systest.Sut.StartChainWithCosmovisor(t)
		requireCurrentPointsTo(t, "genesis")

		// we create a wrapper for the current branch binary which sets up the manual upgrade
		wrapperPath := createWrapper(t, upgradeName, upgradeHeight, currentBranchBinary)

		// schedule manual upgrade to latest version
		systest.Sut.ExecCosmovisor(
			t,
			true,
			"add-upgrade",
			upgradeName,
			wrapperPath,
			fmt.Sprintf("--upgrade-height=%d", upgradeHeight),
		)

		systest.Sut.AwaitBlockHeight(t, upgradeHeight+1, 60*time.Second)

		requireCurrentPointsTo(t, fmt.Sprintf("upgrades/%s", upgradeName))

		// smoke test that new version runs
		cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
		got := cli.Run("tx", "protocolpool", "fund-community-pool", "100stake", "--from=node0")
		systest.RequireTxSuccess(t, got)
	})
}

func requireCurrentPointsTo(t *testing.T, expected string) {
	t.Helper()
	for i := 0; i < systest.Sut.NodesCount(); i++ {
		curSymLink := filepath.Join(systest.Sut.NodeDir(i), "cosmovisor", "current")
		resolved, err := os.Readlink(curSymLink)
		require.NoError(t, err, "failed to read current symlink for node %d", i)
		require.Equal(t, expected, resolved, "current symlink for node %d does not point to expected directory", i)
	}
}

func createWrapper(t *testing.T, upgradeName string, upgradeHeight int64, binary string) string {
	t.Helper()
	plan := upgradetypes.Plan{
		Name:   upgradeName,
		Height: upgradeHeight,
	}
	str, err := (&jsonpb.Marshaler{}).MarshalToString(&plan)
	require.NoError(t, err, "failed to marshal upgrade plan to JSON")

	wrapperTxt := fmt.Sprintf(`#!/usr/bin/env bash
set -e
SIMAPP_MANUAL_UPGRADE='%s' exec %s "$@"`, str, binary)
	wrapperPath := filepath.Join(systest.WorkDir, "testnet", fmt.Sprintf("%s.sh", upgradeName))
	wrapperPath, err = filepath.Abs(wrapperPath)
	require.NoError(t, err, "failed to get absolute path for manual upgrade script")
	err = os.WriteFile(wrapperPath, []byte(wrapperTxt), 0o755)
	require.NoError(t, err, "failed to write manual upgrade script")
	return wrapperPath
}
