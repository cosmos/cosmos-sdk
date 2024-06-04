package math

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDecFromString(t *testing.T) {
	specs := map[string]struct {
		src         string
		constraints []SetupConstraint
		exp         Dec
		expErr      error
	}{
		"simple decimal": {
			src: "1",
			exp: NewDecFromInt64(1),
		},
		"simple negative decimal": {
			src: "-1",
			exp: NewDecFromInt64(-1),
		},
		"valid decimal with decimal places": {
			src: "1.234",
			exp: NewDecWithPrec(1234, -3),
		},
		"valid negative decimal": {
			src: "-1.234",
			exp: NewDecWithPrec(-1234, -3),
		},
		"min decimal": {
			src: "-" + strings.Repeat("9", 34),
			exp: must(NewDecWithPrec(-1, 34).Add(NewDecFromInt64(1))),
		},
		"max decimal": {
			// todo:  src: strings.Repeat("9", 34),
			exp: must(NewDecWithPrec(1, 34).Sub(NewDecFromInt64(1))),
		},
		"precision too high": {
			src:    "." + strings.Repeat("9", 35),
			expErr: ErrInvalidDecString,
		},
		"decimal too big": {
			// todo: src:    strings.Repeat("9", 35), // 10^100000+10
			expErr: ErrInvalidDecString,
		},
		"decimal too small": {
			src:    strings.Repeat("9", 35), // -10^100000+0.99999999999999999... +1
			expErr: ErrInvalidDecString,
		},
		"valid decimal with leading zero": {
			src: "01234",
			exp: NewDecWithPrec(1234, 0),
		},
		"valid decimal without leading zero": {
			src: ".1234",
			exp: NewDecWithPrec(1234, -4),
		},

		"valid decimal without trailing digits": {
			src: "123.",
			exp: NewDecWithPrec(123, 0),
		},

		"valid negative decimal without leading zero": {
			src: "-.1234",
			exp: NewDecWithPrec(-1234, -4),
		},

		"valid negative decimal without trailing digits": {
			src: "-123.",
			exp: NewDecWithPrec(-123, 0),
		},

		"decimal with scientific notation": {
			src: "1.23e4",
			exp: NewDecWithPrec(123, 2),
		},
		"negative decimal with scientific notation": {
			src: "-1.23e4",
			exp: NewDecWithPrec(-123, 2),
		},
		"with setup constraint": {
			src:         "-1",
			constraints: []SetupConstraint{AssertNotNegative()},
			expErr:      ErrInvalidDecString,
		},
		"empty string": {
			src:    "",
			expErr: ErrInvalidDecString,
		},
		"NaN": {
			src:    "NaN",
			expErr: ErrInvalidDecString,
		},
		"random string": {
			src:    "1foo",
			expErr: ErrInvalidDecString,
		},
		"Infinity": {
			src:    "Infinity",
			expErr: ErrInfiniteString,
		},
		"Inf": {
			src:    "Inf",
			expErr: ErrInfiniteString,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got, gotErr := NewDecFromString(spec.src, spec.constraints...)
			if spec.expErr != nil {
				require.ErrorIs(t, gotErr, spec.expErr, got.String())
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.exp.String(), got.String())
		})
	}
}

func TestIsFinite(t *testing.T) {
	a, err := NewDecFromString("1.5")
	require.NoError(t, err)
	require.True(t, a.IsFinite())
}

func TestReduce(t *testing.T) {
	a, err := NewDecFromString("1.30000")
	require.NoError(t, err)
	b, n := a.Reduce()
	require.Equal(t, 4, n)
	require.True(t, a.Equal(b))
	require.Equal(t, "1.3", b.String())
}

func TestMulExactGood(t *testing.T) {
	a, err := NewDecFromString("1.000001")
	require.NoError(t, err)
	b := NewDecWithPrec(1, 6)
	c, err := a.MulExact(b)
	require.NoError(t, err)
	d, err := c.Int64()
	require.NoError(t, err)
	require.Equal(t, int64(1000001), d)
}

func TestMulExactBad(t *testing.T) {
	a, err := NewDecFromString("1.000000000000000000000000000000000000123456789")
	require.NoError(t, err)
	b := NewDecWithPrec(1, 10)
	_, err = a.MulExact(b)
	require.ErrorIs(t, err, ErrUnexpectedRounding)
}

func TestQuoExactGood(t *testing.T) {
	a, err := NewDecFromString("1000001")
	require.NoError(t, err)
	b := NewDecWithPrec(1, 6)
	c, err := a.QuoExact(b)
	require.NoError(t, err)
	require.Equal(t, "1.000001000000000000000000000000000", c.String())
}

func TestQuoExactBad(t *testing.T) {
	a, err := NewDecFromString("1000000000000000000000000000000000000123456789")
	require.NoError(t, err)
	b := NewDecWithPrec(1, 10)
	_, err = a.QuoExact(b)
	require.ErrorIs(t, err, ErrUnexpectedRounding)
}

func TestToBigInt(t *testing.T) {
	i1 := "1000000000000000000000000000000000000123456789"
	tcs := []struct {
		intStr  string
		out     string
		isError error
	}{
		{i1, i1, nil},
		{"1000000000000000000000000000000000000123456789.00000000", i1, nil},
		{"123.456e6", "123456000", nil},
		{"12345.6", "", ErrNonIntegeral},
	}
	for idx, tc := range tcs {
		a, err := NewDecFromString(tc.intStr)
		require.NoError(t, err)
		b, err := a.BigInt()
		if tc.isError == nil {
			require.NoError(t, err, "test_%d", idx)
			require.Equal(t, tc.out, b.String(), "test_%d", idx)
		} else {
			require.ErrorIs(t, err, tc.isError, "test_%d", idx)
		}
	}
}

func TestToSdkInt(t *testing.T) {
	i1 := "1000000000000000000000000000000000000123456789"
	tcs := []struct {
		intStr string
		out    string
	}{
		{i1, i1},
		{"1000000000000000000000000000000000000123456789.00000000", i1},
		{"123.456e6", "123456000"},
		{"123.456e1", "1234"},
		{"123.456", "123"},
		{"123.956", "123"},
		{"-123.456", "-123"},
		{"-123.956", "-123"},
		{"-0.956", "0"},
		{"-0.9", "0"},
	}
	for idx, tc := range tcs {
		a, err := NewDecFromString(tc.intStr)
		require.NoError(t, err)
		b := a.SdkIntTrim()
		require.Equal(t, tc.out, b.String(), "test_%d", idx)
	}
}

func TestInfDecString(t *testing.T) {
	_, err := NewDecFromString("iNf")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInfiniteString)
}

//func TestDecToLegacyDec(t *testing.T) {
//	dec := NewDecFromInt64(123)
//
//	legacyDec, err := DecToLegacyDec(dec)
//	require.NoError(t, err)
//
//	expected, _ := LegacyNewDecFromStr("123.000000000000000000")
//	require.True(t, legacyDec.Equal(expected))
//}

func must[T any](r T, err error) T {
	if err != nil {
		panic(err)
	}
	return r
}
