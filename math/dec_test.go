package math

import (
	"fmt"
	"math"
	"strconv"
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
			src: strings.Repeat("9", 34),
			exp: must(NewDecWithPrec(1, 34).Sub(NewDecFromInt64(1))),
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
		src    int64
		exp    string
		expErr error
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
			src: math.MaxInt32,
			exp: strconv.Itoa(math.MaxInt32),
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
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(1234, -3),
			exp: NewDecWithPrec(2468, -3),
		},
		"1.234 + 123 = 124.234": {
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecFromInt64(123),
			exp: NewDecWithPrec(124234, -3),
		},
		"1.234 + -123 = -121.766": {
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecFromInt64(-123),
			exp: must(NewDecFromString("-121.766")),
		},
		"1.234 + -1.234 = 0": {
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(-1234, -3),
			exp: NewDecWithPrec(0, -3),
		},
		"-1.234 + -1.234 = -2.468": {
			x:   NewDecWithPrec(-1234, -3),
			y:   NewDecWithPrec(-1234, -3),
			exp: NewDecWithPrec(-2468, -3),
		},
		"1e100000 + 9e900000 -> Err": {
			x:      NewDecWithPrec(1, 100_000),
			y:      NewDecWithPrec(9, 900_000),
			expErr: ErrInvalidDec,
		},
		"1e100000 + -9e900000 -> Err": {
			x:      NewDecWithPrec(1, 100_000),
			y:      NewDecWithPrec(9, 900_000),
			expErr: ErrInvalidDec,
		},
		"1e100000 + 1e^-1 -> err": {
			x:      NewDecWithPrec(1, 100_000),
			y:      NewDecWithPrec(1, -1),
			expErr: ErrInvalidDec,
		},
		"1e100000 + -1e^-1 -> err": {
			x:      NewDecWithPrec(1, 100_000),
			y:      NewDecWithPrec(-1, -1),
			expErr: ErrInvalidDec,
		},
		"1e100000 + 1 -> 100..1": {
			x:   NewDecWithPrec(1, 100_000),
			y:   NewDecFromInt64(1),
			exp: must(NewDecWithPrec(1, 100_000).Add(NewDecFromInt64(1))),
		},
		"1e100000 + 0 -> err": {
			x:      NewDecWithPrec(1, 100_001),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDec,
		},
		"-1e100000 + 0 -> err": {
			x:      NewDecWithPrec(-1, 100_001),
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
		x           Dec
		y           Dec
		exp         Dec
		expErr      error
		constraints []SetupConstraint
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
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(1234, -3),
			exp: NewDecWithPrec(0, -3),
		},
		"1.234 - 123 = -121.766": {
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecFromInt64(123),
			exp: NewDecWithPrec(-121766, -3),
		},
		"1.234 - -123 = 124.234": {
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecFromInt64(-123),
			exp: NewDecWithPrec(124234, -3),
		},
		"1.234 - -1.234 = 2.468": {
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(-1234, -3),
			exp: NewDecWithPrec(2468, -3),
		},
		"-1.234 - -1.234 = 2.468": {
			x:   NewDecWithPrec(-1234, -3),
			y:   NewDecWithPrec(-1234, -3),
			exp: NewDecWithPrec(0, -3),
		},
		"1 - 0.999 = 0.001 - rounding after comma": {
			x:   NewDecFromInt64(1),
			y:   NewDecWithPrec(999, -3),
			exp: NewDecWithPrec(1, -3),
		},
		"1e100000 - 1^-1 -> Err": {
			x:      NewDecWithPrec(1, 100_000),
			y:      NewDecWithPrec(1, -1),
			expErr: ErrInvalidDec,
		},
		"1e100000 - 1^1-> Err": {
			x:      NewDecWithPrec(1, 100_000),
			y:      NewDecWithPrec(1, -1),
			expErr: ErrInvalidDec,
		},
		"upper exp limit exceeded": {
			x:      NewDecWithPrec(1, 100_001),
			y:      NewDecWithPrec(1, 100_001),
			expErr: ErrInvalidDec,
		},
		"lower exp limit exceeded": {
			x:      NewDecWithPrec(1, -100_001),
			y:      NewDecWithPrec(1, -100_001),
			expErr: ErrInvalidDec,
		},
		"1e100000 - 1 = 999..9": {
			x:   NewDecWithPrec(1, 100_000),
			y:   NewDecFromInt64(1),
			exp: must(NewDecFromString(strings.Repeat("9", 100_000))),
		},
		"1e100000 - 0 = 1e100000": {
			x:   NewDecWithPrec(1, 100_000),
			y:   NewDecFromInt64(0111),
			exp: must(NewDecFromString("1e100000")),
		},
		"1e100001 - 0 -> err": {
			x:      NewDecWithPrec(1, 100_001),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDec,
		},
		"1e100000 - -1 -> 100..1": {
			x:   NewDecWithPrec(1, 100_000),
			y:   NewDecFromInt64(-1),
			exp: must(NewDecFromString("1" + strings.Repeat("0", 99_999) + "1")),
		},
		"1e-100000 - 0 = 1e-100000": {
			x:   NewDecWithPrec(1, -100_000),
			y:   NewDecFromInt64(0),
			exp: must(NewDecFromString("1e-100000")),
		},
		"1e-100001 - 0 -> err": {
			x:      NewDecWithPrec(1, -100_001),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDec,
		},
		"1e-100000 - -1 -> 0.000..01": {
			x:   NewDecWithPrec(1, -100_000),
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
			assert.True(t, spec.exp.Equal(got))
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
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(1234, -3),
			exp: must(NewDecFromString("1.000000000000000000000000000000000")),
		},
		"-1.234 / 1234 = -1": {
			x:   NewDecWithPrec(-1234, -3),
			y:   NewDecWithPrec(1234, -3),
			exp: must(NewDecFromString("-1.000000000000000000000000000000000")),
		},
		"1.234 / -123 = 1.0100": {
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecFromInt64(-123),
			exp: must(NewDecFromString("-0.01003252032520325203252032520325203")),
		},
		"1.234 / -1.234 = -1": {
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(-1234, -3),
			exp: must(NewDecFromString("-1.000000000000000000000000000000000")),
		},
		"-1.234 / -1.234 = 1": {
			x:   NewDecWithPrec(-1234, -3),
			y:   NewDecWithPrec(-1234, -3),
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
			x:   NewDecWithPrec(9, -34),
			y:   NewDecFromInt64(10),
			exp: must(NewDecFromString("9e-35")),
		},
		"9e-35 / 10 = 9e-36 - no rounding": {
			x:   NewDecWithPrec(9, -35),
			y:   NewDecFromInt64(10),
			exp: must(NewDecFromString("9e-36")),
		},
		"high precision - min/0.1": {
			x:   NewDecWithPrec(1, -100_000),
			y:   NewDecWithPrec(1, -1),
			exp: NewDecWithPrec(1, -99_999),
		},
		"high precision - min/1": {
			x:   NewDecWithPrec(1, -100_000),
			y:   NewDecWithPrec(1, 0),
			exp: NewDecWithPrec(1, -100_000),
		},
		"high precision - min/10": {
			x:      NewDecWithPrec(1, -100_000),
			y:      NewDecWithPrec(1, 1),
			expErr: ErrInvalidDec,
		},
		"high precision - <_min/0.1": {
			x:   NewDecWithPrec(1, -100_001),
			y:   NewDecWithPrec(1, -1),
			exp: NewDecWithPrec(1, -100_000),
		},
		"high precision - <_min/1": {
			x:      NewDecWithPrec(1, -100_001),
			y:      NewDecWithPrec(1, 0),
			expErr: ErrInvalidDec,
		},
		"high precision - <_min/10": {
			x:      NewDecWithPrec(1, -100_001),
			y:      NewDecWithPrec(1, 1),
			expErr: ErrInvalidDec,
		},
		"high precision - min/-0.1": {
			x:   NewDecWithPrec(1, -100_000),
			y:   NewDecWithPrec(-1, -1),
			exp: NewDecWithPrec(-1, -99_999),
		},
		"high precision - min/-1": {
			x:   NewDecWithPrec(1, -100_000),
			y:   NewDecWithPrec(-1, 0),
			exp: NewDecWithPrec(-1, -100_000),
		},
		"high precision - min/-10": {
			x:      NewDecWithPrec(1, -100_000),
			y:      NewDecWithPrec(-1, 1),
			expErr: ErrInvalidDec,
		},
		"high precision - <_min/-0.1": {
			x:   NewDecWithPrec(1, -100_001),
			y:   NewDecWithPrec(-1, -1),
			exp: NewDecWithPrec(-1, -100_000),
		},
		"high precision - <_min/-1": {
			x:      NewDecWithPrec(1, -100_001),
			y:      NewDecWithPrec(-1, 0),
			expErr: ErrInvalidDec,
		},
		"high precision - <_min/-10": {
			x:      NewDecWithPrec(1, -100_001),
			y:      NewDecWithPrec(-1, 1),
			expErr: ErrInvalidDec,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got, gotErr := spec.x.Quo(spec.y)
			if name == "1.234 / -123 = 1.111" {
				fmt.Println("got", got)
			}

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
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(1234, -3),
			exp: must(NewDecFromString("1.000000000000000000000000000000000")),
		},
		"-1.234 / 1234 = -121.766": {
			x:   NewDecWithPrec(-1234, -3),
			y:   NewDecWithPrec(1234, -3),
			exp: must(NewDecFromString("-1.000000000000000000000000000000000")),
		},
		"1.234 / -123 -> Err": {
			x:      NewDecWithPrec(1234, -3),
			y:      NewDecFromInt64(-123),
			expErr: ErrUnexpectedRounding,
		},
		"1.234 / -1.234 = -1": {
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(-1234, -3),
			exp: must(NewDecFromString("-1.000000000000000000000000000000000")),
		},
		"-1.234 / -1.234 = 1": {
			x:   NewDecWithPrec(-1234, -3),
			y:   NewDecWithPrec(-1234, -3),
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
			x:   NewDecWithPrec(9, -34),
			y:   NewDecFromInt64(10),
			exp: must(NewDecFromString("0.00000000000000000000000000000000009000000000000000000000000000000000")),
		},
		"9e-35 / 10 = 9e-36 - no rounding": {
			x:   NewDecWithPrec(9, -35),
			y:   NewDecFromInt64(10),
			exp: must(NewDecFromString("9e-36")),
		},
		"high precision - min/0.1": {
			x:   NewDecWithPrec(1, -100_000),
			y:   NewDecWithPrec(1, -1),
			exp: NewDecWithPrec(1, -99_999),
		},
		"high precision - min/1": {
			x:   NewDecWithPrec(1, -100_000),
			y:   NewDecWithPrec(1, 0),
			exp: NewDecWithPrec(1, -100_000),
		},
		"high precision - min/10": {
			x:      NewDecWithPrec(1, -100_000),
			y:      NewDecWithPrec(1, 1),
			expErr: ErrInvalidDec,
		},
		"high precision - <_min/0.1": {
			x:   NewDecWithPrec(1, -100_001),
			y:   NewDecWithPrec(1, -1),
			exp: NewDecWithPrec(1, -100_000),
		},
		"high precision - <_min/1": {
			x:      NewDecWithPrec(1, -100_001),
			y:      NewDecWithPrec(1, 0),
			expErr: ErrInvalidDec,
		},
		"high precision - <_min/10 -> Err": {
			x:      NewDecWithPrec(1, -100_001),
			y:      NewDecWithPrec(1, 1),
			expErr: ErrInvalidDec,
		},
		"high precision - min/-0.1": {
			x:   NewDecWithPrec(1, -100_000),
			y:   NewDecWithPrec(-1, -1),
			exp: NewDecWithPrec(-1, -99_999),
		},
		"high precision - min/-1": {
			x:   NewDecWithPrec(1, -100_000),
			y:   NewDecWithPrec(-1, 0),
			exp: NewDecWithPrec(-1, -100_000),
		},
		"high precision - min/-10 -> Err": {
			x:      NewDecWithPrec(1, -100_000),
			y:      NewDecWithPrec(-1, 1),
			expErr: ErrInvalidDec,
		},
		"high precision - <_min/-0.1": {
			x:   NewDecWithPrec(1, -100_001),
			y:   NewDecWithPrec(-1, -1),
			exp: NewDecWithPrec(-1, -100_000),
		},
		"high precision - <_min/-1": {
			x:      NewDecWithPrec(1, -100_001),
			y:      NewDecWithPrec(-1, 0),
			expErr: ErrInvalidDec,
		},
		"high precision - <_min/-10": {
			x:      NewDecWithPrec(1, -100_001),
			y:      NewDecWithPrec(-1, 1),
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
			exp: NewDecFromInt64(5),
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
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(1234, -3),
			exp: NewDecFromInt64(1),
		},
		"-1.234 / 1234 = -121.766": {
			x:   NewDecWithPrec(-1234, -3),
			y:   NewDecWithPrec(1234, -3),
			exp: NewDecFromInt64(-1),
		},
		"1.234 / -1.234 = 2.468": {
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(-1234, -3),
			exp: NewDecFromInt64(-1),
		},
		"-1.234 / -1.234 = 1": {
			x:   NewDecWithPrec(-1234, -3),
			y:   NewDecWithPrec(-1234, -3),
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
			x:   NewDecWithPrec(1, -100_000),
			y:   NewDecWithPrec(1, -1),
			exp: NewDecFromInt64(0),
		},
		"high precision - <_min/-1 -> Err": {
			x:      NewDecWithPrec(1, -100_001),
			y:      NewDecWithPrec(-1, 0),
			expErr: ErrInvalidDec,
		},
		"high precision - <_min/-10 -> Err": {
			x:      NewDecWithPrec(1, -100_001),
			y:      NewDecWithPrec(-1, 1),
			expErr: ErrInvalidDec,
		},
		"1e000 / 1 -> Err": {
			x:      NewDecWithPrec(1, 100_000),
			y:      NewDecFromInt64(1),
			expErr: ErrInvalidDec,
		},
		"1e100000 - 1^1 -> Err": {
			x:      NewDecWithPrec(1, 100_000),
			y:      NewDecWithPrec(1, -1),
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
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecFromInt64(1),
			exp: NewDecWithPrec(234, -3),
		},
		"1.234 / 0.1 = 0.034": {
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(1, -1),
			exp: NewDecWithPrec(34, -3),
		},
		"1.234 / 1.1 = 0.134": {
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(11, -1),
			exp: NewDecWithPrec(134, -3),
		},
		"10 / 0 -> Err": {
			x:      NewDecFromInt64(10),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDec,
		},
		"-1e0000 / 9e0000 = 1e0000": {
			x:   NewDecWithPrec(-1, 100_000),
			y:   NewDecWithPrec(9, 100_000),
			exp: NewDecWithPrec(-1, 100_000),
		},
		"1e0000 / 9e0000 = 1e0000": {
			x:   NewDecWithPrec(1, 100_000),
			y:   NewDecWithPrec(9, 100_000),
			exp: NewDecWithPrec(1, 100_000),
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
			src: NewDecWithPrec(1234, -1),
			exp: 1,
		},
		"two decimal places": {
			src: NewDecWithPrec(12345, -2),
			exp: 2,
		},
		"three decimal places": {
			src: NewDecWithPrec(123456, -3),
			exp: 3,
		},
		"trailing zeros": {
			src: NewDecWithPrec(123400, -4),
			exp: 4,
		},
		"zero value": {
			src: NewDecFromInt64(0),
			exp: 0,
		},
		"negative value": {
			src: NewDecWithPrec(-12345, -3),
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
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(1234, -3),
			exp: 0,
		},
		"1.234 > 1.233": {
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(1233, -3),
			exp: 1,
		},
		"1.233 < 1.234": {
			x:   NewDecWithPrec(1233, -3),
			y:   NewDecWithPrec(1234, -3),
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
		expErr    error
	}{
		"positive value": {
			src:       "10",
			exp:       "10",
			decPlaces: 1,
			expErr:    ErrInvalidDec,
		},
		"negative value": {
			src:       "-10",
			exp:       "-10",
			decPlaces: 1,
			expErr:    ErrInvalidDec,
		},
		"positive decimal": {
			src:       "1.30000",
			exp:       "1.3",
			decPlaces: 4,
			expErr:    ErrInvalidDec,
		},
		"negative decimal": {
			src:       "-1.30000",
			exp:       "-1.3",
			decPlaces: 4,
			expErr:    ErrInvalidDec,
		},
		"zero decimal and decimal places": {
			src:       "0.00000",
			exp:       "0",
			decPlaces: 0,
			expErr:    ErrInvalidDec,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			src, _ := NewDecFromString(spec.src)
			got, gotErr := src.Reduce()
			require.Equal(t, spec.exp, got.String())
			if spec.expErr != nil {
				require.Equal(t, spec.decPlaces, gotErr)
			}
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
			x:   NewDecWithPrec(11, -1),
			y:   NewDecWithPrec(11, -1),
			exp: NewDecWithPrec(121, -2),
		},
		"1.000 * 1.000 = 1.000000": {
			x:   NewDecWithPrec(1000, -3),
			y:   NewDecWithPrec(1000, -3),
			exp: must(NewDecFromString("1.000000")),
		},
		"0.0000001 * 0.0000001 = 0": {
			x:   NewDecWithPrec(00000001, -7),
			y:   NewDecWithPrec(00000001, -7),
			exp: NewDecWithPrec(1, -14),
		},
		"1.000000000000000000000000000000000000123456789 * 0.000001 = 0.000000000100000000000000000000000000000123456789": {
			x:      must(NewDecFromString("1.0000000000000000000000000000000000000123456789")),
			y:      NewDecWithPrec(1, -6),
			expErr: ErrUnexpectedRounding,
		},
		"1000001 * 1.000001 = 1000002.000001": {
			x:   NewDecFromInt64(1000001),
			y:   NewDecWithPrec(1000001, -6),
			exp: must(NewDecFromString("1000002.000001")),
		},
		"1000000000000000000000000000000000000123456789 * 100000000000 ": {
			x:      must(NewDecFromString("1000000000000000000000000000000000000123456789")),
			y:      NewDecWithPrec(1, 6),
			expErr: ErrUnexpectedRounding,
		},
		"1000001 * 1000000 = 1000001000000 ": {
			x:   NewDecFromInt64(1000001),
			y:   NewDecFromInt64(1000000),
			exp: NewDecFromInt64(1000001000000),
		},
		"1e0000 * 1e0000 -> Err": {
			x:      NewDecWithPrec(1, 100_000),
			y:      NewDecWithPrec(1, 100_000),
			expErr: ErrInvalidDec,
		},
		"1e0000 * 1 = 1e0000": {
			x:   NewDecWithPrec(1, 100_000),
			y:   NewDecWithPrec(1, 0),
			exp: NewDecWithPrec(1, 100_000),
		},
		"1e100000 * 9 = 9e100000": {
			x:   NewDecWithPrec(1, 100_000),
			y:   NewDecFromInt64(9),
			exp: NewDecWithPrec(9, 100_000),
		},
		"1e100000 * 10 = err": {
			x:      NewDecWithPrec(1, 100_000),
			y:      NewDecWithPrec(1, 1),
			expErr: ErrInvalidDec,
		},
		"1e0000 * -1 = 1e0000": {
			x:   NewDecWithPrec(1, 100_000),
			y:   NewDecWithPrec(-1, 0),
			exp: NewDecWithPrec(1, 100_000),
		},
		"1e100000 * -9 = 9e100000": {
			x:   NewDecWithPrec(1, 100_000),
			y:   NewDecFromInt64(-9),
			exp: NewDecWithPrec(9, 100_000),
		},
		"1e100000 * -10 = err": {
			x:      NewDecWithPrec(1, 100_000),
			y:      NewDecWithPrec(-1, 1),
			expErr: ErrInvalidDec,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got, gotErr := spec.x.MulExact(spec.y)
			if spec.expErr != nil {
				require.ErrorIs(t, gotErr, spec.expErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.exp, got)
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
	require.ErrorIs(t, err, ErrInvalidDec)
}

func must[T any](r T, err error) T {
	if err != nil {
		panic(err)
	}
	return r
}
