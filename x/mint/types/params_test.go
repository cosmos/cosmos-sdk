package types

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
)

func TestDefaultParams_ValidateOK(t *testing.T) {
	t.Parallel()
	require.NoError(t, DefaultParams().Validate())
}

func TestNewParams_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		params      Params
		wantErr     bool
		errContains string
	}{
		{
			name: "valid typical config",
			params: NewParams(
				"uatom",
				sdkmath.LegacyMustNewDecFromStr("0.13"),
				sdkmath.LegacyMustNewDecFromStr("0.20"),
				sdkmath.LegacyMustNewDecFromStr("0.07"),
				sdkmath.LegacyMustNewDecFromStr("0.67"),
				uint64(60*60*8766/5),
			),
			wantErr:     false,
			errContains: "",
		},
		{
			name: "max < min inflation",
			params: NewParams(
				"uatom",
				sdkmath.LegacyMustNewDecFromStr("0.10"),
				sdkmath.LegacyMustNewDecFromStr("0.05"),
				sdkmath.LegacyMustNewDecFromStr("0.06"),
				sdkmath.LegacyMustNewDecFromStr("0.67"),
				1,
			),
			wantErr:     true,
			errContains: "must be greater than or equal to min inflation",
		},
		{
			name: "invalid denom",
			params: NewParams(
				"",
				sdkmath.LegacyMustNewDecFromStr("0.10"),
				sdkmath.LegacyMustNewDecFromStr("0.20"),
				sdkmath.LegacyMustNewDecFromStr("0.07"),
				sdkmath.LegacyMustNewDecFromStr("0.67"),
				1,
			),
			wantErr:     true,
			errContains: "",
		},
		{
			name: "goal bonded > 1",
			params: NewParams(
				"uatom",
				sdkmath.LegacyMustNewDecFromStr("0.10"),
				sdkmath.LegacyMustNewDecFromStr("0.20"),
				sdkmath.LegacyMustNewDecFromStr("0.07"),
				sdkmath.LegacyMustNewDecFromStr("1.01"),
				1,
			),
			wantErr:     true,
			errContains: "",
		},
		{
			name: "blocks per year zero",
			params: NewParams(
				"uatom",
				sdkmath.LegacyMustNewDecFromStr("0.10"),
				sdkmath.LegacyMustNewDecFromStr("0.20"),
				sdkmath.LegacyMustNewDecFromStr("0.07"),
				sdkmath.LegacyMustNewDecFromStr("0.67"),
				0,
			),
			wantErr:     true,
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.params.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.ErrorContains(t, err, tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateMintDenom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		denom       string
		wantErr     bool
		errContains string
	}{
		{name: "blank", denom: "", wantErr: true, errContains: "cannot be blank"},
		{name: "spaces", denom: "   ", wantErr: true, errContains: "cannot be blank"},
		{name: "valid lower", denom: "uatom", wantErr: false, errContains: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateMintDenom(tt.denom)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.ErrorContains(t, err, tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateInflationRateChange(t *testing.T) {
	t.Parallel()

	var nilDec sdkmath.LegacyDec // zero value => IsNil() == true

	tests := []struct {
		name        string
		val         sdkmath.LegacyDec
		wantErr     bool
		errContains string
	}{
		{name: "nil", val: nilDec, wantErr: true, errContains: "cannot be nil"},
		{name: "negative", val: sdkmath.LegacyMustNewDecFromStr("-0.01"), wantErr: true, errContains: "cannot be negative"},
		{name: "too large > 1", val: sdkmath.LegacyMustNewDecFromStr("1.01"), wantErr: true, errContains: "too large"},
		{name: "zero ok", val: sdkmath.LegacyMustNewDecFromStr("0"), wantErr: false, errContains: ""},
		{name: "one ok", val: sdkmath.LegacyMustNewDecFromStr("1.0"), wantErr: false, errContains: ""},
		{name: "mid ok", val: sdkmath.LegacyMustNewDecFromStr("0.13"), wantErr: false, errContains: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateInflationRateChange(tt.val)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.ErrorContains(t, err, tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateInflationMax(t *testing.T) {
	t.Parallel()

	var nilDec sdkmath.LegacyDec

	tests := []struct {
		name        string
		val         sdkmath.LegacyDec
		wantErr     bool
		errContains string
	}{
		{name: "nil", val: nilDec, wantErr: true, errContains: "cannot be nil"},
		{name: "negative", val: sdkmath.LegacyMustNewDecFromStr("-0.01"), wantErr: true, errContains: "cannot be negative"},
		{name: "too large > 1", val: sdkmath.LegacyMustNewDecFromStr("1.1"), wantErr: true, errContains: "too large"},
		{name: "zero ok", val: sdkmath.LegacyMustNewDecFromStr("0"), wantErr: false, errContains: ""},
		{name: "one ok", val: sdkmath.LegacyMustNewDecFromStr("1"), wantErr: false, errContains: ""},
		{name: "typical ok", val: sdkmath.LegacyMustNewDecFromStr("0.20"), wantErr: false, errContains: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateInflationMax(tt.val)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.ErrorContains(t, err, tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateInflationMin(t *testing.T) {
	t.Parallel()

	var nilDec sdkmath.LegacyDec

	tests := []struct {
		name        string
		val         sdkmath.LegacyDec
		wantErr     bool
		errContains string
	}{
		{name: "nil", val: nilDec, wantErr: true, errContains: "cannot be nil"},
		{name: "negative", val: sdkmath.LegacyMustNewDecFromStr("-0.01"), wantErr: true, errContains: "cannot be negative"},
		{name: "too large > 1", val: sdkmath.LegacyMustNewDecFromStr("1.1"), wantErr: true, errContains: "too large"},
		{name: "zero ok", val: sdkmath.LegacyMustNewDecFromStr("0"), wantErr: false, errContains: ""},
		{name: "one ok", val: sdkmath.LegacyMustNewDecFromStr("1"), wantErr: false, errContains: ""},
		{name: "typical ok", val: sdkmath.LegacyMustNewDecFromStr("0.07"), wantErr: false, errContains: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateInflationMin(tt.val)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.ErrorContains(t, err, tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateGoalBonded(t *testing.T) {
	t.Parallel()

	var nilDec sdkmath.LegacyDec

	tests := []struct {
		name        string
		val         sdkmath.LegacyDec
		wantErr     bool
		errContains string
	}{
		{name: "nil", val: nilDec, wantErr: true, errContains: "cannot be nil"},
		{name: "negative", val: sdkmath.LegacyMustNewDecFromStr("-0.01"), wantErr: true, errContains: "must be positive"},
		{name: "zero", val: sdkmath.LegacyMustNewDecFromStr("0"), wantErr: true, errContains: "must be positive"},
		{name: "too large > 1", val: sdkmath.LegacyMustNewDecFromStr("1.0000000001"), wantErr: true, errContains: "too large"},
		{name: "exactly one ok", val: sdkmath.LegacyMustNewDecFromStr("1.0"), wantErr: false, errContains: ""},
		{name: "typical ok", val: sdkmath.LegacyMustNewDecFromStr("0.67"), wantErr: false, errContains: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateGoalBonded(tt.val)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.ErrorContains(t, err, tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateBlocksPerYear(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		val         uint64
		wantErr     bool
		errContains string
	}{
		{name: "zero", val: 0, wantErr: true, errContains: "must be positive"},
		{name: "one ok", val: 1, wantErr: false, errContains: ""},
		{name: "maxInt64 ok", val: uint64(math.MaxInt64), wantErr: false, errContains: ""},
		{name: "maxInt64+1 too large", val: uint64(math.MaxInt64) + 1, wantErr: true, errContains: "too large"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateBlocksPerYear(tt.val)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.ErrorContains(t, err, tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
