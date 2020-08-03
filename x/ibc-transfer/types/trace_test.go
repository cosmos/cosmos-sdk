package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDenomTrace(t *testing.T) {
	testCases := []struct {
		name     string
		denom    string
		expTrace DenomTrace
	}{
		{"empty denom", "", DenomTrace{}},
		{"base denom", "uatom", DenomTrace{BaseDenom: "uatom"}},
		{"trace info", "transfer/channelToA/uatom", DenomTrace{BaseDenom: "uatom", Trace: "transfer/channelToA"}},
	}

	for _, tc := range testCases {
		trace := ParseDenomTrace(tc.denom)
		require.Equal(t, tc.expTrace, trace, tc.name)
	}
}

func TestDenomTrace_IBCDenom(t *testing.T) {
	testCases := []struct {
		name     string
		trace    DenomTrace
		expDenom string
	}{
		{"base denom", DenomTrace{BaseDenom: "uatom"}, "uatom"},
		{"trace info", DenomTrace{BaseDenom: "uatom", Trace: "transfer/channelToA"}, "ibc/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A2"},
	}

	for _, tc := range testCases {
		denom := tc.trace.IBCDenom()
		require.Equal(t, tc.expDenom, denom, tc.name)
	}
}

func TestDenomTrace_RemovePrefix(t *testing.T) {
	testCases := []struct {
		name     string
		trace    DenomTrace
		expTrace string
	}{
		{"no trace", DenomTrace{BaseDenom: "uatom"}, ""},
		{"single trace info", DenomTrace{BaseDenom: "uatom", Trace: "transfer/channelToA"}, ""},
		{"multiple trace info", DenomTrace{BaseDenom: "uatom", Trace: "transfer/channelToA/transfer/channelToB"}, "transfer/channelToB"},
	}

	for _, tc := range testCases {
		tc.trace.RemovePrefix()
		require.Equal(t, tc.expTrace, tc.trace.Trace, tc.name)
	}
}

func TestDenomTrace_Validate(t *testing.T) {
	testCases := []struct {
		name     string
		trace    DenomTrace
		expError bool
	}{
		{"base denom only", DenomTrace{BaseDenom: "uatom"}, false},
		{"empty DenomTrace", DenomTrace{}, true},
		{"valid single trace info", DenomTrace{BaseDenom: "uatom", Trace: "transfer/channelToA"}, false},
		{"valid multiple trace info", DenomTrace{BaseDenom: "uatom", Trace: "transfer/channelToA/transfer/channelToB"}, false},
		{"single trace identifier", DenomTrace{BaseDenom: "uatom", Trace: "transfer"}, true},
		{"invalid port ID", DenomTrace{BaseDenom: "uatom", Trace: "(transfer)/channelToA"}, true},
		{"invalid channel ID", DenomTrace{BaseDenom: "uatom", Trace: "transfer/(channelToA)"}, true},
		{"empty base denom with trace", DenomTrace{BaseDenom: "", Trace: "transfer/channelToA"}, true},
	}

	for _, tc := range testCases {
		err := tc.trace.Validate()
		if tc.expError {
			require.Error(t, err, tc.name)
			continue
		}
		require.NoError(t, err, tc.name)
	}
}
