//go:build system_test && linux

package systemtests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	systest "cosmossdk.io/systemtests"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

func TestChainUpgrade(t *testing.T) {
	// Scenario:
	// start a legacy chain with some state
	// when a chain upgrade proposal is executed
	// then the chain upgrades successfully
	systest.Sut.StopChain()

	legacyBinary := FetchExecutable(t, "0.52.0-beta.3")
	t.Logf("+++ legacy binary: %s\n", legacyBinary)
	currentBranchBinary := systest.Sut.ExecBinary()
	currentInitializer := systest.Sut.TestnetInitializer()
	systest.Sut.SetExecBinary(legacyBinary)
	systest.Sut.SetTestnetInitializer(systest.InitializerWithBinary(legacyBinary, systest.Sut))
	systest.Sut.SetupChain()
	votingPeriod := 5 * time.Second // enough time to vote
	systest.Sut.ModifyGenesisJSON(t, systest.SetGovVotingPeriod(t, votingPeriod))

	const (
		upgradeHeight int64 = 22
		upgradeName         = "v052-to-v2" // must match UpgradeName in simapp/upgrades.go
	)

	systest.Sut.StartChain(t, fmt.Sprintf("--comet.halt-height=%d", upgradeHeight+1))

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
}`, govAddr, upgradeName, upgradeHeight)
	proposalID := cli.SubmitAndVoteGovProposal(proposal)
	t.Logf("current_height: %d\n", systest.Sut.CurrentHeight())
	raw := cli.CustomQuery("q", "gov", "proposal", proposalID)
	t.Log(raw)

	systest.Sut.AwaitBlockHeight(t, upgradeHeight-1, 60*time.Second)
	t.Logf("current_height: %d\n", systest.Sut.CurrentHeight())
	raw = cli.CustomQuery("q", "gov", "proposal", proposalID)
	proposalStatus := gjson.Get(raw, "proposal.status").String()
	require.Equal(t, "PROPOSAL_STATUS_PASSED", proposalStatus, raw) // PROPOSAL_STATUS_PASSED

	t.Log("waiting for upgrade info")
	systest.Sut.AwaitUpgradeInfo(t)
	systest.Sut.StopChain()

	t.Log("Upgrade height was reached. Upgrading chain")
	systest.Sut.SetExecBinary(currentBranchBinary)
	systest.Sut.SetTestnetInitializer(currentInitializer)
	systest.Sut.StartChain(t)
	cli = systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	// smoke test that new version runs
	ownerAddr := cli.GetKeyAddr("node0")
	got := cli.Run("tx", "accounts", "init", "continuous-locking-account", `{"end_time":"2034-01-22T11:38:15.116127Z", "owner":"`+ownerAddr+`"}`, "--from=node0")
	systest.RequireTxSuccess(t, got)
	got = cli.Run("tx", "protocolpool", "fund-community-pool", "100stake", "--from=node0")
	systest.RequireTxSuccess(t, got)
}

const cacheDir = "binaries"

// FetchExecutable to download and extract tar.gz for linux
func FetchExecutable(t *testing.T, version string) string {
	// use local cache
	cacheFolder := filepath.Join(systest.WorkDir, cacheDir)
	err := os.MkdirAll(cacheFolder, 0o777)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	cacheFile := filepath.Join(cacheFolder, fmt.Sprintf("%s_%s", systest.GetExecutableName(), version))
	if _, err := os.Stat(cacheFile); err == nil {
		return cacheFile
	}
	destFile := cacheFile
	t.Log("+++ version not in cache, downloading from docker image")
	systest.MustRunShellCmd(t, "docker", "pull", "ghcr.io/cosmos/simapp:"+version)
	systest.MustRunShellCmd(t, "docker", "create", "--name=ci_temp", "ghcr.io/cosmos/simapp:"+version)
	systest.MustRunShellCmd(t, "docker", "cp", "ci_temp:/usr/bin/simd", destFile)
	return destFile
}
