//go:build system_test

package systemtests

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	systest "github.com/cosmos/cosmos-sdk/tools/systemtests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	safetyMsgSend      = "/cosmos.bank.v1beta1.MsgSend"
	safetyMsgMultiSend = "/cosmos.bank.v1beta1.MsgMultiSend"
	safetyDisabledErr  = "msg type disabled by x/safety"
	safetyVotingPeriod = 12 * time.Second
	safetyFees         = "--fees=1stake"
	safetyMsgExec      = "/cosmos.authz.v1beta1.MsgExec"
)

func TestSafetyOperatorTripAndReset(t *testing.T) {
	// Scenario:
	// given a running chain where x/safety has no operator
	// when gov passes a proposal naming a test key as operator with scope /cosmos.bank.*
	// then the operator can trip MsgSend, which blocks bank sends (also nested in authz exec),
	// and resetting it lets the same sends through again.
	// ModifyGenesisJSON resets the chain, keyring included, so it runs before keys are added
	systest.Sut.ModifyGenesisJSON(t, safetyShortGovPeriods(t))
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	govAddr := sdk.AccAddress(address.Module("gov")).String()

	operatorAddr := cli.AddKey("safety-operator")
	senderAddr := cli.AddKey("safety-sender")
	recipientAddr := cli.AddKey("safety-recipient")
	strangerAddr := cli.AddKey("safety-stranger")
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", operatorAddr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", senderAddr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", strangerAddr, "10000000stake"},
	)
	systest.Sut.StartChain(t)

	// default genesis: gov is the authority, no operator, nothing disabled.
	// The default scope is checked in TestSafetyDefaultGenesisQueries.
	params := cli.CustomQuery("q", "safety", "params")
	require.Equal(t, govAddr, gjson.Get(params, "params.authority").String(), params)
	require.Empty(t, gjson.Get(params, "params.operator").String(), params)
	require.Empty(t, safetyDisabledList(t, cli), "disabled list at genesis")
	require.True(t, safetyIsAllowed(t, cli, safetyMsgSend))

	// no operator yet: the key cannot trip
	safetyRequireTxRejected(t, cli, "unauthorized",
		"tx", "safety", "trip", safetyMsgSend, "--from", operatorAddr)

	// gov names the operator through the autocli proposal command
	newParams := fmt.Sprintf(`{"authority":%q,"operator":%q,"scope":["/cosmos.bank.*"]}`, govAddr, operatorAddr)
	safetySubmitAndPassProposal(t, cli,
		"tx", "safety", "update-params-proposal", newParams,
		"--title", "name safety operator", "--summary", "grant /cosmos.bank.* scope",
		"--deposit", "100000000stake", "--from", "node0")

	params = cli.CustomQuery("q", "safety", "params")
	require.Equal(t, govAddr, gjson.Get(params, "params.authority").String(), params)
	require.Equal(t, operatorAddr, gjson.Get(params, "params.operator").String(), params)
	require.Equal(t, []string{"/cosmos.bank.*"}, safetyStrings(gjson.Get(params, "params.scope")), params)

	// rejections that circuit accepted: empty list, out of scope, unknown,
	// and a sender that is not the operator
	safetyRequireTxRejected(t, cli, "pattern list is empty",
		"tx", "safety", "trip", "--from", operatorAddr)
	safetyRequireTxRejected(t, cli, "not covered by operator scope",
		"tx", "safety", "trip", "/cosmos.staking.v1beta1.MsgDelegate", "--from", operatorAddr)
	safetyRequireTxRejected(t, cli, "matches no registered unprotected msg type",
		"tx", "safety", "trip", "/cosmos.bank.v1beta1.MsgSnd", "--from", operatorAddr)
	safetyRequireTxRejected(t, cli, "unauthorized",
		"tx", "safety", "trip", safetyMsgSend, "--from", strangerAddr)
	require.Empty(t, safetyDisabledList(t, cli), "failed trips must not change state")

	// set up an authz grant while sends still work, for the nested check below
	rsp := cli.Run("tx", "authz", "grant", operatorAddr, "send", "--spend-limit", "1000stake", "--from", senderAddr)
	safetyRequireCommittedOK(t, rsp)

	// operator trips two patterns in one tx, each its own positional arg
	rsp = cli.Run("tx", "safety", "trip", safetyMsgSend, safetyMsgMultiSend, "--from", operatorAddr)
	txResult := safetyRequireCommittedOK(t, rsp)
	require.True(t, gjson.Get(txResult, `events.#(type=="cosmos.safety.v1.EventTrip")`).Exists(), txResult)

	require.Equal(t, []string{safetyMsgMultiSend, safetyMsgSend}, safetyDisabledList(t, cli))
	require.False(t, safetyIsAllowed(t, cli, safetyMsgSend))
	require.False(t, safetyIsAllowed(t, cli, safetyMsgMultiSend))
	require.True(t, safetyIsAllowed(t, cli, "/cosmos.staking.v1beta1.MsgDelegate"))
	require.True(t, safetyIsAllowed(t, cli, "/cosmos.safety.v1.MsgReset"))

	// the REST gateway agrees
	list := safetyREST(t, "/cosmos/safety/v1/disabled_list")
	require.Equal(t, []string{safetyMsgMultiSend, safetyMsgSend}, safetyStrings(gjson.Get(list, "disabled_patterns")), list)
	allowed := safetyREST(t, "/cosmos/safety/v1/is_allowed?msg_type_url="+url.QueryEscape(safetyMsgSend))
	require.True(t, gjson.Get(allowed, "allowed").Exists(), allowed)
	require.False(t, gjson.Get(allowed, "allowed").Bool(), allowed)

	// disabled-list paginates
	page := cli.CustomQuery("q", "safety", "disabled-list", "--page-limit", "1", "--page-count-total")
	require.Equal(t, []string{safetyMsgMultiSend}, safetyStrings(gjson.Get(page, "disabled_patterns")), page)
	require.NotEmpty(t, gjson.Get(page, "pagination.next_key").String(), page)
	require.Equal(t, int64(2), gjson.Get(page, "pagination.total").Int(), page)
	page = cli.CustomQuery("q", "safety", "disabled-list", "--page-limit", "1", "--page-offset", "1")
	require.Equal(t, []string{safetyMsgSend}, safetyStrings(gjson.Get(page, "disabled_patterns")), page)

	// tripping an already disabled pattern fails
	safetyRequireTxRejected(t, cli, "already disabled",
		"tx", "safety", "trip", safetyMsgSend, "--from", operatorAddr)

	// a bank send fails in CheckTx and pays no fee
	senderBalance := cli.QueryBalance(senderAddr, "stake")
	out, _ := cli.WithRunErrorsIgnored().RunOnly("tx", "bank", "send", senderAddr, recipientAddr, "100stake", safetyFees)
	safetyRequireCheckTxRejected(t, out, safetyDisabledErr)
	systest.Sut.AwaitNextBlock(t)
	require.Equal(t, senderBalance, cli.QueryBalance(senderAddr, "stake"), "rejected tx must not charge fees")
	require.Zero(t, cli.QueryBalance(recipientAddr, "stake"))

	// the same send wrapped in authz exec is rejected too
	execFile := safetyGenerateSendTx(t, cli, senderAddr, recipientAddr, "50stake")
	out, _ = cli.WithRunErrorsIgnored().RunOnly("tx", "authz", "exec", execFile, "--from", operatorAddr, safetyFees)
	safetyRequireCheckTxRejected(t, out, safetyDisabledErr)

	// the operator cannot reset a pattern that is not disabled, and a stranger cannot reset at all
	safetyRequireTxRejected(t, cli, "not in disabled list",
		"tx", "safety", "reset", "/cosmos.bank.*", "--from", operatorAddr)
	safetyRequireTxRejected(t, cli, "unauthorized",
		"tx", "safety", "reset", safetyMsgSend, "--from", strangerAddr)

	// operator resets both patterns in one tx
	rsp = cli.Run("tx", "safety", "reset", safetyMsgSend, safetyMsgMultiSend, "--from", operatorAddr)
	txResult = safetyRequireCommittedOK(t, rsp)
	require.True(t, gjson.Get(txResult, `events.#(type=="cosmos.safety.v1.EventReset")`).Exists(), txResult)
	require.Empty(t, safetyDisabledList(t, cli))
	require.True(t, safetyIsAllowed(t, cli, safetyMsgSend))

	// the send now succeeds, directly and through authz
	rsp = cli.Run("tx", "bank", "send", senderAddr, recipientAddr, "100stake")
	safetyRequireCommittedOK(t, rsp)
	require.Equal(t, int64(100), cli.QueryBalance(recipientAddr, "stake"))

	rsp = cli.Run("tx", "authz", "exec", execFile, "--from", operatorAddr)
	safetyRequireCommittedOK(t, rsp)
	require.Equal(t, int64(150), cli.QueryBalance(recipientAddr, "stake"))
}

func TestSafetyTripAllGovResets(t *testing.T) {
	// Scenario:
	// given a running chain whose genesis names an operator with scope *
	// when the operator trips *
	// then ordinary msgs are blocked, and gov can still submit, vote on and
	// execute a proposal that resets *.

	// the operator address goes into genesis, so derive it from a fixed seed and
	// add the key again after ModifyGenesisJSON resets the keyring
	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	govAddr := sdk.AccAddress(address.Module("gov")).String()
	operatorAddr := cli.AddKeyFromSeed("safety-operator", testSeed)

	systest.Sut.ModifyGenesisJSON(t,
		safetyShortGovPeriods(t),
		safetySetGenesis(t, operatorAddr, []string{"*"}, nil),
	)
	require.Equal(t, operatorAddr, cli.AddKeyFromSeed("safety-operator", testSeed))
	recipientAddr := cli.AddKey("safety-recipient")
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", operatorAddr, "10000000stake"},
	)
	systest.Sut.StartChain(t)

	params := cli.CustomQuery("q", "safety", "params")
	require.Equal(t, operatorAddr, gjson.Get(params, "params.operator").String(), params)
	require.Equal(t, []string{"*"}, safetyStrings(gjson.Get(params, "params.scope")), params)

	// protected msgs cannot be tripped, even with scope *
	for _, pattern := range []string{"/cosmos.gov.v1.MsgVote", "/cosmos.safety.*", "/cosmos.safety.v1.MsgReset"} {
		safetyRequireTxRejected(t, cli, "matches no registered unprotected msg type",
			"tx", "safety", "trip", pattern, "--from", operatorAddr)
	}

	rsp := cli.Run("tx", "safety", "trip", "*", "--from", operatorAddr)
	safetyRequireCommittedOK(t, rsp)

	require.Equal(t, []string{"*"}, safetyDisabledList(t, cli))
	require.False(t, safetyIsAllowed(t, cli, safetyMsgSend))
	require.False(t, safetyIsAllowed(t, cli, "/cosmos.staking.v1beta1.MsgDelegate"))
	require.False(t, safetyIsAllowed(t, cli, "/cosmos.authz.v1beta1.MsgGrant"))
	// only an exact trip blocks a wrapper such as authz MsgExec
	require.True(t, safetyIsAllowed(t, cli, safetyMsgExec))
	for _, protected := range []string{
		"/cosmos.gov.v1.MsgSubmitProposal",
		"/cosmos.gov.v1.MsgVote",
		"/cosmos.safety.v1.MsgReset",
		"/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
		"/cosmos.consensus.v1.MsgUpdateParams",
	} {
		require.True(t, safetyIsAllowed(t, cli, protected), protected)
	}

	out, _ := cli.WithRunErrorsIgnored().RunOnly("tx", "bank", "send", "node0", recipientAddr, "100stake", safetyFees)
	safetyRequireCheckTxRejected(t, out, safetyDisabledErr)

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
	require.True(t, safetyIsAllowed(t, cli, safetyMsgSend))

	rsp = cli.Run("tx", "bank", "send", "node0", recipientAddr, "100stake")
	safetyRequireCommittedOK(t, rsp)
	require.Equal(t, int64(100), cli.QueryBalance(recipientAddr, "stake"))
}

func TestSafetyAuthzAuthorityResetsUnderWildcard(t *testing.T) {
	// Scenario:
	// given a running chain whose x/safety authority acts through an authz grantee
	// when the operator trips *
	// then authz exec of a bank send is rejected, the grantee resets * for the
	// authority through authz exec, and an exact trip of MsgExec closes that path.

	// the authority address goes into genesis, so derive it from a fixed seed
	// and add the key again after ModifyGenesisJSON resets the keyring
	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	authorityAddr := cli.AddKeyFromSeed("safety-authority", testSeed)
	systest.Sut.ModifyGenesisJSON(t, func(genesis []byte) []byte {
		genesis, err := sjson.SetBytes(genesis, "app_state.safety.params.authority", authorityAddr)
		require.NoError(t, err)
		return genesis
	})
	require.Equal(t, authorityAddr, cli.AddKeyFromSeed("safety-authority", testSeed))
	operatorAddr := cli.AddKey("safety-operator")
	granteeAddr := cli.AddKey("safety-grantee")
	recipientAddr := cli.AddKey("safety-recipient")
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", authorityAddr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", operatorAddr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", granteeAddr, "10000000stake"},
	)
	systest.Sut.StartChain(t)

	// the authority signs directly: it names the operator, then lets the
	// grantee reset on its behalf
	newParams := fmt.Sprintf(`{"authority":%q,"operator":%q,"scope":["*"]}`, authorityAddr, operatorAddr)
	safetyRequireCommittedOK(t, cli.Run("tx", "safety", "update-params-proposal", newParams, "--no-proposal", "--from", authorityAddr))
	params := cli.CustomQuery("q", "safety", "params")
	require.Equal(t, authorityAddr, gjson.Get(params, "params.authority").String(), params)
	require.Equal(t, operatorAddr, gjson.Get(params, "params.operator").String(), params)
	safetyRequireCommittedOK(t, cli.Run("tx", "authz", "grant", granteeAddr, "generic",
		"--msg-type", "/cosmos.safety.v1.MsgReset", "--from", authorityAddr))

	safetyRequireCommittedOK(t, cli.Run("tx", "safety", "trip", "*", "--from", operatorAddr))
	require.True(t, safetyIsAllowed(t, cli, safetyMsgExec))
	require.False(t, safetyIsAllowed(t, cli, safetyMsgSend))

	// MsgExec runs, but the send inside it is rejected in CheckTx
	sendFile := safetyGenerateSendTx(t, cli, authorityAddr, recipientAddr, "100stake")
	out, _ := cli.WithRunErrorsIgnored().RunOnly("tx", "authz", "exec", sendFile, "--from", granteeAddr, safetyFees)
	safetyRequireCheckTxRejected(t, out, safetyDisabledErr)

	// the grantee resets * for the authority
	resetFile := safetyGenerateTx(t, cli, "reset.json", "/cosmos.safety.v1.MsgReset",
		"tx", "safety", "reset", "*", "--from", authorityAddr)
	safetyRequireCommittedOK(t, cli.Run("tx", "authz", "exec", resetFile, "--from", granteeAddr))
	require.Empty(t, safetyDisabledList(t, cli))

	// an exact trip of MsgExec, accepted beside *, closes the path until the
	// operator resets it
	safetyRequireCommittedOK(t, cli.Run("tx", "safety", "trip", "*", safetyMsgExec, "--from", operatorAddr))
	require.False(t, safetyIsAllowed(t, cli, safetyMsgExec))
	out, _ = cli.WithRunErrorsIgnored().RunOnly("tx", "authz", "exec", resetFile, "--from", granteeAddr, safetyFees)
	safetyRequireCheckTxRejected(t, out, safetyDisabledErr)
	require.Equal(t, []string{"*", safetyMsgExec}, safetyDisabledList(t, cli))

	safetyRequireCommittedOK(t, cli.Run("tx", "safety", "reset", safetyMsgExec, "--from", operatorAddr))
	safetyRequireCommittedOK(t, cli.Run("tx", "authz", "exec", resetFile, "--from", granteeAddr))
	require.Empty(t, safetyDisabledList(t, cli))

	safetyRequireCommittedOK(t, cli.Run("tx", "bank", "send", authorityAddr, recipientAddr, "100stake"))
	require.Equal(t, int64(100), cli.QueryBalance(recipientAddr, "stake"))
}

// The default genesis scope is empty, so the operator can do nothing
// until the authority grants scope.
func TestSafetyDefaultGenesisQueries(t *testing.T) {
	// Scenario:
	// given a running chain with the default x/safety genesis
	// when the safety queries run
	// then params show gov as authority, no operator and an empty scope,
	// the disabled list is empty and every msg is allowed.
	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	govAddr := sdk.AccAddress(address.Module("gov")).String()
	systest.Sut.StartChain(t)

	require.Empty(t, safetyDisabledList(t, cli))
	require.True(t, safetyIsAllowed(t, cli, safetyMsgSend))
	require.True(t, safetyIsAllowed(t, cli, "/cosmos.gov.v1.MsgVote"))

	params := cli.CustomQuery("q", "safety", "params")
	require.Equal(t, govAddr, gjson.Get(params, "params.authority").String(), params)
	require.Empty(t, gjson.Get(params, "params.operator").String(), params)
	require.Empty(t, safetyStrings(gjson.Get(params, "params.scope")), "default scope must be empty: %s", params)

	// the same queries through the REST gateway routes under /cosmos/safety/v1/
	params = safetyREST(t, "/cosmos/safety/v1/params")
	require.Equal(t, govAddr, gjson.Get(params, "params.authority").String(), params)
	require.Empty(t, gjson.Get(params, "params.operator").String(), params)
	list := safetyREST(t, "/cosmos/safety/v1/disabled_list?pagination.count_total=true")
	require.Empty(t, safetyStrings(gjson.Get(list, "disabled_patterns")), list)
	require.Equal(t, int64(0), gjson.Get(list, "pagination.total").Int(), list)
	allowed := safetyREST(t, "/cosmos/safety/v1/is_allowed?msg_type_url="+url.QueryEscape(safetyMsgSend))
	require.True(t, gjson.Get(allowed, "allowed").Bool(), allowed)

	// a type URL without its leading slash is rejected, not answered
	systest.GetRequestWithHeaders(t,
		systest.Sut.APIAddress()+"/cosmos/safety/v1/is_allowed?msg_type_url="+url.QueryEscape(strings.TrimPrefix(safetyMsgSend, "/")),
		nil, http.StatusBadRequest)
}

// safetyREST runs a GET against the node's REST gateway and returns the body.
func safetyREST(t *testing.T, path string) string {
	t.Helper()
	return string(systest.GetRequest(t, systest.Sut.APIAddress()+path))
}

// SafetyDecorator opens any msg exposing GetMsgs() ([]sdk.Msg, error), such as
// gov v1 MsgSubmitProposal, and rejects a proposal with a tripped msg at
// submission.
func TestSafetyGovProposalWithDisabledMsgRejectedAtSubmission(t *testing.T) {
	// Scenario:
	// given a running chain whose genesis disables MsgSend
	// when a gov proposal with a MsgSend is submitted
	// then CheckTx rejects the submission.
	systest.Sut.ModifyGenesisJSON(t, safetySetGenesis(t, "", nil, []string{safetyMsgSend}))
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	govAddr := sdk.AccAddress(address.Module("gov")).String()
	recipientAddr := cli.AddKey("safety-recipient")
	systest.Sut.StartChain(t)

	// genesis state was applied
	require.Equal(t, []string{safetyMsgSend}, safetyDisabledList(t, cli))
	require.False(t, safetyIsAllowed(t, cli, safetyMsgSend))

	proposal := fmt.Sprintf(`{
 "messages": [
  {
   "@type": "/cosmos.bank.v1beta1.MsgSend",
   "from_address": %q,
   "to_address": %q,
   "amount": [{"denom": "stake", "amount": "1"}]
  }
 ],
 "metadata": "ipfs://CID",
 "deposit": "100000000stake",
 "title": "send from gov",
 "summary": "contains a disabled msg"
}`, govAddr, recipientAddr)
	proposalFile := filepath.Join(t.TempDir(), "proposal.json")
	require.NoError(t, os.WriteFile(proposalFile, []byte(proposal), 0o600))

	out, _ := cli.WithRunErrorsIgnored().RunOnly("tx", "gov", "submit-proposal", proposalFile, "--from", "node0", safetyFees)
	safetyRequireCheckTxRejected(t, out, safetyDisabledErr)
}

// safetyShortGovPeriods shortens gov voting so a proposal passes within the test.
func safetyShortGovPeriods(t *testing.T) systest.GenesisMutator {
	t.Helper()
	voting := systest.SetGovVotingPeriod(t, safetyVotingPeriod)
	expedited := systest.SetGovExpeditedVotingPeriod(t, safetyVotingPeriod-2*time.Second)
	return func(genesis []byte) []byte { return expedited(voting(genesis)) }
}

// safetySetGenesis sets the x/safety operator, scope and disabled patterns in
// genesis. The authority stays the default.
func safetySetGenesis(t *testing.T, operator string, scope, disabled []string) systest.GenesisMutator {
	t.Helper()
	return func(genesis []byte) []byte {
		if scope == nil {
			scope = []string{}
		}
		if disabled == nil {
			disabled = []string{}
		}
		var err error
		genesis, err = sjson.SetBytes(genesis, "app_state.safety.params.operator", operator)
		require.NoError(t, err)
		genesis, err = sjson.SetBytes(genesis, "app_state.safety.params.scope", scope)
		require.NoError(t, err)
		genesis, err = sjson.SetBytes(genesis, "app_state.safety.disabled_patterns", disabled)
		require.NoError(t, err)
		return genesis
	}
}

func safetyDisabledList(t *testing.T, cli *systest.CLIWrapper) []string {
	t.Helper()
	raw := safetyTrimJSON(cli.CustomQuery("q", "safety", "disabled-list"))
	require.True(t, gjson.Valid(raw), raw)
	return safetyStrings(gjson.Get(raw, "disabled_patterns"))
}

func safetyIsAllowed(t *testing.T, cli *systest.CLIWrapper, typeURL string) bool {
	t.Helper()
	raw := safetyTrimJSON(cli.CustomQuery("q", "safety", "is-allowed", typeURL))
	require.True(t, gjson.Valid(raw), raw)
	return gjson.Get(raw, "allowed").Bool()
}

// safetyTrimJSON strips any non-JSON lines the binary prints before its JSON
// output, such as runtime warnings.
func safetyTrimJSON(out string) string {
	for strings.HasPrefix(out, "WARNING:") {
		_, rest, found := strings.Cut(out, "\n")
		if !found {
			return ""
		}
		out = strings.TrimSpace(rest)
	}
	return out
}

func safetyStrings(r gjson.Result) []string {
	out := []string{}
	for _, v := range r.Array() {
		out = append(out, v.String())
	}
	return out
}

// safetyRequireCommittedOK requires the tx returned by CLIWrapper.Run to be
// committed with code 0 and returns the committed tx result.
func safetyRequireCommittedOK(t *testing.T, rsp string) string {
	t.Helper()
	systest.RequireTxSuccess(t, rsp)
	return rsp
}

// safetyRequireCheckTxRejected requires a broadcast to have failed before
// entering the mempool, with errMsg in the output.
func safetyRequireCheckTxRejected(t *testing.T, out, errMsg string) {
	t.Helper()
	if js := safetyTrimJSON(out); gjson.Valid(js) && gjson.Get(js, "code").Exists() {
		require.NotZero(t, gjson.Get(js, "code").Int(), out)
	}
	require.Contains(t, out, errMsg)
}

// safetyRequireTxRejected runs a tx and requires it to fail at any stage:
// client-side validation, CheckTx, or execution. When errMsg is set, the
// failure output must contain it.
func safetyRequireTxRejected(t *testing.T, cli *systest.CLIWrapper, errMsg string, args ...string) {
	t.Helper()
	out, _ := cli.WithRunErrorsIgnored().RunOnly(append(args, safetyFees)...)
	out = safetyTrimJSON(out)
	if gjson.Valid(out) && gjson.Get(out, "code").Int() == 0 && gjson.Get(out, "txhash").Exists() {
		// accepted into the mempool: execution must fail
		res, found := cli.AwaitTxCommitted(out)
		require.True(t, found, "tx not committed: %s", out)
		require.NotZero(t, gjson.Get(res, "code").Int(), "tx succeeded but must fail: %s", res)
		out = res
	}
	if gjson.Valid(out) && gjson.Get(out, "code").Exists() {
		require.NotZero(t, gjson.Get(out, "code").Int(), out)
	} else {
		require.NotEmpty(t, out)
	}
	if errMsg != "" {
		require.Contains(t, out, errMsg)
	}
}

// safetySubmitAndPassProposal submits a gov proposal with the given tx
// command, votes yes from every validator and waits until it has passed.
func safetySubmitAndPassProposal(t *testing.T, cli *systest.CLIWrapper, args ...string) string {
	t.Helper()
	rsp := cli.Run(args...)
	systest.RequireTxSuccess(t, rsp)

	raw := cli.CustomQuery("q", "gov", "proposals")
	ids := gjson.Get(raw, "proposals.#.id").Array()
	require.NotEmpty(t, ids, raw)
	proposalID := ids[len(ids)-1].String()

	for i := 0; i < systest.Sut.NodesCount(); i++ {
		out, _ := cli.RunOnly("tx", "gov", "vote", proposalID, "yes", "--from", fmt.Sprintf("node%d", i), safetyFees)
		systest.RequireTxSuccess(t, out)
	}

	var status string
	require.Eventually(t, func() bool {
		raw := cli.CustomQuery("q", "gov", "proposal", proposalID)
		status = gjson.Get(raw, "proposal.status").String()
		return status == "PROPOSAL_STATUS_PASSED" || status == "PROPOSAL_STATUS_FAILED" || status == "PROPOSAL_STATUS_REJECTED"
	}, safetyVotingPeriod+10*systest.Sut.BlockTime(), systest.Sut.BlockTime(), "proposal %s did not finish voting", proposalID)
	require.Equal(t, "PROPOSAL_STATUS_PASSED", status, cli.CustomQuery("q", "gov", "proposal", proposalID))
	return proposalID
}

// safetyGenerateSendTx writes an unsigned bank send to a file for authz exec.
func safetyGenerateSendTx(t *testing.T, cli *systest.CLIWrapper, from, to, amount string) string {
	t.Helper()
	return safetyGenerateTx(t, cli, "send.json", safetyMsgSend, "tx", "bank", "send", from, to, amount)
}

// safetyGenerateTx runs a tx command with --generate-only, checks the
// unsigned tx contains msgType, and writes it to a file for authz exec.
func safetyGenerateTx(t *testing.T, cli *systest.CLIWrapper, name, msgType string, args ...string) string {
	t.Helper()
	out := safetyTrimJSON(cli.RunCommandWithArgs(append(args, "--generate-only", "--chain-id="+cli.ChainID())...))
	require.True(t, gjson.Valid(out), out)
	require.True(t, strings.Contains(out, msgType), out)
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(out), 0o600))
	return path
}
