//go:build system_test

package systemtests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	systest "github.com/cosmos/cosmos-sdk/tools/systemtests"
)

func TestSafetyRejectedMsgs(t *testing.T) {
	// Scenario:
	// given a running chain whose x/safety authority is a test key and whose
	// genesis disables MsgSend and staking MsgDelegate
	// when the authority names an operator, first with an empty scope and then
	// with scope /cosmos.bank.*, and the authority, operator and a stranger send
	// trips, resets and param updates that break a rule
	// then every one of them fails, a batch with one bad pattern changes
	// nothing, and only the valid half of each batch succeeds on its own.

	// the authority address goes into genesis, so derive it from a fixed seed
	// and add the key again after ModifyGenesisJSON resets the keyring
	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	authorityAddr := cli.AddKeyFromSeed("safety-authority", testSeed)
	systest.Sut.ModifyGenesisJSON(t,
		safetySetGenesis(t, "", nil, []string{safetyMsgSend, safetyMsgDelegate}),
		func(genesis []byte) []byte {
			genesis, err := sjson.SetBytes(genesis, "app_state.safety.params.authority", authorityAddr)
			require.NoError(t, err)
			return genesis
		},
	)
	require.Equal(t, authorityAddr, cli.AddKeyFromSeed("safety-authority", testSeed))
	operatorAddr := cli.AddKey("safety-operator")
	strangerAddr := cli.AddKey("safety-stranger")
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", authorityAddr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", operatorAddr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", strangerAddr, "10000000stake"},
	)
	systest.Sut.StartChain(t)
	genesisList := []string{safetyMsgSend, safetyMsgDelegate}
	require.Equal(t, genesisList, safetyDisabledList(t))

	// the authority names the operator with an empty scope, and the event
	// names the signer and the new params
	emptyScope := fmt.Sprintf(`{"authority":%q,"operator":%q,"scope":[]}`, authorityAddr, operatorAddr)
	txResult := safetyTx(t, cli, authorityAddr, safetyUpdateParams(authorityAddr, emptyScope))
	require.JSONEq(t, fmt.Sprintf("%q", authorityAddr),
		safetyEventAttr(t, txResult, "cosmos.safety.v1.EventParamsUpdated", "authority"))
	eventParams := safetyEventAttr(t, txResult, "cosmos.safety.v1.EventParamsUpdated", "params")
	require.Equal(t, authorityAddr, gjson.Get(eventParams, "authority").String(), eventParams)
	require.Equal(t, operatorAddr, gjson.Get(eventParams, "operator").String(), eventParams)
	require.Empty(t, safetyStrings(gjson.Get(eventParams, "scope")), eventParams)
	params := safetyParams(t)
	require.Equal(t, operatorAddr, gjson.Get(params, "params.operator").String(), params)
	require.Empty(t, safetyStrings(gjson.Get(params, "params.scope")), params)

	// an operator with an empty scope can neither trip nor reset anything
	safetyRequireTxRejected(t, cli, "not covered by operator scope", operatorAddr, safetyTrip(operatorAddr, safetyMsgMultiSend))
	safetyRequireTxRejected(t, cli, "not covered by operator scope", operatorAddr, safetyTrip(operatorAddr, "*"))
	safetyRequireTxRejected(t, cli, "not covered by operator scope", operatorAddr, safetyReset(operatorAddr, safetyMsgSend))
	require.Equal(t, genesisList, safetyDisabledList(t))

	bankScope := fmt.Sprintf(`{"authority":%q,"operator":%q,"scope":["/cosmos.bank.*"]}`, authorityAddr, operatorAddr)
	safetyTx(t, cli, authorityAddr, safetyUpdateParams(authorityAddr, bankScope))

	tooLong := "/" + strings.Repeat("a", 256)
	specs := map[string]struct {
		from   string
		msg    string
		errMsg string
	}{
		"authority cannot trip": {
			from: authorityAddr, msg: safetyTrip(authorityAddr, safetyMsgMultiSend), errMsg: "is not the operator",
		},
		"operator cannot update params": {
			from: operatorAddr, msg: safetyUpdateParams(operatorAddr, bankScope), errMsg: "invalid authority",
		},
		"stranger cannot update params": {
			from: strangerAddr, msg: safetyUpdateParams(strangerAddr, bankScope), errMsg: "invalid authority",
		},
		"authority cannot set a malformed scope": {
			from:   authorityAddr,
			msg:    safetyUpdateParams(authorityAddr, fmt.Sprintf(`{"authority":%q,"operator":%q,"scope":["/cosmos.bank*"]}`, authorityAddr, operatorAddr)),
			errMsg: "invalid pattern",
		},
		"trip without a leading slash": {
			from: operatorAddr, msg: safetyTrip(operatorAddr, "cosmos.bank.*"), errMsg: "invalid pattern",
		},
		"trip of a prefix without a segment boundary": {
			from: operatorAddr, msg: safetyTrip(operatorAddr, "/cosmos.bank*"), errMsg: "invalid pattern",
		},
		"trip of an empty segment": {
			from: operatorAddr, msg: safetyTrip(operatorAddr, "/cosmos..bank.*"), errMsg: "invalid pattern",
		},
		"trip of a bare slash": {
			from: operatorAddr, msg: safetyTrip(operatorAddr, "/"), errMsg: "invalid pattern",
		},
		"trip of an empty pattern": {
			from: operatorAddr, msg: safetyTrip(operatorAddr, ""), errMsg: "invalid pattern",
		},
		"trip of a pattern over 256 bytes": {
			from: operatorAddr, msg: safetyTrip(operatorAddr, tooLong), errMsg: "invalid pattern",
		},
		"trip of a duplicate pattern": {
			from: operatorAddr, msg: safetyTrip(operatorAddr, safetyMsgMultiSend, safetyMsgMultiSend), errMsg: "duplicate pattern",
		},
		"trip of an in-scope prefix that matches no msg": {
			from: operatorAddr, msg: safetyTrip(operatorAddr, "/cosmos.bank.v9.*"), errMsg: "matches no registered unprotected msg type",
		},
		"batch trip with an out of scope pattern": {
			from: operatorAddr, msg: safetyTrip(operatorAddr, safetyMsgMultiSend, "/cosmos.staking.v1beta1.MsgUndelegate"),
			errMsg: "not covered by operator scope",
		},
		"batch trip with a typo": {
			from: operatorAddr, msg: safetyTrip(operatorAddr, safetyMsgMultiSend, "/cosmos.bank.v1beta1.MsgSnd"),
			errMsg: "matches no registered unprotected msg type",
		},
		"batch trip with an already disabled pattern": {
			from: operatorAddr, msg: safetyTrip(operatorAddr, safetyMsgMultiSend, safetyMsgSend), errMsg: "already disabled",
		},
		"batch trip with a malformed pattern": {
			from: operatorAddr, msg: safetyTrip(operatorAddr, safetyMsgMultiSend, "/cosmos.bank*"), errMsg: "invalid pattern",
		},
		"batch reset with a pattern not disabled": {
			from: operatorAddr, msg: safetyReset(operatorAddr, safetyMsgSend, safetyMsgMultiSend), errMsg: "not in disabled list",
		},
		"reset of a prefix does not reach an exact entry": {
			from: operatorAddr, msg: safetyReset(operatorAddr, "/cosmos.bank.*"), errMsg: "not in disabled list",
		},
		"operator reset outside its scope": {
			from: operatorAddr, msg: safetyReset(operatorAddr, safetyMsgDelegate), errMsg: "not covered by operator scope",
		},
		"stranger reset": {
			from: strangerAddr, msg: safetyReset(strangerAddr, safetyMsgDelegate), errMsg: "is neither the authority nor the operator",
		},
		"reset of a malformed pattern": {
			from: authorityAddr, msg: safetyReset(authorityAddr, "/cosmos.bank*"), errMsg: "invalid pattern",
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			safetyRequireTxRejected(t, cli, spec.errMsg, spec.from, spec.msg)
		})
	}

	// nothing above changed state
	require.Equal(t, genesisList, safetyDisabledList(t))
	params = safetyParams(t)
	require.Equal(t, authorityAddr, gjson.Get(params, "params.authority").String(), params)
	require.Equal(t, operatorAddr, gjson.Get(params, "params.operator").String(), params)
	require.Equal(t, []string{"/cosmos.bank.*"}, safetyStrings(gjson.Get(params, "params.scope")), params)

	// the valid half of each rejected batch succeeds on its own, so the batches
	// failed on their bad pattern alone
	safetyTx(t, cli, operatorAddr, safetyTrip(operatorAddr, safetyMsgMultiSend))
	require.Equal(t, []string{safetyMsgMultiSend, safetyMsgSend, safetyMsgDelegate}, safetyDisabledList(t))
	safetyTx(t, cli, operatorAddr, safetyReset(operatorAddr, safetyMsgSend, safetyMsgMultiSend))

	// the authority resets the entry outside the operator scope
	txResult = safetyTx(t, cli, authorityAddr, safetyReset(authorityAddr, safetyMsgDelegate))
	require.JSONEq(t, fmt.Sprintf("%q", authorityAddr), safetyEventAttr(t, txResult, "cosmos.safety.v1.EventReset", "sender"))
	require.Empty(t, safetyDisabledList(t))
}
