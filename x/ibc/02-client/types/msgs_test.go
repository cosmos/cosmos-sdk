package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
)

func TestMsgCreateClientValidateBasic(t *testing.T) {
	cs := tendermint.ConsensusState{}
	privKey := secp256k1.GenPrivKey()
	signer := sdk.AccAddress(privKey.PubKey().Address())
	testMsgs := []types.MsgCreateClient{
		types.NewMsgCreateClient(exported.ClientTypeTendermint, exported.ClientTypeTendermint, cs, signer), // valid msg
		types.NewMsgCreateClient("badClient", exported.ClientTypeTendermint, cs, signer),                   // invalid client id
		types.NewMsgCreateClient("goodChain", "bad_type", cs, signer),                                      // invalid client type
		types.NewMsgCreateClient("goodChain", exported.ClientTypeTendermint, nil, signer),                  // nil Consensus State
		types.NewMsgCreateClient("goodChain", exported.ClientTypeTendermint, cs, sdk.AccAddress{}),         // empty signer
	}

	cases := []struct {
		msg     types.MsgCreateClient
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
	testMsgs := []types.MsgUpdateClient{
		types.NewMsgUpdateClient(exported.ClientTypeTendermint, tendermint.Header{}, signer),           // valid msg
		types.NewMsgUpdateClient("badClient", tendermint.Header{}, signer),                             // bad client id
		types.NewMsgUpdateClient(exported.ClientTypeTendermint, nil, signer),                           // nil Header
		types.NewMsgUpdateClient(exported.ClientTypeTendermint, tendermint.Header{}, sdk.AccAddress{}), // empty address
	}

	cases := []struct {
		msg     types.MsgUpdateClient
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
