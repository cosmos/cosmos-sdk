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
		{"valid default params", args{DefaultSendEnabledParam()}, false},

		{"valid empty denom send enabled", args{NewSendEnabledParam("", true)}, false},
		{"valid empty denom send disabled", args{NewSendEnabledParam("", false)}, false},

		{"valid denom send enabled", args{NewSendEnabledParam(sdk.DefaultBondDenom, true)}, false},
		{"valid denom send disabled", args{NewSendEnabledParam(sdk.DefaultBondDenom, false)}, false},

		{"invalid denom send enabled", args{NewSendEnabledParam("FOO", true)}, true},
		{"invalid denom send disabled", args{NewSendEnabledParam("FOO", false)}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantErr, validateSendEnabledParam(tt.args.i) != nil)
		})
	}
}

func Test_sendParamEqual(t *testing.T) {
	paramsA := DefaultSendEnabledParam()
	paramsB := NewSendEnabledParam("", true)
	paramsC := NewSendEnabledParam("foodenom", false)

	ok := paramsA.Equal(paramsB)
	require.True(t, ok)

	ok = paramsA.Equal(paramsC)
	require.False(t, ok)
}

func Test_sendParamString(t *testing.T) {
	// verify denom attribute is omitted when empty
	noDenomString := "send_enabled: true\n"
	noDenom := NewSendEnabledParam("", true)

	require.Equal(t, noDenomString, noDenom.String())

	paramString := "denom: foo\nsend_enabled: false\n"
	param := NewSendEnabledParam("foo", false)

	require.Equal(t, paramString, param.String())
}

func Test_validateSendEnabledParams(t *testing.T) {
	params := DefaultParams()

	// default case is all denoms are enabled for sending
	require.True(t, params.SendEnabledParams.Enabled(sdk.DefaultBondDenom))
	require.True(t, params.SendEnabledParams.Enabled("foodenom"))

	sendParams := NewSendEnabledParams(
		NewSendEnabledParam("", false),
		NewSendEnabledParam("foodenom", true),
	)

	require.NoError(t, validateSendEnabledParams(sendParams))
	require.True(t, sendParams.Enabled("foodenom"))
	require.False(t, sendParams.Enabled(sdk.DefaultBondDenom))

	sendParams = sendParams.SetSendEnabledParam("foodenom", false).SetSendEnabledParam("", true)

	require.NoError(t, validateSendEnabledParams(sendParams))
	require.False(t, sendParams.Enabled("foodenom"))
	require.True(t, sendParams.Enabled(sdk.DefaultBondDenom))

	sendParams = sendParams.SetSendEnabledParam("foodenom", true)
	require.True(t, sendParams.Enabled("foodenom"))

	sendParams = sendParams.SetSendEnabledParam("foodenom", false)
	require.False(t, sendParams.Enabled("foodenom"))

	require.True(t, sendParams.Enabled("foodenom2"))
	sendParams = sendParams.SetSendEnabledParam("foodenom2", false)
	require.True(t, sendParams.Enabled(""))
	require.True(t, sendParams.Enabled(sdk.DefaultBondDenom))
	require.False(t, sendParams.Enabled("foodenom2"))

	sendParams = NewSendEnabledParams(
		NewSendEnabledParam("", true),
		NewSendEnabledParam("foodenom", false),
		NewSendEnabledParam("foodenom", true), // this is not allowed
	)

	// fails due to duplicate entries.
	require.Error(t, validateSendEnabledParams(sendParams))

	// fails due to invalid type
	require.Error(t, validateSendEnabledParams(DefaultSendEnabledParam()))
	require.Error(t, validateSendEnabledParams(NewSendEnabledParams(NewSendEnabledParam("INVALIDDENOM", true))))
}
