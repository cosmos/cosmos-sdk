package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/secp256k1"
	cmn "github.com/tendermint/tendermint/libs/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
)

var _ evidenceexported.Evidence = mockEvidence{}
var _ evidenceexported.Evidence = mockBadEvidence{}

const mockStr = "mock"

// mock GoodEvidence
type mockEvidence struct{}

// Implement Evidence interface
func (me mockEvidence) Route() string        { return mockStr }
func (me mockEvidence) Type() string         { return mockStr }
func (me mockEvidence) String() string       { return mockStr }
func (me mockEvidence) Hash() cmn.HexBytes   { return cmn.HexBytes([]byte(mockStr)) }
func (me mockEvidence) ValidateBasic() error { return nil }
func (me mockEvidence) GetHeight() int64     { return 3 }

// mock bad evidence
type mockBadEvidence struct {
	mockEvidence
}

// Override ValidateBasic
func (mbe mockBadEvidence) ValidateBasic() error {
	return errors.ErrInvalidEvidence
}

func TestMsgCreateClientValidateBasic(t *testing.T) {
	cs := tendermint.ConsensusState{}
	privKey := secp256k1.GenPrivKey()
	signer := sdk.AccAddress(privKey.PubKey().Address())
	testMsgs := []MsgCreateClient{
		NewMsgCreateClient(exported.ClientTypeTendermint, exported.ClientTypeTendermint, cs, signer), // valid msg
		NewMsgCreateClient("badClient", exported.ClientTypeTendermint, cs, signer),                   // invalid client id
		NewMsgCreateClient("goodChain", "bad_type", cs, signer),                                      // invalid client type
		NewMsgCreateClient("goodChain", exported.ClientTypeTendermint, nil, signer),                  // nil Consensus State
		NewMsgCreateClient("goodChain", exported.ClientTypeTendermint, cs, sdk.AccAddress{}),         // empty signer
	}

	cases := []struct {
		msg     MsgCreateClient
		expPass bool
		errMsg  string
	}{
		{testMsgs[0], true, ""},
		{testMsgs[1], false, "invalid client id passed"},
		{testMsgs[2], false, "unregistered client type passed"},
		{testMsgs[3], false, "Nil Consensus State in msg passed"},
		{testMsgs[4], false, "Empty address passed"},
	}

	for i, tc := range cases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			require.Nil(t, err, "Msg %d failed: %v", i, err)
		} else {
			require.NotNil(t, err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}

func TestMsgUpdateClient(t *testing.T) {
	privKey := secp256k1.GenPrivKey()
	signer := sdk.AccAddress(privKey.PubKey().Address())
	testMsgs := []MsgUpdateClient{
		NewMsgUpdateClient(exported.ClientTypeTendermint, tendermint.Header{}, signer),           // valid msg
		NewMsgUpdateClient("badClient", tendermint.Header{}, signer),                             // bad client id
		NewMsgUpdateClient(exported.ClientTypeTendermint, nil, signer),                           // nil Header
		NewMsgUpdateClient(exported.ClientTypeTendermint, tendermint.Header{}, sdk.AccAddress{}), // empty address
	}

	cases := []struct {
		msg     MsgUpdateClient
		expPass bool
		errMsg  string
	}{
		{testMsgs[0], true, ""},
		{testMsgs[1], false, "invalid client id passed"},
		{testMsgs[2], false, "Nil Header passed"},
		{testMsgs[3], false, "Empty address passed"},
	}

	for i, tc := range cases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			require.Nil(t, err, "Msg %d failed: %v", i, err)
		} else {
			require.NotNil(t, err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}

func TestMsgSubmitMisbehaviour(t *testing.T) {
	privKey := secp256k1.GenPrivKey()
	signer := sdk.AccAddress(privKey.PubKey().Address())
	testMsgs := []MsgSubmitMisbehaviour{
		NewMsgSubmitMisbehaviour(exported.ClientTypeTendermint, mockEvidence{}, signer),           // valid msg
		NewMsgSubmitMisbehaviour("badClient", mockEvidence{}, signer),                             // bad client id
		NewMsgSubmitMisbehaviour(exported.ClientTypeTendermint, nil, signer),                      // nil evidence
		NewMsgSubmitMisbehaviour(exported.ClientTypeTendermint, mockBadEvidence{}, signer),        // invalid evidence
		NewMsgSubmitMisbehaviour(exported.ClientTypeTendermint, mockEvidence{}, sdk.AccAddress{}), // empty signer
	}

	cases := []struct {
		msg     MsgSubmitMisbehaviour
		expPass bool
		errMsg  string
	}{
		{testMsgs[0], true, ""},
		{testMsgs[1], false, "invalid client id passed"},
		{testMsgs[2], false, "Nil Evidence passed"},
		{testMsgs[3], false, "Invalid Evidence passed"},
		{testMsgs[4], false, "Empty address passed"},
	}

	for i, tc := range cases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			require.Nil(t, err, "Msg %d failed: %v", i, err)
		} else {
			require.NotNil(t, err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}
