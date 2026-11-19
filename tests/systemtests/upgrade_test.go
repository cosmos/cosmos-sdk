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

// upgradeSources lists the supported upgrade paths to main. v0.55 is the
// default; set SYSTEMTEST_UPGRADE_FROM to run another.
var upgradeSources = map[string]upgradeSource{
	"v0.55": {binaryDir: "v0.55", upgradeName: "v055-to-v056", hasMaxSupply: true, hasKeyRotationFee: true},
}

func currentUpgradeSource(t *testing.T) upgradeSource {
	t.Helper()
	from := os.Getenv("SYSTEMTEST_UPGRADE_FROM")
	if from == "" {
		from = "v0.55"
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

	// Create a fee grant that expires after the upgrade. The EndBlocker must
	// remove it on the upgraded chain without halting; see the x/feegrant
	// v2 to v3 store migration for the failure this guards against.
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

	// let the votes commit before node0 signs more transactions
	systest.Sut.AwaitNBlocks(t, 2)
	seeded := seedLegacyState(t, cli)

	systest.Sut.AwaitBlockHeight(t, upgradeHeight-1, 60*time.Second)
	t.Logf("current_height: %d\n", systest.Sut.CurrentHeight())
	raw = cli.CustomQuery("q", "gov", "proposal", proposalID)
	proposalStatus := gjson.Get(raw, "proposal.status").String()
	require.Equal(t, "PROPOSAL_STATUS_PASSED", proposalStatus, raw)

	// the grants must still be live at the upgrade, so their expiry runs on the new binary
	require.True(t, time.Now().Before(seeded.expiration), "grants expired before the upgrade; raise expiryMargin")
	raw = cli.CustomQuery("q", "feegrant", "grant", seeded.granter, seeded.grantee)
	require.True(t, gjson.Get(raw, "allowance").Exists(), raw)

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

	// decode every collections entry the legacy binary and the migrations
	// wrote. The seeded grants must still be in state, or the decode would
	// pass on entries the EndBlockers already pruned.
	require.True(t, time.Now().Before(seeded.expiration), "seeded grants expired before the decode; raise expiryMargin")
	raw = cli.CustomQuery("q", "feegrant", "grant", seeded.granter, seeded.grantee)
	require.True(t, gjson.Get(raw, "allowance").Exists(), raw)
	raw = cli.CustomQuery("q", "authz", "grants", seeded.granter, seeded.grantee)
	require.NotEmpty(t, gjson.Get(raw, "grants").Array(), raw)
	systest.Sut.StopChain()
	for _, out := range systest.Sut.ForEachNodeExecAndWait(t, []string{"debug", "decode-state"}) {
		require.Contains(t, out[0], "all collections state decodes", out[0])
	}
	systest.Sut.StartChain(t)

	assertSeededStateAfterExpiry(t, cli, seeded)

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

// expiryMargin is how long after the estimated upgrade time the seeded grants
// expire. The grants must expire on the upgraded binary, so the EndBlocker of
// the new version prunes state that the legacy binary wrote, and must outlive
// the restart and the decode-state run that follow the upgrade. The fee grant
// TestChainUpgrade creates first expires about 20 blocks plus 30s after the
// upgrade, so the margin stays well below that.
const expiryMargin = 40 * time.Second

type seededState struct {
	granter, grantee string
	validator        string
	depositProposal  string
	expiration       time.Time
}

// seedLegacyState writes state on the legacy binary that the new binary must
// read back: a feegrant BasicAllowance and an authz grant that expire after the
// upgrade height, a delegation, and a deposit on a proposal that stays in its
// deposit period.
func seedLegacyState(t *testing.T, cli *systest.CLIWrapper) seededState {
	t.Helper()
	granter := cli.GetKeyAddr("node0")
	grantee := cli.AddKey("upgrade-grantee")

	blocksToUpgrade := upgradeHeight - systest.Sut.CurrentHeight()
	require.Positive(t, blocksToUpgrade)
	expiration := time.Now().Add(time.Duration(blocksToUpgrade)*systest.Sut.BlockTime() + expiryMargin).UTC().Truncate(time.Second)
	t.Logf("seeding grants that expire at %s", expiration)

	rsp := cli.Run("tx", "feegrant", "grant", granter, grantee,
		"--spend-limit=1000stake", "--expiration="+expiration.Format(time.RFC3339), "--from=node0")
	systest.RequireTxSuccess(t, rsp)

	rsp = cli.Run("tx", "authz", "grant", grantee, "send",
		"--spend-limit=1000stake", fmt.Sprintf("--expiration=%d", expiration.Unix()), "--from=node0")
	systest.RequireTxSuccess(t, rsp)

	raw := cli.CustomQuery("q", "staking", "validators")
	validator := gjson.Get(raw, "validators.0.operator_address").String()
	require.NotEmpty(t, validator, raw)
	rsp = cli.Run("tx", "staking", "delegate", validator, "1000000stake", "--from=node1")
	systest.RequireTxSuccess(t, rsp)

	// a deposit below the minimum keeps the proposal in its deposit period
	depositProposal := `{
 "messages": [],
 "metadata": "ipfs://CID",
 "deposit": "100000stake",
 "title": "deposit across upgrade",
 "summary": "testing"
}`
	rsp = cli.SubmitGovProposal(depositProposal, "--from=node1")
	systest.RequireTxSuccess(t, rsp)
	raw = cli.CustomQuery("q", "gov", "proposals", "--depositor", cli.GetKeyAddr("node1"))
	proposals := gjson.Get(raw, "proposals.#.id").Array()
	require.NotEmpty(t, proposals, raw)

	return seededState{
		granter:         granter,
		grantee:         grantee,
		validator:       validator,
		depositProposal: proposals[len(proposals)-1].String(),
		expiration:      expiration,
	}
}

// assertSeededStateAfterExpiry waits on the upgraded chain until the seeded
// grants have expired, requires the chain to keep producing blocks, and
// requires the EndBlockers to have pruned the expired grants while the
// delegation and the deposit remain.
func assertSeededStateAfterExpiry(t *testing.T, cli *systest.CLIWrapper, seeded seededState) {
	t.Helper()
	t.Logf("waiting until %s for the seeded grants to expire", seeded.expiration)
	for time.Now().Before(seeded.expiration) {
		systest.Sut.AwaitNextBlock(t)
	}
	// blocks past the expiry show the EndBlockers ran over the expired entries
	systest.Sut.AwaitNBlocks(t, 3)

	raw := cli.WithRunErrorsIgnored().CustomQuery("q", "feegrant", "grant", seeded.granter, seeded.grantee)
	require.False(t, gjson.Get(raw, "allowance").Exists(), "expired fee allowance still present: %s", raw)
	raw = cli.CustomQuery("q", "feegrant", "grants-by-grantee", seeded.grantee)
	require.Empty(t, gjson.Get(raw, "allowances").Array(), raw)

	raw = cli.CustomQuery("q", "authz", "grants", seeded.granter, seeded.grantee)
	require.Empty(t, gjson.Get(raw, "grants").Array(), raw)

	raw = cli.CustomQuery("q", "staking", "delegation", cli.GetKeyAddr("node1"), seeded.validator)
	require.True(t, gjson.Get(raw, "delegation_response.balance.amount").Exists(), raw)

	raw = cli.CustomQuery("q", "gov", "deposits", seeded.depositProposal)
	require.NotEmpty(t, gjson.Get(raw, "deposits").Array(), raw)
}
