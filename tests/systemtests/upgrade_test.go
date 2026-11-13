//go:build system_test

package systemtests

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	systest "github.com/cosmos/cosmos-sdk/tools/systemtests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	testSeed            = "scene learn remember glide apple expand quality spawn property shoe lamp carry upset blossom draft reject aim file trash miss script joy only measure"
	upgradeHeight int64 = 22
)

// upgradeSource describes a legacy release that TestChainUpgrade upgrades from.
type upgradeSource struct {
	// binaryDir is the directory under binaries/ that holds the legacy simd.
	binaryDir string
	// upgradeName must match a handler registered in simapp/upgrades.go.
	upgradeName string
	// hasMaxSupply reports whether the legacy x/mint params include max_supply.
	hasMaxSupply bool
	// hasKeyRotationFee reports whether the legacy x/staking params include key_rotation_fee.
	hasKeyRotationFee bool
}

// upgradeSources lists the supported upgrade paths to v0.55.x. v0.53 is the
// default; set SYSTEMTEST_UPGRADE_FROM to run another.
var upgradeSources = map[string]upgradeSource{
	"v0.53":   {binaryDir: "v0.53", upgradeName: "v053-to-v055"},
	"v0.54":   {binaryDir: "v0.54", upgradeName: "v054-to-v055", hasMaxSupply: true},
	"v0.55.0": {binaryDir: "v0.55.0", upgradeName: "v055-to-v0551", hasMaxSupply: true, hasKeyRotationFee: true},
}

func currentUpgradeSource(t *testing.T) upgradeSource {
	t.Helper()
	from := os.Getenv("SYSTEMTEST_UPGRADE_FROM")
	if from == "" {
		from = "v0.53"
	}
	src, ok := upgradeSources[from]
	require.True(t, ok, "unknown SYSTEMTEST_UPGRADE_FROM %q", from)
	return src
}

func TestChainUpgrade(t *testing.T) {
	// Scenario:
	// start a legacy chain with some state, including a fee grant that
	// expires after the upgrade
	// when a chain upgrade proposal is executed
	// then the chain upgrades successfully and keeps producing blocks past
	// the fee grant expiry
	src := currentUpgradeSource(t)
	systest.Sut.StopChain()

	currentBranchBinary := systest.Sut.ExecBinary()
	currentInitializer := systest.Sut.TestnetInitializer()

	legacyBinary := systest.WorkDir + "/binaries/" + src.binaryDir + "/simd"
	systest.Sut.SetExecBinary(legacyBinary)
	systest.Sut.SetTestnetInitializer(systest.InitializerWithBinary(legacyBinary, systest.Sut))
	systest.Sut.SetupChain()
	require.Equal(t, src.hasMaxSupply, gjson.Get(systest.Sut.ReadGenesisJSON(t), "app_state.mint.params.max_supply").Exists())

	votingPeriod := 5 * time.Second // enough time to vote
	systest.Sut.ModifyGenesisJSON(t, systest.SetGovVotingPeriod(t, votingPeriod))

	systest.Sut.StartChain(t, fmt.Sprintf("--halt-height=%d", upgradeHeight+1))

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	// Create a fee grant that expires after the upgrade. Grants created by the
	// standalone x/feegrant module (v0.53.x) store an empty expiry queue value,
	// which the v2 to v3 store migration rewrites. Without it, the EndBlocker
	// fails and the chain halts when the grant expires.
	granter := cli.GetKeyAddr("node0")
	grantee := cli.AddKey("feegrantee")
	blocksToUpgrade := upgradeHeight - systest.Sut.CurrentHeight()
	grantExpiry := time.Now().Add(time.Duration(blocksToUpgrade+20)*systest.Sut.BlockTime() + 30*time.Second).UTC()
	rsp := cli.RunAndWait("tx", "feegrant", "grant", granter, grantee,
		"--spend-limit=1000stake",
		"--expiration="+grantExpiry.Format(time.RFC3339),
		"--from=node0",
	)
	systest.RequireTxSuccess(t, rsp)

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
}`, govAddr, src.upgradeName, upgradeHeight)
	proposalID := cli.SubmitAndVoteGovProposal(proposal)
	t.Logf("current_height: %d\n", systest.Sut.CurrentHeight())
	raw := cli.CustomQuery("q", "gov", "proposal", proposalID)
	t.Log(raw)

	systest.Sut.AwaitBlockHeight(t, upgradeHeight-1, 60*time.Second)
	t.Logf("current_height: %d\n", systest.Sut.CurrentHeight())
	raw = cli.CustomQuery("q", "gov", "proposal", proposalID)
	proposalStatus := gjson.Get(raw, "proposal.status").String()
	require.Equal(t, "PROPOSAL_STATUS_PASSED", proposalStatus, raw)

	legacyStakingParamsRaw := cli.CustomQuery("q", "staking", "params")
	require.Equal(t, src.hasKeyRotationFee,
		gjson.Get(legacyStakingParamsRaw, "params.key_rotation_fee").Exists(),
		legacyStakingParamsRaw,
	)

	t.Log("waiting for upgrade info")
	systest.Sut.AwaitUpgradeInfo(t)
	systest.Sut.StopChain()

	t.Log("Upgrade height was reached. Upgrading chain")
	systest.Sut.SetExecBinary(currentBranchBinary)
	systest.Sut.SetTestnetInitializer(currentInitializer)
	systest.Sut.StartChain(t)

	require.True(t, upgradeHeight+1 <= systest.Sut.CurrentHeight())
	cli = systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	mintParamsRaw := cli.CustomQuery("q", "mint", "params")
	require.True(t,
		gjson.Get(mintParamsRaw, "params.max_supply").Exists() || gjson.Get(mintParamsRaw, "params.maxSupply").Exists(),
		mintParamsRaw,
	)

	stakingParamsRaw := cli.CustomQuery("q", "staking", "params")
	bondDenom := gjson.Get(stakingParamsRaw, "params.bond_denom")
	require.True(t, bondDenom.Exists(), stakingParamsRaw)

	keyRotationFee := gjson.Get(stakingParamsRaw, "params.key_rotation_fee")
	require.True(t, keyRotationFee.Exists(), stakingParamsRaw)
	require.Equal(t, "1000000", keyRotationFee.Get("amount").String(), stakingParamsRaw)
	require.Equal(t, bondDenom.String(), keyRotationFee.Get("denom").String(), stakingParamsRaw)

	regex, err := regexp.Compile("DBG this is a debug level message to test that verbose logging mode has properly been enabled during a chain upgrade")
	require.NoError(t, err)
	require.Equal(t, systest.Sut.NodesCount(), systest.Sut.FindLogMessage(regex))

	// smoke test that new version runs
	cli = systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	got := cli.Run("tx", "distribution", "fund-community-pool", "100stake", "--from=node0")
	systest.RequireTxSuccess(t, got)

	// the chain must keep producing blocks past the fee grant expiry, and the
	// EndBlocker must have removed the expired grant
	require.True(t, time.Now().Before(grantExpiry), "fee grant expired before the upgrade; raise its expiry")
	time.Sleep(time.Until(grantExpiry))
	systest.Sut.AwaitNBlocks(t, 3)
	rsp = cli.WithRunErrorsIgnored().CustomQuery("q", "feegrant", "grant", granter, grantee)
	require.Contains(t, rsp, "not found", rsp)
}
