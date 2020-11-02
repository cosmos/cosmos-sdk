package types

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/cockroachdb/apd/v2"
)

func apdQuoFirst(ctx apd.Context) func(x, y int64) (bool, string) {
	return func(x, y int64) (bool, string) {
		a := apd.New(x, 1)
		b := apd.New(y, 1)
		c := apd.New(0, 1)
		_, err := ctx.Quo(c, a, b)
		if err != nil {
			panic(err)
		}
		_, err = ctx.Mul(c, c, b)
		if err != nil {
			panic(err)
		}
		_, err = ctx.Sub(c, c, a)
		if err != nil {
			panic(err)
		}
		return c.IsZero(), c.String()
	}
}

func apdMulFirst(ctx apd.Context) func(x, y int64) (bool, string) {
	return func(x, y int64) (bool, string) {
		a := apd.New(x, 1)
		b := apd.New(y, 1)
		c := apd.New(0, 1)
		_, err := ctx.Mul(c, a, b)
		if err != nil {
			panic(err)
		}
		_, err = ctx.Quo(c, c, b)
		if err != nil {
			panic(err)
		}
		_, err = ctx.Sub(c, c, a)
		if err != nil {
			panic(err)
		}
		return c.IsZero(), c.String()
	}
}

var decimal32Context = apd.Context{
	Precision:   7,
	MaxExponent: 96,
	MinExponent: -95,
	Traps:       apd.DefaultTraps,
	Rounding:    apd.RoundHalfEven,
}

var decimal64Context = apd.Context{
	Precision:   16,
	MaxExponent: 384,
	MinExponent: -383,
	Traps:       apd.DefaultTraps,
	Rounding:    apd.RoundHalfEven,
}

var decimal128Context = apd.Context{
	Precision:   34,
	MaxExponent: 6144,
	MinExponent: -6143,
	Traps:       apd.DefaultTraps,
	Rounding:    apd.RoundHalfEven,
}

func TestDecErr(t *testing.T) {
	cases := []struct {
		name string
		f    func(x, y int64) (bool, string)
	}{
		{
			"big.Rat (x/y)*y",
			func(x, y int64) (bool, string) {
				a := big.NewRat(x, 1)
				b := big.NewRat(y, 1)
				c := big.NewRat(0, 1)
				c = c.Quo(a, b)
				c = c.Mul(c, b)
				c = c.Sub(c, a)
				zero := big.NewRat(0, 1)
				return c.Cmp(zero) == 0, c.FloatString(10)
			},
		},
		{
			"big.Rat (x*y)/y",
			func(x, y int64) (bool, string) {
				a := big.NewRat(x, 1)
				b := big.NewRat(y, 1)
				c := big.NewRat(0, 1)
				c = c.Mul(a, b)
				c = c.Quo(c, b)
				c = c.Sub(c, a)
				zero := big.NewRat(0, 1)
				return c.Cmp(zero) == 0, c.FloatString(10)
			},
		},
		{
			"sdk.Dec (x/y)*y",
			func(x, y int64) (bool, string) {
				a := NewDec(x)
				b := NewDec(y)
				c := a.Quo(b).Mul(b).Sub(a)
				return c.IsZero(), c.String()
			},
		},
		{
			"sdk.Dec (x*y)/y",
			func(x, y int64) (bool, string) {
				a := NewDec(x)
				b := NewDec(y)
				c := a.Mul(b).Quo(b).Sub(a)
				return c.IsZero(), c.String()
			},
		},
		{
			"float32 (x/y)*y",
			func(x, y int64) (bool, string) {
				a := float32(x)
				b := float32(y)
				c := a / b
				d := c * b
				e := d - a
				return e == float32(0), fmt.Sprintf("%e", e)
			},
		},
		{
			"float32 (x*y)/y",
			func(x, y int64) (bool, string) {
				a := float32(x)
				b := float32(y)
				c := a * b
				d := c / b
				e := d - a
				return e == float32(0), fmt.Sprintf("%e", e)
			},
		},
		{
			"float64 (x/y)*y", func(x, y int64) (bool, string) {
				a := float64(x)
				b := float64(y)
				c := a / b
				d := c * b
				e := d - a
				return e == float64(0), fmt.Sprintf("%e", e)
			},
		},
		{
			"float64 (x*y)/y",
			func(x, y int64) (bool, string) {
				a := float64(x)
				b := float64(y)
				c := a * b
				d := c / b
				e := d - a
				return e == float64(0), fmt.Sprintf("%e", e)
			},
		},
		{
			"float128 (x/y)*y",
			func(x, y int64) (bool, string) {
				a := big.NewFloat(0)
				a.SetInt64(x)
				a.SetPrec(113)
				b := big.NewFloat(0)
				b.SetInt64(x)
				b.SetPrec(113)
				c := big.NewFloat(0)
				c.SetPrec(113)
				c = c.Quo(a, b)
				c = c.Mul(c, b)
				c = c.Sub(c, a)
				zero := big.NewFloat(0)
				return c.Cmp(zero) == 0, c.String()
			},
		},
		{
			"float128 (x*y)/y",
			func(x, y int64) (bool, string) {
				a := big.NewFloat(0)
				a.SetInt64(x)
				a.SetPrec(113)
				b := big.NewFloat(0)
				b.SetInt64(x)
				b.SetPrec(113)
				c := big.NewFloat(0)
				c.SetPrec(113)
				c = c.Mul(a, b)
				c = c.Quo(c, b)
				c = c.Sub(c, a)
				zero := big.NewFloat(0)
				return c.Cmp(zero) == 0, c.String()
			},
		},
		{
			"decimal32 (x/y)*y",
			apdQuoFirst(decimal32Context),
		},
		{
			"decimal32 (x*y)/y", apdMulFirst(decimal32Context),
		},
		{
			"decimal64 (x/y)*y", apdQuoFirst(decimal64Context),
		},
		{
			"decimal64 (x*y)/y", apdMulFirst(decimal64Context),
		},
		{
			"decimal128 (x/y)*y", apdQuoFirst(decimal128Context),
		},
		{
			"decimal128 (x*y)/y", apdMulFirst(decimal128Context),
		},
	}

	numTests := 100000

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := rand.New(rand.NewSource(0))
			xs := make([]int64, numTests)
			ys := make([]int64, numTests)
			for i := 0; i < numTests; i++ {
				xs[i] = int64(r.Int31())
				ys[i] = int64(r.Int31())
			}

			numCorrect := 0
			errs := make([]string, numTests)

			for i := 0; i < numTests; i++ {
				var correct bool
				correct, errs[i] = tc.f(xs[i], ys[i])
				if correct {
					numCorrect++
				}
			}

			totalErr := apd.New(0, 1)
			for i := 0; i < numTests; i++ {
				errF, _, err := decimal128Context.NewFromString(errs[i])
				if err != nil {
					panic(err)
				}
				errF = errF.Abs(errF)
				x := apd.New(0, 1)
				x.SetInt64(xs[i])
				_, err = decimal128Context.Quo(errF, errF, x)
				if err != nil {
					panic(err)
				}
				_, err = decimal128Context.Add(totalErr, totalErr, errF)
				if err != nil {
					panic(err)
				}
			}
			n := apd.New(0, 1)
			n.SetInt64(int64(numTests))
			_, err := decimal128Context.Quo(totalErr, totalErr, n)
			if err != nil {
				panic(err)
			}

			totalErrF, err := totalErr.Float64()
			if err != nil {
				panic(err)
			}

			t.Logf("%s Out of %d trials: %.2f%% Correct, %.2e Average Error",
				tc.name,
				numTests,
				float64(numCorrect)/float64(numTests)*float64(100),
				totalErrF,
			)
		})
	}
}

func Benchmark(b *testing.B) {
	parseFloat64 := func(f float64) interface{} { return f }

	parseDec := func(f float64) interface{} {
		res, err := NewDecFromStr(fmt.Sprintf("%f", f))
		if err != nil {
			panic(fmt.Errorf("%v: %f", err, f))
		}
		return res
	}

	parseBigRat := func(f float64) interface{} {
		r := big.NewRat(0, 1)
		return r.SetFloat64(f)
	}

	parseBigFloat128 := func(f float64) interface{} {
		res := big.NewFloat(f)
		res.SetPrec(113)
		return res
	}

	parseDecimal128 := func(f float64) interface{} {
		res := apd.New(0, 1)
		res, err := res.SetFloat64(f)
		if err != nil {
			panic(err)
		}
		return res
	}

	benchmarks := []struct {
		name  string
		parse func(f float64) interface{}
		f     func(x, y interface{}) interface{}
	}{
		{
			"float64.+", parseFloat64,
			func(x, y interface{}) interface{} { return x.(float64) + y.(float64) },
		},
		{
			"float64.-", parseFloat64,
			func(x, y interface{}) interface{} { return x.(float64) - y.(float64) },
		},
		{
			"float64.*", parseFloat64,
			func(x, y interface{}) interface{} { return x.(float64) * y.(float64) },
		},
		{
			"float64./", parseFloat64,
			func(x, y interface{}) interface{} { return x.(float64) / y.(float64) },
		},
		{
			"sdk.Dec.Add", parseDec,
			func(x, y interface{}) interface{} { return x.(Dec).Add(y.(Dec)) },
		},
		{
			"sdk.Dec.Sub", parseDec,
			func(x, y interface{}) interface{} { return x.(Dec).Sub(y.(Dec)) },
		},
		{
			"sdk.Dec.Mul", parseDec,
			func(x, y interface{}) interface{} { return x.(Dec).Quo(y.(Dec)) },
		},
		{
			"sdk.Dec.Quo", parseDec,
			func(x, y interface{}) interface{} { return x.(Dec).Quo(y.(Dec)) },
		},
		{
			"big.Rat.Add", parseBigRat,
			func(x, y interface{}) interface{} { return x.(*big.Rat).Add(x.(*big.Rat), y.(*big.Rat)) },
		},
		{
			"big.Rat.Sub", parseBigRat,
			func(x, y interface{}) interface{} { return x.(*big.Rat).Sub(x.(*big.Rat), y.(*big.Rat)) },
		},
		{
			"big.Rat.Mul", parseBigRat,
			func(x, y interface{}) interface{} { return x.(*big.Rat).Mul(x.(*big.Rat), y.(*big.Rat)) },
		},
		{
			"big.Rat.Quo", parseBigRat,
			func(x, y interface{}) interface{} { return x.(*big.Rat).Quo(x.(*big.Rat), y.(*big.Rat)) },
		},
		{
			"big.Float128.Add", parseBigFloat128,
			func(x, y interface{}) interface{} { return x.(*big.Float).Add(x.(*big.Float), y.(*big.Float)) },
		},
		{
			"big.Float128.Sub", parseBigFloat128,
			func(x, y interface{}) interface{} { return x.(*big.Float).Sub(x.(*big.Float), y.(*big.Float)) },
		},
		{
			"big.Float128.Mul", parseBigFloat128,
			func(x, y interface{}) interface{} { return x.(*big.Float).Mul(x.(*big.Float), y.(*big.Float)) },
		},
		{
			"big.Float128.Quo", parseBigFloat128,
			func(x, y interface{}) interface{} { return x.(*big.Float).Quo(x.(*big.Float), y.(*big.Float)) },
		},
		{
			"apd.Decimal128.Add", parseDecimal128,
			func(x, y interface{}) interface{} {
				_, err := decimal128Context.Add(x.(*apd.Decimal), x.(*apd.Decimal), y.(*apd.Decimal))
				if err != nil {
					panic(err)
				}
				return x
			},
		},
		{
			"apd.Decimal128.Sub", parseDecimal128,
			func(x, y interface{}) interface{} {
				_, err := decimal128Context.Sub(x.(*apd.Decimal), x.(*apd.Decimal), y.(*apd.Decimal))
				if err != nil {
					panic(err)
				}
				return x
			},
		},
		{
			"apd.Decimal128.Mul", parseDecimal128,
			func(x, y interface{}) interface{} {
				_, err := decimal128Context.Mul(x.(*apd.Decimal), x.(*apd.Decimal), y.(*apd.Decimal))
				if err != nil {
					panic(err)
				}
				return x
			},
		},
		{
			"apd.Decimal128.Quo", parseDecimal128,
			func(x, y interface{}) interface{} {
				_, err := decimal128Context.Quo(x.(*apd.Decimal), x.(*apd.Decimal), y.(*apd.Decimal))
				if err != nil {
					panic(err)
				}
				return x
			},
		},
	}

	opCounts := []int{2, 33, 257}

	for _, opCount := range opCounts {
		for _, bm := range benchmarks {
			b.Run(fmt.Sprintf("%d Ops: %s", opCount-1, bm.name), func(b *testing.B) {
				b.StopTimer()
				r := rand.New(rand.NewSource(0))
				xs := make([][]interface{}, b.N)
				for i := 0; i < b.N; i++ {
					xs[i] = make([]interface{}, opCount)
					for j := 0; j < opCount; j++ {
						a := r.Int63()
						b := math.Pow10(r.Intn(10))
						c := float64(a) / b
						xs[i][j] = bm.parse(c)
					}
				}

				b.StartTimer()
				for i := 0; i < b.N; i++ {
					a := xs[i][0]
					for j := 1; j < opCount; j++ {
						a = bm.f(a, xs[i][j])
					}
				}
			})
		}
	}
}
