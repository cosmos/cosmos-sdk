package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgSendRoute(t *testing.T) {
	addr1 := sdk.AccAddress([]byte("from"))
	addr2 := sdk.AccAddress([]byte("to"))
	coins := sdk.NewCoins(sdk.NewInt64Coin("atom", 10))
	var msg = NewMsgSend(addr1, addr2, coins)

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "send")
}

func TestMsgSendValidation(t *testing.T) {
	addr1 := sdk.AccAddress([]byte("from________________"))
	addr2 := sdk.AccAddress([]byte("to__________________"))
	addrEmpty := sdk.AccAddress([]byte(""))
	addrLong := sdk.AccAddress([]byte("Purposefully long address"))

	atom123 := sdk.NewCoins(sdk.NewInt64Coin("atom", 123))
	atom0 := sdk.NewCoins(sdk.NewInt64Coin("atom", 0))
	atom123eth123 := sdk.NewCoins(sdk.NewInt64Coin("atom", 123), sdk.NewInt64Coin("eth", 123))
	atom123eth0 := sdk.Coins{sdk.NewInt64Coin("atom", 123), sdk.NewInt64Coin("eth", 0)}

	cases := []struct {
		expectedErr string // empty means no error expected
		msg         *MsgSend
	}{
		{"", NewMsgSend(addr1, addr2, atom123)},                                // valid send
		{"", NewMsgSend(addr1, addr2, atom123eth123)},                          // valid send with multiple coins
		{"", NewMsgSend(addrLong, addr2, atom123)},                             // valid send with long addr sender
		{"", NewMsgSend(addr1, addrLong, atom123)},                             // valid send with long addr recipient
		{": invalid coins", NewMsgSend(addr1, addr2, atom0)},                   // non positive coin
		{"123atom,0eth: invalid coins", NewMsgSend(addr1, addr2, atom123eth0)}, // non positive coin in multicoins
		{"Invalid sender address (empty address string is not allowed): invalid address", NewMsgSend(addrEmpty, addr2, atom123)},
		{"Invalid recipient address :(decoding bech32 failed: invalid bech32 string length 0): invalid address", NewMsgSend(addr1, addrEmpty, atom123)},
	}

	for _, tc := range cases {
		err := tc.msg.ValidateBasic()
		if tc.expectedErr == "" {
			require.Nil(t, err)
		} else {
			require.EqualError(t, err, tc.expectedErr)
		}
	}
}

func TestMsgSendGetSignBytes(t *testing.T) {
	addr1 := sdk.AccAddress([]byte("input"))
	addr2 := sdk.AccAddress([]byte("output"))
	coins := sdk.NewCoins(sdk.NewInt64Coin("atom", 10))
	var msg = NewMsgSend(addr1, addr2, coins)
	res := msg.GetSignBytes()

	expected := `{"type":"cosmos-sdk/MsgSend","value":{"amount":[{"amount":"10","denom":"atom"}],"from_address":"cosmos1d9h8qat57ljhcm","to_address":"cosmos1da6hgur4wsmpnjyg"}}`
	require.Equal(t, expected, string(res))
}

func TestMsgSendGetSigners(t *testing.T) {
	var msg = NewMsgSend(sdk.AccAddress([]byte("input111111111111111")), sdk.AccAddress{}, sdk.NewCoins())
	res := msg.GetSigners()
	// TODO: fix this !
	require.Equal(t, fmt.Sprintf("%v", res), "[696E707574313131313131313131313131313131]")
}
