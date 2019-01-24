package bank

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestNewMsgSend(t *testing.T) {}

func TestMsgSendRoute(t *testing.T) {
	// Construct a MsgSend
	addr1 := sdk.AccAddress([]byte("input"))
	addr2 := sdk.AccAddress([]byte("output"))
	coins := sdk.Coins{sdk.NewInt64Coin("atom", 10)}
	var msg = MsgSend{
		Inputs:  []Input{NewInput(addr1, coins)},
		Outputs: []Output{NewOutput(addr2, coins)},
	}

	// TODO some failures for bad result
	require.Equal(t, msg.Route(), "bank")
}

func TestInputValidation(t *testing.T) {
	addr1 := sdk.AccAddress([]byte{1, 2})
	addr2 := sdk.AccAddress([]byte{7, 8})
	someCoins := sdk.Coins{sdk.NewInt64Coin("atom", 123)}
	multiCoins := sdk.Coins{sdk.NewInt64Coin("atom", 123), sdk.NewInt64Coin("eth", 20)}

	var emptyAddr sdk.AccAddress
	emptyCoins := sdk.Coins{}
	emptyCoins2 := sdk.Coins{sdk.NewInt64Coin("eth", 0)}
	someEmptyCoins := sdk.Coins{sdk.NewInt64Coin("eth", 10), sdk.NewInt64Coin("atom", 0)}
	unsortedCoins := sdk.Coins{sdk.NewInt64Coin("eth", 1), sdk.NewInt64Coin("atom", 1)}

	cases := []struct {
		valid bool
		txIn  Input
	}{
		// auth works with different apps
		{true, NewInput(addr1, someCoins)},
		{true, NewInput(addr2, someCoins)},
		{true, NewInput(addr2, multiCoins)},

		{false, NewInput(emptyAddr, someCoins)},  // empty address
		{false, NewInput(addr1, emptyCoins)},     // invalid coins
		{false, NewInput(addr1, emptyCoins2)},    // invalid coins
		{false, NewInput(addr1, someEmptyCoins)}, // invalid coins
		{false, NewInput(addr1, unsortedCoins)},  // unsorted coins
	}

	for i, tc := range cases {
		err := tc.txIn.ValidateBasic()
		if tc.valid {
			require.Nil(t, err, "%d: %+v", i, err)
		} else {
			require.NotNil(t, err, "%d", i)
		}
	}
}

func TestOutputValidation(t *testing.T) {
	addr1 := sdk.AccAddress([]byte{1, 2})
	addr2 := sdk.AccAddress([]byte{7, 8})
	someCoins := sdk.Coins{sdk.NewInt64Coin("atom", 123)}
	multiCoins := sdk.Coins{sdk.NewInt64Coin("atom", 123), sdk.NewInt64Coin("eth", 20)}

	var emptyAddr sdk.AccAddress
	emptyCoins := sdk.Coins{}
	emptyCoins2 := sdk.Coins{sdk.NewInt64Coin("eth", 0)}
	someEmptyCoins := sdk.Coins{sdk.NewInt64Coin("eth", 10), sdk.NewInt64Coin("atom", 0)}
	unsortedCoins := sdk.Coins{sdk.NewInt64Coin("eth", 1), sdk.NewInt64Coin("atom", 1)}

	cases := []struct {
		valid bool
		txOut Output
	}{
		// auth works with different apps
		{true, NewOutput(addr1, someCoins)},
		{true, NewOutput(addr2, someCoins)},
		{true, NewOutput(addr2, multiCoins)},

		{false, NewOutput(emptyAddr, someCoins)},  // empty address
		{false, NewOutput(addr1, emptyCoins)},     // invalid coins
		{false, NewOutput(addr1, emptyCoins2)},    // invalid coins
		{false, NewOutput(addr1, someEmptyCoins)}, // invalid coins
		{false, NewOutput(addr1, unsortedCoins)},  // unsorted coins
	}

	for i, tc := range cases {
		err := tc.txOut.ValidateBasic()
		if tc.valid {
			require.Nil(t, err, "%d: %+v", i, err)
		} else {
			require.NotNil(t, err, "%d", i)
		}
	}
}

func TestMsgSendValidation(t *testing.T) {
	addr1 := sdk.AccAddress([]byte{1, 2})
	addr2 := sdk.AccAddress([]byte{7, 8})
	atom123 := sdk.Coins{sdk.NewInt64Coin("atom", 123)}
	atom124 := sdk.Coins{sdk.NewInt64Coin("atom", 124)}
	eth123 := sdk.Coins{sdk.NewInt64Coin("eth", 123)}
	atom123eth123 := sdk.Coins{sdk.NewInt64Coin("atom", 123), sdk.NewInt64Coin("eth", 123)}

	input1 := NewInput(addr1, atom123)
	input2 := NewInput(addr1, eth123)
	output1 := NewOutput(addr2, atom123)
	output2 := NewOutput(addr2, atom124)
	output3 := NewOutput(addr2, eth123)
	outputMulti := NewOutput(addr2, atom123eth123)

	var emptyAddr sdk.AccAddress

	cases := []struct {
		valid bool
		tx    MsgSend
	}{
		{false, MsgSend{}},                           // no input or output
		{false, MsgSend{Inputs: []Input{input1}}},    // just input
		{false, MsgSend{Outputs: []Output{output1}}}, // just output
		{false, MsgSend{
			Inputs:  []Input{NewInput(emptyAddr, atom123)}, // invalid input
			Outputs: []Output{output1}}},
		{false, MsgSend{
			Inputs:  []Input{input1},
			Outputs: []Output{{emptyAddr, atom123}}}, // invalid output
		},
		{false, MsgSend{
			Inputs:  []Input{input1},
			Outputs: []Output{output2}}, // amounts dont match
		},
		{false, MsgSend{
			Inputs:  []Input{input1},
			Outputs: []Output{output3}}, // amounts dont match
		},
		{false, MsgSend{
			Inputs:  []Input{input1},
			Outputs: []Output{outputMulti}}, // amounts dont match
		},
		{false, MsgSend{
			Inputs:  []Input{input2},
			Outputs: []Output{output1}}, // amounts dont match
		},

		{true, MsgSend{
			Inputs:  []Input{input1},
			Outputs: []Output{output1}},
		},
		{true, MsgSend{
			Inputs:  []Input{input1, input2},
			Outputs: []Output{outputMulti}},
		},
	}

	for i, tc := range cases {
		err := tc.tx.ValidateBasic()
		if tc.valid {
			require.Nil(t, err, "%d: %+v", i, err)
		} else {
			require.NotNil(t, err, "%d", i)
		}
	}
}

func TestMsgSendGetSignBytes(t *testing.T) {
	addr1 := sdk.AccAddress([]byte("input"))
	addr2 := sdk.AccAddress([]byte("output"))
	coins := sdk.Coins{sdk.NewInt64Coin("atom", 10)}
	var msg = MsgSend{
		Inputs:  []Input{NewInput(addr1, coins)},
		Outputs: []Output{NewOutput(addr2, coins)},
	}
	res := msg.GetSignBytes()

	expected := `{"type":"cosmos-sdk/Send","value":{"inputs":[{"address":"cosmos1d9h8qat57ljhcm","coins":[{"amount":"10","denom":"atom"}]}],"outputs":[{"address":"cosmos1da6hgur4wsmpnjyg","coins":[{"amount":"10","denom":"atom"}]}]}}`
	require.Equal(t, expected, string(res))
}

func TestMsgSendGetSigners(t *testing.T) {
	var msg = MsgSend{
		Inputs: []Input{
			NewInput(sdk.AccAddress([]byte("input1")), nil),
			NewInput(sdk.AccAddress([]byte("input2")), nil),
			NewInput(sdk.AccAddress([]byte("input3")), nil),
		},
	}
	res := msg.GetSigners()
	// TODO: fix this !
	require.Equal(t, fmt.Sprintf("%v", res), "[696E70757431 696E70757432 696E70757433]")
}

/*
// what to do w/ this test?
func TestMsgSendSigners(t *testing.T) {
	signers := []sdk.AccAddress{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	someCoins := sdk.Coins{sdk.NewInt64Coin("atom", 123)}
	inputs := make([]Input, len(signers))
	for i, signer := range signers {
		inputs[i] = NewInput(signer, someCoins)
	}
	tx := NewMsgSend(inputs, nil)

	require.Equal(t, signers, tx.Signers())
}
*/
