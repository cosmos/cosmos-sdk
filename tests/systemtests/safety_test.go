//go:build system_test

package systemtests

import (
	"encoding/json"
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
	params := safetyParams(t)
	require.Equal(t, govAddr, gjson.Get(params, "params.authority").String(), params)
	require.Empty(t, gjson.Get(params, "params.operator").String(), params)
	require.Empty(t, safetyDisabledList(t), "disabled list at genesis")
	require.True(t, safetyIsAllowed(t, safetyMsgSend))

	// no operator yet: the key cannot trip
	safetyRequireTxRejected(t, cli, "unauthorized", operatorAddr, safetyTrip(operatorAddr, safetyMsgSend))

	// gov names the operator through a proposal
	newParams := fmt.Sprintf(`{"authority":%q,"operator":%q,"scope":["/cosmos.bank.*"]}`, govAddr, operatorAddr)
	safetySubmitAndPassProposal(t, cli, "tx", "gov", "submit-proposal",
		safetyProposalFile(t, safetyUpdateParams(govAddr, newParams), "name safety operator", "grant /cosmos.bank.* scope"),
		"--from", "node0")

	params = safetyParams(t)
	require.Equal(t, govAddr, gjson.Get(params, "params.authority").String(), params)
	require.Equal(t, operatorAddr, gjson.Get(params, "params.operator").String(), params)
	require.Equal(t, []string{"/cosmos.bank.*"}, safetyStrings(gjson.Get(params, "params.scope")), params)

	// rejections that circuit accepted: empty list, out of scope, unknown,
	// and a sender that is not the operator
	safetyRequireTxRejected(t, cli, "pattern list is empty", operatorAddr, safetyTrip(operatorAddr))
	safetyRequireTxRejected(t, cli, "not covered by operator scope",
		operatorAddr, safetyTrip(operatorAddr, "/cosmos.staking.v1beta1.MsgDelegate"))
	safetyRequireTxRejected(t, cli, "matches no registered unprotected msg type",
		operatorAddr, safetyTrip(operatorAddr, "/cosmos.bank.v1beta1.MsgSnd"))
	safetyRequireTxRejected(t, cli, "unauthorized", strangerAddr, safetyTrip(strangerAddr, safetyMsgSend))
	require.Empty(t, safetyDisabledList(t), "failed trips must not change state")

	// set up an authz grant while sends still work, for the nested check below
	rsp := cli.Run("tx", "authz", "grant", operatorAddr, "send", "--spend-limit", "1000stake", "--from", senderAddr)
	safetyRequireCommittedOK(t, rsp)

	// operator trips two patterns in one tx
	txResult := safetyTx(t, cli, operatorAddr, safetyTrip(operatorAddr, safetyMsgSend, safetyMsgMultiSend))
	require.True(t, gjson.Get(txResult, `events.#(type=="cosmos.safety.v1.EventTrip")`).Exists(), txResult)

	require.Equal(t, []string{safetyMsgMultiSend, safetyMsgSend}, safetyDisabledList(t))
	require.False(t, safetyIsAllowed(t, safetyMsgSend))
	require.False(t, safetyIsAllowed(t, safetyMsgMultiSend))
	require.True(t, safetyIsAllowed(t, "/cosmos.staking.v1beta1.MsgDelegate"))
	require.True(t, safetyIsAllowed(t, "/cosmos.safety.v1.MsgReset"))

	// the disabled list paginates
	page := safetyREST(t, "/cosmos/safety/v1/disabled_list?pagination.limit=1&pagination.count_total=true")
	require.Equal(t, []string{safetyMsgMultiSend}, safetyStrings(gjson.Get(page, "disabled_patterns")), page)
	require.NotEmpty(t, gjson.Get(page, "pagination.next_key").String(), page)
	require.Equal(t, int64(2), gjson.Get(page, "pagination.total").Int(), page)
	page = safetyREST(t, "/cosmos/safety/v1/disabled_list?pagination.limit=1&pagination.offset=1")
	require.Equal(t, []string{safetyMsgSend}, safetyStrings(gjson.Get(page, "disabled_patterns")), page)

	// tripping an already disabled pattern fails
	safetyRequireTxRejected(t, cli, "already disabled", operatorAddr, safetyTrip(operatorAddr, safetyMsgSend))

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
	safetyRequireTxRejected(t, cli, "not in disabled list", operatorAddr, safetyReset(operatorAddr, "/cosmos.bank.*"))
	safetyRequireTxRejected(t, cli, "unauthorized", strangerAddr, safetyReset(strangerAddr, safetyMsgSend))

	// operator resets both patterns in one tx
	txResult = safetyTx(t, cli, operatorAddr, safetyReset(operatorAddr, safetyMsgSend, safetyMsgMultiSend))
	require.True(t, gjson.Get(txResult, `events.#(type=="cosmos.safety.v1.EventReset")`).Exists(), txResult)
	require.Empty(t, safetyDisabledList(t))
	require.True(t, safetyIsAllowed(t, safetyMsgSend))

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

	params := safetyParams(t)
	require.Equal(t, operatorAddr, gjson.Get(params, "params.operator").String(), params)
	require.Equal(t, []string{"*"}, safetyStrings(gjson.Get(params, "params.scope")), params)

	// protected msgs cannot be tripped, even with scope *
	for _, pattern := range []string{"/cosmos.gov.v1.MsgVote", "/cosmos.safety.*", "/cosmos.safety.v1.MsgReset"} {
		safetyRequireTxRejected(t, cli, "matches no registered unprotected msg type", operatorAddr, safetyTrip(operatorAddr, pattern))
	}

	safetyTx(t, cli, operatorAddr, safetyTrip(operatorAddr, "*"))

	require.Equal(t, []string{"*"}, safetyDisabledList(t))
	require.False(t, safetyIsAllowed(t, safetyMsgSend))
	require.False(t, safetyIsAllowed(t, "/cosmos.staking.v1beta1.MsgDelegate"))
	require.False(t, safetyIsAllowed(t, "/cosmos.authz.v1beta1.MsgGrant"))
	// only an exact trip blocks a wrapper such as authz MsgExec
	require.True(t, safetyIsAllowed(t, safetyMsgExec))
	for _, protected := range []string{
		"/cosmos.gov.v1.MsgSubmitProposal",
		"/cosmos.gov.v1.MsgVote",
		"/cosmos.safety.v1.MsgReset",
		"/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
		"/cosmos.consensus.v1.MsgUpdateParams",
	} {
		require.True(t, safetyIsAllowed(t, protected), protected)
	}

	out, _ := cli.WithRunErrorsIgnored().RunOnly("tx", "bank", "send", "node0", recipientAddr, "100stake", safetyFees)
	safetyRequireCheckTxRejected(t, out, safetyDisabledErr)

	// gov submits, votes on and executes a reset of * while * is tripped
	safetySubmitAndPassProposal(t, cli, "tx", "gov", "submit-proposal",
		safetyProposalFile(t, safetyReset(govAddr, "*"), "reset safety", "clear the * trip"), "--from", "node0")

	require.Empty(t, safetyDisabledList(t))
	require.True(t, safetyIsAllowed(t, safetyMsgSend))

	rsp := cli.Run("tx", "bank", "send", "node0", recipientAddr, "100stake")
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
	safetyTx(t, cli, authorityAddr, safetyUpdateParams(authorityAddr, newParams))
	params := safetyParams(t)
	require.Equal(t, authorityAddr, gjson.Get(params, "params.authority").String(), params)
	require.Equal(t, operatorAddr, gjson.Get(params, "params.operator").String(), params)
	safetyRequireCommittedOK(t, cli.Run("tx", "authz", "grant", granteeAddr, "generic",
		"--msg-type", "/cosmos.safety.v1.MsgReset", "--from", authorityAddr))

	safetyTx(t, cli, operatorAddr, safetyTrip(operatorAddr, "*"))
	require.True(t, safetyIsAllowed(t, safetyMsgExec))
	require.False(t, safetyIsAllowed(t, safetyMsgSend))

	// MsgExec runs, but the send inside it is rejected in CheckTx
	sendFile := safetyGenerateSendTx(t, cli, authorityAddr, recipientAddr, "100stake")
	out, _ := cli.WithRunErrorsIgnored().RunOnly("tx", "authz", "exec", sendFile, "--from", granteeAddr, safetyFees)
	safetyRequireCheckTxRejected(t, out, safetyDisabledErr)

	// the grantee resets * for the authority
	resetFile := safetyUnsignedTx(t, safetyReset(authorityAddr, "*"))
	safetyRequireCommittedOK(t, cli.Run("tx", "authz", "exec", resetFile, "--from", granteeAddr))
	require.Empty(t, safetyDisabledList(t))

	// an exact trip of MsgExec, accepted beside *, closes the path until the
	// operator resets it
	safetyTx(t, cli, operatorAddr, safetyTrip(operatorAddr, "*", safetyMsgExec))
	require.False(t, safetyIsAllowed(t, safetyMsgExec))
	out, _ = cli.WithRunErrorsIgnored().RunOnly("tx", "authz", "exec", resetFile, "--from", granteeAddr, safetyFees)
	safetyRequireCheckTxRejected(t, out, safetyDisabledErr)
	require.Equal(t, []string{"*", safetyMsgExec}, safetyDisabledList(t))

	safetyTx(t, cli, operatorAddr, safetyReset(operatorAddr, safetyMsgExec))
	safetyRequireCommittedOK(t, cli.Run("tx", "authz", "exec", resetFile, "--from", granteeAddr))
	require.Empty(t, safetyDisabledList(t))

	safetyRequireCommittedOK(t, cli.Run("tx", "bank", "send", authorityAddr, recipientAddr, "100stake"))
	require.Equal(t, int64(100), cli.QueryBalance(recipientAddr, "stake"))
}

func TestSafetyGenericCLI(t *testing.T) {
	// Scenario:
	// given a running chain whose x/safety authority is a test key
	// when the authority and operator use the generic `tx safety` and
	// `q safety` commands derived from the module's gRPC services
	// then update-params names the operator, trip takes several patterns as a
	// comma list or a repeated flag, a tripped bank send is refused, and reset
	// lets it through again.

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
	strangerAddr := cli.AddKey("safety-stranger")
	recipientAddr := cli.AddKey("safety-recipient")
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", authorityAddr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", operatorAddr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", strangerAddr, "10000000stake"},
	)
	systest.Sut.StartChain(t)

	params := cli.CustomQuery("q", "safety", "params")
	require.Equal(t, authorityAddr, gjson.Get(params, "params.authority").String(), params)
	require.Empty(t, gjson.Get(params, "params.operator").String(), params)

	// the authority names the operator; --from fills the authority field
	newParams := fmt.Sprintf(`{"authority":%q,"operator":%q,"scope":["/cosmos.bank.*","/cosmos.staking.*"]}`,
		authorityAddr, operatorAddr)
	safetyRequireCommittedOK(t, cli.Run("tx", "safety", "update-params", "--params", newParams, "--from", authorityAddr))
	params = cli.CustomQuery("q", "safety", "params")
	require.Equal(t, operatorAddr, gjson.Get(params, "params.operator").String(), params)
	require.Equal(t, []string{"/cosmos.bank.*", "/cosmos.staking.*"}, safetyStrings(gjson.Get(params, "params.scope")), params)

	// a key that is not the operator cannot trip
	out, _ := cli.WithRunErrorsIgnored().RunOnly("tx", "safety", "trip", "--patterns", safetyMsgSend, "--from", strangerAddr, safetyFees)
	safetyRequireRejected(t, cli, out, "unauthorized")
	require.Empty(t, safetyCLIDisabledList(t, cli))

	// --patterns takes a comma-separated list and can repeat; both forms mix
	safetyRequireCommittedOK(t, cli.Run("tx", "safety", "trip",
		"--patterns", safetyMsgSend+","+safetyMsgMultiSend, "--patterns", "/cosmos.staking.*",
		"--from", operatorAddr))
	require.Equal(t, []string{safetyMsgMultiSend, safetyMsgSend, "/cosmos.staking.*"}, safetyCLIDisabledList(t, cli))
	require.False(t, safetyCLIIsAllowed(t, cli, safetyMsgSend))
	require.False(t, safetyCLIIsAllowed(t, cli, "/cosmos.staking.v1beta1.MsgDelegate"))
	require.True(t, safetyCLIIsAllowed(t, cli, "/cosmos.gov.v1.MsgVote"))

	out, _ = cli.WithRunErrorsIgnored().RunOnly("tx", "bank", "send", operatorAddr, recipientAddr, "100stake", safetyFees)
	safetyRequireCheckTxRejected(t, out, safetyDisabledErr)
	require.Zero(t, cli.QueryBalance(recipientAddr, "stake"))

	// the operator resets one pattern, then the authority resets the rest
	safetyRequireCommittedOK(t, cli.Run("tx", "safety", "reset", "--patterns", "/cosmos.staking.*", "--from", operatorAddr))
	require.Equal(t, []string{safetyMsgMultiSend, safetyMsgSend}, safetyCLIDisabledList(t, cli))
	safetyRequireCommittedOK(t, cli.Run("tx", "safety", "reset",
		"--patterns", safetyMsgSend, "--patterns", safetyMsgMultiSend, "--from", authorityAddr))
	require.Empty(t, safetyCLIDisabledList(t, cli))
	require.True(t, safetyCLIIsAllowed(t, cli, safetyMsgSend))

	safetyRequireCommittedOK(t, cli.Run("tx", "bank", "send", operatorAddr, recipientAddr, "100stake"))
	require.Equal(t, int64(100), cli.QueryBalance(recipientAddr, "stake"))
}

// safetyCLIDisabledList returns the disabled patterns from `q safety disabled-list`.
func safetyCLIDisabledList(t *testing.T, cli *systest.CLIWrapper) []string {
	t.Helper()
	list := cli.CustomQuery("q", "safety", "disabled-list")
	require.True(t, gjson.Valid(list), list)
	return safetyStrings(gjson.Get(list, "disabled_patterns"))
}

// safetyCLIIsAllowed asks `q safety is-allowed` whether typeURL may execute.
// The command prints {} for false, since JSON output omits default values.
func safetyCLIIsAllowed(t *testing.T, cli *systest.CLIWrapper, typeURL string) bool {
	t.Helper()
	res := cli.CustomQuery("q", "safety", "is-allowed", "--msg-type-url", typeURL)
	require.True(t, gjson.Valid(res), res)
	return gjson.Get(res, "allowed").Bool()
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
	govAddr := sdk.AccAddress(address.Module("gov")).String()
	systest.Sut.StartChain(t)

	require.Empty(t, safetyDisabledList(t))
	require.True(t, safetyIsAllowed(t, safetyMsgSend))
	require.True(t, safetyIsAllowed(t, "/cosmos.gov.v1.MsgVote"))

	params := safetyParams(t)
	require.Equal(t, govAddr, gjson.Get(params, "params.authority").String(), params)
	require.Empty(t, gjson.Get(params, "params.operator").String(), params)
	require.Empty(t, safetyStrings(gjson.Get(params, "params.scope")), "default scope must be empty: %s", params)

	list := safetyREST(t, "/cosmos/safety/v1/disabled_list?pagination.count_total=true")
	require.Empty(t, safetyStrings(gjson.Get(list, "disabled_patterns")), list)
	require.Equal(t, int64(0), gjson.Get(list, "pagination.total").Int(), list)

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

// The ante circuit breaker check opens any msg exposing
// GetMsgs() ([]sdk.Msg, error), such as gov v1 MsgSubmitProposal, and rejects
// a proposal with a tripped msg at submission.
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
	require.Equal(t, []string{safetyMsgSend}, safetyDisabledList(t))
	require.False(t, safetyIsAllowed(t, safetyMsgSend))

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

// safetyParams returns the x/safety params from the REST gateway.
func safetyParams(t *testing.T) string {
	t.Helper()
	return safetyREST(t, "/cosmos/safety/v1/params")
}

// safetyDisabledList returns the disabled patterns from the REST gateway.
func safetyDisabledList(t *testing.T) []string {
	t.Helper()
	list := safetyREST(t, "/cosmos/safety/v1/disabled_list")
	return safetyStrings(gjson.Get(list, "disabled_patterns"))
}

// safetyIsAllowed asks the REST gateway whether typeURL may execute.
func safetyIsAllowed(t *testing.T, typeURL string) bool {
	t.Helper()
	res := safetyREST(t, "/cosmos/safety/v1/is_allowed?msg_type_url="+url.QueryEscape(typeURL))
	require.True(t, gjson.Get(res, "allowed").Exists(), res)
	return gjson.Get(res, "allowed").Bool()
}

// safetyTrip returns a MsgTrip as tx JSON.
func safetyTrip(operator string, patterns ...string) string {
	return safetyPatternsMsg("/cosmos.safety.v1.MsgTrip", "operator", operator, patterns)
}

// safetyReset returns a MsgReset as tx JSON.
func safetyReset(sender string, patterns ...string) string {
	return safetyPatternsMsg("/cosmos.safety.v1.MsgReset", "sender", sender, patterns)
}

func safetyPatternsMsg(typeURL, signerField, signer string, patterns []string) string {
	if patterns == nil {
		patterns = []string{}
	}
	bz, err := json.Marshal(map[string]any{"@type": typeURL, signerField: signer, "patterns": patterns})
	if err != nil {
		panic(err)
	}
	return string(bz)
}

// safetyUpdateParams returns a MsgUpdateParams as tx JSON. params is the
// full params object as JSON.
func safetyUpdateParams(authority, params string) string {
	return fmt.Sprintf(`{"@type":"/cosmos.safety.v1.MsgUpdateParams","authority":%q,"params":%s}`, authority, params)
}

// safetyUnsignedTx writes an unsigned tx with msgs, each given as JSON, to a
// file and returns its path, for tx sign or authz exec.
func safetyUnsignedTx(t *testing.T, msgs ...string) string {
	t.Helper()
	tx := fmt.Sprintf(`{"body":{"messages":[%s]},"auth_info":{"fee":{"amount":[{"denom":"stake","amount":"1"}],"gas_limit":"200000"}},"signatures":[]}`,
		strings.Join(msgs, ","))
	path := filepath.Join(t.TempDir(), "unsigned.json")
	require.NoError(t, os.WriteFile(path, []byte(tx), 0o600))
	return path
}

// safetyBroadcast signs an unsigned tx with msgs as from, broadcasts it in
// sync mode and returns the broadcast output without waiting for the tx to
// commit.
func safetyBroadcast(t *testing.T, cli *systest.CLIWrapper, from string, msgs ...string) string {
	t.Helper()
	signed := safetyTrimJSON(cli.RunCommandWithArgs(cli.WithTargetNodeFlags(
		"tx", "sign", safetyUnsignedTx(t, msgs...), "--from", from, "--chain-id", cli.ChainID(),
	)...))
	require.True(t, gjson.Valid(signed), signed)
	path := filepath.Join(t.TempDir(), "signed.json")
	require.NoError(t, os.WriteFile(path, []byte(signed), 0o600))
	out, _ := cli.WithRunErrorsIgnored().RunOnly("tx", "broadcast", path)
	return out
}

// safetyTx signs and broadcasts a tx with msgs as from, requires it to commit
// with code 0, and returns the committed tx result.
func safetyTx(t *testing.T, cli *systest.CLIWrapper, from string, msgs ...string) string {
	t.Helper()
	out := safetyBroadcast(t, cli, from, msgs...)
	systest.RequireTxSuccess(t, out)
	res, found := cli.AwaitTxCommitted(out)
	require.True(t, found, "tx not committed: %s", out)
	return safetyRequireCommittedOK(t, res)
}

// safetyProposalFile writes a gov proposal with msg, given as JSON, and the
// full deposit to a file and returns its path.
func safetyProposalFile(t *testing.T, msg, title, summary string) string {
	t.Helper()
	proposal := fmt.Sprintf(`{"messages":[%s],"metadata":"ipfs://CID","deposit":"100000000stake","title":%q,"summary":%q}`,
		msg, title, summary)
	path := filepath.Join(t.TempDir(), "proposal.json")
	require.NoError(t, os.WriteFile(path, []byte(proposal), 0o600))
	return path
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

// safetyRequireTxRejected signs and broadcasts a tx with msgs as from and
// requires it to fail in CheckTx or execution. When errMsg is set, the failure
// output must contain it.
func safetyRequireTxRejected(t *testing.T, cli *systest.CLIWrapper, errMsg, from string, msgs ...string) {
	t.Helper()
	safetyRequireRejected(t, cli, safetyBroadcast(t, cli, from, msgs...), errMsg)
}

// safetyRequireRejected requires the broadcast output out to show a tx that
// failed in CheckTx or execution. When errMsg is set, the failure output must
// contain it.
func safetyRequireRejected(t *testing.T, cli *systest.CLIWrapper, out, errMsg string) {
	t.Helper()
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
	out := safetyTrimJSON(cli.RunCommandWithArgs("tx", "bank", "send", from, to, amount, "--generate-only", "--chain-id="+cli.ChainID()))
	require.True(t, gjson.Valid(out), out)
	require.True(t, strings.Contains(out, safetyMsgSend), out)
	path := filepath.Join(t.TempDir(), "send.json")
	require.NoError(t, os.WriteFile(path, []byte(out), 0o600))
	return path
}
