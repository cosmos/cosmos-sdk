/*
//go:build system_test
*/
package systemtests

import (
	"fmt"
	"testing"
	"time"

	systest "cosmossdk.io/systemtests"
)

type initAccount struct {
	address string
	balance string
}

func createLegacyBinary(t *testing.T, extraAccounts ...initAccount) (*systest.CLIWrapper, *systest.SystemUnderTest) {
	//// Now we're going to switch to a v.50 chain.
	legacyBinary := systest.WorkDir + "/binaries/v0.50/simd"

	// setup the v50 chain. v53 made some changes to testnet command, so we'll have to adjust here.
	// this only uses 1 node.
	legacySut := systest.NewSystemUnderTest("simd", systest.Verbose, 1, 1*time.Second)
	// we need to explicitly set this here as the constructor infers the exec binary is in the "binaries" directory.
	legacySut.SetExecBinary(legacyBinary)
	legacySut.SetTestnetInitializer(systest.LegacyInitializerWithBinary(legacyBinary, legacySut))
	legacySut.SetupChain()
	v50CLI := systest.NewCLIWrapper(t, legacySut, systest.Verbose)
	v50CLI.AddKeyFromSeed("account1", testSeed)

	var extraArgs [][]string
	for _, extraAccount := range extraAccounts {
		extraArgs = append(extraArgs, []string{"genesis", "add-genesis-account", extraAccount.address, extraAccount.balance})
	}

	legacySut.ModifyGenesisCLI(t,
		// add some bogus accounts because the v53 chain had 4 nodes which takes account numbers 1-4.
		[]string{"genesis", "add-genesis-account", v50CLI.AddKey("foo"), "10000000000stake"},
		[]string{"genesis", "add-genesis-account", v50CLI.AddKey("bar"), "10000000000stake"},
		[]string{"genesis", "add-genesis-account", v50CLI.AddKey("baz"), "10000000000stake"},
		extraArgs...,
	)

	return v50CLI, legacySut
}

/*
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
		upgradeName         = "v050-to-v053" // must match UpgradeName in simapp/upgrades.go
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
*/
