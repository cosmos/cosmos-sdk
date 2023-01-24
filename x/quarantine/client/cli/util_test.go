package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine"

	. "github.com/cosmos/cosmos-sdk/x/quarantine/testutil"
)

func TestExampleAddress(t *testing.T) {
	// The only thing to really test here is that it's consistent.
	// So just do a few known values (known because the actual values
	// are logged out then the tests are fixed).
	tests := []struct {
		name string
		exp  sdk.AccAddress
	}{
		{
			name: "",
			exp: sdk.AccAddress{
				227, 176, 196, 66, 152, 252, 28, 20, 154, 251,
				244, 200, 153, 111, 185, 36, 39, 174, 65, 228,
			},
		},
		{
			name: "addr",
			exp: sdk.AccAddress{
				162, 89, 60, 152, 108, 134, 202, 50, 58, 51,
				157, 119, 129, 204, 4, 187, 47, 209, 90, 195,
			},
		},
		{
			name: "exampleAddr1",
			exp: sdk.AccAddress{
				199, 131, 86, 61, 89, 233, 25, 212, 30, 112,
				118, 234, 154, 3, 90, 170, 220, 156, 233, 166,
			},
		},
	}

	for _, tc := range tests {
		name := tc.name
		if len(name) == 0 {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			act := exampleAddr(tc.name)
			t.Log(strings.ReplaceAll(fmt.Sprintf("%v", []byte(act)), " ", ", "))
			assert.Equal(t, tc.exp, act, "exampleAddr(%q) result", tc.name)
		})
	}
}

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		argName string
		exp     string
		expErr  []string
	}{
		{
			name:    "exampleAddr1",
			addr:    exampleAddr1.String(),
			argName: "does not matter",
			exp:     exampleAddr1.String(),
		},
		{
			name:    "empty addr",
			addr:    "",
			argName: "this is a field",
			exp:     "",
			expErr:  []string{"invalid this is a field:", "invalid address", "empty address string is not allowed"},
		},
		{
			name:    "bad address",
			addr:    "bad1reallynotgood",
			argName: "something something updog",
			exp:     "",
			expErr:  []string{"invalid something something updog", "invalid address", "decoding bech32 failed"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			var err error
			testFunc := func() {
				act, err = validateAddress(tc.addr, tc.argName)
			}
			require.NotPanics(t, testFunc, "validateAddress")
			AssertErrorContents(t, err, tc.expErr, "validateAddress error")
			assert.Equal(t, tc.exp, act, "validateAddress result")
		})
	}
}

func TestParseAutoResponseUpdatesFromArgs(t *testing.T) {
	newARE := func(from string, response quarantine.AutoResponse) *quarantine.AutoResponseUpdate {
		return &quarantine.AutoResponseUpdate{
			FromAddress: from,
			Response:    response,
		}
	}
	addr0 := MakeTestAddr("parufa", 0).String()
	addr1 := MakeTestAddr("parufa", 1).String()
	addr2 := MakeTestAddr("parufa", 2).String()
	addr3 := MakeTestAddr("parufa", 3).String()
	addr4 := MakeTestAddr("parufa", 4).String()
	addr5 := MakeTestAddr("parufa", 5).String()

	tests := []struct {
		name       string
		args       []string
		startIndex int
		exp        []*quarantine.AutoResponseUpdate
		expErr     []string
	}{
		{
			name:       "start 1 not an auto-response",
			args:       []string{"arg1", addr0},
			startIndex: 1,
			expErr:     []string{"invalid arg 2", "invalid auto-response", fmt.Sprintf("%q", addr0)},
		},
		{
			name:       "start 1 two auto-responses in a row",
			args:       []string{"arg1", "a", "decline", addr0},
			startIndex: 1,
			expErr:     []string{"invalid arg 3", `no from_address args provided after auto-response 1: "a"`},
		},
		{
			name:       "start 1 two auto-responses in a row but not first",
			args:       []string{"arg1", "accept", addr0, "d", "u", addr1},
			startIndex: 1,
			expErr:     []string{"invalid arg 5", `no from_address args provided after auto-response 2: "d"`},
		},
		{
			name:       "start 1 ends with auto-response",
			args:       []string{"arg1", "unspecified", addr0, "accept"},
			startIndex: 1,
			expErr:     []string{"invalid arg 4", `last arg cannot be an auto-response, got: "accept"`},
		},
		{
			name:       "start 1 first address is bad",
			args:       []string{"arg1", "accept", "not1address234567"},
			startIndex: 1,
			expErr: []string{
				`unknown arg 3 "not1address234567"`, `auto-response 1 "accept"`,
				"from_address 1", "invalid address", "decoding bech32 failed",
			},
		},
		{
			name:       "start 1 bad address in the middle",
			args:       []string{"arg1", "a", addr0, addr1, "d", addr2, "bad1notgonnadecode", addr3},
			startIndex: 1,
			expErr: []string{
				`unknown arg 7 "bad1notgonnadecode"`, `auto-response 2 "d"`,
				"from_address 2", "invalid address", "decoding bech32 failed",
			},
		},
		{
			name:       "start 1 auto-response address",
			args:       []string{"arg1", "decline", addr0},
			startIndex: 1,
			exp: []*quarantine.AutoResponseUpdate{
				newARE(addr0, quarantine.AUTO_RESPONSE_DECLINE),
			},
		},
		{
			name:       "start 1 complex",
			args:       []string{"arg1", "a", addr0, addr1, "u", addr2, "d", addr3, addr4, addr5},
			startIndex: 1,
			exp: []*quarantine.AutoResponseUpdate{
				newARE(addr0, quarantine.AUTO_RESPONSE_ACCEPT),
				newARE(addr1, quarantine.AUTO_RESPONSE_ACCEPT),
				newARE(addr2, quarantine.AUTO_RESPONSE_UNSPECIFIED),
				newARE(addr3, quarantine.AUTO_RESPONSE_DECLINE),
				newARE(addr4, quarantine.AUTO_RESPONSE_DECLINE),
				newARE(addr5, quarantine.AUTO_RESPONSE_DECLINE),
			},
		},
		// 3 addr
		// 3 addr
		// 3 ar ar addr
		// 3 ar addr ar ar addr
		// 3 ar addr ar
		// 3 ar bad-addr
		// 3 ar addr addr ar addr bad-addr addr
		// 3 ar addr addr ar addr ar addr addr addr
		{
			name:       "start 3 not an auto-response",
			args:       []string{"arg1", "arg2", "arg3", addr0},
			startIndex: 3,
			expErr:     []string{"invalid arg 4", "invalid auto-response", fmt.Sprintf("%q", addr0)},
		},
		{
			name:       "start 3 two auto-responses in a row",
			args:       []string{"arg1", "arg2", "arg3", "a", "decline", addr0},
			startIndex: 3,
			expErr:     []string{"invalid arg 5", `no from_address args provided after auto-response 1: "a"`},
		},
		{
			name:       "start 3 two auto-responses in a row but not first",
			args:       []string{"arg1", "arg2", "arg3", "accept", addr0, "d", "u", addr1},
			startIndex: 3,
			expErr:     []string{"invalid arg 7", `no from_address args provided after auto-response 2: "d"`},
		},
		{
			name:       "start 3 ends with auto-response",
			args:       []string{"arg1", "arg2", "arg3", "unspecified", addr0, "accept"},
			startIndex: 3,
			expErr:     []string{"invalid arg 6", `last arg cannot be an auto-response, got: "accept"`},
		},
		{
			name:       "start 3 first address is bad",
			args:       []string{"arg1", "arg2", "arg3", "accept", "not1address234567"},
			startIndex: 3,
			expErr: []string{
				`unknown arg 5 "not1address234567"`, `auto-response 1 "accept"`,
				"from_address 1", "invalid address", "decoding bech32 failed",
			},
		},
		{
			name:       "start 3 bad address in the middle",
			args:       []string{"arg1", "arg2", "arg3", "a", addr0, addr1, "d", addr2, "bad1notgonnadecode", addr3},
			startIndex: 3,
			expErr: []string{
				`unknown arg 9 "bad1notgonnadecode"`, `auto-response 2 "d"`,
				"from_address 2", "invalid address", "decoding bech32 failed",
			},
		},
		{
			name:       "start 3 auto-response address",
			args:       []string{"arg1", "arg2", "arg3", "decline", addr0},
			startIndex: 3,
			exp: []*quarantine.AutoResponseUpdate{
				newARE(addr0, quarantine.AUTO_RESPONSE_DECLINE),
			},
		},
		{
			name:       "start 3 complex",
			args:       []string{"arg1", "arg2", "arg3", "a", addr0, addr1, "u", addr2, "d", addr3, addr4, addr5},
			startIndex: 3,
			exp: []*quarantine.AutoResponseUpdate{
				newARE(addr0, quarantine.AUTO_RESPONSE_ACCEPT),
				newARE(addr1, quarantine.AUTO_RESPONSE_ACCEPT),
				newARE(addr2, quarantine.AUTO_RESPONSE_UNSPECIFIED),
				newARE(addr3, quarantine.AUTO_RESPONSE_DECLINE),
				newARE(addr4, quarantine.AUTO_RESPONSE_DECLINE),
				newARE(addr5, quarantine.AUTO_RESPONSE_DECLINE),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act []*quarantine.AutoResponseUpdate
			var err error
			testFunc := func() {
				act, err = ParseAutoResponseUpdatesFromArgs(tc.args, tc.startIndex)
			}
			require.NotPanics(t, testFunc, "ParseAutoResponseUpdatesFromArgs")
			AssertErrorContents(t, err, tc.expErr, "ParseAutoResponseUpdatesFromArgs error")
			assert.Equal(t, tc.exp, act, "ParseAutoResponseUpdatesFromArgs result")
		})
	}
}

func TestParseAutoResponseArg(t *testing.T) {
	tests := []struct {
		arg   string
		expAR quarantine.AutoResponse
		expB  bool
	}{
		{arg: "accept", expAR: quarantine.AUTO_RESPONSE_ACCEPT, expB: true},
		{arg: "ACCEPT", expAR: quarantine.AUTO_RESPONSE_ACCEPT, expB: true},
		{arg: "Accept", expAR: quarantine.AUTO_RESPONSE_ACCEPT, expB: true},
		{arg: "aCcePt", expAR: quarantine.AUTO_RESPONSE_ACCEPT, expB: true},
		{arg: "a", expAR: quarantine.AUTO_RESPONSE_ACCEPT, expB: true},
		{arg: "A", expAR: quarantine.AUTO_RESPONSE_ACCEPT, expB: true},
		{arg: "1", expAR: quarantine.AUTO_RESPONSE_ACCEPT, expB: true},
		{arg: "decline", expAR: quarantine.AUTO_RESPONSE_DECLINE, expB: true},
		{arg: "DECLINE", expAR: quarantine.AUTO_RESPONSE_DECLINE, expB: true},
		{arg: "Decline", expAR: quarantine.AUTO_RESPONSE_DECLINE, expB: true},
		{arg: "dEcliNe", expAR: quarantine.AUTO_RESPONSE_DECLINE, expB: true},
		{arg: "d", expAR: quarantine.AUTO_RESPONSE_DECLINE, expB: true},
		{arg: "D", expAR: quarantine.AUTO_RESPONSE_DECLINE, expB: true},
		{arg: "2", expAR: quarantine.AUTO_RESPONSE_DECLINE, expB: true},
		{arg: "unspecified", expAR: quarantine.AUTO_RESPONSE_UNSPECIFIED, expB: true},
		{arg: "UNSPECIFIED", expAR: quarantine.AUTO_RESPONSE_UNSPECIFIED, expB: true},
		{arg: "Unspecified", expAR: quarantine.AUTO_RESPONSE_UNSPECIFIED, expB: true},
		{arg: "uNspecIfiEd", expAR: quarantine.AUTO_RESPONSE_UNSPECIFIED, expB: true},
		{arg: "u", expAR: quarantine.AUTO_RESPONSE_UNSPECIFIED, expB: true},
		{arg: "U", expAR: quarantine.AUTO_RESPONSE_UNSPECIFIED, expB: true},
		{arg: "off", expAR: quarantine.AUTO_RESPONSE_UNSPECIFIED, expB: true},
		{arg: "OFF", expAR: quarantine.AUTO_RESPONSE_UNSPECIFIED, expB: true},
		{arg: "Off", expAR: quarantine.AUTO_RESPONSE_UNSPECIFIED, expB: true},
		{arg: "oFf", expAR: quarantine.AUTO_RESPONSE_UNSPECIFIED, expB: true},
		{arg: "ofF", expAR: quarantine.AUTO_RESPONSE_UNSPECIFIED, expB: true},
		{arg: "o", expAR: quarantine.AUTO_RESPONSE_UNSPECIFIED, expB: true},
		{arg: "O", expAR: quarantine.AUTO_RESPONSE_UNSPECIFIED, expB: true},
		{arg: "0", expAR: quarantine.AUTO_RESPONSE_UNSPECIFIED, expB: true},
		{arg: "", expAR: quarantine.AUTO_RESPONSE_UNSPECIFIED, expB: false},
		{arg: "accepta", expAR: quarantine.AUTO_RESPONSE_UNSPECIFIED, expB: false},
		{arg: "declined", expAR: quarantine.AUTO_RESPONSE_UNSPECIFIED, expB: false},
		{arg: "uoff", expAR: quarantine.AUTO_RESPONSE_UNSPECIFIED, expB: false},
		{arg: "something else", expAR: quarantine.AUTO_RESPONSE_UNSPECIFIED, expB: false},
	}

	for _, tc := range tests {
		name := tc.arg
		if len(name) == 0 {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			actAr, actB := ParseAutoResponseArg(tc.arg)
			assert.Equal(t, tc.expAR, actAr, "ParseAutoResponseArg response")
			assert.Equal(t, tc.expB, actB, "ParseAutoResponseArg bool")
		})
	}
}
