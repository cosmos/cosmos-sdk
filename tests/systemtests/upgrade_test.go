//go:build system_test

package systemtests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	systest "cosmossdk.io/systemtests"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	testSeed            = "scene learn remember glide apple expand quality spawn property shoe lamp carry upset blossom draft reject aim file trash miss script joy only measure"
	upgradeHeight int64 = 22
	upgradeName         = "v050-to-v053" // must match UpgradeName in simapp/upgrades.go
)

type initAccount struct {
	address string
	balance string
}

func createLegacyBinary(t *testing.T, extraAccounts ...initAccount) (*systest.CLIWrapper, *systest.SystemUnderTest) {
	t.Helper()

	legacyBinary := systest.WorkDir + "/binaries/v0.50/simd"

	//// Now we're going to switch to a v.50 chain.
	t.Logf("+++ legacy binary: %s\n", legacyBinary)

	// setup the v50 chain. v53 made some changes to testnet command, so we'll have to adjust here.
	// this only uses 1 node.
	legacySut := systest.NewSystemUnderTest("simd", systest.Verbose, 1, 1*time.Second)
	// we need to explicitly set this here as the constructor infers the exec binary is in the "binaries" directory.
	legacySut.SetExecBinary(legacyBinary)
	legacySut.SetTestnetInitializer(systest.LegacyInitializerWithBinary(legacyBinary, legacySut))
	legacySut.SetupChain()
	v50CLI := systest.NewCLIWrapper(t, legacySut, systest.Verbose)
	v50CLI.AddKeyFromSeed("account1", testSeed)

	// Typically, SystemUnderTest will create a node with 4 validators. In the legacy setup, we create run a single validator network.
	// This means we need to add 3 more accounts in order to make further account additions map to the same account number in state
	modifications := [][]string{
		{"genesis", "add-genesis-account", v50CLI.AddKey("foo"), "10000000000stake"},
		{"genesis", "add-genesis-account", v50CLI.AddKey("bar"), "10000000000stake"},
		{"genesis", "add-genesis-account", v50CLI.AddKey("baz"), "10000000000stake"},
	}
	for _, extraAccount := range extraAccounts {
		modifications = append(modifications, []string{"genesis", "add-genesis-account", extraAccount.address, extraAccount.balance})
	}

	legacySut.ModifyGenesisCLI(t,
		modifications...,
	)

	return v50CLI, legacySut
}

func TestChainUpgrade(t *testing.T) {
	// Scenario:
	// start a legacy chain with some state
	// when a chain upgrade proposal is executed
	// then the chain upgrades successfully
	systest.Sut.StopChain()

	currentBranchBinary := systest.Sut.ExecBinary()
	currentInitializer := systest.Sut.TestnetInitializer()

	legacyBinary := systest.WorkDir + "/binaries/v0.50/simd"
	systest.Sut.SetExecBinary(legacyBinary)
	systest.Sut.SetTestnetInitializer(systest.NewModifyConfigYamlInitializer(legacyBinary, systest.Sut))
	systest.Sut.SetupChain()

	votingPeriod := 5 * time.Second // enough time to vote
	systest.Sut.ModifyGenesisJSON(t, systest.SetGovVotingPeriod(t, votingPeriod))

	systest.Sut.StartChain(t, fmt.Sprintf("--halt-height=%d", upgradeHeight+1))

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
	require.Equal(t, "PROPOSAL_STATUS_PASSED", proposalStatus, raw)

	t.Log("waiting for upgrade info")
	systest.Sut.AwaitUpgradeInfo(t)
	systest.Sut.StopChain()

	t.Log("Upgrade height was reached. Upgrading chain")
	systest.Sut.SetExecBinary(currentBranchBinary)
	systest.Sut.SetTestnetInitializer(currentInitializer)
	systest.Sut.StartChain(t)

	require.Equal(t, upgradeHeight+1, systest.Sut.CurrentHeight())

	// smoke test that new version runs
	cli = systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	got := cli.Run("tx", "protocolpool", "fund-community-pool", "100stake", "--from=node0")
	systest.RequireTxSuccess(t, got)
}
