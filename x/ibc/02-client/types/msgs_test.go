package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func TestMsgCreateClientValidateBasic(t *testing.T) {
	cs := tendermint.ConsensusState{
		Root:             commitment.NewRoot([]byte("root")),
		ValidatorSetHash: []byte("hash"),
	}
	privKey := secp256k1.GenPrivKey()
	signer := sdk.AccAddress(privKey.PubKey().Address())

	cases := []struct {
		msg     types.MsgCreateClient
		expPass bool
		errMsg  string
	}{
		{types.NewMsgCreateClient(exported.ClientTypeTendermint, exported.ClientTypeTendermint, cs, signer), true, "success msg should pass"},
		{types.NewMsgCreateClient("BADCHAIN", exported.ClientTypeTendermint, cs, signer), false, "invalid client id passed"},
		{types.NewMsgCreateClient("goodchain", "invalid_client_type", cs, signer), false, "unregistered client type passed"},
		{types.NewMsgCreateClient("goodchain", exported.ClientTypeTendermint, nil, signer), false, "nil Consensus State in msg passed"},
		{types.NewMsgCreateClient("goodchain", exported.ClientTypeTendermint, tendermint.ConsensusState{}, signer), false, "invalid Consensus State in msg passed"},
		{types.NewMsgCreateClient("goodchain", exported.ClientTypeTendermint, cs, nil), false, "Empty address passed"},
	}

	for i, tc := range cases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			require.NoError(t, err, "Msg %d failed: %v", i, err)
		} else {
			require.Error(t, err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}

func TestMsgUpdateClient(t *testing.T) {
	privKey := secp256k1.GenPrivKey()
	signer := sdk.AccAddress(privKey.PubKey().Address())

	cases := []struct {
		msg     types.MsgUpdateClient
		expPass bool
		errMsg  string
	}{
		{types.NewMsgUpdateClient(exported.ClientTypeTendermint, tendermint.Header{}, signer), true, "success msg should pass"},
		{types.NewMsgUpdateClient("badClient", tendermint.Header{}, signer), false, "invalid client id passed"},
		{types.NewMsgUpdateClient(exported.ClientTypeTendermint, nil, signer), false, "Nil Header passed"},
		{types.NewMsgUpdateClient(exported.ClientTypeTendermint, tendermint.Header{}, nil), false, "Empty address passed"},
	}

	for i, tc := range cases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			require.NoError(t, err, "Msg %d failed: %v", i, err)
		} else {
			require.Error(t, err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}
