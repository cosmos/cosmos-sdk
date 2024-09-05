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

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

func TestChainUpgrade(t *testing.T) {
	// Scenario:
	// start a legacy chain with some state
	// when a chain upgrade proposal is executed
	// then the chain upgrades successfully
	sut.StopChain()

	legacyBinary := FetchExecutable(t, "v0.50")
	t.Logf("+++ legacy binary: %s\n", legacyBinary)
	currentBranchBinary := sut.execBinary
	currentInitializer := sut.testnetInitializer
	sut.SetExecBinary(legacyBinary)
	sut.SetTestnetInitializer(NewModifyConfigYamlInitializer(legacyBinary, sut))
	sut.SetupChain()
	votingPeriod := 5 * time.Second // enough time to vote
	sut.ModifyGenesisJSON(t, SetGovVotingPeriod(t, votingPeriod))

	const (
		upgradeHeight int64 = 22
		upgradeName         = "v050-to-v051"
	)

	sut.StartChain(t, fmt.Sprintf("--halt-height=%d", upgradeHeight))

	cli := NewCLIWrapper(t, sut, verbose)
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
	t.Logf("current_height: %d\n", sut.currentHeight)
	raw := cli.CustomQuery("q", "gov", "proposal", proposalID)
	t.Log(raw)

	sut.AwaitBlockHeight(t, upgradeHeight-1, 60*time.Second)
	t.Logf("current_height: %d\n", sut.currentHeight)
	raw = cli.CustomQuery("q", "gov", "proposal", proposalID)
	proposalStatus := gjson.Get(raw, "proposal.status").String()
	require.Equal(t, "PROPOSAL_STATUS_PASSED", proposalStatus, raw) // PROPOSAL_STATUS_PASSED

	t.Log("waiting for upgrade info")
	sut.AwaitUpgradeInfo(t)
	sut.StopChain()

	t.Log("Upgrade height was reached. Upgrading chain")
	sut.SetExecBinary(currentBranchBinary)
	sut.SetTestnetInitializer(currentInitializer)
	sut.StartChain(t)
	cli = NewCLIWrapper(t, sut, verbose)

	// smoke test that new version runs
	ownerAddr := cli.GetKeyAddr(defaultSrcAddr)
	got := cli.Run("tx", "accounts", "init", "continuous-locking-account", `{"end_time":"2034-01-22T11:38:15.116127Z", "owner":"`+ownerAddr+`"}`, "--from="+defaultSrcAddr)
	RequireTxSuccess(t, got)
	got = cli.Run("tx", "protocolpool", "fund-community-pool", "100stake", "--from="+defaultSrcAddr)
	RequireTxSuccess(t, got)
}

const cacheDir = "binaries"

// FetchExecutable to download and extract tar.gz for linux
func FetchExecutable(t *testing.T, version string) string {
	// use local cache
	cacheFolder := filepath.Join(WorkDir, cacheDir)
	err := os.MkdirAll(cacheFolder, 0o777)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	cacheFile := filepath.Join(cacheFolder, fmt.Sprintf("%s_%s", execBinaryName, version))
	if _, err := os.Stat(cacheFile); err == nil {
		return cacheFile
	}
	destFile := cacheFile
	t.Log("+++ version not in cache, downloading from docker image")
	MustRunShellCmd(t, "docker", "pull", "ghcr.io/cosmos/simapp:"+version)
	MustRunShellCmd(t, "docker", "create", "--name=ci_temp", "ghcr.io/cosmos/simapp:"+version)
	MustRunShellCmd(t, "docker", "cp", "ci_temp:/usr/bin/simd", destFile)
	return destFile
}
