//go:build system_test

package systemtests

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	systest "github.com/cosmos/cosmos-sdk/tools/systemtests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const safetyMsgDelegate = "/cosmos.staking.v1beta1.MsgDelegate"

func TestSafetyEnforcementPaths(t *testing.T) {
	// Scenario:
	// given a running chain whose genesis names an operator with scope *
	// when the operator trips msgs and accounts send them
	// then a msg tripped earlier in the same tx fails at the router in
	// FinalizeBlock and pays its fee, a msg tripped before the tx fails CheckTx
	// and keeps fee and sequence, a prefix trip blocks only its package,
	// simulation fails, a reset in the same tx does not admit the msg after it,
	// and protected msgs run under *.

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
	senderAddr := cli.AddKey("safety-sender")
	recipientAddr := cli.AddKey("safety-recipient")
	otherAddr := cli.AddKey("safety-other")
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", operatorAddr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", senderAddr, "10000000stake"},
	)
	systest.Sut.StartChain(t)
	valAddr := gjson.Get(cli.CustomQuery("q", "staking", "validators"), "validators.0.operator_address").String()
	require.NotEmpty(t, valAddr)

	// FinalizeBlock: the send passes CheckTx because nothing is tripped yet, then
	// the trip before it in the same tx makes the router refuse it. The tx fails
	// as a whole, so the trip is rolled back, but the fee is paid and the
	// sequence used.
	operatorBalance := cli.QueryBalance(operatorAddr, "stake")
	operatorSeq := safetySequence(t, operatorAddr)
	out := safetyBroadcast(t, cli, operatorAddr,
		safetyTrip(operatorAddr, safetyMsgSend), safetySendMsg(operatorAddr, recipientAddr, "10stake"))
	systest.RequireTxSuccess(t, out)
	res, found := cli.AwaitTxCommitted(out)
	require.True(t, found, "tx not committed: %s", out)
	upgradeRequireSafetyRejected(t, res, safetyMsgSend+" by "+safetyMsgSend)
	require.Equal(t, operatorBalance-1, cli.QueryBalance(operatorAddr, "stake"), "fee must be paid")
	require.Equal(t, operatorSeq+1, safetySequence(t, operatorAddr), "sequence must be used")
	require.Zero(t, cli.QueryBalance(recipientAddr, "stake"))
	require.Empty(t, safetyDisabledList(t), "the failed tx must roll back its trip")

	// a prefix trip, and the event names the operator and the patterns
	txResult := safetyTx(t, cli, operatorAddr, safetyTrip(operatorAddr, "/cosmos.bank.*"))
	require.JSONEq(t, fmt.Sprintf("%q", operatorAddr), safetyEventAttr(t, txResult, "cosmos.safety.v1.EventTrip", "operator"))
	require.JSONEq(t, `["/cosmos.bank.*"]`, safetyEventAttr(t, txResult, "cosmos.safety.v1.EventTrip", "patterns"))
	require.Equal(t, []string{"/cosmos.bank.*"}, safetyDisabledList(t))
	for _, typeURL := range []string{safetyMsgSend, safetyMsgMultiSend} {
		res := safetyREST(t, "/cosmos/safety/v1/is_allowed?msg_type_url="+url.QueryEscape(typeURL))
		require.False(t, gjson.Get(res, "allowed").Bool(), res)
		require.Equal(t, "/cosmos.bank.*", gjson.Get(res, "blocked_by").String(), res)
	}
	require.True(t, safetyIsAllowed(t, safetyMsgDelegate))

	// the preview of the trip pages through the same sorted list
	all := safetyStrings(gjson.Get(safetyREST(t, "/cosmos/safety/v1/matching_msgs?pattern="+url.QueryEscape("/cosmos.bank.*")), "msg_type_urls"))
	require.Contains(t, all, safetyMsgSend)
	require.Contains(t, all, safetyMsgMultiSend)
	require.IsNonDecreasing(t, all)
	page := safetyREST(t, "/cosmos/safety/v1/matching_msgs?pattern="+url.QueryEscape("/cosmos.bank.*")+
		"&pagination.limit=1&pagination.count_total=true")
	require.Equal(t, all[:1], safetyStrings(gjson.Get(page, "msg_type_urls")), page)
	require.Equal(t, int64(len(all)), gjson.Get(page, "pagination.total").Int(), page)
	page = safetyREST(t, "/cosmos/safety/v1/matching_msgs?pattern="+url.QueryEscape("/cosmos.bank.*")+
		"&pagination.offset=1&pagination.limit=100")
	require.Equal(t, all[1:], safetyStrings(gjson.Get(page, "msg_type_urls")), page)

	// CheckTx: send and multi-send are refused, keep the fee and the sequence
	senderBalance := cli.QueryBalance(senderAddr, "stake")
	senderSeq := safetySequence(t, senderAddr)
	out, _ = cli.WithRunErrorsIgnored().RunOnly("tx", "bank", "send", senderAddr, recipientAddr, "10stake", safetyFees)
	upgradeRequireSafetyRejected(t, out, safetyMsgSend+" by /cosmos.bank.*")
	out, _ = cli.WithRunErrorsIgnored().RunOnly("tx", "bank", "multi-send", senderAddr, recipientAddr, otherAddr, "10stake", safetyFees)
	upgradeRequireSafetyRejected(t, out, safetyMsgMultiSend+" by /cosmos.bank.*")

	// simulation, which --gas auto runs before signing, fails the same way and
	// broadcasts nothing
	out, _ = cli.WithRunErrorsIgnored().RunOnly("tx", "bank", "send", senderAddr, recipientAddr, "10stake", "--gas", "auto", safetyFees)
	require.Contains(t, out, safetyMsgSend+" by /cosmos.bank.*: "+safetyDisabledErr)
	require.NotContains(t, out, "txhash", out)

	systest.Sut.AwaitNextBlock(t)
	require.Equal(t, senderBalance, cli.QueryBalance(senderAddr, "stake"), "refused txs must not charge fees")
	require.Equal(t, senderSeq, safetySequence(t, senderAddr), "refused txs must keep the sequence")
	require.Zero(t, cli.QueryBalance(recipientAddr, "stake"))
	require.Zero(t, cli.QueryBalance(otherAddr, "stake"))

	// the prefix leaves staking alone
	safetyRequireCommittedOK(t, cli.Run("tx", "staking", "delegate", valAddr, "1000stake", "--from", senderAddr))

	// the ante check reads the disabled list as it stands before the tx, so a
	// reset does not admit a send later in the same tx
	out = safetyBroadcast(t, cli, operatorAddr,
		safetyReset(operatorAddr, "/cosmos.bank.*"), safetySendMsg(operatorAddr, recipientAddr, "10stake"))
	upgradeRequireSafetyRejected(t, out, safetyMsgSend+" by /cosmos.bank.*")
	require.Equal(t, []string{"/cosmos.bank.*"}, safetyDisabledList(t))

	// under * the protected set still runs: gov submit and vote, and a proposal
	// that cancels an upgrade, updates consensus params and updates x/safety
	// params executes while * stands
	safetyTx(t, cli, operatorAddr, safetyTrip(operatorAddr, "*"))
	out, _ = cli.WithRunErrorsIgnored().RunOnly("tx", "staking", "delegate", valAddr, "1000stake", "--from", senderAddr, safetyFees)
	upgradeRequireSafetyRejected(t, out, safetyMsgDelegate+" by *")

	consensus := cli.CustomQuery("q", "consensus", "params")
	require.True(t, gjson.Get(consensus, "params.block").Exists(), consensus)
	const newMaxGas = "123456789"
	updateConsensus := `{"@type":"/cosmos.consensus.v1.MsgUpdateParams","authority":"` + govAddr + `"}`
	for _, field := range []string{"block", "evidence", "validator", "abci"} {
		if v := gjson.Get(consensus, "params."+field); v.Exists() {
			var err error
			updateConsensus, err = sjson.SetRaw(updateConsensus, field, v.Raw)
			require.NoError(t, err)
		}
	}
	updateConsensus, err := sjson.Set(updateConsensus, "block.max_gas", newMaxGas)
	require.NoError(t, err)
	wideScope := fmt.Sprintf(`{"authority":%q,"operator":%q,"scope":["/cosmos.bank.*","*"]}`, govAddr, operatorAddr)
	proposalMsgs := strings.Join([]string{
		fmt.Sprintf(`{"@type":"/cosmos.upgrade.v1beta1.MsgCancelUpgrade","authority":%q}`, govAddr),
		updateConsensus,
		safetyUpdateParams(govAddr, wideScope),
	}, ",")
	safetySubmitAndPassProposal(t, cli, "tx", "gov", "submit-proposal",
		safetyProposalFile(t, proposalMsgs, "protected msgs under *", "cancel upgrade, consensus and safety params"),
		"--from", "node0", "--gas", "400000")
	require.Equal(t, []string{"*", "/cosmos.bank.*"}, safetyDisabledList(t), "the proposal must execute while * stands")
	consensus = cli.CustomQuery("q", "consensus", "params")
	require.Equal(t, newMaxGas, gjson.Get(consensus, "params.block.max_gas").String(), consensus)
	params := safetyParams(t)
	require.Equal(t, []string{"/cosmos.bank.*", "*"}, safetyStrings(gjson.Get(params, "params.scope")), params)

	// the operator resets both entries in one tx, and the event names the sender
	// and the patterns
	txResult = safetyTx(t, cli, operatorAddr, safetyReset(operatorAddr, "*", "/cosmos.bank.*"))
	require.JSONEq(t, fmt.Sprintf("%q", operatorAddr), safetyEventAttr(t, txResult, "cosmos.safety.v1.EventReset", "sender"))
	require.JSONEq(t, `["*","/cosmos.bank.*"]`, safetyEventAttr(t, txResult, "cosmos.safety.v1.EventReset", "patterns"))
	require.Empty(t, safetyDisabledList(t))

	safetyRequireCommittedOK(t, cli.Run("tx", "bank", "send", senderAddr, recipientAddr, "10stake"))
	require.Equal(t, int64(10), cli.QueryBalance(recipientAddr, "stake"))
}

// safetySendMsg returns a bank MsgSend as tx JSON.
func safetySendMsg(from, to, amount string) string {
	coin, err := sdk.ParseCoinNormalized(amount)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf(`{"@type":%q,"from_address":%q,"to_address":%q,"amount":[{"denom":%q,"amount":%q}]}`,
		safetyMsgSend, from, to, coin.Denom, coin.Amount.String())
}

// safetySequence returns the account sequence of addr from the REST gateway.
func safetySequence(t *testing.T, addr string) int64 {
	t.Helper()
	res := safetyREST(t, "/cosmos/auth/v1beta1/accounts/"+addr)
	seq := gjson.Get(res, "account.sequence")
	require.True(t, seq.Exists(), res)
	return seq.Int()
}

// safetyEventAttr returns the value of attribute key on the first event of
// eventType in a committed tx result. Typed events encode values as JSON.
func safetyEventAttr(t *testing.T, txResult, eventType, key string) string {
	t.Helper()
	event := gjson.Get(txResult, fmt.Sprintf(`events.#(type==%q)`, eventType))
	require.True(t, event.Exists(), "no %s event: %s", eventType, txResult)
	value := event.Get(fmt.Sprintf(`attributes.#(key==%q).value`, key))
	require.True(t, value.Exists(), "no %s attribute on %s: %s", key, eventType, txResult)
	return value.String()
}
