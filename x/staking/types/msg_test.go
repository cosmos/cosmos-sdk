package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

var coinPos = sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)

func TestMsgDecode(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	types.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	// firstly we start testing the pubkey serialization

	pk1bz, err := cdc.MarshalInterface(pk1)
	require.NoError(t, err)
	var pkUnmarshaled cryptotypes.PubKey
	err = cdc.UnmarshalInterface(pk1bz, &pkUnmarshaled)
	require.NoError(t, err)
	require.True(t, pk1.Equals(pkUnmarshaled.(*ed25519.PubKey)))

	// now let's try to serialize the whole message

	commission1 := types.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec())
	msg, err := types.NewMsgCreateValidator(valAddr1.String(), pk1, coinPos, types.Description{}, commission1, math.OneInt())
	require.NoError(t, err)
	msgSerialized, err := cdc.MarshalInterface(msg)
	require.NoError(t, err)

	var msgUnmarshaled sdk.Msg
	err = cdc.UnmarshalInterface(msgSerialized, &msgUnmarshaled)
	require.NoError(t, err)
	msg2, ok := msgUnmarshaled.(*types.MsgCreateValidator)
	require.True(t, ok)
	require.True(t, msg.Value.IsEqual(msg2.Value))
	require.True(t, msg.Pubkey.Equal(msg2.Pubkey))
}
