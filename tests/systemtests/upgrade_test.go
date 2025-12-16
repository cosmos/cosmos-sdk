//go:build system_test

package systemtests

import (
	"fmt"
	"regexp"
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
	upgradeName         = "v053-to-v054" // must match UpgradeName in simapp/upgrades.go
)

func TestChainUpgrade(t *testing.T) {
	// Scenario:
	// start a legacy chain with some state
	// when a chain upgrade proposal is executed
	// then the chain upgrades successfully
	systest.Sut.StopChain()

	currentBranchBinary := systest.Sut.ExecBinary()
	currentInitializer := systest.Sut.TestnetInitializer()

	legacyBinary := systest.WorkDir + "/binaries/v0.53/simd"
	systest.Sut.SetExecBinary(legacyBinary)
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

	require.True(t, upgradeHeight+1 <= systest.Sut.CurrentHeight())

	regex, err := regexp.Compile("DBG this is a debug level message to test that verbose logging mode has properly been enabled during a chain upgrade")
	require.NoError(t, err)
	require.Equal(t, systest.Sut.NodesCount(), systest.Sut.FindLogMessage(regex))

	// smoke test that new version runs
	cli = systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	got := cli.Run("tx", "protocolpool", "fund-community-pool", "100stake", "--from=node0")
	systest.RequireTxSuccess(t, got)
}
