package valuerenderer_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/valuerenderer"
)

// TODO add more test cases
func TestFormatCoin(t *testing.T) {


	// TODO add test case to convert from mregen to uregen
	tt := []struct {
		name   string
		dvr    valuerenderer.DefaultValueRenderer
		coin   types.Coin
		expRes string
		expErr bool
	}{
		{
			"convert 1000000uregen to 1regen",
			valuerenderer.NewDefaultValueRendererWithDenom("regen"),
			types.NewCoin("uregen", types.NewInt(int64(1000000))),
			"1regen",
			false,
		},
		{
			"convert 1000000000uregen to 1000regen",
			valuerenderer.NewDefaultValueRendererWithDenom("regen"),
			types.NewCoin("uregen", types.NewInt(int64(1000000000))),
			"1,000regen",
			false,
		},
		{
			"convert 23000000mregen to 1000regen",
			valuerenderer.NewDefaultValueRendererWithDenom("regen"),
			types.NewCoin("mregen", types.NewInt(int64(23000000))),
			"23,000regen",
			false,
		},
		{
			"convert 23000000mregen to 23000000000uregen",
			valuerenderer.NewDefaultValueRendererWithDenom("uregen"),
			types.NewCoin("mregen", types.NewInt(int64(23000000))),
			"23000000000uregen",
			false,
		},

	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.dvr.Format(tc.coin)
			require.NoError(t, err)
			require.Equal(t, tc.expRes, res)
		})
	}
}

func TestFormatDec(t *testing.T) {
	var (
		d valuerenderer.DefaultValueRenderer
	)
	// TODO add more cases and error cases
	tt := []struct {
		name   string
		input  types.Dec
		expRes string
		expErr bool
	}{
		{
			"Decimal, no error",
			types.NewDecFromIntWithPrec(types.NewInt(1000000), 2), // 10000.000000000000000000
			"10,000.000000000000000000",
			false,
		},

		//{"invalid string input panic", "qwerty", "", true, true},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := d.Format(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expRes, res)
		})
	}
}

func TestFormatInt(t *testing.T) {
	var (
		d valuerenderer.DefaultValueRenderer
	)
	// TODO add more cases and error cases
	tt := []struct {
		name   string
		input  types.Int
		expRes string
		expErr bool
	}{
		{
			"1000000",
			types.NewInt(1000000),
			"1,000,000",
			false,
		},
		{
			"100",
			types.NewInt(100),
			"100",
			false,
		},
		{
			"23232345476756",
			types.NewInt(23232345476756),
			"23,232,345,476,756",
			false,
		},

		//{"invalid string input panic", "qwerty", "", true, true},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := d.Format(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expRes, res)
		})
	}
}

// TODO add more test cases
func TestParseString(t *testing.T) {
	re := regexp.MustCompile(`\d+[mu]?regen`)
	dvr := valuerenderer.NewDefaultValueRenderer()

	tt := []struct {
		str           string
		satisfyRegExp bool
		expErr        bool
	}{
		{"", false, true},
		{"10regen", true, false},
		{"1,000,000", false, false},
		{"323,000,000", false, false},
		{"1mregen", true, false},
		{"500uregen", true, false},
		{"1,500,000,000regen", true, false},
		{"394,382,328uregen", true, false},
	}

	for _, tc := range tt {
		t.Run(tc.str, func(t *testing.T) {
			x, err := dvr.Parse(tc.str)
			if tc.expErr {
				require.Error(t, err)
				require.Nil(t, x)
				return
			}

			if tc.satisfyRegExp {
				require.NoError(t, err)
				coin, ok := x.(types.Coin)
				require.True(t, ok)
				require.NotNil(t, coin)
				require.True(t, re.MatchString(tc.str))
			} else {
				require.NoError(t, err)
				u, ok := x.(types.Uint)
				require.True(t, ok)
				require.NotNil(t, u)
			}
		})
	}
}
