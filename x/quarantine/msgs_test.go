package quarantine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	. "github.com/cosmos/cosmos-sdk/x/quarantine"
	. "github.com/cosmos/cosmos-sdk/x/quarantine/testutil"
)

func TestNewMsgOptIn(t *testing.T) {
	testAddr0 := MakeTestAddr("nmoi", 0)
	testAddr1 := MakeTestAddr("nmoi", 1)

	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		expected *MsgOptIn
	}{
		{
			name:     "addr 0",
			toAddr:   testAddr0,
			expected: &MsgOptIn{ToAddress: testAddr0.String()},
		},
		{
			name:     "addr 1",
			toAddr:   testAddr1,
			expected: &MsgOptIn{ToAddress: testAddr1.String()},
		},
		{
			name:     "nil",
			toAddr:   nil,
			expected: &MsgOptIn{ToAddress: ""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewMsgOptIn(tc.toAddr)
			assert.Equal(t, tc.expected, actual, "NewMsgOptIn")
		})
	}
}

func TestMsgOptIn_ValidateBasic(t *testing.T) {
	addr := MakeTestAddr("moivb", 0).String()

	tests := []struct {
		name          string
		addr          string
		expectedInErr []string
	}{
		{
			name:          "addr",
			addr:          addr,
			expectedInErr: nil,
		},
		{
			name:          "bad",
			addr:          "not an actual address",
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "empty",
			addr:          "",
			expectedInErr: []string{"invalid to address"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := MsgOptIn{ToAddress: tc.addr}
			msg := MsgOptIn{ToAddress: tc.addr}
			err := msg.ValidateBasic()
			AssertErrorContents(t, err, tc.expectedInErr, "ValidateBasic")
			assert.Equal(t, msgOrig, msg, "MsgOptIn before and after")
		})
	}
}

func TestMsgOptIn_GetSigners(t *testing.T) {
	addr := MakeTestAddr("moigs", 0)

	tests := []struct {
		name     string
		addr     string
		expected []sdk.AccAddress
	}{
		{
			name:     "addr",
			addr:     addr.String(),
			expected: []sdk.AccAddress{addr},
		},
		{
			name:     "bad",
			addr:     "not an actual address",
			expected: []sdk.AccAddress{nil},
		},
		{
			name:     "empty",
			addr:     "",
			expected: []sdk.AccAddress{{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := MsgOptIn{ToAddress: tc.addr}
			msg := MsgOptIn{ToAddress: tc.addr}
			actual := msg.GetSigners()
			assert.Equal(t, tc.expected, actual, "GetSigners")
			assert.Equal(t, msgOrig, msg, "MsgOptIn before and after")
		})
	}
}

func TestNewMsgOptOut(t *testing.T) {
	testAddr0 := MakeTestAddr("nmoo", 0)
	testAddr1 := MakeTestAddr("nmoo", 1)

	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		expected *MsgOptOut
	}{
		{
			name:     "addr 0",
			toAddr:   testAddr0,
			expected: &MsgOptOut{ToAddress: testAddr0.String()},
		},
		{
			name:     "addr 1",
			toAddr:   testAddr1,
			expected: &MsgOptOut{ToAddress: testAddr1.String()},
		},
		{
			name:     "nil",
			toAddr:   nil,
			expected: &MsgOptOut{ToAddress: ""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewMsgOptOut(tc.toAddr)
			assert.Equal(t, tc.expected, actual, "NewMsgOptOut")
		})
	}
}

func TestMsgOptOut_ValidateBasic(t *testing.T) {
	addr := MakeTestAddr("moovb", 0).String()

	tests := []struct {
		name          string
		addr          string
		expectedInErr []string
	}{
		{
			name:          "addr",
			addr:          addr,
			expectedInErr: nil,
		},
		{
			name:          "bad",
			addr:          "not an actual address",
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "empty",
			addr:          "",
			expectedInErr: []string{"invalid to address"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := MsgOptOut{ToAddress: tc.addr}
			msg := MsgOptOut{ToAddress: tc.addr}
			err := msg.ValidateBasic()
			AssertErrorContents(t, err, tc.expectedInErr, "ValidateBasic")
			assert.Equal(t, msgOrig, msg, "MsgOptOut before and after")
		})
	}
}

func TestMsgOptOut_GetSigners(t *testing.T) {
	addr := MakeTestAddr("moogs", 0)

	tests := []struct {
		name     string
		addr     string
		expected []sdk.AccAddress
	}{
		{
			name:     "addr",
			addr:     addr.String(),
			expected: []sdk.AccAddress{addr},
		},
		{
			name:     "bad",
			addr:     "not an actual address",
			expected: []sdk.AccAddress{nil},
		},
		{
			name:     "empty",
			addr:     "",
			expected: []sdk.AccAddress{{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := MsgOptOut{ToAddress: tc.addr}
			msg := MsgOptOut{ToAddress: tc.addr}
			actual := msg.GetSigners()
			assert.Equal(t, tc.expected, actual, "GetSigners")
			assert.Equal(t, msgOrig, msg, "MsgOptOut before and after")
		})
	}
}

func TestNewMsgAccept(t *testing.T) {
	testAddr0 := MakeTestAddr("nma", 0)
	testAddr1 := MakeTestAddr("nma", 1)

	tests := []struct {
		name      string
		toAddr    sdk.AccAddress
		fromAddrs []string
		permanent bool
		expected  *MsgAccept
	}{
		{
			name:      "control",
			toAddr:    testAddr0,
			fromAddrs: []string{testAddr1.String()},
			permanent: false,
			expected: &MsgAccept{
				ToAddress:     testAddr0.String(),
				FromAddresses: []string{testAddr1.String()},
				Permanent:     false,
			},
		},
		{
			name:      "nil toAddr",
			toAddr:    nil,
			fromAddrs: []string{testAddr1.String()},
			permanent: false,
			expected: &MsgAccept{
				ToAddress:     "",
				FromAddresses: []string{testAddr1.String()},
				Permanent:     false,
			},
		},
		{
			name:      "nil fromAddrsStrs",
			toAddr:    testAddr1,
			fromAddrs: nil,
			permanent: false,
			expected: &MsgAccept{
				ToAddress:     testAddr1.String(),
				FromAddresses: nil,
				Permanent:     false,
			},
		},
		{
			name:      "empty fromAddrsStrs",
			toAddr:    testAddr1,
			fromAddrs: []string{},
			permanent: false,
			expected: &MsgAccept{
				ToAddress:     testAddr1.String(),
				FromAddresses: []string{},
				Permanent:     false,
			},
		},
		{
			name:      "three bad fromAddrsStrs",
			toAddr:    testAddr1,
			fromAddrs: []string{"one", "two", "three"},
			permanent: false,
			expected: &MsgAccept{
				ToAddress:     testAddr1.String(),
				FromAddresses: []string{"one", "two", "three"},
				Permanent:     false,
			},
		},
		{
			name:      "permanent",
			toAddr:    testAddr1,
			fromAddrs: []string{testAddr0.String()},
			permanent: true,
			expected: &MsgAccept{
				ToAddress:     testAddr1.String(),
				FromAddresses: []string{testAddr0.String()},
				Permanent:     true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewMsgAccept(tc.toAddr, tc.fromAddrs, tc.permanent)
			assert.Equal(t, tc.expected, actual, "NewMsgAccept")
		})
	}
}

func TestMsgAccept_ValidateBasic(t *testing.T) {
	testAddr0 := MakeTestAddr("mavb", 0).String()
	testAddr1 := MakeTestAddr("mavb", 1).String()
	testAddr2 := MakeTestAddr("mavb", 2).String()

	tests := []struct {
		name          string
		toAddr        string
		fromAddrs     []string
		permanent     bool
		expectedInErr []string
	}{
		{
			name:          "control",
			toAddr:        testAddr0,
			fromAddrs:     []string{testAddr1},
			permanent:     false,
			expectedInErr: nil,
		},
		{
			name:          "permanent",
			toAddr:        testAddr0,
			fromAddrs:     []string{testAddr1},
			permanent:     true,
			expectedInErr: nil,
		},
		{
			name:          "empty to address",
			toAddr:        "",
			fromAddrs:     []string{testAddr1},
			permanent:     false,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "bad to address",
			toAddr:        "this address isn't",
			fromAddrs:     []string{testAddr0},
			permanent:     false,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "nil from addresses",
			toAddr:        testAddr1,
			fromAddrs:     nil,
			permanent:     false,
			expectedInErr: []string{"at least one from address is required", "unknown address"},
		},
		{
			name:          "empty from addresses",
			toAddr:        testAddr1,
			fromAddrs:     []string{},
			permanent:     false,
			expectedInErr: []string{"at least one from address is required", "unknown address"},
		},
		{
			name:          "bad from address",
			toAddr:        testAddr0,
			fromAddrs:     []string{"this one is a tunic"},
			permanent:     false,
			expectedInErr: []string{"invalid from address[0]"},
		},
		{
			name:          "bad third from address",
			toAddr:        testAddr0,
			fromAddrs:     []string{testAddr1, testAddr2, "Michael Jackson (he's bad)"},
			permanent:     false,
			expectedInErr: []string{"invalid from address[2]"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := MsgAccept{
				ToAddress:     tc.toAddr,
				FromAddresses: MakeCopyOfStringSlice(tc.fromAddrs),
				Permanent:     tc.permanent,
			}
			msg := MsgAccept{
				ToAddress:     tc.toAddr,
				FromAddresses: tc.fromAddrs,
				Permanent:     tc.permanent,
			}
			err := msg.ValidateBasic()
			AssertErrorContents(t, err, tc.expectedInErr, "ValidateBasic")
			assert.Equal(t, msgOrig, msg, "MsgAccept before and after")
		})
	}
}

func TestMsgAccept_GetSigners(t *testing.T) {
	testAddr0 := MakeTestAddr("mags", 0)
	testAddr1 := MakeTestAddr("mags", 1)
	testAddr2 := MakeTestAddr("mags", 2)

	tests := []struct {
		name      string
		toAddr    string
		fromAddrs []string
		permanent bool
		expected  []sdk.AccAddress
	}{
		{
			name:      "control",
			toAddr:    testAddr0.String(),
			fromAddrs: []string{testAddr1.String()},
			permanent: false,
			expected:  []sdk.AccAddress{testAddr0},
		},
		{
			name:      "permanent",
			toAddr:    testAddr0.String(),
			fromAddrs: []string{testAddr1.String()},
			permanent: true,
			expected:  []sdk.AccAddress{testAddr0},
		},
		{
			name:      "empty to address",
			toAddr:    "",
			fromAddrs: []string{testAddr1.String()},
			permanent: false,
			expected:  []sdk.AccAddress{{}},
		},
		{
			name:      "bad to address",
			toAddr:    "this address isn't",
			fromAddrs: []string{testAddr0.String()},
			permanent: false,
			expected:  []sdk.AccAddress{nil},
		},
		{
			name:      "empty from addresses",
			toAddr:    testAddr1.String(),
			fromAddrs: []string{},
			permanent: false,
			expected:  []sdk.AccAddress{testAddr1},
		},
		{
			name:      "two from addresses",
			toAddr:    testAddr2.String(),
			fromAddrs: []string{testAddr0.String(), testAddr1.String()},
			permanent: false,
			expected:  []sdk.AccAddress{testAddr2},
		},
		{
			name:      "bad from address",
			toAddr:    testAddr0.String(),
			fromAddrs: []string{"this one is a tunic"},
			permanent: false,
			expected:  []sdk.AccAddress{testAddr0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := MsgAccept{
				ToAddress:     tc.toAddr,
				FromAddresses: MakeCopyOfStringSlice(tc.fromAddrs),
				Permanent:     tc.permanent,
			}
			msg := MsgAccept{
				ToAddress:     tc.toAddr,
				FromAddresses: tc.fromAddrs,
				Permanent:     tc.permanent,
			}
			actual := msg.GetSigners()
			assert.Equal(t, tc.expected, actual, "GetSigners")
			assert.Equal(t, msgOrig, msg, "MsgAccept before and after")
		})
	}
}

func TestNewMsgDecline(t *testing.T) {
	testAddr0 := MakeTestAddr("nmd", 0)
	testAddr1 := MakeTestAddr("nmd", 1)

	tests := []struct {
		name      string
		toAddr    sdk.AccAddress
		fromAddrs []string
		permanent bool
		expected  *MsgDecline
	}{
		{
			name:      "control",
			toAddr:    testAddr0,
			fromAddrs: []string{testAddr1.String()},
			permanent: false,
			expected: &MsgDecline{
				ToAddress:     testAddr0.String(),
				FromAddresses: []string{testAddr1.String()},
				Permanent:     false,
			},
		},
		{
			name:      "nil toAddr",
			toAddr:    nil,
			fromAddrs: []string{testAddr1.String()},
			permanent: false,
			expected: &MsgDecline{
				ToAddress:     "",
				FromAddresses: []string{testAddr1.String()},
				Permanent:     false,
			},
		},
		{
			name:      "nil fromAddrsStrs",
			toAddr:    testAddr1,
			fromAddrs: nil,
			permanent: false,
			expected: &MsgDecline{
				ToAddress:     testAddr1.String(),
				FromAddresses: nil,
				Permanent:     false,
			},
		},
		{
			name:      "empty fromAddrsStrs",
			toAddr:    testAddr1,
			fromAddrs: []string{},
			permanent: false,
			expected: &MsgDecline{
				ToAddress:     testAddr1.String(),
				FromAddresses: []string{},
				Permanent:     false,
			},
		},
		{
			name:      "three bad fromAddrsStrs",
			toAddr:    testAddr1,
			fromAddrs: []string{"one", "two", "three"},
			permanent: false,
			expected: &MsgDecline{
				ToAddress:     testAddr1.String(),
				FromAddresses: []string{"one", "two", "three"},
				Permanent:     false,
			},
		},
		{
			name:      "permanent",
			toAddr:    testAddr1,
			fromAddrs: []string{testAddr0.String()},
			permanent: true,
			expected: &MsgDecline{
				ToAddress:     testAddr1.String(),
				FromAddresses: []string{testAddr0.String()},
				Permanent:     true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewMsgDecline(tc.toAddr, tc.fromAddrs, tc.permanent)
			assert.Equal(t, tc.expected, actual, "NewMsgDecline")
		})
	}
}

func TestMsgDecline_ValidateBasic(t *testing.T) {
	testAddr0 := MakeTestAddr("mdvb", 0).String()
	testAddr1 := MakeTestAddr("mdvb", 1).String()
	testAddr2 := MakeTestAddr("mdvb", 2).String()

	tests := []struct {
		name          string
		toAddr        string
		fromAddrs     []string
		permanent     bool
		expectedInErr []string
	}{
		{
			name:          "control",
			toAddr:        testAddr0,
			fromAddrs:     []string{testAddr1},
			permanent:     false,
			expectedInErr: nil,
		},
		{
			name:          "permanent",
			toAddr:        testAddr0,
			fromAddrs:     []string{testAddr1},
			permanent:     true,
			expectedInErr: nil,
		},
		{
			name:          "empty to address",
			toAddr:        "",
			fromAddrs:     []string{testAddr1},
			permanent:     false,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "bad to address",
			toAddr:        "this address isn't",
			fromAddrs:     []string{testAddr0},
			permanent:     false,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "nil from addresses",
			toAddr:        testAddr1,
			fromAddrs:     nil,
			permanent:     false,
			expectedInErr: []string{"at least one from address is required", "unknown address"},
		},
		{
			name:          "empty from addresses",
			toAddr:        testAddr1,
			fromAddrs:     []string{},
			permanent:     false,
			expectedInErr: []string{"at least one from address is required", "unknown address"},
		},
		{
			name:          "bad from address",
			toAddr:        testAddr0,
			fromAddrs:     []string{"this one is a tunic"},
			permanent:     false,
			expectedInErr: []string{"invalid from address[0]"},
		},
		{
			name:          "bad third from address",
			toAddr:        testAddr0,
			fromAddrs:     []string{testAddr1, testAddr2, "Michael Jackson (he's bad)"},
			permanent:     false,
			expectedInErr: []string{"invalid from address[2]"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := MsgDecline{
				ToAddress:     tc.toAddr,
				FromAddresses: MakeCopyOfStringSlice(tc.fromAddrs),
				Permanent:     tc.permanent,
			}
			msg := MsgDecline{
				ToAddress:     tc.toAddr,
				FromAddresses: tc.fromAddrs,
				Permanent:     tc.permanent,
			}
			err := msg.ValidateBasic()
			AssertErrorContents(t, err, tc.expectedInErr, "ValidateBasic")
			assert.Equal(t, msgOrig, msg, "MsgDecline before and after")
		})
	}
}

func TestMsgDecline_GetSigners(t *testing.T) {
	testAddr0 := MakeTestAddr("mdgs", 0)
	testAddr1 := MakeTestAddr("mdgs", 1)
	testAddr2 := MakeTestAddr("mdgs", 2)

	tests := []struct {
		name      string
		toAddr    string
		fromAddrs []string
		permanent bool
		expected  []sdk.AccAddress
	}{
		{
			name:      "control",
			toAddr:    testAddr0.String(),
			fromAddrs: []string{testAddr1.String()},
			permanent: false,
			expected:  []sdk.AccAddress{testAddr0},
		},
		{
			name:      "permanent",
			toAddr:    testAddr0.String(),
			fromAddrs: []string{testAddr1.String()},
			permanent: true,
			expected:  []sdk.AccAddress{testAddr0},
		},
		{
			name:      "empty to address",
			toAddr:    "",
			fromAddrs: []string{testAddr1.String()},
			permanent: false,
			expected:  []sdk.AccAddress{{}},
		},
		{
			name:      "bad to address",
			toAddr:    "this address isn't",
			fromAddrs: []string{testAddr0.String()},
			permanent: false,
			expected:  []sdk.AccAddress{nil},
		},
		{
			name:      "empty from addresses",
			toAddr:    testAddr1.String(),
			fromAddrs: []string{},
			permanent: false,
			expected:  []sdk.AccAddress{testAddr1},
		},
		{
			name:      "two from addresses",
			toAddr:    testAddr2.String(),
			fromAddrs: []string{testAddr0.String(), testAddr1.String()},
			permanent: false,
			expected:  []sdk.AccAddress{testAddr2},
		},
		{
			name:      "bad from address",
			toAddr:    testAddr0.String(),
			fromAddrs: []string{"this one is a tunic"},
			permanent: false,
			expected:  []sdk.AccAddress{testAddr0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := MsgDecline{
				ToAddress:     tc.toAddr,
				FromAddresses: MakeCopyOfStringSlice(tc.fromAddrs),
				Permanent:     tc.permanent,
			}
			msg := MsgDecline{
				ToAddress:     tc.toAddr,
				FromAddresses: tc.fromAddrs,
				Permanent:     tc.permanent,
			}
			actual := msg.GetSigners()
			assert.Equal(t, tc.expected, actual, "GetSigners")
			assert.Equal(t, msgOrig, msg, "MsgDecline before and after")
		})
	}
}

func TestNewMsgUpdateAutoResponses(t *testing.T) {
	testAddr0 := MakeTestAddr("nmuar", 0)
	testAddr1 := MakeTestAddr("nmuar", 1)
	testAddr2 := MakeTestAddr("nmuar", 2)
	testAddr3 := MakeTestAddr("nmuar", 3)
	testAddr4 := MakeTestAddr("nmuar", 4)
	testAddr5 := MakeTestAddr("nmuar", 5)

	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		updates  []*AutoResponseUpdate
		expected *MsgUpdateAutoResponses
	}{
		{
			name:    "empty updates",
			toAddr:  testAddr0,
			updates: []*AutoResponseUpdate{},
			expected: &MsgUpdateAutoResponses{
				ToAddress: testAddr0.String(),
				Updates:   []*AutoResponseUpdate{},
			},
		},
		{
			name:    "one update no to addr",
			toAddr:  nil,
			updates: []*AutoResponseUpdate{{FromAddress: testAddr2.String(), Response: AUTO_RESPONSE_ACCEPT}},
			expected: &MsgUpdateAutoResponses{
				ToAddress: "",
				Updates:   []*AutoResponseUpdate{{FromAddress: testAddr2.String(), Response: AUTO_RESPONSE_ACCEPT}},
			},
		},
		{
			name:    "one update accept",
			toAddr:  testAddr1,
			updates: []*AutoResponseUpdate{{FromAddress: testAddr2.String(), Response: AUTO_RESPONSE_ACCEPT}},
			expected: &MsgUpdateAutoResponses{
				ToAddress: testAddr1.String(),
				Updates:   []*AutoResponseUpdate{{FromAddress: testAddr2.String(), Response: AUTO_RESPONSE_ACCEPT}},
			},
		},
		{
			name:    "one update decline",
			toAddr:  testAddr2,
			updates: []*AutoResponseUpdate{{FromAddress: testAddr1.String(), Response: AUTO_RESPONSE_DECLINE}},
			expected: &MsgUpdateAutoResponses{
				ToAddress: testAddr2.String(),
				Updates:   []*AutoResponseUpdate{{FromAddress: testAddr1.String(), Response: AUTO_RESPONSE_DECLINE}},
			},
		},
		{
			name:    "one update unspecified",
			toAddr:  testAddr0,
			updates: []*AutoResponseUpdate{{FromAddress: testAddr2.String(), Response: AUTO_RESPONSE_UNSPECIFIED}},
			expected: &MsgUpdateAutoResponses{
				ToAddress: testAddr0.String(),
				Updates:   []*AutoResponseUpdate{{FromAddress: testAddr2.String(), Response: AUTO_RESPONSE_UNSPECIFIED}},
			},
		},
		{
			name:    "one update unspecified",
			toAddr:  testAddr0,
			updates: []*AutoResponseUpdate{{FromAddress: testAddr2.String(), Response: AUTO_RESPONSE_UNSPECIFIED}},
			expected: &MsgUpdateAutoResponses{
				ToAddress: testAddr0.String(),
				Updates:   []*AutoResponseUpdate{{FromAddress: testAddr2.String(), Response: AUTO_RESPONSE_UNSPECIFIED}},
			},
		},
		{
			name:   "five updates",
			toAddr: testAddr0,
			updates: []*AutoResponseUpdate{
				{FromAddress: testAddr1.String(), Response: AUTO_RESPONSE_ACCEPT},
				{FromAddress: testAddr2.String(), Response: AUTO_RESPONSE_DECLINE},
				{FromAddress: testAddr3.String(), Response: AUTO_RESPONSE_ACCEPT},
				{FromAddress: testAddr4.String(), Response: AUTO_RESPONSE_UNSPECIFIED},
				{FromAddress: testAddr5.String(), Response: AUTO_RESPONSE_ACCEPT},
			},
			expected: &MsgUpdateAutoResponses{
				ToAddress: testAddr0.String(),
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddr1.String(), Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr2.String(), Response: AUTO_RESPONSE_DECLINE},
					{FromAddress: testAddr3.String(), Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr4.String(), Response: AUTO_RESPONSE_UNSPECIFIED},
					{FromAddress: testAddr5.String(), Response: AUTO_RESPONSE_ACCEPT},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewMsgUpdateAutoResponses(tc.toAddr, tc.updates)
			assert.Equal(t, tc.expected, actual, "NewMsgUpdateAutoResponses")
		})
	}
}

func TestMsgUpdateAutoResponses_ValidateBasic(t *testing.T) {
	testAddr0 := MakeTestAddr("muarvb", 0).String()
	testAddr1 := MakeTestAddr("muarvb", 1).String()
	testAddr2 := MakeTestAddr("muarvb", 2).String()
	testAddr3 := MakeTestAddr("muarvb", 3).String()
	testAddr4 := MakeTestAddr("muarvb", 4).String()
	testAddr5 := MakeTestAddr("muarvb", 5).String()

	tests := []struct {
		name          string
		orig          MsgUpdateAutoResponses
		expectedInErr []string
	}{
		{
			name: "control accept",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddr1, Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: nil,
		},
		{
			name: "control decline",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddr2, Response: AUTO_RESPONSE_DECLINE},
				},
			},
			expectedInErr: nil,
		},
		{
			name: "control unspecified",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddr3, Response: AUTO_RESPONSE_UNSPECIFIED},
				},
			},
			expectedInErr: nil,
		},
		{
			name: "bad to address",
			orig: MsgUpdateAutoResponses{
				ToAddress: "not really that bad",
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddr1, Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid to address"},
		},
		{
			name: "empty to address",
			orig: MsgUpdateAutoResponses{
				ToAddress: "",
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddr1, Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid to address"},
		},
		{
			name: "nil updates",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates:   nil,
			},
			expectedInErr: []string{"invalid value", "no updates"},
		},
		{
			name: "empty updates",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates:   []*AutoResponseUpdate{},
			},
			expectedInErr: []string{"invalid value", "no updates"},
		},
		{
			name: "one update bad from address",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*AutoResponseUpdate{
					{FromAddress: "Okay, I'm bad again.", Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid update 1", "invalid from address"},
		},
		{
			name: "one update empty from address",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*AutoResponseUpdate{
					{FromAddress: "", Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid update 1", "invalid from address"},
		},
		{
			name: "one update negative resp",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddr1, Response: -1},
				},
			},
			expectedInErr: []string{"invalid update 1", "unknown auto-response value: -1"},
		},
		{
			name: "one update resp too large",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddr2, Response: 900},
				},
			},
			expectedInErr: []string{"invalid update 1", "unknown auto-response value: 900"},
		},
		{
			name: "five updates third bad from address",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddr1, Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr2, Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: "still not good", Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr4, Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr5, Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid update 3", "invalid from address"},
		},
		{
			name: "five updates fourth empty from address",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddr1, Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr2, Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr3, Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: "", Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr5, Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid update 4", "invalid from address"},
		},
		{
			name: "five updates first negative resp",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddr1, Response: -88},
					{FromAddress: testAddr2, Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr3, Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr4, Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr5, Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid update 1", "unknown auto-response value: -88"},
		},
		{
			name: "five update last resp too large",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddr1, Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr2, Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr3, Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr4, Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr5, Response: 55},
				},
			},
			expectedInErr: []string{"invalid update 5", "unknown auto-response value: 55"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := MsgUpdateAutoResponses{
				ToAddress: tc.orig.ToAddress,
				Updates:   nil,
			}
			if tc.orig.Updates != nil {
				msg.Updates = []*AutoResponseUpdate{}
				for _, update := range tc.orig.Updates {
					msg.Updates = append(msg.Updates, &AutoResponseUpdate{
						FromAddress: update.FromAddress,
						Response:    update.Response,
					})
				}
			}
			err := msg.ValidateBasic()
			AssertErrorContents(t, err, tc.expectedInErr, "ValidateBasic")
			assert.Equal(t, tc.orig, msg, "MsgUpdateAutoResponses before and after")
		})
	}
}

func TestMsgUpdateAutoResponses_GetSigners(t *testing.T) {
	testAddr0 := MakeTestAddr("muargs", 0)
	testAddr1 := MakeTestAddr("muargs", 1)
	testAddr2 := MakeTestAddr("muargs", 2)

	tests := []struct {
		name     string
		orig     MsgUpdateAutoResponses
		expected []sdk.AccAddress
	}{
		{
			name: "control",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddr0.String(),
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddr1.String(), Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expected: []sdk.AccAddress{testAddr0},
		},
		{
			name: "bad addr",
			orig: MsgUpdateAutoResponses{
				ToAddress: "bad bad bad",
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddr2.String(), Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expected: []sdk.AccAddress{nil},
		},
		{
			name: "empty addr",
			orig: MsgUpdateAutoResponses{
				ToAddress: "",
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddr1.String(), Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expected: []sdk.AccAddress{{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := MsgUpdateAutoResponses{
				ToAddress: tc.orig.ToAddress,
				Updates:   nil,
			}
			if tc.orig.Updates != nil {
				msg.Updates = []*AutoResponseUpdate{}
				for _, update := range tc.orig.Updates {
					msg.Updates = append(msg.Updates, &AutoResponseUpdate{
						FromAddress: update.FromAddress,
						Response:    update.Response,
					})
				}
			}
			actual := msg.GetSigners()
			assert.Equal(t, tc.expected, actual, "GetSigners")
			assert.Equal(t, tc.orig, msg, "MsgUpdateAutoResponses before and after")
		})
	}
}
