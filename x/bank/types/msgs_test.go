package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgSendRoute(t *testing.T) {
	addr1 := sdk.AccAddress([]byte("from"))
	addr2 := sdk.AccAddress([]byte("to"))
	coins := sdk.NewCoins(sdk.NewInt64Coin("atom", 10))
	msg := NewMsgSend(addr1, addr2, coins)

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
		{"invalid from address: empty address string is not allowed: invalid address", NewMsgSend(addrEmpty, addr2, atom123)},
		{"invalid to address: empty address string is not allowed: invalid address", NewMsgSend(addr1, addrEmpty, atom123)},
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
	msg := NewMsgSend(addr1, addr2, coins)
	res := msg.GetSignBytes()

	expected := `{"type":"cosmos-sdk/MsgSend","value":{"amount":[{"amount":"10","denom":"atom"}],"from_address":"cosmos1d9h8qat57ljhcm","to_address":"cosmos1da6hgur4wsmpnjyg"}}`
	require.Equal(t, expected, string(res))
}

func TestMsgMultiSendRoute(t *testing.T) {
	// Construct a MsgSend
	addr1 := sdk.AccAddress([]byte("input"))
	addr2 := sdk.AccAddress([]byte("output"))
	coins := sdk.NewCoins(sdk.NewInt64Coin("atom", 10))
	msg := MsgMultiSend{
		Inputs:  []Input{NewInput(addr1, coins)},
		Outputs: []Output{NewOutput(addr2, coins)},
	}

	// TODO some failures for bad result
	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "multisend")
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

func TestMsgMultiSendValidation(t *testing.T) {
	addr1 := sdk.AccAddress([]byte("_______alice________"))
	addr2 := sdk.AccAddress([]byte("________bob_________"))
	atom123 := sdk.NewCoins(sdk.NewInt64Coin("atom", 123))
	atom124 := sdk.NewCoins(sdk.NewInt64Coin("atom", 124))
	eth123 := sdk.NewCoins(sdk.NewInt64Coin("eth", 123))
	atom123eth123 := sdk.NewCoins(sdk.NewInt64Coin("atom", 123), sdk.NewInt64Coin("eth", 123))

	input1 := NewInput(addr1, atom123)
	input2 := NewInput(addr1, eth123)
	inputMulti := NewInput(addr1, atom123.Add(atom124...))
	output1 := NewOutput(addr2, atom123)
	output2 := NewOutput(addr2, atom124)
	outputMulti := NewOutput(addr2, atom123eth123)

	var emptyAddr sdk.AccAddress

	cases := []struct {
		valid bool
		tx    MsgMultiSend
	}{
		{false, MsgMultiSend{}},                           // no input or output
		{false, MsgMultiSend{Inputs: []Input{input1}}},    // just input
		{false, MsgMultiSend{Outputs: []Output{output1}}}, // just output
		{
			false,
			MsgMultiSend{
				Inputs:  []Input{NewInput(emptyAddr, atom123)}, // invalid input
				Outputs: []Output{output1},
			},
		},
		{
			false,
			MsgMultiSend{
				Inputs:  []Input{input1},
				Outputs: []Output{{emptyAddr.String(), atom123}}, // invalid output
			},
		},
		{
			false,
			MsgMultiSend{
				Inputs:  []Input{input1},
				Outputs: []Output{output2}, // amounts dont match
			},
		},
		{
			true,
			MsgMultiSend{
				Inputs:  []Input{input1},
				Outputs: []Output{output1}, // one-to-one
			},
		},
		{
			true,
			MsgMultiSend{
				Inputs:  []Input{input1, input2},
				Outputs: []Output{outputMulti}, // many-to-one
			},
		},
		{
			true,
			MsgMultiSend{
				Inputs:  []Input{inputMulti},
				Outputs: []Output{output1, output2}, // one-to-many
			},
		},
		{
			false,
			MsgMultiSend{
				Inputs:  []Input{input1, input2},
				Outputs: []Output{output1, output2}, // many-to-many not allowed
			},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			err := tc.tx.ValidateBasic()
			if tc.valid {
				require.NoError(t, err, "%d: %+v", i, err)
			} else {
				require.Error(t, err, "%d", i)
			}
		})
	}
}

func TestValidateInputsOutputs(t *testing.T) {
	addr1 := sdk.AccAddress("_______alice________")
	addr2 := sdk.AccAddress("________bob_________")
	addr3 := sdk.AccAddress("_______carol________")
	addr4 := sdk.AccAddress("_______dave_________")

	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name    string
		inputs  []Input
		outputs []Output
		expErr  string
	}{
		{
			name:    "no inputs",
			inputs:  nil,
			outputs: []Output{NewOutput(addr1, cz("123atom"))},
			expErr:  ErrNoInputs.Error(),
		},
		{
			name:    "no outputs",
			inputs:  []Input{NewInput(addr1, cz("123atom"))},
			outputs: nil,
			expErr:  ErrNoOutputs.Error(),
		},
		{
			name:    "many to many",
			inputs:  []Input{NewInput(addr1, cz("123atom")), NewInput(addr2, cz("456eth"))},
			outputs: []Output{NewOutput(addr2, cz("123atom")), NewOutput(addr1, cz("456eth"))},
			expErr:  ErrManyToMany.Error(),
		},
		{
			name:    "one input invalid",
			inputs:  []Input{{Address: "nope", Coins: cz("123atom")}},
			outputs: []Output{NewOutput(addr2, cz("123atom"))},
			expErr:  "invalid input address: decoding bech32 failed: invalid bech32 string length 4: invalid address",
		},
		{
			name: "three inputs: first invalid",
			inputs: []Input{
				{Address: "nope", Coins: cz("100atom")},
				NewInput(addr1, cz("20atom")),
				NewInput(addr2, cz("3atom")),
			},
			outputs: []Output{NewOutput(addr3, cz("123atom"))},
			expErr:  "invalid input address: decoding bech32 failed: invalid bech32 string length 4: invalid address",
		},
		{
			name: "three inputs: last invalid",
			inputs: []Input{
				NewInput(addr1, cz("100atom")),
				NewInput(addr2, cz("20atom")),
				{Address: "nope", Coins: cz("3atom")},
			},
			outputs: []Output{NewOutput(addr3, cz("123atom"))},
			expErr:  "invalid input address: decoding bech32 failed: invalid bech32 string length 4: invalid address",
		},
		{
			name: "negative coins in input",
			inputs: []Input{
				NewInput(addr1, cz("124atom")),
				NewInput(addr1, sdk.Coins{sdk.Coin{Denom: "atom", Amount: sdkmath.NewInt(-1)}}),
			},
			outputs: []Output{NewOutput(addr2, cz("123atom"))},
			expErr:  "-1atom: invalid coins",
		},
		{
			name:    "one output invalid",
			inputs:  []Input{NewInput(addr1, cz("123atom"))},
			outputs: []Output{{Address: "nope", Coins: cz("123atom")}},
			expErr:  "invalid output address: decoding bech32 failed: invalid bech32 string length 4: invalid address",
		},
		{
			name:   "three outputs: first invalid",
			inputs: []Input{NewInput(addr1, cz("123atom"))},
			outputs: []Output{
				{Address: "nope", Coins: cz("100atom")},
				NewOutput(addr2, cz("20atom")),
				NewOutput(addr3, cz("3atom")),
			},
			expErr: "invalid output address: decoding bech32 failed: invalid bech32 string length 4: invalid address",
		},
		{
			name:   "three outputs: last invalid",
			inputs: []Input{NewInput(addr1, cz("123atom"))},
			outputs: []Output{
				NewOutput(addr2, cz("100atom")),
				NewOutput(addr3, cz("20atom")),
				{Address: "nope", Coins: cz("3atom")},
			},
			expErr: "invalid output address: decoding bech32 failed: invalid bech32 string length 4: invalid address",
		},
		{
			name:   "negative coins in output",
			inputs: []Input{NewInput(addr1, cz("123atom"))},
			outputs: []Output{
				NewOutput(addr2, cz("124atom")),
				NewOutput(addr3, sdk.Coins{sdk.Coin{Denom: "atom", Amount: sdkmath.NewInt(-1)}}),
			},
			expErr: "-1atom: invalid coins",
		},
		{
			name:    "amount mismatch one denom too much input",
			inputs:  []Input{NewInput(addr1, cz("124atom"))},
			outputs: []Output{NewOutput(addr2, cz("123atom"))},
			expErr:  ErrInputOutputMismatch.Error(),
		},
		{
			name:    "amount mismatch one denom too much output",
			inputs:  []Input{NewInput(addr1, cz("123atom"))},
			outputs: []Output{NewOutput(addr2, cz("124atom"))},
			expErr:  ErrInputOutputMismatch.Error(),
		},
		{
			name:    "amount mismatch different denoms",
			inputs:  []Input{NewInput(addr1, cz("123atom"))},
			outputs: []Output{NewOutput(addr2, cz("123eth"))},
			expErr:  ErrInputOutputMismatch.Error(),
		},
		{
			name:    "amount mismatch input has extra denom",
			inputs:  []Input{NewInput(addr1, cz("123atom,123eth"))},
			outputs: []Output{NewOutput(addr2, cz("123atom"))},
			expErr:  ErrInputOutputMismatch.Error(),
		},
		{
			name:    "amount mismatch output has extra denom",
			inputs:  []Input{NewInput(addr1, cz("123atom"))},
			outputs: []Output{NewOutput(addr2, cz("123atom,123eth"))},
			expErr:  ErrInputOutputMismatch.Error(),
		},
		{
			name:    "amount mismatch first denom okay second not",
			inputs:  []Input{NewInput(addr1, cz("123atom,124eth"))},
			outputs: []Output{NewOutput(addr2, cz("123atom,123eth"))},
			expErr:  ErrInputOutputMismatch.Error(),
		},
		{
			name:    "amount mismatch second denom okay first not",
			inputs:  []Input{NewInput(addr1, cz("124atom,123eth"))},
			outputs: []Output{NewOutput(addr2, cz("123atom,123eth"))},
			expErr:  ErrInputOutputMismatch.Error(),
		},
		{
			name:    "amount mismatch two denoms in each but different denoms",
			inputs:  []Input{NewInput(addr1, cz("123atom")), NewInput(addr2, cz("321eth"))},
			outputs: []Output{NewOutput(addr3, cz("123atom,321esh"))},
			expErr:  ErrInputOutputMismatch.Error(),
		},
		{
			name:   "one to three amount mismatch",
			inputs: []Input{NewInput(addr1, cz("124atom,123eth"))},
			outputs: []Output{
				NewOutput(addr2, cz("100atom,3eth")),
				NewOutput(addr3, cz("20atom,20eth")),
				NewOutput(addr4, cz("3atom,100eth")),
			},
			expErr: ErrInputOutputMismatch.Error(),
		},
		{
			name: "three to one amount mismatch",
			inputs: []Input{
				NewInput(addr1, cz("100atom,3eth")),
				NewInput(addr2, cz("20atom,20eth")),
				NewInput(addr3, cz("4atom,100eth")),
			},
			outputs: []Output{NewOutput(addr4, cz("123atom,123eth"))},
			expErr:  ErrInputOutputMismatch.Error(),
		},
		{
			name:    "one to one okay",
			inputs:  []Input{NewInput(addr1, cz("123atom"))},
			outputs: []Output{NewOutput(addr1, cz("123atom"))},
		},
		{
			name:   "one to many okay",
			inputs: []Input{NewInput(addr1, cz("123atom"))},
			outputs: []Output{
				NewOutput(addr2, cz("100atom")),
				NewOutput(addr3, cz("20atom")),
				NewOutput(addr4, cz("3atom")),
			},
		},
		{
			name: "many to one okay",
			inputs: []Input{
				NewInput(addr1, cz("100atom")),
				NewInput(addr2, cz("20atom")),
				NewInput(addr3, cz("3atom")),
			},
			outputs: []Output{NewOutput(addr4, cz("123atom"))},
		},
	}
	//
	// amounts mismatch

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateInputsOutputs(tc.inputs, tc.outputs)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "ValidateInputsOutputs error")
			} else {
				assert.NoError(t, err, "ValidateInputsOutputs error")
			}
		})
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
	addrs := make([]string, 3)
	inputs := make([]Input, 3)
	for i, v := range []string{"input111111111111111", "input222222222222222", "input333333333333333"} {
		addr := sdk.AccAddress([]byte(v))
		inputs[i] = NewInput(addr, nil)
		addrs[i] = addr.String()
	}
	msg := NewMsgMultiSend(inputs, nil)

	res := msg.GetSigners()
	for i, signer := range res {
		require.Equal(t, signer.String(), addrs[i])
	}
}

func TestMsgSendGetSigners(t *testing.T) {
	from := sdk.AccAddress([]byte("input111111111111111"))
	msg := NewMsgSend(from, sdk.AccAddress{}, sdk.NewCoins())
	res := msg.GetSigners()
	require.Equal(t, 1, len(res))
	require.True(t, from.Equals(res[0]))
}

func TestUpdateDenomMetadataGetSignBytes(t *testing.T) {
	//from := sdk.AccAddress("input")
	//coins := sdk.NewCoins(sdk.NewInt64Coin("atom", 10))
	msg := MsgUpdateDenomMetadata{
		Title:       "title",
		Description: "description",
		Metadata: &Metadata{
			Name:        "diamondback",
			Symbol:      "DB",
			Description: "The native staking token",
			DenomUnits: []*DenomUnit{
				{"udiamondback", uint32(0), []string{"microdiamondback"}},
			},
		},
	}
	res := msg.GetSignBytes()

	expected := `{"type":"cosmos-sdk/MsgUpdateDenomMetadata","value":{"description":"description","metadata":{"denom_units":[{"aliases":["microdiamondback"],"denom":"udiamondback"}],"description":"The native staking token","name":"diamondback","symbol":"DB"},"title":"title"}}`
	require.Equal(t, expected, string(res))
}

func TestUpdateDenomMetadataGetSigners(t *testing.T) {
	from := sdk.AccAddress("cosmos1d9h8qat57ljhcm")
	title := "Proposal Title"
	description := "Proposal description"
	metadata := Metadata{}
	msg := NewMsgUpdateDenomMetadata(from.String(), title, description, &metadata)
	res := msg.GetSigners()
	require.Equal(t, 1, len(res))
	require.True(t, from.Equals(res[0]))
}
