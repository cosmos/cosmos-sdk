package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/stretchr/testify/require"
)

func TestNewMsgConnectionOpenInit(t *testing.T) {

	type TestCase = struct {
		msg      MsgConnectionOpenInit
		expected bool
		errMsg   string
	}

	prefix := commitment.NewPrefix([]byte("storePrefixKey"))
	signer, _ := sdk.AccAddressFromBech32("cosmos1ckgw5d7jfj7wwxjzs9fdrdev9vc8dzcw3n2lht")

	testMsgs := []MsgConnectionOpenInit{
		NewMsgConnectionOpenInit("gaia/conn1", "clienttogaiaa", "connectiontogaia", "clienttogaia", prefix, signer),
		NewMsgConnectionOpenInit("ibcconngaia", "gaia/iris", "connectiontogaia", "clienttogaia", prefix, signer),
		NewMsgConnectionOpenInit("ibcconngaia", "clienttogaia", "gaia/conn1", "clienttogaia", prefix, signer),
		NewMsgConnectionOpenInit("ibcconngaia", "clienttogaia", "connectiontogaia", "gaia/conn1", prefix, signer),
		NewMsgConnectionOpenInit("ibcconngaia", "clienttogaia", "connectiontogaia", "clienttogaia", nil, signer),
		NewMsgConnectionOpenInit("ibcconngaia", "clienttogaia", "connectiontogaia", "clienttogaia", prefix, nil),
		NewMsgConnectionOpenInit("ibcconngaia", "clienttogaia", "connectiontogaia", "clienttogaia", prefix, signer),
	}

	var testCases = []TestCase{
		{testMsgs[0], false, "invalid connection ID"},
		{testMsgs[1], false, "invalid client ID"},
		{testMsgs[2], false, "invalid counterparty client ID"},
		{testMsgs[3], false, "invalid counterparty connection ID"},
		{testMsgs[4], false, "empty counterparty prefix"},
		{testMsgs[5], false, "empty singer"},
		{testMsgs[6], true, "success"},
	}

	for i, tc := range testCases {
		require.Equal(t, tc.expected, tc.msg.ValidateBasic() == nil, fmt.Sprintf("case: %d,msg: %s,", i, tc.errMsg))
	}
}

func TestNewMsgConnectionOpenTry(t *testing.T) {
	type TestCase = struct {
		msg      MsgConnectionOpenTry
		expected bool
		errMsg   string
	}

	prefix := commitment.NewPrefix([]byte("storePrefixKey"))
	signer, _ := sdk.AccAddressFromBech32("cosmos1ckgw5d7jfj7wwxjzs9fdrdev9vc8dzcw3n2lht")

	testMsgs := []MsgConnectionOpenTry{
		NewMsgConnectionOpenTry("gaia/conn1", "clienttogaiaa", "connectiontogaia", "clienttogaia", prefix, []string{"1.0.0"}, commitment.Proof{}, commitment.Proof{}, 10, 10, signer),
		NewMsgConnectionOpenTry("ibcconngaia", "gaia/iris", "connectiontogaia", "clienttogaia", prefix, []string{"1.0.0"}, commitment.Proof{}, commitment.Proof{}, 10, 10, signer),
		NewMsgConnectionOpenTry("ibcconngaia", "clienttogaiaa", "ibc/gaia", "clienttogaia", prefix, []string{"1.0.0"}, commitment.Proof{}, commitment.Proof{}, 10, 10, signer),
		NewMsgConnectionOpenTry("ibcconngaia", "clienttogaiaa", "connectiontogaia", "gaia/conn1", prefix, []string{"1.0.0"}, commitment.Proof{}, commitment.Proof{}, 10, 10, signer),
		NewMsgConnectionOpenTry("ibcconngaia", "clienttogaiaa", "connectiontogaia", "clienttogaia", nil, []string{"1.0.0"}, commitment.Proof{}, commitment.Proof{}, 10, 10, signer),
		NewMsgConnectionOpenTry("ibcconngaia", "clienttogaiaa", "connectiontogaia", "clienttogaia", prefix, []string{}, commitment.Proof{}, commitment.Proof{}, 10, 10, signer),
		NewMsgConnectionOpenTry("ibcconngaia", "clienttogaiaa", "connectiontogaia", "clienttogaia", prefix, []string{"1.0.0"}, nil, commitment.Proof{}, 10, 10, signer),
		NewMsgConnectionOpenTry("ibcconngaia", "clienttogaiaa", "connectiontogaia", "clienttogaia", prefix, []string{"1.0.0"}, commitment.Proof{}, nil, 10, 10, signer),
		NewMsgConnectionOpenTry("ibcconngaia", "clienttogaiaa", "connectiontogaia", "clienttogaia", prefix, []string{"1.0.0"}, commitment.Proof{}, commitment.Proof{}, 0, 10, signer),
		NewMsgConnectionOpenTry("ibcconngaia", "clienttogaiaa", "connectiontogaia", "clienttogaia", prefix, []string{"1.0.0"}, commitment.Proof{}, commitment.Proof{}, 10, 0, signer),
		NewMsgConnectionOpenTry("ibcconngaia", "clienttogaiaa", "connectiontogaia", "clienttogaia", prefix, []string{"1.0.0"}, commitment.Proof{}, commitment.Proof{}, 10, 10, nil),
		NewMsgConnectionOpenTry("ibcconngaia", "clienttogaiaa", "connectiontogaia", "clienttogaia", prefix, []string{"1.0.0"}, commitment.Proof{}, commitment.Proof{}, 10, 10, signer),
	}

	var testCases = []TestCase{
		{testMsgs[0], false, "invalid connection ID"},
		{testMsgs[1], false, "invalid client ID"},
		{testMsgs[2], false, "invalid counterparty connection ID"},
		{testMsgs[3], false, "invalid counterparty client ID"},
		{testMsgs[4], false, "empty counterparty prefix"},
		{testMsgs[5], false, "empty counterpartyVersions"},
		{testMsgs[6], false, "empty proofInit"},
		{testMsgs[7], false, "empty proofConsensus"},
		{testMsgs[8], false, "invalid proofHeight"},
		{testMsgs[9], false, "invalid consensusHeight"},
		{testMsgs[10], false, "empty singer"},
		{testMsgs[11], true, "success"},
	}

	for i, tc := range testCases {
		require.Equal(t, tc.expected, tc.msg.ValidateBasic() == nil, fmt.Sprintf("case: %d,msg: %s,", i, tc.errMsg))
	}
}

func TestNewMsgConnectionOpenAck(t *testing.T) {
	type TestCase = struct {
		msg      MsgConnectionOpenAck
		expected bool
		errMsg   string
	}

	signer, _ := sdk.AccAddressFromBech32("cosmos1ckgw5d7jfj7wwxjzs9fdrdev9vc8dzcw3n2lht")

	testMsgs := []MsgConnectionOpenAck{
		NewMsgConnectionOpenAck("gaia/conn1", commitment.Proof{}, commitment.Proof{}, 10, 10, "1.0.0", signer),
		NewMsgConnectionOpenAck("ibcconngaia", nil, commitment.Proof{}, 10, 10, "1.0.0", signer),
		NewMsgConnectionOpenAck("ibcconngaia", commitment.Proof{}, nil, 10, 10, "1.0.0", signer),
		NewMsgConnectionOpenAck("ibcconngaia", commitment.Proof{}, commitment.Proof{}, 0, 10, "1.0.0", signer),
		NewMsgConnectionOpenAck("ibcconngaia", commitment.Proof{}, commitment.Proof{}, 10, 0, "1.0.0", signer),
		NewMsgConnectionOpenAck("ibcconngaia", commitment.Proof{}, commitment.Proof{}, 10, 10, "", signer),
		NewMsgConnectionOpenAck("ibcconngaia", commitment.Proof{}, commitment.Proof{}, 10, 10, "1.0.0", nil),
		NewMsgConnectionOpenAck("ibcconngaia", commitment.Proof{}, commitment.Proof{}, 10, 10, "1.0.0", signer),
	}
	var testCases = []TestCase{
		{testMsgs[0], false, "invalid connection ID"},
		{testMsgs[1], false, "empty proofTry"},
		{testMsgs[2], false, "empty proofConsensus"},
		{testMsgs[3], false, "invalid proofHeight"},
		{testMsgs[4], false, "invalid consensusHeight"},
		{testMsgs[5], false, "invalid version"},
		{testMsgs[6], false, "empty signer"},
		{testMsgs[7], true, "success"},
	}

	for i, tc := range testCases {
		require.Equal(t, tc.expected, tc.msg.ValidateBasic() == nil, fmt.Sprintf("case: %d,msg: %s,", i, tc.errMsg))
	}
}

func TestNewMsgConnectionOpenConfirm(t *testing.T) {
	type TestCase = struct {
		msg      MsgConnectionOpenConfirm
		expected bool
		errMsg   string
	}

	signer, _ := sdk.AccAddressFromBech32("cosmos1ckgw5d7jfj7wwxjzs9fdrdev9vc8dzcw3n2lht")

	testMsgs := []MsgConnectionOpenConfirm{
		NewMsgConnectionOpenConfirm("gaia/conn1", commitment.Proof{}, 10, signer),
		NewMsgConnectionOpenConfirm("ibcconngaia", nil, 10, signer),
		NewMsgConnectionOpenConfirm("ibcconngaia", commitment.Proof{}, 0, signer),
		NewMsgConnectionOpenConfirm("ibcconngaia", commitment.Proof{}, 10, nil),
		NewMsgConnectionOpenConfirm("ibcconngaia", commitment.Proof{}, 10, signer),
	}

	var testCases = []TestCase{
		{testMsgs[0], false, "invalid connection ID"},
		{testMsgs[1], false, "empty proofTry"},
		{testMsgs[2], false, "invalid proofHeight"},
		{testMsgs[3], false, "empty signer"},
		{testMsgs[4], true, "success"},
	}

	for i, tc := range testCases {
		require.Equal(t, tc.expected, tc.msg.ValidateBasic() == nil, fmt.Sprintf("case: %d,msg: %s,", i, tc.errMsg))
	}
}
