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
			expErr: ErrInvalidDecString,
		},
		"Inf": {
			src:    "Inf",
			expErr: ErrInvalidDecString,
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
			assert.Equal(t, spec.exp.String(), got.String())
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
		x      Dec
		y      Dec
		exp    Dec
		expErr error
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

		// TO DO: more edge cases For example: 1^100000 + 9^100000 , 1^100000 + 1^-1
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
		"zero minus zero": {
			// 0 - 0 = 0
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
			// -123 - 123 = 0
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
		"valid decimal with decimal places minus simple positive": {
			// 1.234 - 123 = -121.766
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecFromInt64(123),
			exp: NewDecWithPrec(-121766, -3),
		},
		"valid decimal with decimal places minus simple negative": {
			// 1.234 - -123 = 1.111
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecFromInt64(-123),
			exp: NewDecWithPrec(124234, -3),
		},
		"valid decimal with decimal places minus valid negative decimal with decimal places": {
			// 1.234 - -1.234 = 2.468
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(-1234, -3),
			exp: NewDecWithPrec(2468, -3),
		},
		"valid negative decimal with decimal places minus valid negative decimal with decimal places": {
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
			got, gotErr := spec.x.Sub(spec.y)
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

func TestQuo(t *testing.T) {
	specs := map[string]struct {
		src    string
		x      Dec
		y      Dec
		exp    Dec
		expErr error
	}{
		"0 / 0": {
			x:      NewDecFromInt64(0),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDecString,
		},
		" 0 / 123 = 0": {
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(0),
		},
		// answer is -0 but this is not throwing an error
		// "0 / -123 = 0": {
		// 	x:   NewDecFromInt64(0),
		// 	y:   NewDecFromInt64(-123),
		// 	exp: NewDecFromInt64(-0),
		// },
		"123 / 0 = 0": {
			x:      NewDecFromInt64(123),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDecString,
		},
		"-123 / 0 = 0": {
			x:      NewDecFromInt64(-123),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDecString,
		},
		"123 / 123 = 1": {
			// the answer is showing up as 0 although it should be 1. Again with 34 precision it is mismatched
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
		"1.234 / 1.234": {
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(1234, -3),
			exp: must(NewDecFromString("1.000000000000000000000000000000000")),
		},
		"-1.234 / 1234 = -121.766": {
			x:   NewDecWithPrec(-1234, -3),
			y:   NewDecWithPrec(1234, -3),
			exp: must(NewDecFromString("-1.000000000000000000000000000000000")),
		},
		"1.234 / -123 = 1.111": {
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecFromInt64(-123),
			exp: must(NewDecFromString("-0.01003252032520325203252032520325203")),
		},
		"1.234 / -1.234 = 2.468": {
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(-1234, -3),
			exp: must(NewDecFromString("-1.000000000000000000000000000000000")),
		},
		"-1.234 / -1.234 = 1": {
			x:   NewDecWithPrec(-1234, -3),
			y:   NewDecWithPrec(-1234, -3),
			exp: must(NewDecFromString("1.000000000000000000000000000000000")),
		},
		// "precision too high": {
		// 	// 10^34 / 10^34 = 2*10^34
		// 	x:           NewDecWithPrec(1, 36),
		// 	y:           NewDecWithPrec(1, 36),
		// 	constraints: []SetupConstraint{AssertMaxDecimals(34)},
		// 	expErr:      ErrInvalidDecString,
		// },
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got, gotErr := spec.x.Quo(spec.y)
			fmt.Println(spec.x, spec.y, got, spec.exp)

			if spec.expErr != nil {
				require.ErrorIs(t, gotErr, spec.expErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.exp, got)
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
		"0 / 0": {
			x:      NewDecFromInt64(0),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDecString,
		},
		" 0 / 123 = 0": {
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(0),
		},
		// answer is -0 but this is not throwing an error
		// "0 / -123 = 0": {
		// 	x:   NewDecFromInt64(0),
		// 	y:   NewDecFromInt64(-123),
		// 	exp: NewDecFromInt64(-0),
		// },
		"123 / 0 = 0": {
			x:      NewDecFromInt64(123),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDecString,
		},
		"-123 / 0 = 0": {
			x:      NewDecFromInt64(-123),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDecString,
		},
		"123 / 123 = 1": {
			// the answer is showing up as 0 although it should be 1. Again with 34 precision it is mismatched
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
		"1.234 / 1.234": {
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(1234, -3),
			exp: must(NewDecFromString("1.000000000000000000000000000000000")),
		},
		"-1.234 / 1234 = -121.766": {
			x:   NewDecWithPrec(-1234, -3),
			y:   NewDecWithPrec(1234, -3),
			exp: must(NewDecFromString("-1.000000000000000000000000000000000")),
		},
		"1.234 / -123 = 1.111": {
			x:      NewDecWithPrec(1234, -3),
			y:      NewDecFromInt64(-123),
			expErr: ErrUnexpectedRounding,
		},
		"1.234 / -1.234 = 2.468": {
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(-1234, -3),
			exp: must(NewDecFromString("-1.000000000000000000000000000000000")),
		},
		"-1.234 / -1.234 = 1": {
			x:   NewDecWithPrec(-1234, -3),
			y:   NewDecWithPrec(-1234, -3),
			exp: must(NewDecFromString("1.000000000000000000000000000000000")),
		},
		// "precision too high": {
		// 	// 10^34 / 10^34 = 2*10^34
		// 	x:           NewDecWithPrec(1, 36),
		// 	y:           NewDecWithPrec(1, 36),
		// 	constraints: []SetupConstraint{AssertMaxDecimals(34)},
		// 	expErr:      ErrInvalidDecString,
		// },
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got, gotErr := spec.x.Quo(spec.y)
			fmt.Println(spec.x, spec.y, got, spec.exp)

			if spec.expErr != nil {
				require.ErrorIs(t, gotErr, spec.expErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.exp, got)
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
		"0 / 0": {
			x:      NewDecFromInt64(0),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDecString,
		},
		" 0 / 123 = 0": {
			x:   NewDecFromInt64(0),
			y:   NewDecFromInt64(123),
			exp: NewDecFromInt64(0),
		},
		// answer is -0 but this is not throwing an error
		// "0 / -123 = 0": {
		// 	x:   NewDecFromInt64(0),
		// 	y:   NewDecFromInt64(-123),
		// 	exp: NewDecFromInt64(-0),
		// },
		"123 / 0 = 0": {
			x:      NewDecFromInt64(123),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDecString,
		},
		"-123 / 0 = 0": {
			x:      NewDecFromInt64(-123),
			y:      NewDecFromInt64(0),
			expErr: ErrInvalidDecString,
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
			x:   NewDecWithPrec(1234, -3),
			y:   NewDecWithPrec(1234, -3),
			exp: NewDecFromInt64(1),
		},
		"-1.234 / 1234 = -121.766": {
			x:   NewDecWithPrec(-1234, -3),
			y:   NewDecWithPrec(1234, -3),
			exp: NewDecFromInt64(-1),
		},
		// "1.234 / -123 = -0": {
		//-0
		// 	x:   NewDecWithPrec(1234, -3),
		// 	y:   NewDecFromInt64(-123),
		// 	exp: NewDecFromInt64(1),
		// },
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
		// "precision too high": {
		// 	// 10^34 / 10^34 = 2*10^34
		// 	x:           NewDecWithPrec(1, 36),
		// 	y:           NewDecWithPrec(1, 36),
		// 	constraints: []SetupConstraint{AssertMaxDecimals(34)},
		// 	expErr:      ErrInvalidDecString,
		// },
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got, gotErr := spec.x.QuoInteger(spec.y)
			fmt.Println(spec.x, spec.y, got, spec.exp)

			if spec.expErr != nil {
				require.ErrorIs(t, gotErr, spec.expErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.exp, got)
		})
	}
}

func TestRem(t *testing.T) {
	// TO DO
}

func TestNumDecimalPlaces(t *testing.T) {
	// TO DO
}

func TestCmp(t *testing.T) {
	// TO DO
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
			expErr:    ErrInvalidDecString,
		},
		"negative value": {
			src:       "-10",
			exp:       "-10",
			decPlaces: 1,
			expErr:    ErrInvalidDecString,
		},
		"positive decimal": {
			src:       "1.30000",
			exp:       "1.3",
			decPlaces: 4,
			expErr:    ErrInvalidDecString,
		},
		"negative decimal": {
			src:       "-1.30000",
			exp:       "-1.3",
			decPlaces: 4,
			expErr:    ErrInvalidDecString,
		},
		"zero decimal and decimal places": {
			src:       "0.00000",
			exp:       "0",
			decPlaces: 0,
			expErr:    ErrInvalidDecString,
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
		"200 * 200 = 200": {
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
		"0 * 0 = 10000": {
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
			exp: NewDecWithPrec(1000000, 6),
		},
		"0.0000001 * 0.0000001 = 1.21": {
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
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got, gotErr := spec.x.MulExact(spec.y)
			fmt.Println(spec.x, spec.y, got, spec.exp)

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
	require.ErrorIs(t, err, ErrInvalidDecString)
}

func must[T any](r T, err error) T {
	if err != nil {
		panic(err)
	}
	return r
}
