package sanction_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/testutil"
)

func TestNewMsgSanction(t *testing.T) {
	tests := []struct {
		name      string
		authority string
		addrs     []sdk.AccAddress
		exp       *sanction.MsgSanction
	}{
		{
			name:      "empty nil",
			authority: "",
			addrs:     nil,
			exp: &sanction.MsgSanction{
				Addresses: nil,
				Authority: "",
			},
		},
		{
			name:      "empty empty",
			authority: "",
			addrs:     []sdk.AccAddress{},
			exp: &sanction.MsgSanction{
				Addresses: nil,
				Authority: "",
			},
		},
		{
			name:      "just authority provided",
			authority: "cartman",
			addrs:     nil,
			exp: &sanction.MsgSanction{
				Addresses: nil,
				Authority: "cartman",
			},
		},
		{
			name:      "three addresses provided",
			authority: "authority",
			addrs: []sdk.AccAddress{
				sdk.AccAddress("testaddr0___________"),
				sdk.AccAddress("testaddr1___________"),
				sdk.AccAddress("testaddr2___________"),
			},
			exp: &sanction.MsgSanction{
				Addresses: []string{
					sdk.AccAddress("testaddr0___________").String(),
					sdk.AccAddress("testaddr1___________").String(),
					sdk.AccAddress("testaddr2___________").String(),
				},
				Authority: "authority",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var msg *sanction.MsgSanction
			testFunc := func() {
				msg = sanction.NewMsgSanction(tc.authority, tc.addrs...)
			}
			require.NotPanics(t, testFunc, "NewMsgSanction")
			assert.Equal(t, tc.exp, msg, "NewMsgSanction result")
		})
	}
}

func TestMsgSanction_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  *sanction.MsgSanction
		exp  []string
	}{
		{
			name: "control",
			msg: &sanction.MsgSanction{
				Addresses: []string{
					sdk.AccAddress("addr0_______________").String(),
					sdk.AccAddress("addr1_______________").String(),
					sdk.AccAddress("addr2_______________").String(),
					sdk.AccAddress("addr3_______________").String(),
					sdk.AccAddress("addr4_______________").String(),
				},
				Authority: sdk.AccAddress("authority___________").String(),
			},
			exp: nil,
		},
		{
			name: "empty authority",
			msg: &sanction.MsgSanction{
				Addresses: []string{
					sdk.AccAddress("addr0_______________").String(),
					sdk.AccAddress("addr1_______________").String(),
					sdk.AccAddress("addr2_______________").String(),
					sdk.AccAddress("addr3_______________").String(),
					sdk.AccAddress("addr4_______________").String(),
				},
				Authority: "",
			},
			exp: []string{"invalid address", "authority", `""`, "empty address string is not allowed"},
		},
		{
			name: "bad authority",
			msg: &sanction.MsgSanction{
				Addresses: []string{
					sdk.AccAddress("addr0_______________").String(),
					sdk.AccAddress("addr1_______________").String(),
					sdk.AccAddress("addr2_______________").String(),
					sdk.AccAddress("addr3_______________").String(),
					sdk.AccAddress("addr4_______________").String(),
				},
				Authority: "bad1authority",
			},
			exp: []string{"invalid address", "authority", `"bad1authority"`, "decoding bech32 failed"},
		},
		{
			name: "bad first addr",
			msg: &sanction.MsgSanction{
				Addresses: []string{
					"bad1firstaddr",
					sdk.AccAddress("addr1_______________").String(),
					sdk.AccAddress("addr2_______________").String(),
					sdk.AccAddress("addr3_______________").String(),
					sdk.AccAddress("addr4_______________").String(),
				},
				Authority: sdk.AccAddress("authority___________").String(),
			},
			exp: []string{"invalid address", "addresses[0]", `"bad1firstaddr"`, "decoding bech32 failed"},
		},
		{
			name: "bad third addr",
			msg: &sanction.MsgSanction{
				Addresses: []string{
					sdk.AccAddress("addr0_______________").String(),
					sdk.AccAddress("addr1_______________").String(),
					"bad1thirdaddr",
					sdk.AccAddress("addr3_______________").String(),
					sdk.AccAddress("addr4_______________").String(),
				},
				Authority: sdk.AccAddress("authority___________").String(),
			},
			exp: []string{"invalid address", "addresses[2]", `"bad1thirdaddr"`, "decoding bech32 failed"},
		},
		{
			name: "bad first addr",
			msg: &sanction.MsgSanction{
				Addresses: []string{
					sdk.AccAddress("addr0_______________").String(),
					sdk.AccAddress("addr1_______________").String(),
					sdk.AccAddress("addr2_______________").String(),
					sdk.AccAddress("addr3_______________").String(),
					"bad1fifthaddr",
				},
				Authority: sdk.AccAddress("authority___________").String(),
			},
			exp: []string{"invalid address", "addresses[4]", `"bad1fifthaddr"`, "decoding bech32 failed"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.msg.ValidateBasic()
			}
			require.NotPanics(t, testFunc, "ValidateBasic")
			testutil.AssertErrorContents(t, err, tc.exp, "ValidateBasic result.")
		})
	}
}

func TestMsgSanction_GetSigners(t *testing.T) {
	tests := []struct {
		name  string
		msg   *sanction.MsgSanction
		exp   []sdk.AccAddress
		panic string
	}{
		{
			name:  "empty authority",
			msg:   &sanction.MsgSanction{Authority: ""},
			panic: "empty address string is not allowed",
		},
		{
			name:  "bad authority",
			msg:   &sanction.MsgSanction{Authority: "nope1notauthority"},
			panic: "decoding bech32 failed: invalid character not part of charset: 111",
		},
		{
			name: "good authority",
			msg:  &sanction.MsgSanction{Authority: sdk.AccAddress("testauthority_______").String()},
			exp:  []sdk.AccAddress{sdk.AccAddress("testauthority_______")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []sdk.AccAddress
			testFunc := func() {
				actual = tc.msg.GetSigners()
			}
			if len(tc.panic) > 0 {
				require.PanicsWithError(t, tc.panic, testFunc, "GetSigners()")
			} else {
				require.NotPanics(t, testFunc, "GetSigners()")
				assert.Equal(t, tc.exp, actual, "GetSigners result")
			}
		})
	}
}

func TestNewMsgUnsanction(t *testing.T) {
	tests := []struct {
		name      string
		authority string
		addrs     []sdk.AccAddress
		exp       *sanction.MsgUnsanction
	}{
		{
			name:      "empty nil",
			authority: "",
			addrs:     nil,
			exp: &sanction.MsgUnsanction{
				Addresses: nil,
				Authority: "",
			},
		},
		{
			name:      "empty empty",
			authority: "",
			addrs:     []sdk.AccAddress{},
			exp: &sanction.MsgUnsanction{
				Addresses: nil,
				Authority: "",
			},
		},
		{
			name:      "just authority provided",
			authority: "cartman",
			addrs:     nil,
			exp: &sanction.MsgUnsanction{
				Addresses: nil,
				Authority: "cartman",
			},
		},
		{
			name:      "three addresses provided",
			authority: "authority",
			addrs: []sdk.AccAddress{
				sdk.AccAddress("testaddr0___________"),
				sdk.AccAddress("testaddr1___________"),
				sdk.AccAddress("testaddr2___________"),
			},
			exp: &sanction.MsgUnsanction{
				Addresses: []string{
					sdk.AccAddress("testaddr0___________").String(),
					sdk.AccAddress("testaddr1___________").String(),
					sdk.AccAddress("testaddr2___________").String(),
				},
				Authority: "authority",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var msg *sanction.MsgUnsanction
			testFunc := func() {
				msg = sanction.NewMsgUnsanction(tc.authority, tc.addrs...)
			}
			require.NotPanics(t, testFunc, "NewMsgUnsanction")
			assert.Equal(t, tc.exp, msg, "NewMsgUnsanction result")
		})
	}
}

func TestMsgUnsanction_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  *sanction.MsgUnsanction
		exp  []string
	}{
		{
			name: "control",
			msg: &sanction.MsgUnsanction{
				Addresses: []string{
					sdk.AccAddress("addr0_______________").String(),
					sdk.AccAddress("addr1_______________").String(),
					sdk.AccAddress("addr2_______________").String(),
					sdk.AccAddress("addr3_______________").String(),
					sdk.AccAddress("addr4_______________").String(),
				},
				Authority: sdk.AccAddress("authority___________").String(),
			},
			exp: nil,
		},
		{
			name: "empty authority",
			msg: &sanction.MsgUnsanction{
				Addresses: []string{
					sdk.AccAddress("addr0_______________").String(),
					sdk.AccAddress("addr1_______________").String(),
					sdk.AccAddress("addr2_______________").String(),
					sdk.AccAddress("addr3_______________").String(),
					sdk.AccAddress("addr4_______________").String(),
				},
				Authority: "",
			},
			exp: []string{"invalid address", "authority", `""`, "empty address string is not allowed"},
		},
		{
			name: "bad authority",
			msg: &sanction.MsgUnsanction{
				Addresses: []string{
					sdk.AccAddress("addr0_______________").String(),
					sdk.AccAddress("addr1_______________").String(),
					sdk.AccAddress("addr2_______________").String(),
					sdk.AccAddress("addr3_______________").String(),
					sdk.AccAddress("addr4_______________").String(),
				},
				Authority: "bad1authority",
			},
			exp: []string{"invalid address", "authority", `"bad1authority"`, "decoding bech32 failed"},
		},
		{
			name: "bad first addr",
			msg: &sanction.MsgUnsanction{
				Addresses: []string{
					"bad1firstaddr",
					sdk.AccAddress("addr1_______________").String(),
					sdk.AccAddress("addr2_______________").String(),
					sdk.AccAddress("addr3_______________").String(),
					sdk.AccAddress("addr4_______________").String(),
				},
				Authority: sdk.AccAddress("authority___________").String(),
			},
			exp: []string{"invalid address", "addresses[0]", `"bad1firstaddr"`, "decoding bech32 failed"},
		},
		{
			name: "bad third addr",
			msg: &sanction.MsgUnsanction{
				Addresses: []string{
					sdk.AccAddress("addr0_______________").String(),
					sdk.AccAddress("addr1_______________").String(),
					"bad1thirdaddr",
					sdk.AccAddress("addr3_______________").String(),
					sdk.AccAddress("addr4_______________").String(),
				},
				Authority: sdk.AccAddress("authority___________").String(),
			},
			exp: []string{"invalid address", "addresses[2]", `"bad1thirdaddr"`, "decoding bech32 failed"},
		},
		{
			name: "bad first addr",
			msg: &sanction.MsgUnsanction{
				Addresses: []string{
					sdk.AccAddress("addr0_______________").String(),
					sdk.AccAddress("addr1_______________").String(),
					sdk.AccAddress("addr2_______________").String(),
					sdk.AccAddress("addr3_______________").String(),
					"bad1fifthaddr",
				},
				Authority: sdk.AccAddress("authority___________").String(),
			},
			exp: []string{"invalid address", "addresses[4]", `"bad1fifthaddr"`, "decoding bech32 failed"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.msg.ValidateBasic()
			}
			require.NotPanics(t, testFunc, "ValidateBasic")
			testutil.AssertErrorContents(t, err, tc.exp, "ValidateBasic result.")
		})
	}
}

func TestMsgUnsanction_GetSigners(t *testing.T) {
	tests := []struct {
		name  string
		msg   *sanction.MsgUnsanction
		exp   []sdk.AccAddress
		panic string
	}{
		{
			name:  "empty authority",
			msg:   &sanction.MsgUnsanction{Authority: ""},
			panic: "empty address string is not allowed",
		},
		{
			name:  "bad authority",
			msg:   &sanction.MsgUnsanction{Authority: "nope1notauthority"},
			panic: "decoding bech32 failed: invalid character not part of charset: 111",
		},
		{
			name: "good authority",
			msg:  &sanction.MsgUnsanction{Authority: sdk.AccAddress("testauthority_______").String()},
			exp:  []sdk.AccAddress{sdk.AccAddress("testauthority_______")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []sdk.AccAddress
			testFunc := func() {
				actual = tc.msg.GetSigners()
			}
			if len(tc.panic) > 0 {
				require.PanicsWithError(t, tc.panic, testFunc, "GetSigners()")
			} else {
				require.NotPanics(t, testFunc, "GetSigners()")
				assert.Equal(t, tc.exp, actual, "GetSigners result")
			}
		})
	}
}

func TestNewMsgUpdateParams(t *testing.T) {
	// cz is a short way of defining sdk.Coins for these tests.
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name             string
		authority        string
		minDepSanction   sdk.Coins
		minDepUnsanction sdk.Coins
		expected         *sanction.MsgUpdateParams
	}{
		{
			name:             "control",
			authority:        "auth-str",
			minDepSanction:   cz("1acoin,2bcoin"),
			minDepUnsanction: cz("5acoin,3bcoin,1ccoin"),
			expected: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   cz("1acoin,2bcoin"),
					ImmediateUnsanctionMinDeposit: cz("5acoin,3bcoin,1ccoin"),
				},
				Authority: "auth-str",
			},
		},
		{
			name:             "empty authority",
			authority:        "",
			minDepSanction:   cz("1acoin,2bcoin"),
			minDepUnsanction: cz("5acoin,3bcoin,1ccoin"),
			expected: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   cz("1acoin,2bcoin"),
					ImmediateUnsanctionMinDeposit: cz("5acoin,3bcoin,1ccoin"),
				},
				Authority: "",
			},
		},
		{
			name:             "nil minDepSanction",
			authority:        "auth-str",
			minDepSanction:   nil,
			minDepUnsanction: cz("5acoin,3bcoin,1ccoin"),
			expected: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   nil,
					ImmediateUnsanctionMinDeposit: cz("5acoin,3bcoin,1ccoin"),
				},
				Authority: "auth-str",
			},
		},
		{
			name:             "empty minDepSanction",
			authority:        "auth-str",
			minDepSanction:   sdk.Coins{},
			minDepUnsanction: cz("5acoin,3bcoin,1ccoin"),
			expected: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.Coins{},
					ImmediateUnsanctionMinDeposit: cz("5acoin,3bcoin,1ccoin"),
				},
				Authority: "auth-str",
			},
		},
		{
			name:             "nil minDepUnsanction",
			authority:        "auth-str",
			minDepSanction:   cz("1acoin,2bcoin"),
			minDepUnsanction: nil,
			expected: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   cz("1acoin,2bcoin"),
					ImmediateUnsanctionMinDeposit: nil,
				},
				Authority: "auth-str",
			},
		},
		{
			name:             "empty minDepUnsanction",
			authority:        "auth-str",
			minDepSanction:   cz("1acoin,2bcoin"),
			minDepUnsanction: sdk.Coins{},
			expected: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   cz("1acoin,2bcoin"),
					ImmediateUnsanctionMinDeposit: sdk.Coins{},
				},
				Authority: "auth-str",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual *sanction.MsgUpdateParams
			testFunc := func() {
				actual = sanction.NewMsgUpdateParams(tc.authority, tc.minDepSanction, tc.minDepUnsanction)
			}
			require.NotPanics(t, testFunc, "NewMsgUpdateParams")
			if !assert.Equal(t, tc.expected, actual, "NewMsgUpdateParams result") && actual != nil {
				// The test already failed, assert each field just to make it easier to find the problem.
				assert.Equal(t, tc.expected.Authority, actual.Authority, "Authority")
				if assert.NotNil(t, actual.Params, "Params") {
					assert.Equal(t, tc.expected.Params.ImmediateSanctionMinDeposit.String(),
						actual.Params.ImmediateSanctionMinDeposit.String(), "ImmediateSanctionMinDeposit")
					assert.Equal(t, tc.expected.Params.ImmediateUnsanctionMinDeposit.String(),
						actual.Params.ImmediateUnsanctionMinDeposit.String(), "ImmediateUnsanctionMinDeposit")
				}
			}
		})
	}
}

func TestMsgUpdateParams_ValidateBasic(t *testing.T) {
	// cz is a short way of defining sdk.Coins for these tests.
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name string
		msg  *sanction.MsgUpdateParams
		exp  []string
	}{
		{
			name: "control",
			msg: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   cz("5dolla"),
					ImmediateUnsanctionMinDeposit: cz("50cent"),
				},
				Authority: sdk.AccAddress("authority_test_addr_").String(),
			},
			exp: nil,
		},
		{
			name: "empty authority",
			msg: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   cz("5dolla"),
					ImmediateUnsanctionMinDeposit: cz("50cent"),
				},
				Authority: "",
			},
			exp: []string{"invalid address", "authority", `""`, "empty address string is not allowed"},
		},
		{
			name: "bad authority",
			msg: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   cz("5dolla"),
					ImmediateUnsanctionMinDeposit: cz("50cent"),
				},
				Authority: "auth1notvalidaddr",
			},
			exp: []string{"invalid address", "authority", `"auth1notvalidaddr"`, "decoding bech32 failed"},
		},
		{
			name: "bad params",
			msg: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.Coins{sdk.NewInt64Coin("dcoin", 1), sdk.NewInt64Coin("dcoin", 2)},
					ImmediateUnsanctionMinDeposit: cz("50cent"),
				},
				Authority: sdk.AccAddress("authority_test_addr_").String(),
			},
			exp: []string{"invalid params", "invalid immediate sanction min deposit", "duplicate denomination dcoin"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.msg.ValidateBasic()
			}
			require.NotPanics(t, testFunc, "ValidateBasic")
			testutil.AssertErrorContents(t, err, tc.exp, "ValidateBasic result.")
		})
	}
}

func TestMsgUpdateParams_GetSigners(t *testing.T) {
	tests := []struct {
		name  string
		msg   *sanction.MsgUpdateParams
		exp   []sdk.AccAddress
		panic string
	}{
		{
			name:  "empty authority",
			msg:   &sanction.MsgUpdateParams{Authority: ""},
			panic: "empty address string is not allowed",
		},
		{
			name:  "bad authority",
			msg:   &sanction.MsgUpdateParams{Authority: "nope1notauthority"},
			panic: "decoding bech32 failed: invalid character not part of charset: 111",
		},
		{
			name: "good authority",
			msg:  &sanction.MsgUpdateParams{Authority: sdk.AccAddress("testauthority_______").String()},
			exp:  []sdk.AccAddress{sdk.AccAddress("testauthority_______")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []sdk.AccAddress
			testFunc := func() {
				actual = tc.msg.GetSigners()
			}
			if len(tc.panic) > 0 {
				require.PanicsWithError(t, tc.panic, testFunc, "GetSigners()")
			} else {
				require.NotPanics(t, testFunc, "GetSigners()")
				assert.Equal(t, tc.exp, actual, "GetSigners result")
			}
		})
	}
}
