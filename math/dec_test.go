package math

import (
	"fmt"
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

func TestNewDecFromInt64(t *testing.T) {
	specs := map[string]struct {
		src       int64
		constants []SetupConstraint
		exp       string
		expErr    error
	}{
		"zero value": {
			src: 0,
			exp: "0",
		},
		"positive value": {
			src: 123,
			exp: "123",
		},
		"negative value": {
			src: -123,
			exp: "-123",
		},
		"max value": {
			src: 9223372036854775807,
			exp: "9223372036854775807",
		},
		"min value": {
			src: -9223372036854775808,
			exp: "-9223372036854775808",
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got := NewDecFromInt64(spec.src)
			require.NoError(t, spec.expErr)
			assert.Equal(t, spec.exp, got.String())
		})
	}
}

func TestAdd(t *testing.T) {
	specs := map[string]struct {
		x           Dec
		y           Dec
		constraints []SetupConstraint
		exp         Dec
		expErr      error
	}{
		"zero add zero": {
			// 0 + 0 = 0
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(0),
			exp: NewDecFromInt64(0),
		},
		"zero add simple positive": {
			// 0 + 123 = 123
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(123),
		},
		"zero and simple negative": {
			// 0 + -123 = -123
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(-123),
			exp: NewDecFromInt64(-123),
		},
		"simple positive add simple positive": {
			// 123 + 123 = 246
			x:   NewDecFromInt64(123),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(246),
		},
		"simple negative add simple positive": {
			// -123 + 123 = 0
			x:   NewDecFromInt64(-123),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(0),
		},
		"simple negative add simple negative": {
			// -123 + -123 = -246
			x:   NewDecFromInt64(-123),
			y:   NewDecFromInt64(-123),
			exp: NewDecFromInt64(-246),
		},
		"valid decimal with decimal places add valid decimal with decimal places": {
			// 1.234 + 1.234 = 2.468
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(1234, -3),
			exp: NewDecWithPrec(2468, -3),
		},
		"valid decimal with decimal places and simple positive": {
			// 1.234 + 123 = 124.234
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecFromInt64(123),
			exp: NewDecWithPrec(124234, -3),
		},
		"valid decimal with decimal places and simple negative": {
			// 1.234 + -123 = 1.111
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecFromInt64(-123),
			exp: NewDecWithPrec(111, -3),
		},

		"valid decimal with decimal places add valid negative decimal with decimal places": {
			// 1.234 + -1.234 = 0
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(-1234, -3),
			exp: NewDecWithPrec(0, -3),
		},
		"valid negative decimal with decimal places add valid negative decimal with decimal places": {
			// -1.234 + -1.234 = -2.468
			x:   NewDecWithPrec(-1234, -3),
			y:   NewDecWithPrec(-1234, -3),
			exp: NewDecWithPrec(-2468, -3),
		},
		// "precision too high": {
		// 	// 10^34 + 10^34 = 2*10^34
		// 	x:           NewDecWithPrec(1, 36),
		// 	y:           NewDecWithPrec(1, 36),
		// 	constraints: []SetupConstraint{AssertMaxDecimals(34)},
		// 	expErr:      ErrInvalidDecString,
		// },
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got, gotErr := spec.x.Add(spec.y, spec.constraints...)
			if spec.expErr != nil {
				require.ErrorIs(t, gotErr, spec.expErr, got)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, spec.exp, got)
		})
	}
}

func TestSub(t *testing.T) {
	specs := map[string]struct {
		x           Dec
		y           Dec
		constraints []SetupConstraint
		exp         Dec
		expErr      error
	}{
		"zero minus zero": {
			// 0 + 0 = 0
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(0),
			exp: NewDecFromInt64(0),
		},
		"zero minus simple positive": {
			// 0 - 123 = -123
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(-123),
		},
		"zero minus simple negative": {
			// 0 - -123 = 123
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(-123),
			exp: NewDecFromInt64(123),
		},
		"simple positive minus simple positive": {
			// 123 - 123 = 0
			x:   NewDecFromInt64(123),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(0),
		},
		"simple negative minus simple positive": {
			// -123 + 123 = 0
			x:   NewDecFromInt64(-123),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(-246),
		},
		"simple negative minus simple negative": {
			// -123 - -123 = 0
			x:   NewDecFromInt64(-123),
			y:   NewDecFromInt64(-123),
			exp: NewDecFromInt64(0),
		},
		"valid decimal with decimal places add valid decimal with decimal places": {
			// 1.234 - 1.234 = 0.000
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(1234, -3),
			exp: NewDecWithPrec(0, -3),
		},
		"valid decimal with decimal places and simple positive": {
			// 1.234 - 123 = -121.766
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecFromInt64(123),
			exp: NewDecWithPrec(-121766, -3),
		},
		"valid decimal with decimal places and simple negative": {
			// 1.234 - -123 = 1.111
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecFromInt64(-123),
			exp: NewDecWithPrec(124234, -3),
		},
		"valid decimal with decimal places add valid negative decimal with decimal places": {
			// 1.234 - -1.234 = 2.468
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(-1234, -3),
			exp: NewDecWithPrec(2468, -3),
		},
		"valid negative decimal with decimal places add valid negative decimal with decimal places": {
			// -1.234 - -1.234 = 2.468
			x:   NewDecWithPrec(-1234, -3),
			y:   NewDecWithPrec(-1234, -3),
			exp: NewDecWithPrec(0, -3),
		},
		// "precision too high": {
		// 	// 10^34 - 10^34 = 2*10^34
		// 	x:           NewDecWithPrec(1, 36),
		// 	y:           NewDecWithPrec(1, 36),
		// 	constraints: []SetupConstraint{AssertMaxDecimals(34)},
		// 	expErr:      ErrInvalidDecString,
		// },
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got, gotErr := spec.x.Sub(spec.y, spec.constraints...)
			fmt.Println(got)
			if spec.expErr != nil {
				require.ErrorIs(t, gotErr, spec.expErr, got)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, spec.exp, got)
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
