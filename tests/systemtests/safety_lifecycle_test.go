//go:build system_test

package systemtests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	systest "github.com/cosmos/cosmos-sdk/tools/systemtests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

func TestSafetyGovManagesOperator(t *testing.T) {
	// Scenario:
	// given a running chain whose genesis names an operator with scope
	// /cosmos.bank.* and /cosmos.staking.*, and the operator has tripped one
	// pattern in each
	// when gov tries to trip, narrows the scope, then replaces the operator and
	// resets the entry the old scope covered
	// then the gov trip fails, narrowing keeps both entries but takes the
	// staking one out of the operator's reach, and the old operator loses all
	// power while the new one trips and resets.

	// the operator address goes into genesis, so derive it from a fixed seed and
	// add the key again after ModifyGenesisJSON resets the keyring
	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	govAddr := sdk.AccAddress(address.Module("gov")).String()
	oldOperator := cli.AddKeyFromSeed("safety-operator", testSeed)
	systest.Sut.ModifyGenesisJSON(t,
		safetyShortGovPeriods(t),
		safetySetGenesis(t, oldOperator, []string{"/cosmos.bank.*", "/cosmos.staking.*"}, nil),
	)
	require.Equal(t, oldOperator, cli.AddKeyFromSeed("safety-operator", testSeed))
	newOperator := cli.AddKey("safety-new-operator")
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", oldOperator, "10000000stake"},
		[]string{"genesis", "add-genesis-account", newOperator, "10000000stake"},
	)
	systest.Sut.StartChain(t)

	safetyTx(t, cli, oldOperator, safetyTrip(oldOperator, safetyMsgSend, "/cosmos.staking.*"))
	tripped := []string{safetyMsgSend, "/cosmos.staking.*"}
	require.Equal(t, tripped, safetyDisabledList(t))

	// gov is the authority but not the operator, so a trip it passes fails at
	// execution
	_, proposal := safetySubmitAndAwaitProposal(t, cli, "tx", "gov", "submit-proposal",
		safetyProposalFile(t, safetyTrip(govAddr, safetyMsgMultiSend), "gov trip", "the authority cannot trip"),
		"--from", "node0")
	require.Equal(t, "PROPOSAL_STATUS_FAILED", gjson.Get(proposal, "proposal.status").String(), proposal)
	require.Contains(t, gjson.Get(proposal, "proposal.failed_reason").String(), "is not the operator", proposal)
	require.Equal(t, tripped, safetyDisabledList(t))
	require.True(t, safetyIsAllowed(t, safetyMsgMultiSend))

	// gov narrows the scope to bank; both entries stay
	bankOnly := fmt.Sprintf(`{"authority":%q,"operator":%q,"scope":["/cosmos.bank.*"]}`, govAddr, oldOperator)
	safetySubmitAndPassProposal(t, cli, "tx", "gov", "submit-proposal",
		safetyProposalFile(t, safetyUpdateParams(govAddr, bankOnly), "narrow safety scope", "bank only"),
		"--from", "node0")
	params := safetyParams(t)
	require.Equal(t, []string{"/cosmos.bank.*"}, safetyStrings(gjson.Get(params, "params.scope")), params)
	require.Equal(t, tripped, safetyDisabledList(t), "narrowing the scope must keep existing entries")
	require.False(t, safetyIsAllowed(t, safetyMsgDelegate))

	// the operator can no longer reach the staking entry
	safetyRequireTxRejected(t, cli, "not covered by operator scope", oldOperator, safetyReset(oldOperator, "/cosmos.staking.*"))
	safetyRequireTxRejected(t, cli, "not covered by operator scope", oldOperator,
		safetyTrip(oldOperator, "/cosmos.staking.v1beta1.MsgUndelegate"))
	require.Equal(t, tripped, safetyDisabledList(t))

	// gov replaces the operator and resets the staking entry in one proposal
	newParams := fmt.Sprintf(`{"authority":%q,"operator":%q,"scope":["/cosmos.bank.*"]}`, govAddr, newOperator)
	safetySubmitAndPassProposal(t, cli, "tx", "gov", "submit-proposal",
		safetyProposalFile(t, safetyUpdateParams(govAddr, newParams)+","+safetyReset(govAddr, "/cosmos.staking.*"),
			"replace safety operator", "new operator, reset staking"),
		"--from", "node0")
	params = safetyParams(t)
	require.Equal(t, newOperator, gjson.Get(params, "params.operator").String(), params)
	require.Equal(t, []string{safetyMsgSend}, safetyDisabledList(t))
	require.True(t, safetyIsAllowed(t, safetyMsgDelegate))

	// the old operator can neither trip nor reset
	safetyRequireTxRejected(t, cli, "is not the operator", oldOperator, safetyTrip(oldOperator, safetyMsgMultiSend))
	safetyRequireTxRejected(t, cli, "is neither the authority nor the operator", oldOperator, safetyReset(oldOperator, safetyMsgSend))
	require.Equal(t, []string{safetyMsgSend}, safetyDisabledList(t))

	// the new operator resets the old operator's trip and trips its own
	safetyTx(t, cli, newOperator, safetyReset(newOperator, safetyMsgSend))
	safetyTx(t, cli, newOperator, safetyTrip(newOperator, safetyMsgMultiSend))
	require.Equal(t, []string{safetyMsgMultiSend}, safetyDisabledList(t))
	require.True(t, safetyIsAllowed(t, safetyMsgSend))
	require.False(t, safetyIsAllowed(t, safetyMsgMultiSend))
}

func TestSafetyStateSurvivesRestartAndExport(t *testing.T) {
	// Scenario:
	// given a running chain whose x/safety authority is a test key, with an
	// operator that has tripped an exact and a prefix pattern
	// when the chain restarts, then exports its state, and a new chain starts
	// from the exported x/safety genesis
	// then params and the disabled list are unchanged at each step, the trips
	// still refuse msgs, and the authority resets them on the new chain.

	// the authority address goes into genesis, so derive it from a fixed seed
	// and add the key again after ModifyGenesisJSON resets the keyring
	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	authorityAddr := cli.AddKeyFromSeed("safety-authority", testSeed)
	setAuthority := func(genesis []byte) []byte {
		genesis, err := sjson.SetBytes(genesis, "app_state.safety.params.authority", authorityAddr)
		require.NoError(t, err)
		return genesis
	}
	systest.Sut.ModifyGenesisJSON(t, setAuthority)
	require.Equal(t, authorityAddr, cli.AddKeyFromSeed("safety-authority", testSeed))
	operatorAddr := cli.AddKey("safety-operator")
	recipientAddr := cli.AddKey("safety-recipient")
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", authorityAddr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", operatorAddr, "10000000stake"},
	)
	systest.Sut.StartChain(t)

	newParams := fmt.Sprintf(`{"authority":%q,"operator":%q,"scope":["/cosmos.bank.*","/cosmos.staking.*"]}`,
		authorityAddr, operatorAddr)
	safetyTx(t, cli, authorityAddr, safetyUpdateParams(authorityAddr, newParams))
	safetyTx(t, cli, operatorAddr, safetyTrip(operatorAddr, safetyMsgSend, "/cosmos.staking.*"))
	wantList := []string{safetyMsgSend, "/cosmos.staking.*"}
	wantParams := gjson.Get(safetyParams(t), "params").Raw
	require.Equal(t, wantList, safetyDisabledList(t))

	requireState := func(t *testing.T) {
		t.Helper()
		require.JSONEq(t, wantParams, gjson.Get(safetyParams(t), "params").Raw)
		require.Equal(t, wantList, safetyDisabledList(t))
		require.False(t, safetyIsAllowed(t, safetyMsgSend))
		require.False(t, safetyIsAllowed(t, safetyMsgDelegate))
		require.True(t, safetyIsAllowed(t, safetyMsgMultiSend))
		out, _ := cli.WithRunErrorsIgnored().RunOnly("tx", "bank", "send", "node0", recipientAddr, "100stake", safetyFees)
		upgradeRequireSafetyRejected(t, out, safetyMsgSend+" by "+safetyMsgSend)
	}

	// restart
	systest.Sut.StopChain()
	systest.Sut.StartChain(t)
	requireState(t)

	// export the safety state from a stopped node
	systest.Sut.StopChain()
	exportFile := filepath.Join(t.TempDir(), "export.json")
	systest.MustRunShellCmd(t, systest.Sut.ExecBinary(), "export",
		"--home", systest.Sut.NodeDir(0), "--modules-to-export", "safety", "--output-document", exportFile)
	exported, err := os.ReadFile(exportFile)
	require.NoError(t, err)
	safetyState := gjson.GetBytes(exported, "app_state.safety")
	require.True(t, safetyState.Exists(), string(exported))
	require.JSONEq(t, wantParams, safetyState.Get("params").Raw)
	require.Equal(t, wantList, safetyStrings(safetyState.Get("disabled_patterns")))

	// import it into a fresh chain; the keyring resets with it, so only the
	// seeded authority key comes back
	systest.Sut.ModifyGenesisJSON(t, func(genesis []byte) []byte {
		genesis, err := sjson.SetRawBytes(genesis, "app_state.safety", []byte(safetyState.Raw))
		require.NoError(t, err)
		return genesis
	})
	require.Equal(t, authorityAddr, cli.AddKeyFromSeed("safety-authority", testSeed))
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", authorityAddr, "10000000stake"},
	)
	systest.Sut.StartChain(t)
	cli = systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	requireState(t)

	safetyTx(t, cli, authorityAddr, safetyReset(authorityAddr, wantList...))
	require.Empty(t, safetyDisabledList(t))
	safetyRequireCommittedOK(t, cli.Run("tx", "bank", "send", "node0", recipientAddr, "100stake"))
	require.Equal(t, int64(100), cli.QueryBalance(recipientAddr, "stake"))
}

// safetySubmitAndAwaitProposal submits a gov proposal with the given tx
// command, votes yes from every validator, waits until voting ends whatever
// the outcome, and returns the proposal id and the final proposal query.
func safetySubmitAndAwaitProposal(t *testing.T, cli *systest.CLIWrapper, args ...string) (string, string) {
	t.Helper()
	systest.RequireTxSuccess(t, cli.Run(args...))

	raw := cli.CustomQuery("q", "gov", "proposals")
	ids := gjson.Get(raw, "proposals.#.id").Array()
	require.NotEmpty(t, ids, raw)
	proposalID := ids[len(ids)-1].String()

	for i := 0; i < systest.Sut.NodesCount(); i++ {
		out, _ := cli.RunOnly("tx", "gov", "vote", proposalID, "yes", "--from", fmt.Sprintf("node%d", i), safetyFees)
		systest.RequireTxSuccess(t, out)
	}

	require.Eventually(t, func() bool {
		raw = cli.CustomQuery("q", "gov", "proposal", proposalID)
		status := gjson.Get(raw, "proposal.status").String()
		return !strings.HasSuffix(status, "_VOTING_PERIOD") && !strings.HasSuffix(status, "_DEPOSIT_PERIOD")
	}, safetyVotingPeriod+10*systest.Sut.BlockTime(), systest.Sut.BlockTime(), "proposal %s did not finish voting", proposalID)
	return proposalID, raw
}
