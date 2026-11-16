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
	// then the chain upgrades successfully, keeps producing blocks past the
	// fee grant expiry, and x/safety, which the upgrade adds, trips and resets
	// msgs
	cli, grant := upgradeLegacyChain(t)
	govAddr := sdk.AccAddress(address.Module("gov")).String()

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

	// x/safety is new in this version: the upgrade adds its store and runs its
	// InitGenesis with the default params
	safetyParamsRaw := safetyParams(t)
	require.Equal(t, govAddr, gjson.Get(safetyParamsRaw, "params.authority").String(), safetyParamsRaw)
	require.Empty(t, gjson.Get(safetyParamsRaw, "params.operator").String(), safetyParamsRaw)
	require.Empty(t, safetyDisabledList(t))
	require.True(t, safetyIsAllowed(t, safetyMsgSend))

	regex, err := regexp.Compile("DBG this is a debug level message to test that verbose logging mode has properly been enabled during a chain upgrade")
	require.NoError(t, err)
	require.Equal(t, systest.Sut.NodesCount(), systest.Sut.FindLogMessage(regex))

	// smoke test that new version runs
	cli = systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	got := cli.Run("tx", "distribution", "fund-community-pool", "100stake", "--from=node0")
	systest.RequireTxSuccess(t, got)

	// the chain must keep producing blocks past the fee grant expiry, and the
	// EndBlocker must have removed the expired grant
	require.True(t, time.Now().Before(grant.expiry), "fee grant expired before the upgrade; raise its expiry")
	time.Sleep(time.Until(grant.expiry))
	systest.Sut.AwaitNBlocks(t, 3)
	rsp := cli.WithRunErrorsIgnored().CustomQuery("q", "feegrant", "grant", grant.granter, grant.grantee)
	require.Contains(t, rsp, "not found", rsp)

	safetyAfterUpgrade(t, cli, govAddr)
}

// upgradeFeeGrant is the fee grant upgradeLegacyChain creates on the legacy
// chain, set to expire after the upgrade.
type upgradeFeeGrant struct {
	granter, grantee string
	expiry           time.Time
}

// safetyAfterUpgrade has gov name an x/safety operator on the upgraded chain
// and the operator trip and reset msgs. The trips block those msgs in CheckTx
// and survive a restart, gov still submits and votes, and the operator and
// gov resets reopen them.
func safetyAfterUpgrade(t *testing.T, cli *systest.CLIWrapper, govAddr string) {
	t.Helper()
	// keys added after the upgrade, funded by bank sends that x/safety allows
	operatorAddr := cli.AddKey("safety-operator")
	recipientAddr := cli.AddKey("safety-recipient")
	cli.FundAddress(operatorAddr, "10000000stake")
	require.Equal(t, int64(10000000), cli.QueryBalance(operatorAddr, "stake"))

	// gov names the operator with scope /cosmos.bank.*
	newParams := fmt.Sprintf(`{"authority":%q,"operator":%q,"scope":["/cosmos.bank.*"]}`, govAddr, operatorAddr)
	safetySubmitAndPassProposal(t, cli, "tx", "gov", "submit-proposal",
		safetyProposalFile(t, safetyUpdateParams(govAddr, newParams), "name safety operator", "grant /cosmos.bank.* scope"),
		"--from", "node0")
	params := safetyParams(t)
	require.Equal(t, operatorAddr, gjson.Get(params, "params.operator").String(), params)
	require.Equal(t, []string{"/cosmos.bank.*"}, safetyStrings(gjson.Get(params, "params.scope")), params)

	safetyRequireTxRejected(t, cli, "not covered by operator scope",
		operatorAddr, safetyTrip(operatorAddr, "/cosmos.distribution.v1beta1.MsgFundCommunityPool"))

	// the operator trips MsgSend
	safetyTx(t, cli, operatorAddr, safetyTrip(operatorAddr, safetyMsgSend))
	require.Equal(t, []string{safetyMsgSend}, safetyDisabledList(t))
	require.False(t, safetyIsAllowed(t, safetyMsgSend))

	out, _ := cli.WithRunErrorsIgnored().RunOnly("tx", "bank", "send", "node0", recipientAddr, "100stake", safetyFees)
	upgradeRequireSafetyRejected(t, out, safetyMsgSend+" by "+safetyMsgSend)
	require.Zero(t, cli.QueryBalance(recipientAddr, "stake"))

	// other msgs still run, and gov still submits and votes: gov widens the
	// operator scope to *
	safetyRequireCommittedOK(t, cli.Run("tx", "distribution", "fund-community-pool", "100stake", "--from", "node0"))
	newParams = fmt.Sprintf(`{"authority":%q,"operator":%q,"scope":["*"]}`, govAddr, operatorAddr)
	safetySubmitAndPassProposal(t, cli, "tx", "gov", "submit-proposal",
		safetyProposalFile(t, safetyUpdateParams(govAddr, newParams), "widen safety scope", "grant * scope"),
		"--from", "node0")
	params = safetyParams(t)
	require.Equal(t, []string{"*"}, safetyStrings(gjson.Get(params, "params.scope")), params)

	// the operator resets MsgSend and the send goes through
	safetyTx(t, cli, operatorAddr, safetyReset(operatorAddr, safetyMsgSend))
	require.Empty(t, safetyDisabledList(t))
	require.True(t, safetyIsAllowed(t, safetyMsgSend))
	safetyRequireCommittedOK(t, cli.Run("tx", "bank", "send", "node0", recipientAddr, "100stake"))
	require.Equal(t, int64(100), cli.QueryBalance(recipientAddr, "stake"))

	// the operator trips *, which now blocks the community pool fund too
	safetyTx(t, cli, operatorAddr, safetyTrip(operatorAddr, "*"))
	out, _ = cli.WithRunErrorsIgnored().RunOnly("tx", "distribution", "fund-community-pool", "100stake", "--from", "node0", safetyFees)
	upgradeRequireSafetyRejected(t, out, "/cosmos.distribution.v1beta1.MsgFundCommunityPool by *")

	// the trip survives a restart
	systest.Sut.StopChain()
	systest.Sut.StartChain(t)
	cli = systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	require.Equal(t, []string{"*"}, safetyDisabledList(t))
	out, _ = cli.WithRunErrorsIgnored().RunOnly("tx", "bank", "send", "node0", recipientAddr, "100stake", safetyFees)
	upgradeRequireSafetyRejected(t, out, safetyMsgSend+" by *")

	// gov submits, votes on and executes a reset of * while * is tripped
	safetySubmitAndPassProposal(t, cli, "tx", "gov", "submit-proposal",
		safetyProposalFile(t, safetyReset(govAddr, "*"), "reset safety", "clear the * trip"), "--from", "node0")

	require.Empty(t, safetyDisabledList(t))
	safetyRequireCommittedOK(t, cli.Run("tx", "bank", "send", "node0", recipientAddr, "100stake"))
	require.Equal(t, int64(200), cli.QueryBalance(recipientAddr, "stake"))
	safetyRequireCommittedOK(t, cli.Run("tx", "distribution", "fund-community-pool", "100stake", "--from", "node0"))
}

// upgradeLegacyChain starts a chain on the legacy binary of the current
// upgrade source, creates a fee grant that expires after the upgrade, passes a
// gov software upgrade, and restarts the chain on the current binary past the
// upgrade height.
func upgradeLegacyChain(t *testing.T) (*systest.CLIWrapper, upgradeFeeGrant) {
	t.Helper()
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
	return systest.NewCLIWrapper(t, systest.Sut, systest.Verbose), upgradeFeeGrant{granter: granter, grantee: grantee, expiry: grantExpiry}
}

// upgradeRequireSafetyRejected requires a broadcast to have failed in CheckTx
// with the x/safety ErrMsgDisabled code, and its log to name the blocked msg
// type URL and the disabled pattern as "<type URL> by <pattern>".
func upgradeRequireSafetyRejected(t *testing.T, out, blockedBy string) {
	t.Helper()
	safetyRequireCheckTxRejected(t, out, safetyDisabledErr)
	js := safetyTrimJSON(out)
	require.True(t, gjson.Valid(js), out)
	require.Equal(t, "safety", gjson.Get(js, "codespace").String(), out)
	require.Equal(t, int64(10), gjson.Get(js, "code").Int(), out)
	require.Contains(t, gjson.Get(js, "raw_log").String(), blockedBy, out)
}
