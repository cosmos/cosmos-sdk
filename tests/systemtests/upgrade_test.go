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

	systest "github.com/cosmos/cosmos-sdk/tools/systemtests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	testSeed            = "scene learn remember glide apple expand quality spawn property shoe lamp carry upset blossom draft reject aim file trash miss script joy only measure"
	upgradeHeight int64 = 22
	upgradeName         = "v054-to-v055" // must match UpgradeName in simapp/upgrades.go
)

func TestChainUpgrade(t *testing.T) {
	// Scenario:
	// start a legacy chain with some state
	// when a chain upgrade proposal is executed
	// then the chain upgrades successfully
	cli := upgradeLegacyChain(t)
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
	safetyParamsRaw := cli.CustomQuery("q", "safety", "params")
	require.Equal(t, govAddr, gjson.Get(safetyParamsRaw, "params.authority").String(), safetyParamsRaw)
	require.Empty(t, gjson.Get(safetyParamsRaw, "params.operator").String(), safetyParamsRaw)
	require.Empty(t, safetyDisabledList(t, cli))
	require.True(t, safetyIsAllowed(t, cli, safetyMsgSend))

	regex, err := regexp.Compile("DBG this is a debug level message to test that verbose logging mode has properly been enabled during a chain upgrade")
	require.NoError(t, err)
	require.Equal(t, systest.Sut.NodesCount(), systest.Sut.FindLogMessage(regex))

	// smoke test that new version runs
	cli = systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	got := cli.Run("tx", "distribution", "fund-community-pool", "100stake", "--from=node0")
	systest.RequireTxSuccess(t, got)
}

func TestSafetyAfterChainUpgrade(t *testing.T) {
	// Scenario:
	// given a legacy chain upgraded to the current binary, which adds x/safety
	// when gov names an operator and the operator trips and resets msgs
	// then the trips block those msgs in CheckTx and survive a restart, gov
	// still submits and votes, and the operator and gov resets reopen them.
	cli := upgradeLegacyChain(t)
	govAddr := sdk.AccAddress(address.Module("gov")).String()

	// keys added after the upgrade, funded by bank sends that x/safety allows
	operatorAddr := cli.AddKey("safety-operator")
	recipientAddr := cli.AddKey("safety-recipient")
	cli.FundAddress(operatorAddr, "10000000stake")
	require.Equal(t, int64(10000000), cli.QueryBalance(operatorAddr, "stake"))

	// gov names the operator with scope /cosmos.bank.*
	newParams := fmt.Sprintf(`{"authority":%q,"operator":%q,"scope":["/cosmos.bank.*"]}`, govAddr, operatorAddr)
	safetySubmitAndPassProposal(t, cli,
		"tx", "safety", "update-params-proposal", newParams,
		"--title", "name safety operator", "--summary", "grant /cosmos.bank.* scope",
		"--deposit", "100000000stake", "--from", "node0")
	params := cli.CustomQuery("q", "safety", "params")
	require.Equal(t, operatorAddr, gjson.Get(params, "params.operator").String(), params)
	require.Equal(t, []string{"/cosmos.bank.*"}, safetyStrings(gjson.Get(params, "params.scope")), params)

	safetyRequireTxRejected(t, cli, "not covered by operator scope",
		"tx", "safety", "trip", "/cosmos.distribution.v1beta1.MsgFundCommunityPool", "--from", operatorAddr)

	// the operator trips MsgSend
	safetyRequireCommittedOK(t, cli.Run("tx", "safety", "trip", safetyMsgSend, "--from", operatorAddr))
	require.Equal(t, []string{safetyMsgSend}, safetyDisabledList(t, cli))
	require.False(t, safetyIsAllowed(t, cli, safetyMsgSend))

	out, _ := cli.WithRunErrorsIgnored().RunOnly("tx", "bank", "send", "node0", recipientAddr, "100stake", safetyFees)
	upgradeRequireSafetyRejected(t, out, safetyMsgSend+" by "+safetyMsgSend)
	require.Zero(t, cli.QueryBalance(recipientAddr, "stake"))

	// other msgs still run, and gov still submits and votes: gov widens the
	// operator scope to *
	safetyRequireCommittedOK(t, cli.Run("tx", "distribution", "fund-community-pool", "100stake", "--from", "node0"))
	newParams = fmt.Sprintf(`{"authority":%q,"operator":%q,"scope":["*"]}`, govAddr, operatorAddr)
	safetySubmitAndPassProposal(t, cli,
		"tx", "safety", "update-params-proposal", newParams,
		"--title", "widen safety scope", "--summary", "grant * scope",
		"--deposit", "100000000stake", "--from", "node0")
	params = cli.CustomQuery("q", "safety", "params")
	require.Equal(t, []string{"*"}, safetyStrings(gjson.Get(params, "params.scope")), params)

	// the operator resets MsgSend and the send goes through
	safetyRequireCommittedOK(t, cli.Run("tx", "safety", "reset", safetyMsgSend, "--from", operatorAddr))
	require.Empty(t, safetyDisabledList(t, cli))
	require.True(t, safetyIsAllowed(t, cli, safetyMsgSend))
	safetyRequireCommittedOK(t, cli.Run("tx", "bank", "send", "node0", recipientAddr, "100stake"))
	require.Equal(t, int64(100), cli.QueryBalance(recipientAddr, "stake"))

	// the operator trips *, which now blocks the community pool fund too
	safetyRequireCommittedOK(t, cli.Run("tx", "safety", "trip", "*", "--from", operatorAddr))
	out, _ = cli.WithRunErrorsIgnored().RunOnly("tx", "distribution", "fund-community-pool", "100stake", "--from", "node0", safetyFees)
	upgradeRequireSafetyRejected(t, out, "/cosmos.distribution.v1beta1.MsgFundCommunityPool by *")

	// the trip survives a restart
	systest.Sut.StopChain()
	systest.Sut.StartChain(t)
	cli = systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	require.Equal(t, []string{"*"}, safetyDisabledList(t, cli))
	out, _ = cli.WithRunErrorsIgnored().RunOnly("tx", "bank", "send", "node0", recipientAddr, "100stake", safetyFees)
	upgradeRequireSafetyRejected(t, out, safetyMsgSend+" by *")

	// gov submits, votes on and executes a reset of * while * is tripped
	proposal := fmt.Sprintf(`{
 "messages": [
  {
   "@type": "/cosmos.safety.v1.MsgReset",
   "sender": %q,
   "patterns": ["*"]
  }
 ],
 "metadata": "ipfs://CID",
 "deposit": "100000000stake",
 "title": "reset safety",
 "summary": "clear the * trip"
}`, govAddr)
	proposalFile := filepath.Join(t.TempDir(), "proposal.json")
	require.NoError(t, os.WriteFile(proposalFile, []byte(proposal), 0o600))
	safetySubmitAndPassProposal(t, cli, "tx", "gov", "submit-proposal", proposalFile, "--from", "node0")

	require.Empty(t, safetyDisabledList(t, cli))
	safetyRequireCommittedOK(t, cli.Run("tx", "bank", "send", "node0", recipientAddr, "100stake"))
	require.Equal(t, int64(200), cli.QueryBalance(recipientAddr, "stake"))
	safetyRequireCommittedOK(t, cli.Run("tx", "distribution", "fund-community-pool", "100stake", "--from", "node0"))
}

// upgradeLegacyChain starts a chain on the v0.54 binary, passes a gov
// software upgrade, and restarts the chain on the current binary past the
// upgrade height.
func upgradeLegacyChain(t *testing.T) *systest.CLIWrapper {
	t.Helper()
	systest.Sut.StopChain()

	currentBranchBinary := systest.Sut.ExecBinary()
	currentInitializer := systest.Sut.TestnetInitializer()

	legacyBinary := systest.WorkDir + "/binaries/v0.54/simd"
	systest.Sut.SetExecBinary(legacyBinary)
	systest.Sut.SetTestnetInitializer(systest.InitializerWithBinary(legacyBinary, systest.Sut))
	systest.Sut.SetupChain()
	require.True(t, gjson.Get(systest.Sut.ReadGenesisJSON(t), "app_state.mint.params.max_supply").Exists())

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

	legacyStakingParamsRaw := cli.CustomQuery("q", "staking", "params")
	require.False(t,
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
	return systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
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
