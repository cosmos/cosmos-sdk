package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
)

func TestValidate(t *testing.T) {
	params := DefaultParams()
	err := params.Validate()
	require.NoError(t, err)

	params2 := NewParams(
		params.MintDenom,
		params.InflationRateChange,
		params.InflationMax,
		params.InflationMin,
		params.GoalBonded,
		params.BlocksPerYear,
		params.MaxSupply,
	)
	err = params2.Validate()
	require.NoError(t, err)
	require.Equal(t, params, params2)

	params.MintDenom = ""
	err = params.Validate()
	require.Error(t, err)

	params.MintDenom = "asd/$%!@#"
	err = params.Validate()
	require.Error(t, err)

	params = DefaultParams()
	params.InflationRateChange = math.LegacyNewDec(123)
	err = params.Validate()
	require.Error(t, err)

	params = DefaultParams()
	params.InflationRateChange = math.LegacyNewDec(-123)
	err = params.Validate()
	require.Error(t, err)

	params = DefaultParams()
	params.InflationRateChange = math.LegacyDec{}
	err = params.Validate()
	require.Error(t, err)

	params = DefaultParams()
	params.InflationMax = math.LegacyNewDec(123)
	err = params.Validate()
	require.Error(t, err)

	params = DefaultParams()
	params.InflationMax = math.LegacyNewDec(-123)
	err = params.Validate()
	require.Error(t, err)

	params = DefaultParams()
	params.InflationMax = math.LegacyDec{}
	err = params.Validate()
	require.Error(t, err)

	params = DefaultParams()
	params.InflationMin = math.LegacyNewDec(123)
	err = params.Validate()
	require.Error(t, err)

	params = DefaultParams()
	params.InflationMin = math.LegacyNewDec(-123)
	err = params.Validate()
	require.Error(t, err)

	params = DefaultParams()
	params.InflationMin = math.LegacyDec{}
	err = params.Validate()
	require.Error(t, err)

	params = DefaultParams()
	params.GoalBonded = math.LegacyNewDec(123)
	err = params.Validate()
	require.Error(t, err)

	params = DefaultParams()
	params.GoalBonded = math.LegacyNewDec(-123)
	err = params.Validate()
	require.Error(t, err)

	params = DefaultParams()
	params.GoalBonded = math.LegacyDec{}
	err = params.Validate()
	require.Error(t, err)

	params = DefaultParams()
	params.BlocksPerYear = 0
	err = params.Validate()
	require.Error(t, err)

	params = DefaultParams()
	params.MaxSupply = math.NewInt(-1)
	err = params.Validate()
	require.Error(t, err)

	params = DefaultParams()
	params.InflationMax = math.LegacyNewDecWithPrec(1, 2)
	params.InflationMin = math.LegacyNewDecWithPrec(2, 2)
	err = params.Validate()
	require.Error(t, err)
}
