package math

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDecFromString(t *testing.T) {
	specs := map[string]struct {
		src    string
		exp    Dec
		expErr error
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
			exp: NewDecWithExp(1234, -3),
		},
		"valid negative decimal": {
			src: "-1.234",
			exp: NewDecWithExp(-1234, -3),
		},
		"min decimal": {
			src: "-" + strings.Repeat("9", 34),
			exp: must(NewDecWithExp(-1, 34).Add(NewDecFromInt64(1))),
		},
		"max decimal": {
			src: strings.Repeat("9", 34),
			exp: must(NewDecWithExp(1, 34).Sub(NewDecFromInt64(1))),
		},
		"too big": {
			src:    strings.Repeat("9", 100_0000),
			expErr: ErrInvalidDec,
		},
		"too small": {
			src:    "-" + strings.Repeat("9", 100_0000),
			expErr: ErrInvalidDec,
		},
		"valid decimal with leading zero": {
			src: "01234",
			exp: NewDecWithExp(1234, 0),
		},
		"valid decimal without leading zero": {
			src: ".1234",
			exp: NewDecWithExp(1234, -4),
		},

		"valid decimal without trailing digits": {
			src: "123.",
			exp: NewDecWithExp(123, 0),
		},

		"valid negative decimal without leading zero": {
			src: "-.1234",
			exp: NewDecWithExp(-1234, -4),
		},
		"valid negative decimal without trailing digits": {
			src: "-123.",
			exp: NewDecWithExp(-123, 0),
		},
		"decimal with scientific notation": {
			src: "1.23e4",
			exp: NewDecWithExp(123, 2),
		},
		"decimal with upper case scientific notation": {
			src: "1.23E+4",
			exp: NewDecWithExp(123, 2),
		},
		"negative decimal with scientific notation": {
			src: "-1.23e4",
			exp: NewDecWithExp(-123, 2),
		},
		"exceed max exp 11E+1000000": {
			src:    "11E+1000000",
			expErr: ErrInvalidDec,
		},
		"exceed min exp 11E-1000000": {
			src:    "11E-1000000",
			expErr: ErrInvalidDec,
		},
		"exceed max exp 1E100001": {
			src:    "1E100001",
			expErr: ErrInvalidDec,
		},
		"exceed min exp 1E-100001": {
			src:    "1E-100001",
			expErr: ErrInvalidDec,
		},
		"empty string": {
			src:    "",
			expErr: ErrInvalidDec,
		},
		"NaN": {
			src:    "NaN",
			expErr: ErrInvalidDec,
		},
		"random string": {
			src:    "1foo",
			expErr: ErrInvalidDec,
		},
		"Infinity": {
			src:    "Infinity",
			expErr: ErrInvalidDec,
		},
		"Inf": {
			src:    "Inf",
			expErr: ErrInvalidDec,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got, gotErr := NewDecFromString(spec.src)
			if spec.expErr != nil {
				require.ErrorIs(t, gotErr, spec.expErr, got.String())
				return
			}
			require.NoError(t, gotErr)
			assert.True(t, spec.exp.Equal(got))
		})
	}
}

func TestNewDecFromInt64(t *testing.T) {
	specs := map[string]struct {
		src int64
		exp string
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
			src: math.MaxInt64,
			exp: "9223372036854.775807E+6",
		},
		"min value": {
			src: math.MinInt64,
			exp: "9223372036854.775808E+6",
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got := NewDecFromInt64(spec.src)
			assert.Equal(t, spec.exp, got.String())
		})
	}
}

func TestAdd(t *testing.T) {
	specs := map[string]struct {
		x      Dec
		y      Dec
		exp    Dec
		expErr error
	}{
		"0 + 0 = 0": {
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(0),
			exp: NewDecFromInt64(0),
		},
		"0 + 123 = 123": {
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(123),
		},
		"0 + -123 = -123": {
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(-123),
			exp: NewDecFromInt64(-123),
		},
		"123 + 123 = 246": {
			x:   NewDecFromInt64(123),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(246),
		},
		"-123 + 123 = 0": {
			x:   NewDecFromInt64(-123),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(0),
		},
		"-123 + -123 = -246": {
			x:   NewDecFromInt64(-123),
			y:   NewDecFromInt64(-123),
			exp: NewDecFromInt64(-246),
		},
		"1.234 + 1.234 = 2.468": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecWithExp(1234, -3),
			exp: NewDecWithExp(2468, -3),
		},
		"1.234 + 123 = 124.234": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecFromInt64(123),
			exp: NewDecWithExp(124234, -3),
		},
		"1.234 + -123 = -121.766": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecFromInt64(-123),
			exp: must(NewDecFromString("-121.766")),
		},
		"1.234 + -1.234 = 0": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecWithExp(-1234, -3),
			exp: NewDecWithExp(0, -3),
		},
		"-1.234 + -1.234 = -2.468": {
			x:   NewDecWithExp(-1234, -3),
			y:   NewDecWithExp(-1234, -3),
			exp: NewDecWithExp(-2468, -3),
		},
		"1e100000 + 9e900000 -> Err": {
			x:      NewDecWithExp(1, 100_000),
			y:      NewDecWithExp(9, 900_000),
			expErr: ErrInvalidDec,
		},
		"1e100000 + -9e900000 -> Err": {
			x:      NewDecWithExp(1, 100_000),
			y:      NewDecWithExp(9, 900_000),
			expErr: ErrInvalidDec,
		},
		"1e100000 + 1e^-1 -> err": {
			x:      NewDecWithExp(1, 100_000),
			y:      NewDecWithExp(1, -1),
			expErr: ErrInvalidDec,
		},
		"1e100000 + -1e^-1 -> err": {
			x:      NewDecWithExp(1, 100_000),
			y:      NewDecWithExp(-1, -1),
			expErr: ErrInvalidDec,
		},
		"1e100000 + 1 -> 100..1": {
			x:   NewDecWithExp(1, 100_000),
			y:   NewDecFromInt64(1),
			exp: must(NewDecWithExp(1, 100_000).Add(NewDecFromInt64(1))),
		},
		"1e100001 + 0 -> err": {
			x:      NewDecWithExp(1, 100_001),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDec,
		},
		"-1e100001 + 0 -> err": {
			x:      NewDecWithExp(1, -100_001),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDec,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got, gotErr := spec.x.Add(spec.y)
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
		x      Dec
		y      Dec
		exp    Dec
		expErr error
	}{
		"0 - 0 = 0": {
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(0),
			exp: NewDecFromInt64(0),
		},
		"0 - 123 = -123": {
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(-123),
		},
		"0 - -123 = 123": {
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(-123),
			exp: NewDecFromInt64(123),
		},
		"123 - 123 = 0": {
			x:   NewDecFromInt64(123),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(0),
		},
		"-123 - 123 = -246": {
			x:   NewDecFromInt64(-123),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(-246),
		},
		"-123 - -123 = 0": {
			x:   NewDecFromInt64(-123),
			y:   NewDecFromInt64(-123),
			exp: NewDecFromInt64(0),
		},
		"1.234 - 1.234 = 0.000": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecWithExp(1234, -3),
			exp: NewDecWithExp(0, -3),
		},
		"1.234 - 123 = -121.766": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecFromInt64(123),
			exp: NewDecWithExp(-121766, -3),
		},
		"1.234 - -123 = 124.234": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecFromInt64(-123),
			exp: NewDecWithExp(124234, -3),
		},
		"1.234 - -1.234 = 2.468": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecWithExp(-1234, -3),
			exp: NewDecWithExp(2468, -3),
		},
		"-1.234 - -1.234 = 2.468": {
			x:   NewDecWithExp(-1234, -3),
			y:   NewDecWithExp(-1234, -3),
			exp: NewDecWithExp(0, -3),
		},
		"1 - 0.999 = 0.001 - rounding after comma": {
			x:   NewDecFromInt64(1),
			y:   NewDecWithExp(999, -3),
			exp: NewDecWithExp(1, -3),
		},
		"1e100000 - 1^-1 -> Err": {
			x:      NewDecWithExp(1, 100_000),
			y:      NewDecWithExp(1, -1),
			expErr: ErrInvalidDec,
		},
		"1e100000 - 1^1-> Err": {
			x:      NewDecWithExp(1, 100_000),
			y:      NewDecWithExp(1, -1),
			expErr: ErrInvalidDec,
		},
		"upper exp limit exceeded": {
			x:      NewDecWithExp(1, 100_001),
			y:      NewDecWithExp(1, 100_001),
			expErr: ErrInvalidDec,
		},
		"lower exp limit exceeded": {
			x:      NewDecWithExp(1, -100_001),
			y:      NewDecWithExp(1, -100_001),
			expErr: ErrInvalidDec,
		},
		"1e100000 - 1 = 999..9": {
			x:   NewDecWithExp(1, 100_000),
			y:   NewDecFromInt64(1),
			exp: must(NewDecFromString(strings.Repeat("9", 100_000))),
		},
		"1e100000 - 0 = 1e100000": {
			x:   NewDecWithExp(1, 100_000),
			y:   NewDecFromInt64(0),
			exp: must(NewDecFromString("1e100000")),
		},
		"1e100001 - 0 -> err": {
			x:      NewDecWithExp(1, 100_001),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDec,
		},
		"1e100000 - -1 -> 100..1": {
			x:      NewDecWithExp(1, 100_000),
			y:      must(NewDecFromString("-9e100000")),
			expErr: ErrInvalidDec,
		},
		"1e-100000 - 0 = 1e-100000": {
			x:   NewDecWithExp(1, -100_000),
			y:   NewDecFromInt64(0),
			exp: must(NewDecFromString("1e-100000")),
		},
		"1e-100001 - 0 -> err": {
			x:      NewDecWithExp(1, -100_001),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDec,
		},
		"1e-100000 - -1 -> 0.000..01": {
			x:   NewDecWithExp(1, -100_000),
			y:   NewDecFromInt64(-1),
			exp: must(NewDecFromString("1." + strings.Repeat("0", 99999) + "1")),
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got, gotErr := spec.x.Sub(spec.y)
			if spec.expErr != nil {
				require.ErrorIs(t, gotErr, spec.expErr)
				return
			}
			require.NoError(t, gotErr)
			assert.True(t, spec.exp.Equal(got), got.String())
		})
	}
}

func TestQuo(t *testing.T) {
	specs := map[string]struct {
		src    string
		x      Dec
		y      Dec
		exp    Dec
		expErr error
	}{
		"0 / 0 -> Err": {
			x:      NewDecFromInt64(0),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDec,
		},
		" 0 / 123 = 0": {
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(0),
		},
		"123 / 0 = 0": {
			x:      NewDecFromInt64(123),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDec,
		},
		"-123 / 0 = 0": {
			x:      NewDecFromInt64(-123),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDec,
		},
		"123 / 123 = 1": {
			x:   NewDecFromInt64(123),
			y:   NewDecFromInt64(123),
			exp: must(NewDecFromString("1.000000000000000000000000000000000")),
		},
		"-123 / 123 = -1": {
			x:   NewDecFromInt64(-123),
			y:   NewDecFromInt64(123),
			exp: must(NewDecFromString("-1.000000000000000000000000000000000")),
		},
		"-123 / -123 = 1": {
			x:   NewDecFromInt64(-123),
			y:   NewDecFromInt64(-123),
			exp: must(NewDecFromString("1.000000000000000000000000000000000")),
		},
		"1.234 / 1.234 = 1": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecWithExp(1234, -3),
			exp: must(NewDecFromString("1.000000000000000000000000000000000")),
		},
		"-1.234 / 1234 = -1": {
			x:   NewDecWithExp(-1234, -3),
			y:   NewDecWithExp(1234, -3),
			exp: must(NewDecFromString("-1.000000000000000000000000000000000")),
		},
		"1.234 / -123 = 1.0100": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecFromInt64(-123),
			exp: must(NewDecFromString("-0.01003252032520325203252032520325203")),
		},
		"1.234 / -1.234 = -1": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecWithExp(-1234, -3),
			exp: must(NewDecFromString("-1.000000000000000000000000000000000")),
		},
		"-1.234 / -1.234 = 1": {
			x:   NewDecWithExp(-1234, -3),
			y:   NewDecWithExp(-1234, -3),
			exp: must(NewDecFromString("1.000000000000000000000000000000000")),
		},
		"3 / -9 = -0.3333...3 - round down": {
			x:   NewDecFromInt64(3),
			y:   NewDecFromInt64(-9),
			exp: must(NewDecFromString("-0.3333333333333333333333333333333333")),
		},
		"4 / 9 = 0.4444...4 - round down": {
			x:   NewDecFromInt64(4),
			y:   NewDecFromInt64(9),
			exp: must(NewDecFromString("0.4444444444444444444444444444444444")),
		},
		"5 / 9 = 0.5555...6 - round up": {
			x:   NewDecFromInt64(5),
			y:   NewDecFromInt64(9),
			exp: must(NewDecFromString("0.5555555555555555555555555555555556")),
		},
		"6 / 9 = 0.6666...7 - round up": {
			x:   NewDecFromInt64(6),
			y:   NewDecFromInt64(9),
			exp: must(NewDecFromString("0.6666666666666666666666666666666667")),
		},
		"7 / 9 = 0.7777...8 - round up": {
			x:   NewDecFromInt64(7),
			y:   NewDecFromInt64(9),
			exp: must(NewDecFromString("0.7777777777777777777777777777777778")),
		},
		"8 / 9 = 0.8888...9 - round up": {
			x:   NewDecFromInt64(8),
			y:   NewDecFromInt64(9),
			exp: must(NewDecFromString("0.8888888888888888888888888888888889")),
		},
		"9e-34 / 10 = 9e-35 - no rounding": {
			x:   NewDecWithExp(9, -34),
			y:   NewDecFromInt64(10),
			exp: must(NewDecFromString("9e-35")),
		},
		"9e-35 / 10 = 9e-36 - no rounding": {
			x:   NewDecWithExp(9, -35),
			y:   NewDecFromInt64(10),
			exp: must(NewDecFromString("9e-36")),
		},
		"high precision - min/0.1": {
			x:   NewDecWithExp(1, -100_000),
			y:   NewDecWithExp(1, -1),
			exp: NewDecWithExp(1, -99_999),
		},
		"high precision - min/1": {
			x:   NewDecWithExp(1, -100_000),
			y:   NewDecWithExp(1, 0),
			exp: NewDecWithExp(1, -100_000),
		},
		"high precision - min/10": {
			x:      NewDecWithExp(1, -100_000),
			y:      NewDecWithExp(1, 1),
			expErr: ErrInvalidDec,
		},
		"high precision - <_min/0.1": {
			x:   NewDecWithExp(1, -100_001),
			y:   NewDecWithExp(1, -1),
			exp: NewDecWithExp(1, -100_000),
		},
		"high precision - <_min/1": {
			x:      NewDecWithExp(1, -100_001),
			y:      NewDecWithExp(1, 0),
			expErr: ErrInvalidDec,
		},
		"high precision - <_min/10": {
			x:      NewDecWithExp(1, -100_001),
			y:      NewDecWithExp(1, 1),
			expErr: ErrInvalidDec,
		},
		"high precision - min/-0.1": {
			x:   NewDecWithExp(1, -100_000),
			y:   NewDecWithExp(-1, -1),
			exp: NewDecWithExp(-1, -99_999),
		},
		"high precision - min/-1": {
			x:   NewDecWithExp(1, -100_000),
			y:   NewDecWithExp(-1, 0),
			exp: NewDecWithExp(-1, -100_000),
		},
		"high precision - min/-10": {
			x:      NewDecWithExp(1, -100_000),
			y:      NewDecWithExp(-1, 1),
			expErr: ErrInvalidDec,
		},
		"high precision - <_min/-0.1": {
			x:   NewDecWithExp(1, -100_001),
			y:   NewDecWithExp(-1, -1),
			exp: NewDecWithExp(-1, -100_000),
		},
		"high precision - <_min/-1": {
			x:      NewDecWithExp(1, -100_001),
			y:      NewDecWithExp(-1, 0),
			expErr: ErrInvalidDec,
		},
		"high precision - <_min/-10": {
			x:      NewDecWithExp(1, -100_001),
			y:      NewDecWithExp(-1, 1),
			expErr: ErrInvalidDec,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got, gotErr := spec.x.Quo(spec.y)
			if spec.expErr != nil {
				require.ErrorIs(t, gotErr, spec.expErr)
				return
			}
			require.NoError(t, gotErr)
			last35 := func(s string) string {
				var x int
				if len(s) < 36 {
					x = 0
				} else {
					x = len(s) - 36
				}
				return fmt.Sprintf("%s(%d)", s[x:], len(s))
			}
			gotReduced, _ := got.Reduce()
			assert.True(t, spec.exp.Equal(gotReduced), "exp %s, got: %s", last35(spec.exp.String()), last35(gotReduced.String()))
		})
	}
}

func TestQuoExact(t *testing.T) {
	specs := map[string]struct {
		src    string
		x      Dec
		y      Dec
		exp    Dec
		expErr error
	}{
		"0 / 0 -> Err": {
			x:      NewDecFromInt64(0),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDec,
		},
		" 0 / 123 = 0": {
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(0),
		},
		"123 / 0 -> Err": {
			x:      NewDecFromInt64(123),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDec,
		},
		"-123 / 0 -> Err": {
			x:      NewDecFromInt64(-123),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDec,
		},
		"123 / 123 = 1": {
			x:   NewDecFromInt64(123),
			y:   NewDecFromInt64(123),
			exp: must(NewDecFromString("1.000000000000000000000000000000000")),
		},
		"-123 / 123 = 1": {
			x:   NewDecFromInt64(-123),
			y:   NewDecFromInt64(123),
			exp: must(NewDecFromString("-1.000000000000000000000000000000000")),
		},
		"-123 / -123 = 1": {
			x:   NewDecFromInt64(-123),
			y:   NewDecFromInt64(-123),
			exp: must(NewDecFromString("1.000000000000000000000000000000000")),
		},
		"1.234 / 1.234 = 1": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecWithExp(1234, -3),
			exp: must(NewDecFromString("1.000000000000000000000000000000000")),
		},
		"-1.234 / 1.234 = -1": {
			x:   NewDecWithExp(-1234, -3),
			y:   NewDecWithExp(1234, -3),
			exp: must(NewDecFromString("-1.000000000000000000000000000000000")),
		},
		"1.234 / -123 -> Err": {
			x:      NewDecWithExp(1234, -3),
			y:      NewDecFromInt64(-123),
			expErr: ErrUnexpectedRounding,
		},
		"1.234 / -1.234 = -1": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecWithExp(-1234, -3),
			exp: must(NewDecFromString("-1.000000000000000000000000000000000")),
		},
		"-1.234 / -1.234 = 1": {
			x:   NewDecWithExp(-1234, -3),
			y:   NewDecWithExp(-1234, -3),
			exp: must(NewDecFromString("1.000000000000000000000000000000000")),
		},
		"3 / -9 -> Err": {
			x:      NewDecFromInt64(3),
			y:      NewDecFromInt64(-9),
			expErr: ErrUnexpectedRounding,
		},
		"4 / 9 -> Err": {
			x:      NewDecFromInt64(4),
			y:      NewDecFromInt64(9),
			expErr: ErrUnexpectedRounding,
		},
		"5 / 9 -> Err": {
			x:      NewDecFromInt64(5),
			y:      NewDecFromInt64(9),
			expErr: ErrUnexpectedRounding,
		},
		"6 / 9 -> Err": {
			x:      NewDecFromInt64(6),
			y:      NewDecFromInt64(9),
			expErr: ErrUnexpectedRounding,
		},
		"7 / 9 -> Err": {
			x:      NewDecFromInt64(7),
			y:      NewDecFromInt64(9),
			expErr: ErrUnexpectedRounding,
		},
		"8 / 9 -> Err": {
			x:      NewDecFromInt64(8),
			y:      NewDecFromInt64(9),
			expErr: ErrUnexpectedRounding,
		},
		"9e-34 / 10 = 9e-35 - no rounding": {
			x:   NewDecWithExp(9, -34),
			y:   NewDecFromInt64(10),
			exp: must(NewDecFromString("0.00000000000000000000000000000000009000000000000000000000000000000000")),
		},
		"9e-35 / 10 = 9e-36 - no rounding": {
			x:   NewDecWithExp(9, -35),
			y:   NewDecFromInt64(10),
			exp: must(NewDecFromString("9e-36")),
		},
		"high precision - min/0.1": {
			x:   NewDecWithExp(1, -100_000),
			y:   NewDecWithExp(1, -1),
			exp: NewDecWithExp(1, -99_999),
		},
		"high precision - min/1": {
			x:   NewDecWithExp(1, -100_000),
			y:   NewDecWithExp(1, 0),
			exp: NewDecWithExp(1, -100_000),
		},
		"high precision - min/10": {
			x:      NewDecWithExp(1, -100_000),
			y:      NewDecWithExp(1, 1),
			expErr: ErrInvalidDec,
		},
		"high precision - <_min/0.1": {
			x:   NewDecWithExp(1, -100_001),
			y:   NewDecWithExp(1, -1),
			exp: NewDecWithExp(1, -100_000),
		},
		"high precision - <_min/1": {
			x:      NewDecWithExp(1, -100_001),
			y:      NewDecWithExp(1, 0),
			expErr: ErrInvalidDec,
		},
		"high precision - <_min/10 -> Err": {
			x:      NewDecWithExp(1, -100_001),
			y:      NewDecWithExp(1, 1),
			expErr: ErrInvalidDec,
		},
		"high precision - min/-0.1": {
			x:   NewDecWithExp(1, -100_000),
			y:   NewDecWithExp(-1, -1),
			exp: NewDecWithExp(-1, -99_999),
		},
		"high precision - min/-1": {
			x:   NewDecWithExp(1, -100_000),
			y:   NewDecWithExp(-1, 0),
			exp: NewDecWithExp(-1, -100_000),
		},
		"high precision - min/-10 -> Err": {
			x:      NewDecWithExp(1, -100_000),
			y:      NewDecWithExp(-1, 1),
			expErr: ErrInvalidDec,
		},
		"high precision - <_min/-0.1": {
			x:   NewDecWithExp(1, -100_001),
			y:   NewDecWithExp(-1, -1),
			exp: NewDecWithExp(-1, -100_000),
		},
		"high precision - <_min/-1": {
			x:      NewDecWithExp(1, -100_001),
			y:      NewDecWithExp(-1, 0),
			expErr: ErrInvalidDec,
		},
		"high precision - <_min/-10": {
			x:      NewDecWithExp(1, -100_001),
			y:      NewDecWithExp(-1, 1),
			expErr: ErrInvalidDec,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got, gotErr := spec.x.QuoExact(spec.y)

			if spec.expErr != nil {
				require.ErrorIs(t, gotErr, spec.expErr)
				return
			}
			require.NoError(t, gotErr)
			assert.True(t, spec.exp.Equal(got))
		})
	}
}

func TestQuoInteger(t *testing.T) {
	specs := map[string]struct {
		src    string
		x      Dec
		y      Dec
		exp    Dec
		expErr error
	}{
		"0 / 0 -> Err": {
			x:      NewDecFromInt64(0),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDec,
		},
		" 0 / 123 = 0": {
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(0),
		},
		"123 / 0 -> Err": {
			x:      NewDecFromInt64(123),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDec,
		},
		"-123 / -> Err": {
			x:      NewDecFromInt64(-123),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDec,
		},
		"123 / 123 = 1": {
			x:   NewDecFromInt64(123),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(1),
		},
		"-123 / 123 = -1": {
			x:   NewDecFromInt64(-123),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(-1),
		},
		"-123 / -123 = 1": {
			x:   NewDecFromInt64(-123),
			y:   NewDecFromInt64(-123),
			exp: NewDecFromInt64(1),
		},
		"1.234 / 1.234": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecWithExp(1234, -3),
			exp: NewDecFromInt64(1),
		},
		"-1.234 / 1234 = -121.766": {
			x:   NewDecWithExp(-1234, -3),
			y:   NewDecWithExp(1234, -3),
			exp: NewDecFromInt64(-1),
		},
		"1.234 / -1.234 = 2.468": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecWithExp(-1234, -3),
			exp: NewDecFromInt64(-1),
		},
		"-1.234 / -1.234 = 1": {
			x:   NewDecWithExp(-1234, -3),
			y:   NewDecWithExp(-1234, -3),
			exp: NewDecFromInt64(1),
		},
		"3 / -9 = 0": {
			x:   NewDecFromInt64(3),
			y:   NewDecFromInt64(-9),
			exp: must(NewDecFromString("0")),
		},
		"8 / 9 = 0": {
			x:   NewDecFromInt64(8),
			y:   NewDecFromInt64(9),
			exp: must(NewDecFromString("0")),
		},
		"high precision - min/0.1": {
			x:   NewDecWithExp(1, -100_000),
			y:   NewDecWithExp(1, -1),
			exp: NewDecFromInt64(0),
		},
		"high precision - <_min/-1 -> Err": {
			x:      NewDecWithExp(1, -100_001),
			y:      NewDecWithExp(-1, 0),
			expErr: ErrInvalidDec,
		},
		"high precision - <_min/-10 -> Err": {
			x:      NewDecWithExp(1, -100_001),
			y:      NewDecWithExp(-1, 1),
			expErr: ErrInvalidDec,
		},
		"1e000 / 1 -> Err": {
			x:      NewDecWithExp(1, 100_000),
			y:      NewDecFromInt64(1),
			expErr: ErrInvalidDec,
		},
		"1e100000 - 1^1 -> Err": {
			x:      NewDecWithExp(1, 100_000),
			y:      NewDecWithExp(1, -1),
			expErr: ErrInvalidDec,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got, gotErr := spec.x.QuoInteger(spec.y)

			if spec.expErr != nil {
				require.ErrorIs(t, gotErr, spec.expErr)
				return
			}
			require.NoError(t, gotErr)
			assert.True(t, spec.exp.Equal(got))
		})
	}
}

func TestModulo(t *testing.T) {
	specs := map[string]struct {
		x      Dec
		y      Dec
		exp    Dec
		expErr error
	}{
		"0 / 123 = 0": {
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(0),
		},
		"123 / 10 = 3": {
			x:   NewDecFromInt64(123),
			y:   NewDecFromInt64(10),
			exp: NewDecFromInt64(3),
		},
		"123 / -10 = 3": {
			x:   NewDecFromInt64(123),
			y:   NewDecFromInt64(-10),
			exp: NewDecFromInt64(3),
		},
		"-123 / 10 = -3": {
			x:   NewDecFromInt64(-123),
			y:   NewDecFromInt64(10),
			exp: NewDecFromInt64(-3),
		},
		"1.234 / 1 = 0.234": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecFromInt64(1),
			exp: NewDecWithExp(234, -3),
		},
		"1.234 / 0.1 = 0.034": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecWithExp(1, -1),
			exp: NewDecWithExp(34, -3),
		},
		"1.234 / 1.1 = 0.134": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecWithExp(11, -1),
			exp: NewDecWithExp(134, -3),
		},
		"10 / 0 -> Err": {
			x:      NewDecFromInt64(10),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDec,
		},
		"-1e0000 / 9e0000 = 1e0000": {
			x:   NewDecWithExp(-1, 100_000),
			y:   NewDecWithExp(9, 100_000),
			exp: NewDecWithExp(-1, 100_000),
		},
		"1e0000 / 9e0000 = 1e0000": {
			x:   NewDecWithExp(1, 100_000),
			y:   NewDecWithExp(9, 100_000),
			exp: NewDecWithExp(1, 100_000),
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got, gotErr := spec.x.Modulo(spec.y)

			if spec.expErr != nil {
				require.ErrorIs(t, gotErr, spec.expErr)
				return
			}
			require.NoError(t, gotErr)
			assert.True(t, spec.exp.Equal(got))
		})
	}
}

func TestNumDecimalPlaces(t *testing.T) {
	specs := map[string]struct {
		src Dec
		exp uint32
	}{
		"integer": {
			src: NewDecFromInt64(123),
			exp: 0,
		},
		"one decimal place": {
			src: NewDecWithExp(1234, -1),
			exp: 1,
		},
		"two decimal places": {
			src: NewDecWithExp(12345, -2),
			exp: 2,
		},
		"three decimal places": {
			src: NewDecWithExp(123456, -3),
			exp: 3,
		},
		"trailing zeros": {
			src: NewDecWithExp(123400, -4),
			exp: 4,
		},
		"zero value": {
			src: NewDecFromInt64(0),
			exp: 0,
		},
		"negative value": {
			src: NewDecWithExp(-12345, -3),
			exp: 3,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got := spec.src.NumDecimalPlaces()
			assert.Equal(t, spec.exp, got)
		})
	}
}

func TestCmp(t *testing.T) {
	specs := map[string]struct {
		x   Dec
		y   Dec
		exp int
	}{
		"0 == 0": {
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(0),
			exp: 0,
		},
		"0 < 123 = -1": {
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(123),
			exp: -1,
		},
		"123 > 0 = 1": {
			x:   NewDecFromInt64(123),
			y:   NewDecFromInt64(0),
			exp: 1,
		},
		"-123 < 0 = -1": {
			x:   NewDecFromInt64(-123),
			y:   NewDecFromInt64(0),
			exp: -1,
		},
		"123 == 123": {
			x:   NewDecFromInt64(123),
			y:   NewDecFromInt64(123),
			exp: 0,
		},
		"-123 == -123": {
			x:   NewDecFromInt64(-123),
			y:   NewDecFromInt64(-123),
			exp: 0,
		},
		"1.234 == 1.234": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecWithExp(1234, -3),
			exp: 0,
		},
		"1.234 > 1.233": {
			x:   NewDecWithExp(1234, -3),
			y:   NewDecWithExp(1233, -3),
			exp: 1,
		},
		"1.233 < 1.234": {
			x:   NewDecWithExp(1233, -3),
			y:   NewDecWithExp(1234, -3),
			exp: -1,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got := spec.x.Cmp(spec.y)
			assert.Equal(t, spec.exp, got)
		})
	}
}

func TestReduce(t *testing.T) {
	specs := map[string]struct {
		src       string
		exp       string
		decPlaces int
	}{
		"positive value": {
			src:       "10",
			exp:       "10",
			decPlaces: 1,
		},
		"negative value": {
			src:       "-10",
			exp:       "-10",
			decPlaces: 1,
		},
		"positive decimal": {
			src:       "1.30000",
			exp:       "1.3",
			decPlaces: 4,
		},
		"negative decimal": {
			src:       "-1.30000",
			exp:       "-1.3",
			decPlaces: 4,
		},
		"zero decimal and decimal places": {
			src:       "0.00000",
			exp:       "0",
			decPlaces: 0,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			src := must(NewDecFromString(spec.src))
			got, gotZerosRemoved := src.Reduce()
			assert.Equal(t, spec.decPlaces, gotZerosRemoved)
			assert.Equal(t, spec.exp, got.String())
		})
	}
}

func TestMulExact(t *testing.T) {
	specs := map[string]struct {
		x      Dec
		y      Dec
		exp    Dec
		expErr error
	}{
		"200 * 200 = 40000": {
			x:   NewDecFromInt64(200),
			y:   NewDecFromInt64(200),
			exp: NewDecFromInt64(40000),
		},
		"-200 * -200 = 40000": {
			x:   NewDecFromInt64(-200),
			y:   NewDecFromInt64(-200),
			exp: NewDecFromInt64(40000),
		},
		"-100 * -100 = 10000": {
			x:   NewDecFromInt64(-100),
			y:   NewDecFromInt64(-100),
			exp: NewDecFromInt64(10000),
		},
		"0 * 0 = 0": {
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(0),
			exp: NewDecFromInt64(0),
		},
		"1.1 * 1.1 = 1.21": {
			x:   NewDecWithExp(11, -1),
			y:   NewDecWithExp(11, -1),
			exp: NewDecWithExp(121, -2),
		},
		"1.000 * 1.000 = 1.000000": {
			x:   NewDecWithExp(1000, -3),
			y:   NewDecWithExp(1000, -3),
			exp: must(NewDecFromString("1.000000")),
		},
		"0.0000001 * 0.0000001 = 0": {
			x:   NewDecWithExp(0o0000001, -7),
			y:   NewDecWithExp(0o0000001, -7),
			exp: NewDecWithExp(1, -14),
		},
		"0.12345678901234567890123456789012345 * 1": {
			x:      must(NewDecFromString("0.12345678901234567890123456789012345")),
			y:      NewDecWithExp(1, 0),
			expErr: ErrUnexpectedRounding,
		},
		"0.12345678901234567890123456789012345 * 0": {
			x:   must(NewDecFromString("0.12345678901234567890123456789012345")),
			y:   NewDecFromInt64(0),
			exp: NewDecFromInt64(0),
		},
		"0.12345678901234567890123456789012345 * 0.1": {
			x:      must(NewDecFromString("0.12345678901234567890123456789012345")),
			y:      NewDecWithExp(1, -1),
			expErr: ErrUnexpectedRounding,
		},
		"1000001 * 1.000001 = 1000002.000001": {
			x:   NewDecFromInt64(1000001),
			y:   NewDecWithExp(1000001, -6),
			exp: must(NewDecFromString("1000002.000001")),
		},
		"1000001 * 1000000 = 1000001000000 ": {
			x:   NewDecFromInt64(1000001),
			y:   NewDecFromInt64(1000000),
			exp: NewDecFromInt64(1000001000000),
		},
		"1e0000 * 1e0000 -> Err": {
			x:      NewDecWithExp(1, 100_000),
			y:      NewDecWithExp(1, 100_000),
			expErr: ErrInvalidDec,
		},
		"1e0000 * 1 = 1e0000": {
			x:   NewDecWithExp(1, 100_000),
			y:   NewDecWithExp(1, 0),
			exp: NewDecWithExp(1, 100_000),
		},
		"1e100000 * 9 = 9e100000": {
			x:   NewDecWithExp(1, 100_000),
			y:   NewDecFromInt64(9),
			exp: NewDecWithExp(9, 100_000),
		},
		"1e100000 * 10 = err": {
			x:      NewDecWithExp(1, 100_000),
			y:      NewDecWithExp(1, 1),
			expErr: ErrInvalidDec,
		},
		"1e0000 * -1 = -1e0000": {
			x:   NewDecWithExp(1, 100_000),
			y:   NewDecWithExp(-1, 0),
			exp: NewDecWithExp(-1, 100_000),
		},
		"1e100000 * -9 = 9e100000": {
			x:   NewDecWithExp(1, 100_000),
			y:   NewDecFromInt64(-9),
			exp: NewDecWithExp(-9, 100_000),
		},
		"1e100000 * -10 = err": {
			x:      NewDecWithExp(1, 100_000),
			y:      NewDecWithExp(-1, 1),
			expErr: ErrInvalidDec,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got, gotErr := spec.x.MulExact(spec.y)
			if spec.expErr != nil {
				require.ErrorIs(t, gotErr, spec.expErr, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.True(t, spec.exp.Equal(got), "exp: %s, got: %s", spec.exp.Text('E'), got.Text('E'))
		})
	}
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
		{"12345.6", "", ErrNonIntegral},
	}
	for idx, tc := range tcs {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			a, err := NewDecFromString(tc.intStr)
			require.NoError(t, err)
			b, err := a.BigInt()
			if tc.isError == nil {
				require.NoError(t, err, "test_%d", idx)
				require.Equal(t, tc.out, b.String(), "test_%d", idx)
			} else {
				require.ErrorIs(t, err, tc.isError, "test_%d", idx)
			}
		})
	}
}

func TestToSdkInt(t *testing.T) {
	maxIntValue := "115792089237316195423570985008687907853269984665640564039457584007913129639935" // 2^256 -1
	tcs := []struct {
		src    string
		exp    string
		expErr bool
	}{
		{src: maxIntValue, exp: maxIntValue},
		{src: "1000000000000000000000000000000000000123456789.00000001", exp: "1000000000000000000000000000000000000123456789"},
		{src: "123.456e6", exp: "123456000"},
		{src: "123.456e1", exp: "1234"},
		{src: "123.456", exp: "123"},
		{src: "123.956", exp: "123"},
		{src: "-123.456", exp: "-123"},
		{src: "-123.956", exp: "-123"},
		{src: "-0.956", exp: "0"},
		{src: "-0.9", exp: "0"},
		{src: "1E-100000", exp: "0"},
		{src: "115792089237316195423570985008687907853269984665640564039457584007913129639936", expErr: true}, // 2^256
		{src: "1E100000", expErr: true},
	}
	for _, tc := range tcs {
		t.Run(fmt.Sprint(tc.src), func(t *testing.T) {
			a, err := NewDecFromString(tc.src)
			require.NoError(t, err)
			b, gotErr := a.SdkIntTrim()
			if tc.expErr {
				require.Error(t, gotErr, "value: %s", b.String())
				return
			}
			require.NoError(t, gotErr)
			require.Equal(t, tc.exp, b.String())
		})
	}
}

func TestInfDecString(t *testing.T) {
	_, err := NewDecFromString("iNf")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidDec)
}

func must[T any](r T, err error) T {
	if err != nil {
		panic(err)
	}
	return r
}

func TestMarshalUnmarshal(t *testing.T) {
	specs := map[string]struct {
		x      Dec
		exp    string
		expErr error
	}{
		"Zero value": {
			x:   NewDecFromInt64(0),
			exp: "0",
		},
		"-0": {
			x:   NewDecFromInt64(-0),
			exp: "0",
		},
		"1 decimal place": {
			x:   must(NewDecFromString("0.1")),
			exp: "0.1",
		},
		"2 decimal places": {
			x:   must(NewDecFromString("0.01")),
			exp: "0.01",
		},
		"3 decimal places": {
			x:   must(NewDecFromString("0.001")),
			exp: "0.001",
		},
		"4 decimal places": {
			x:   must(NewDecFromString("0.0001")),
			exp: "0.0001",
		},
		"5 decimal places": {
			x:   must(NewDecFromString("0.00001")),
			exp: "0.00001",
		},
		"6 decimal places": {
			x:   must(NewDecFromString("0.000001")),
			exp: "1E-6",
		},
		"7 decimal places": {
			x:   must(NewDecFromString("0.0000001")),
			exp: "1E-7",
		},
		"1": {
			x:   must(NewDecFromString("1")),
			exp: "1",
		},
		"12": {
			x:   must(NewDecFromString("12")),
			exp: "12",
		},
		"123": {
			x:   must(NewDecFromString("123")),
			exp: "123",
		},
		"1234": {
			x:   must(NewDecFromString("1234")),
			exp: "1234",
		},
		"12345": {
			x:   must(NewDecFromString("12345")),
			exp: "12345",
		},
		"123456": {
			x:   must(NewDecFromString("123456")),
			exp: "123456",
		},
		"1234567": {
			x:   must(NewDecFromString("1234567")),
			exp: "1.234567E+6",
		},
		"12345678": {
			x:   must(NewDecFromString("12345678")),
			exp: "12.345678E+6",
		},
		"123456789": {
			x:   must(NewDecFromString("123456789")),
			exp: "123.456789E+6",
		},
		"1234567890": {
			x:   must(NewDecFromString("1234567890")),
			exp: "123.456789E+7",
		},
		"12345678900": {
			x:   must(NewDecFromString("12345678900")),
			exp: "123.456789E+8",
		},
		"negative 1 with negative exponent": {
			x:   must(NewDecFromString("-1.000001")),
			exp: "-1.000001",
		},
		"-1.0000001 - negative 1 with negative exponent": {
			x:   must(NewDecFromString("-1.0000001")),
			exp: "-1.0000001",
		},
		"3 decimal places before the comma": {
			x:   must(NewDecFromString("100")),
			exp: "100",
		},
		"4 decimal places before the comma": {
			x:   must(NewDecFromString("1000")),
			exp: "1000",
		},
		"5 decimal places before the comma": {
			x:   must(NewDecFromString("10000")),
			exp: "10000",
		},
		"6 decimal places before the comma": {
			x:   must(NewDecFromString("100000")),
			exp: "100000",
		},
		"7 decimal places before the comma": {
			x:   must(NewDecFromString("1000000")),
			exp: "1E+6",
		},
		"1e100000": {
			x:   NewDecWithExp(1, 100_000),
			exp: "1E+100000",
		},
		"1.1e100000": {
			x:   must(NewDecFromString("1.1e100000")),
			exp: "1.1E+100000",
		},
		"1e100001": {
			x:      NewDecWithExp(1, 100_001),
			expErr: ErrInvalidDec,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			marshaled, gotErr := spec.x.Marshal()
			if spec.expErr != nil {
				require.ErrorIs(t, gotErr, spec.expErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.exp, string(marshaled))
			// and backwards
			unmarshalledDec := new(Dec)
			require.NoError(t, unmarshalledDec.Unmarshal(marshaled))
			assert.Equal(t, spec.exp, unmarshalledDec.String())
			assert.True(t, spec.x.Equal(*unmarshalledDec))
		})
	}
}
