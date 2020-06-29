package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
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
		{"invalid type", args{sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt())}, true},
		{"valid default params", args{*DefaultParams().SendEnabled[0]}, false},

		{"valid empty denom send enabled", args{*NewSendEnabled("", true)}, false},
		{"valid empty denom send disabled", args{*NewSendEnabled("", false)}, false},

		{"valid denom send enabled", args{*NewSendEnabled(sdk.DefaultBondDenom, true)}, false},
		{"valid denom send disabled", args{*NewSendEnabled(sdk.DefaultBondDenom, false)}, false},

		{"invalid denom send enabled", args{*NewSendEnabled("FOO", true)}, true},
		{"invalid denom send disabled", args{*NewSendEnabled("FOO", false)}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantErr, validateSendEnabled(tt.args.i) != nil)
		})
	}
}

func Test_sendParamEqual(t *testing.T) {
	paramsA := DefaultSendEnabled()
	paramsB := NewSendEnabled("", true)
	paramsC := NewSendEnabled("foodenom", false)

	ok := paramsA.Equal(paramsB)
	require.True(t, ok)

	ok = paramsA.Equal(paramsC)
	require.False(t, ok)
}

func Test_sendParamString(t *testing.T) {
	// verify denom attribute is omitted when empty
	noDenomString := "enabled: true\n"
	noDenom := NewSendEnabled("", true)

	require.Equal(t, noDenomString, noDenom.String())

	paramString := "denom: foo\nenabled: false\n"
	param := NewSendEnabled("foo", false)

	require.Equal(t, paramString, param.String())
}

func Test_validateParams(t *testing.T) {
	params := DefaultParams()

	// default params have no error
	require.NoError(t, params.Validate())

	// default case is all denoms are enabled for sending
	require.True(t, params.IsSendEnabled(sdk.DefaultBondDenom))
	require.True(t, params.IsSendEnabled("foodenom"))

	params = params.SetSendEnabledParam("", false).SetSendEnabledParam("foodenom", true)

	require.NoError(t, validateSendEnabledParams(params.SendEnabled))
	require.True(t, params.IsSendEnabled("foodenom"))
	require.False(t, params.IsSendEnabled(sdk.DefaultBondDenom))

	params = params.SetSendEnabledParam("foodenom", false).SetSendEnabledParam("", true)

	require.NoError(t, validateSendEnabledParams(params.SendEnabled))
	require.False(t, params.IsSendEnabled("foodenom"))
	require.True(t, params.IsSendEnabled(sdk.DefaultBondDenom))

	params = params.SetSendEnabledParam("foodenom", true)
	require.True(t, params.IsSendEnabled("foodenom"))

	params = params.SetSendEnabledParam("foodenom", false)
	require.False(t, params.IsSendEnabled("foodenom"))

	require.True(t, params.IsSendEnabled("foodenom2"))
	params = params.SetSendEnabledParam("foodenom2", false)
	require.True(t, params.IsSendEnabled(""))
	require.True(t, params.IsSendEnabled(sdk.DefaultBondDenom))
	require.False(t, params.IsSendEnabled("foodenom2"))

	paramYaml := `send_enabled:
- enabled: true
- denom: foodenom
  enabled: false
- denom: foodenom2
  enabled: false
`
	require.Equal(t, paramYaml, params.String())

	params = NewParams(SendEnabledParams{
		NewSendEnabled("", true),
		NewSendEnabled("foodenom", false),
		NewSendEnabled("foodenom", true), // this is not allowed
	})

	// fails due to duplicate entries.
	require.Error(t, params.Validate())

	// fails due to invalid type
	require.Error(t, validateSendEnabledParams(DefaultSendEnabled()))

	require.Error(t, validateSendEnabledParams(SendEnabledParams{NewSendEnabled("INVALIDDENOM", true)}))
}
