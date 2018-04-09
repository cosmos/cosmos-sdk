package bank

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestNewSendMsg(t *testing.T) {}

func TestSendMsgType(t *testing.T) {
	// Construct a SendMsg
	addr1 := sdk.Address([]byte("input"))
	addr2 := sdk.Address([]byte("output"))
	coins := sdk.Coins{{"atom", 10}}
	var msg = SendMsg{
		Inputs:  []Input{NewInput(addr1, coins)},
		Outputs: []Output{NewOutput(addr2, coins)},
	}

	// TODO some failures for bad result
	assert.Equal(t, msg.Type(), "bank")
}

func TestInputValidation(t *testing.T) {
	addr1 := sdk.Address([]byte{1, 2})
	addr2 := sdk.Address([]byte{7, 8})
	someCoins := sdk.Coins{{"atom", 123}}
	multiCoins := sdk.Coins{{"atom", 123}, {"eth", 20}}

	var emptyAddr sdk.Address
	emptyCoins := sdk.Coins{}
	emptyCoins2 := sdk.Coins{{"eth", 0}}
	someEmptyCoins := sdk.Coins{{"eth", 10}, {"atom", 0}}
	minusCoins := sdk.Coins{{"eth", -34}}
	someMinusCoins := sdk.Coins{{"atom", 20}, {"eth", -34}}
	unsortedCoins := sdk.Coins{{"eth", 1}, {"atom", 1}}

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
		{false, NewInput(addr1, minusCoins)},     // negative coins
		{false, NewInput(addr1, someMinusCoins)}, // negative coins
		{false, NewInput(addr1, unsortedCoins)},  // unsorted coins
	}

	for i, tc := range cases {
		err := tc.txIn.ValidateBasic()
		if tc.valid {
			assert.Nil(t, err, "%d: %+v", i, err)
		} else {
			assert.NotNil(t, err, "%d", i)
		}
	}
}

func TestOutputValidation(t *testing.T) {
	addr1 := sdk.Address([]byte{1, 2})
	addr2 := sdk.Address([]byte{7, 8})
	someCoins := sdk.Coins{{"atom", 123}}
	multiCoins := sdk.Coins{{"atom", 123}, {"eth", 20}}

	var emptyAddr sdk.Address
	emptyCoins := sdk.Coins{}
	emptyCoins2 := sdk.Coins{{"eth", 0}}
	someEmptyCoins := sdk.Coins{{"eth", 10}, {"atom", 0}}
	minusCoins := sdk.Coins{{"eth", -34}}
	someMinusCoins := sdk.Coins{{"atom", 20}, {"eth", -34}}
	unsortedCoins := sdk.Coins{{"eth", 1}, {"atom", 1}}

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
		{false, NewOutput(addr1, minusCoins)},     // negative coins
		{false, NewOutput(addr1, someMinusCoins)}, // negative coins
		{false, NewOutput(addr1, unsortedCoins)},  // unsorted coins
	}

	for i, tc := range cases {
		err := tc.txOut.ValidateBasic()
		if tc.valid {
			assert.Nil(t, err, "%d: %+v", i, err)
		} else {
			assert.NotNil(t, err, "%d", i)
		}
	}
}

func TestSendMsgValidation(t *testing.T) {
	addr1 := sdk.Address([]byte{1, 2})
	addr2 := sdk.Address([]byte{7, 8})
	atom123 := sdk.Coins{{"atom", 123}}
	atom124 := sdk.Coins{{"atom", 124}}
	eth123 := sdk.Coins{{"eth", 123}}
	atom123eth123 := sdk.Coins{{"atom", 123}, {"eth", 123}}

	input1 := NewInput(addr1, atom123)
	input2 := NewInput(addr1, eth123)
	output1 := NewOutput(addr2, atom123)
	output2 := NewOutput(addr2, atom124)
	output3 := NewOutput(addr2, eth123)
	outputMulti := NewOutput(addr2, atom123eth123)

	var emptyAddr sdk.Address

	cases := []struct {
		valid bool
		tx    SendMsg
	}{
		{false, SendMsg{}},                           // no input or output
		{false, SendMsg{Inputs: []Input{input1}}},    // just input
		{false, SendMsg{Outputs: []Output{output1}}}, // just ouput
		{false, SendMsg{
			Inputs:  []Input{NewInput(emptyAddr, atom123)}, // invalid input
			Outputs: []Output{output1}}},
		{false, SendMsg{
			Inputs:  []Input{input1},
			Outputs: []Output{{emptyAddr, atom123}}}, // invalid ouput
		},
		{false, SendMsg{
			Inputs:  []Input{input1},
			Outputs: []Output{output2}}, // amounts dont match
		},
		{false, SendMsg{
			Inputs:  []Input{input1},
			Outputs: []Output{output3}}, // amounts dont match
		},
		{false, SendMsg{
			Inputs:  []Input{input1},
			Outputs: []Output{outputMulti}}, // amounts dont match
		},
		{false, SendMsg{
			Inputs:  []Input{input2},
			Outputs: []Output{output1}}, // amounts dont match
		},

		{true, SendMsg{
			Inputs:  []Input{input1},
			Outputs: []Output{output1}},
		},
		{true, SendMsg{
			Inputs:  []Input{input1, input2},
			Outputs: []Output{outputMulti}},
		},
	}

	for i, tc := range cases {
		err := tc.tx.ValidateBasic()
		if tc.valid {
			assert.Nil(t, err, "%d: %+v", i, err)
		} else {
			assert.NotNil(t, err, "%d", i)
		}
	}
}

func TestSendMsgString(t *testing.T) {
	// Construct a SendMsg
	addr1String := "input"
	addr2String := "output"
	addr1 := sdk.Address([]byte(addr1String))
	addr2 := sdk.Address([]byte(addr2String))
	coins := sdk.Coins{{"atom", 10}}
	var msg = SendMsg{
		Inputs:  []Input{NewInput(addr1, coins)},
		Outputs: []Output{NewOutput(addr2, coins)},
	}

	res := msg.String()
	expected := fmt.Sprintf("SendMsg{[Input{%X,10atom}]->[Output{%X,10atom}]}", addr1String, addr2String)
	// TODO some failures for bad results
	assert.Equal(t, expected, res)
}

func TestSendMsgGet(t *testing.T) {
	addr1 := sdk.Address([]byte("input"))
	addr2 := sdk.Address([]byte("output"))
	coins := sdk.Coins{{"atom", 10}}
	var msg = SendMsg{
		Inputs:  []Input{NewInput(addr1, coins)},
		Outputs: []Output{NewOutput(addr2, coins)},
	}
	res := msg.Get(nil)
	assert.Nil(t, res)
}

func TestSendMsgGetSignBytes(t *testing.T) {
	addr1 := sdk.Address([]byte("input"))
	addr2 := sdk.Address([]byte("output"))
	coins := sdk.Coins{{"atom", 10}}
	var msg = SendMsg{
		Inputs:  []Input{NewInput(addr1, coins)},
		Outputs: []Output{NewOutput(addr2, coins)},
	}
	res := msg.GetSignBytes()
	// TODO bad results
	assert.Equal(t, string(res), `{"inputs":[{"address":"696E707574","coins":[{"denom":"atom","amount":10}]}],"outputs":[{"address":"6F7574707574","coins":[{"denom":"atom","amount":10}]}]}`)
}

func TestSendMsgGetSigners(t *testing.T) {
	var msg = SendMsg{
		Inputs: []Input{
			NewInput(sdk.Address([]byte("input1")), nil),
			NewInput(sdk.Address([]byte("input2")), nil),
			NewInput(sdk.Address([]byte("input3")), nil),
		},
	}
	res := msg.GetSigners()
	// TODO: fix this !
	assert.Equal(t, fmt.Sprintf("%v", res), "[696E70757431 696E70757432 696E70757433]")
}

/*
// what to do w/ this test?
func TestSendMsgSigners(t *testing.T) {
	signers := []sdk.Address{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	someCoins := sdk.Coins{{"atom", 123}}
	inputs := make([]Input, len(signers))
	for i, signer := range signers {
		inputs[i] = NewInput(signer, someCoins)
	}
	tx := NewSendMsg(inputs, nil)

	assert.Equal(t, signers, tx.Signers())
}
*/

// ----------------------------------------
// IssueMsg Tests

func TestNewIssueMsg(t *testing.T) {
	// TODO
}

func TestIssueMsgType(t *testing.T) {
	// Construct an IssueMsg
	addr := sdk.Address([]byte("loan-from-bank"))
	coins := sdk.Coins{{"atom", 10}}
	var msg = IssueMsg{
		Banker:  sdk.Address([]byte("input")),
		Outputs: []Output{NewOutput(addr, coins)},
	}

	// TODO some failures for bad result
	assert.Equal(t, msg.Type(), "bank")
}

func TestIssueMsgValidation(t *testing.T) {
	// TODO
}

func TestIssueMsgString(t *testing.T) {
	addrString := "loan-from-bank"
	bankerString := "input"
	// Construct a IssueMsg
	addr := sdk.Address([]byte(addrString))
	coins := sdk.Coins{{"atom", 10}}
	var msg = IssueMsg{
		Banker:  sdk.Address([]byte(bankerString)),
		Outputs: []Output{NewOutput(addr, coins)},
	}
	res := msg.String()
	expected := fmt.Sprintf("IssueMsg{%X#[Output{%X,10atom}]}", bankerString, addrString)
	assert.Equal(t, expected, res)
}

func TestIssueMsgGet(t *testing.T) {
	addr := sdk.Address([]byte("loan-from-bank"))
	coins := sdk.Coins{{"atom", 10}}
	var msg = IssueMsg{
		Banker:  sdk.Address([]byte("input")),
		Outputs: []Output{NewOutput(addr, coins)},
	}
	res := msg.Get(nil)
	assert.Nil(t, res)
}

func TestIssueMsgGetSignBytes(t *testing.T) {
	addr := sdk.Address([]byte("loan-from-bank"))
	coins := sdk.Coins{{"atom", 10}}
	var msg = IssueMsg{
		Banker:  sdk.Address([]byte("input")),
		Outputs: []Output{NewOutput(addr, coins)},
	}
	res := msg.GetSignBytes()
	// TODO bad results
	assert.Equal(t, string(res), `{"banker":"696E707574","outputs":[{"address":"6C6F616E2D66726F6D2D62616E6B","coins":[{"denom":"atom","amount":10}]}]}`)
}

func TestIssueMsgGetSigners(t *testing.T) {
	var msg = IssueMsg{
		Banker: sdk.Address([]byte("onlyone")),
	}
	res := msg.GetSigners()
	assert.Equal(t, fmt.Sprintf("%v", res), "[6F6E6C796F6E65]")
}
