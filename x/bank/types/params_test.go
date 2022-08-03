package types

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_validateSendEnabledParam(t *testing.T) {
	type args struct {
		i interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"invalid type", args{sdk.NewCoin(sdk.DefaultBondDenom, math.OneInt())}, true},

		{"invalid empty denom send enabled", args{*NewSendEnabled("", true)}, true},
		{"invalid empty denom send disabled", args{*NewSendEnabled("", false)}, true},

		{"valid denom send enabled", args{*NewSendEnabled(sdk.DefaultBondDenom, true)}, false},
		{"valid denom send disabled", args{*NewSendEnabled(sdk.DefaultBondDenom, false)}, false},

		{"invalid denom send enabled", args{*NewSendEnabled("0FOO", true)}, true},
		{"invalid denom send disabled", args{*NewSendEnabled("0FOO", false)}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantErr, validateSendEnabled(tt.args.i) != nil)
		})
	}
}

func Test_sendParamEqual(t *testing.T) {
	paramsA := NewSendEnabled(sdk.DefaultBondDenom, true)
	paramsB := NewSendEnabled(sdk.DefaultBondDenom, true)
	paramsC := NewSendEnabled("foodenom", false)

	ok := paramsA.Equal(paramsB)
	require.True(t, ok)

	ok = paramsA.Equal(paramsC)
	require.False(t, ok)
}

func Test_SendEnabledString(t *testing.T) {
	paramStringTrue := "denom: foo\nenabled: true\n"
	paramTrue := NewSendEnabled("foo", true)
	assert.Equal(t, paramStringTrue, paramTrue.String(), "true")
	paramStringFalse := "denom: bar\nenabled: false\n"
	paramFalse := NewSendEnabled("bar", false)
	assert.Equal(t, paramStringFalse, paramFalse.String(), "false")
}

func Test_ParamsString(t *testing.T) {
	tests := []struct {
		name     string
		params   Params
		expected string
	}{
		{
			name:     "default true empty send enabled",
			params:   Params{[]*SendEnabled{}, true},
			expected: "default_send_enabled: true\nsend_enabled: []\n",
		},
		{
			name:     "default false empty send enabled",
			params:   Params{[]*SendEnabled{}, false},
			expected: "default_send_enabled: false\nsend_enabled: []\n",
		},
		{
			name:     "default true one true send enabled",
			params:   Params{[]*SendEnabled{{"foocoin", true}}, true},
			expected: "default_send_enabled: true\nsend_enabled:\n- denom: foocoin\n  enabled: true\n",
		},
		{
			name:     "default true one false send enabled",
			params:   Params{[]*SendEnabled{{"barcoin", false}}, true},
			expected: "default_send_enabled: true\nsend_enabled:\n- denom: barcoin\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			actual := tc.params.String()
			assert.Equal(tt, tc.expected, actual)
		})
	}
}

func Test_validateParams(t *testing.T) {
	assert.NoError(t, DefaultParams().Validate(), "default")
	assert.NoError(t, NewParams(true).Validate(), "true")
	assert.NoError(t, NewParams(false).Validate(), "false")
	assert.Error(t, Params{[]*SendEnabled{{"foocoing", false}}, true}.Validate(), "with SendEnabled entry")
}

func Test_validateSendEnabledParams(t *testing.T) {
	tests := []struct {
		name string
		arg  interface{}
		exp  string
	}{
		{
			name: "ok",
			arg:  []*SendEnabled{},
			exp:  "",
		},
		{
			name: "has entry",
			arg:  []*SendEnabled{{"foocoin", false}},
			exp:  "",
		},
		{
			name: "not a slice",
			arg:  &SendEnabled{},
			exp:  "invalid parameter type: *types.SendEnabled",
		},
		{
			name: "not a slice of refs",
			arg:  []SendEnabled{},
			exp:  "invalid parameter type: []types.SendEnabled",
		},
		{
			name: "not a slice of send enabled",
			arg:  []*Params{},
			exp:  "invalid parameter type: []*types.Params",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			actual := validateSendEnabledParams(tc.arg)
			if len(tc.exp) == 0 {
				assert.NoError(tt, actual)
			} else {
				assert.EqualError(tt, actual, tc.exp)
			}
		})
	}
}
