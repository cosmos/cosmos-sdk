package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgUnjailGetSignBytes(t *testing.T) {
	addr := sdk.AccAddress("abcd")
	msg := NewMsgUnjail(sdk.ValAddress(addr).String())
	pc := codec.NewProtoCodec(types.NewInterfaceRegistry())
	bytes, err := pc.MarshalAminoJSON(msg)
	require.NoError(t, err)
	require.Equal(
		t,
		`{"type":"cosmos-sdk/MsgUnjail","value":{"address":"cosmosvaloper1v93xxeqhg9nn6"}}`,
		string(bytes),
	)
}
