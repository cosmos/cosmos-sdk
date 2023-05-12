package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func TestMsgSendGetSignBytes(t *testing.T) {
	addr1 := sdk.AccAddress([]byte("input"))
	addr2 := sdk.AccAddress([]byte("output"))
	coins := sdk.NewCoins(sdk.NewInt64Coin("atom", 10))
	msg := NewMsgSend(addr1, addr2, coins)
	res := msg.GetSignBytes()

	expected := `{"type":"cosmos-sdk/MsgSend","value":{"amount":[{"amount":"10","denom":"atom"}],"from_address":"cosmos1d9h8qat57ljhcm","to_address":"cosmos1da6hgur4wsmpnjyg"}}`
	require.Equal(t, expected, string(res))
}

func TestInputValidation(t *testing.T) {
	addr1 := sdk.AccAddress([]byte("_______alice________"))
	addr2 := sdk.AccAddress([]byte("________bob_________"))
	addrEmpty := sdk.AccAddress([]byte(""))
	addrLong := sdk.AccAddress([]byte("Purposefully long address"))

	someCoins := sdk.NewCoins(sdk.NewInt64Coin("atom", 123))
	multiCoins := sdk.NewCoins(sdk.NewInt64Coin("atom", 123), sdk.NewInt64Coin("eth", 20))

	emptyCoins := sdk.NewCoins()
	emptyCoins2 := sdk.NewCoins(sdk.NewInt64Coin("eth", 0))
	someEmptyCoins := sdk.Coins{sdk.NewInt64Coin("eth", 10), sdk.NewInt64Coin("atom", 0)}
	unsortedCoins := sdk.Coins{sdk.NewInt64Coin("eth", 1), sdk.NewInt64Coin("atom", 1)}

	cases := []struct {
		expectedErr string // empty means no error expected
		txIn        Input
	}{
		// auth works with different apps
		{"", NewInput(addr1, someCoins)},
		{"", NewInput(addr2, someCoins)},
		{"", NewInput(addr2, multiCoins)},
		{"", NewInput(addrLong, someCoins)},

		{"invalid input address: empty address string is not allowed: invalid address", NewInput(addrEmpty, someCoins)},
		{": invalid coins", NewInput(addr1, emptyCoins)},                // invalid coins
		{": invalid coins", NewInput(addr1, emptyCoins2)},               // invalid coins
		{"10eth,0atom: invalid coins", NewInput(addr1, someEmptyCoins)}, // invalid coins
		{"1eth,1atom: invalid coins", NewInput(addr1, unsortedCoins)},   // unsorted coins
	}

	for i, tc := range cases {
		err := tc.txIn.ValidateBasic()
		if tc.expectedErr == "" {
			require.Nil(t, err, "%d: %+v", i, err)
		} else {
			require.EqualError(t, err, tc.expectedErr, "%d", i)
		}
	}
}

func TestOutputValidation(t *testing.T) {
	addr1 := sdk.AccAddress([]byte("_______alice________"))
	addr2 := sdk.AccAddress([]byte("________bob_________"))
	addrEmpty := sdk.AccAddress([]byte(""))
	addrLong := sdk.AccAddress([]byte("Purposefully long address"))

	someCoins := sdk.NewCoins(sdk.NewInt64Coin("atom", 123))
	multiCoins := sdk.NewCoins(sdk.NewInt64Coin("atom", 123), sdk.NewInt64Coin("eth", 20))

	emptyCoins := sdk.NewCoins()
	emptyCoins2 := sdk.NewCoins(sdk.NewInt64Coin("eth", 0))
	someEmptyCoins := sdk.Coins{sdk.NewInt64Coin("eth", 10), sdk.NewInt64Coin("atom", 0)}
	unsortedCoins := sdk.Coins{sdk.NewInt64Coin("eth", 1), sdk.NewInt64Coin("atom", 1)}

	cases := []struct {
		expectedErr string // empty means no error expected
		txOut       Output
	}{
		// auth works with different apps
		{"", NewOutput(addr1, someCoins)},
		{"", NewOutput(addr2, someCoins)},
		{"", NewOutput(addr2, multiCoins)},
		{"", NewOutput(addrLong, someCoins)},

		{"invalid output address: empty address string is not allowed: invalid address", NewOutput(addrEmpty, someCoins)},
		{": invalid coins", NewOutput(addr1, emptyCoins)},                // invalid coins
		{": invalid coins", NewOutput(addr1, emptyCoins2)},               // invalid coins
		{"10eth,0atom: invalid coins", NewOutput(addr1, someEmptyCoins)}, // invalid coins
		{"1eth,1atom: invalid coins", NewOutput(addr1, unsortedCoins)},   // unsorted coins
	}

	for i, tc := range cases {
		err := tc.txOut.ValidateBasic()
		if tc.expectedErr == "" {
			require.Nil(t, err, "%d: %+v", i, err)
		} else {
			require.EqualError(t, err, tc.expectedErr, "%d", i)
		}
	}
}

func TestMsgMultiSendGetSignBytes(t *testing.T) {
	addr1 := sdk.AccAddress([]byte("input"))
	addr2 := sdk.AccAddress([]byte("output"))
	coins := sdk.NewCoins(sdk.NewInt64Coin("atom", 10))
	msg := MsgMultiSend{
		Inputs:  []Input{NewInput(addr1, coins)},
		Outputs: []Output{NewOutput(addr2, coins)},
	}
	res := msg.GetSignBytes()

	expected := `{"type":"cosmos-sdk/MsgMultiSend","value":{"inputs":[{"address":"cosmos1d9h8qat57ljhcm","coins":[{"amount":"10","denom":"atom"}]}],"outputs":[{"address":"cosmos1da6hgur4wsmpnjyg","coins":[{"amount":"10","denom":"atom"}]}]}}`
	require.Equal(t, expected, string(res))
}

func TestMsgMultiSendGetSigners(t *testing.T) {
	addr := sdk.AccAddress([]byte("input111111111111111"))
	input := NewInput(addr, nil)
	msg := NewMsgMultiSend(input, nil)

	res := msg.GetSigners()
	require.Equal(t, 1, len(res))
	require.True(t, addr.Equals(res[0]))
}

func TestNewMsgSetSendEnabled(t *testing.T) {
	// Punt. Just setting one to all non-default values and making sure they're as expected.
	msg := NewMsgSetSendEnabled("milton", []*SendEnabled{{"barrycoin", true}}, []string{"billcoin"})
	assert.Equal(t, "milton", msg.Authority, "msg.Authority")
	if assert.Len(t, msg.SendEnabled, 1, "msg.SendEnabled length") {
		assert.Equal(t, "barrycoin", msg.SendEnabled[0].Denom, "msg.SendEnabled[0].Denom")
		assert.True(t, msg.SendEnabled[0].Enabled, "msg.SendEnabled[0].Enabled")
	}
	if assert.Len(t, msg.UseDefaultFor, 1, "msg.UseDefault") {
		assert.Equal(t, "billcoin", msg.UseDefaultFor[0], "msg.UseDefault[0]")
	}
}

func TestMsgSendGetSigners(t *testing.T) {
	from := sdk.AccAddress([]byte("input111111111111111"))
	msg := NewMsgSend(from, sdk.AccAddress{}, sdk.NewCoins())
	res := msg.GetSigners()
	require.Equal(t, 1, len(res))
	require.True(t, from.Equals(res[0]))
}

func TestMsgSetSendEnabledGetSignBytes(t *testing.T) {
	msg := NewMsgSetSendEnabled("cartman", []*SendEnabled{{"casafiestacoin", false}, {"kylecoin", true}}, nil)
	expected := `{"type":"cosmos-sdk/MsgSetSendEnabled","value":{"authority":"cartman","send_enabled":[{"denom":"casafiestacoin"},{"denom":"kylecoin","enabled":true}]}}`
	actualBz := msg.GetSignBytes()
	actual := string(actualBz)
	assert.Equal(t, expected, actual)
}

func TestMsgSetSendEnabledGetSigners(t *testing.T) {
	govModuleAddr := authtypes.NewModuleAddress(govtypes.ModuleName)
	msg := NewMsgSetSendEnabled(govModuleAddr.String(), nil, nil)
	expected := []sdk.AccAddress{govModuleAddr}
	actual := msg.GetSigners()
	assert.Equal(t, expected, actual)
}
