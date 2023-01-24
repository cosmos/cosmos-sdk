package quarantine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/cosmos/cosmos-sdk/x/quarantine"
	. "github.com/cosmos/cosmos-sdk/x/quarantine/testutil"
)

func TestGenesisState_Validate(t *testing.T) {
	testAddr0 := MakeTestAddr("gsv", 0).String()
	testAddr1 := MakeTestAddr("gsv", 1).String()
	badAddr := "this1addressisnaughty"

	goodAutoResponse := &AutoResponseEntry{
		ToAddress:   testAddr0,
		FromAddress: testAddr1,
		Response:    AUTO_RESPONSE_ACCEPT,
	}
	badAutoResponse := &AutoResponseEntry{
		ToAddress:   testAddr0,
		FromAddress: testAddr1,
		Response:    -10,
	}

	goodQuarantinedFunds := &QuarantinedFunds{
		ToAddress:               testAddr0,
		UnacceptedFromAddresses: []string{testAddr1},
		Coins:                   coinMakerOK(),
		Declined:                false,
	}
	badQuarantinedFunds := &QuarantinedFunds{
		ToAddress:               testAddr0,
		UnacceptedFromAddresses: []string{testAddr1},
		Coins:                   coinMakerBad(),
		Declined:                false,
	}

	tests := []struct {
		name    string
		gs      *GenesisState
		expErrs []string
	}{
		{
			name: "control",
			gs: &GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        []*AutoResponseEntry{goodAutoResponse, goodAutoResponse},
				QuarantinedFunds:     []*QuarantinedFunds{goodQuarantinedFunds, goodQuarantinedFunds},
			},
			expErrs: nil,
		},
		{
			name:    "empty",
			gs:      &GenesisState{},
			expErrs: nil,
		},
		{
			name: "bad first addr",
			gs: &GenesisState{
				QuarantinedAddresses: []string{badAddr, testAddr1},
				AutoResponses:        []*AutoResponseEntry{goodAutoResponse, goodAutoResponse},
				QuarantinedFunds:     []*QuarantinedFunds{goodQuarantinedFunds, goodQuarantinedFunds},
			},
			expErrs: []string{"invalid quarantined address[0]"},
		},
		{
			name: "bad second addr",
			gs: &GenesisState{
				QuarantinedAddresses: []string{testAddr0, badAddr},
				AutoResponses:        []*AutoResponseEntry{goodAutoResponse, goodAutoResponse},
				QuarantinedFunds:     []*QuarantinedFunds{goodQuarantinedFunds, goodQuarantinedFunds},
			},
			expErrs: []string{"invalid quarantined address[1]"},
		},
		{
			name: "bad first auto response",
			gs: &GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        []*AutoResponseEntry{badAutoResponse, goodAutoResponse},
				QuarantinedFunds:     []*QuarantinedFunds{goodQuarantinedFunds, goodQuarantinedFunds},
			},
			expErrs: []string{"invalid quarantine auto response entry[0]"},
		},
		{
			name: "bad second auto response",
			gs: &GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        []*AutoResponseEntry{goodAutoResponse, badAutoResponse},
				QuarantinedFunds:     []*QuarantinedFunds{goodQuarantinedFunds, goodQuarantinedFunds},
			},
			expErrs: []string{"invalid quarantine auto response entry[1]"},
		},
		{
			name: "bad first quarantined funds",
			gs: &GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        []*AutoResponseEntry{goodAutoResponse, goodAutoResponse},
				QuarantinedFunds:     []*QuarantinedFunds{badQuarantinedFunds, goodQuarantinedFunds},
			},
			expErrs: []string{"invalid quarantined funds[0]"},
		},
		{
			name: "bad second quarantined funds",
			gs: &GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        []*AutoResponseEntry{goodAutoResponse, goodAutoResponse},
				QuarantinedFunds:     []*QuarantinedFunds{goodQuarantinedFunds, badQuarantinedFunds},
			},
			expErrs: []string{"invalid quarantined funds[1]"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := MakeCopyOfGenesisState(tc.gs)
			var err error
			testFunc := func() {
				err = tc.gs.Validate()
			}
			assert.NotPanics(t, testFunc, "GenesisState.Validate()")
			AssertErrorContents(t, err, tc.expErrs, "Validate")
			assert.Equal(t, orig, tc.gs, "GenesisState before and after Validate")
		})
	}
}

func TestNewGenesisState(t *testing.T) {
	testAddr0 := MakeTestAddr("ngs", 0).String()
	testAddr1 := MakeTestAddr("ngs", 1).String()

	autoResponse := &AutoResponseEntry{
		ToAddress:   testAddr0,
		FromAddress: testAddr1,
		Response:    AUTO_RESPONSE_ACCEPT,
	}

	quarantinedFunds := &QuarantinedFunds{
		ToAddress:               testAddr0,
		UnacceptedFromAddresses: []string{testAddr1},
		Coins:                   coinMakerOK(),
		Declined:                false,
	}

	tests := []struct {
		name  string
		addrs []string
		ars   []*AutoResponseEntry
		qfs   []*QuarantinedFunds
		exp   *GenesisState
	}{
		{
			name:  "control",
			addrs: []string{testAddr0, testAddr1},
			ars:   []*AutoResponseEntry{autoResponse, autoResponse},
			qfs:   []*QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			exp: &GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        []*AutoResponseEntry{autoResponse, autoResponse},
				QuarantinedFunds:     []*QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			},
		},
		{
			name:  "nil addrs",
			addrs: nil,
			ars:   []*AutoResponseEntry{autoResponse, autoResponse},
			qfs:   []*QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			exp: &GenesisState{
				QuarantinedAddresses: nil,
				AutoResponses:        []*AutoResponseEntry{autoResponse, autoResponse},
				QuarantinedFunds:     []*QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			},
		},
		{
			name:  "empty addrs",
			addrs: []string{},
			ars:   []*AutoResponseEntry{autoResponse, autoResponse},
			qfs:   []*QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			exp: &GenesisState{
				QuarantinedAddresses: []string{},
				AutoResponses:        []*AutoResponseEntry{autoResponse, autoResponse},
				QuarantinedFunds:     []*QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			},
		},
		{
			name:  "nil auto responses",
			addrs: []string{testAddr0, testAddr1},
			ars:   nil,
			qfs:   []*QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			exp: &GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        nil,
				QuarantinedFunds:     []*QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			},
		},
		{
			name:  "empty auto responses",
			addrs: []string{testAddr0, testAddr1},
			ars:   []*AutoResponseEntry{},
			qfs:   []*QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			exp: &GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        []*AutoResponseEntry{},
				QuarantinedFunds:     []*QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			},
		},
		{
			name:  "nil quarantined funds",
			addrs: []string{testAddr0, testAddr1},
			ars:   []*AutoResponseEntry{autoResponse, autoResponse},
			qfs:   nil,
			exp: &GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        []*AutoResponseEntry{autoResponse, autoResponse},
				QuarantinedFunds:     nil,
			},
		},
		{
			name:  "empty quarantined funds",
			addrs: []string{testAddr0, testAddr1},
			ars:   []*AutoResponseEntry{autoResponse, autoResponse},
			qfs:   []*QuarantinedFunds{},
			exp: &GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        []*AutoResponseEntry{autoResponse, autoResponse},
				QuarantinedFunds:     []*QuarantinedFunds{},
			},
		},
		{
			name:  "all empty",
			addrs: []string{},
			ars:   []*AutoResponseEntry{},
			qfs:   []*QuarantinedFunds{},
			exp: &GenesisState{
				QuarantinedAddresses: []string{},
				AutoResponses:        []*AutoResponseEntry{},
				QuarantinedFunds:     []*QuarantinedFunds{},
			},
		},
		{
			name:  "DefaultGenesisState",
			addrs: nil,
			ars:   nil,
			qfs:   nil,
			exp:   DefaultGenesisState(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewGenesisState(tc.addrs, tc.ars, tc.qfs)
			assert.Equal(t, tc.exp, actual, "NewGenesisState")
		})
	}
}
