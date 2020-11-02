package types

import (
	"fmt"
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

func BenchmarkDecErr(b *testing.B) {
	cases := map[string]func(x, y int64) (bool, string){
		"big.Rat (x/y)*y": func(x, y int64) (bool, string) {
			a := big.NewRat(x, 1)
			b := big.NewRat(y, 1)
			c := big.NewRat(0, 1)
			c = c.Quo(a, b)
			c = c.Mul(c, b)
			c = c.Sub(c, a)
			zero := big.NewRat(0, 1)
			return c.Cmp(zero) == 0, c.FloatString(10)
		},
		"big.Rat (x*y)/y": func(x, y int64) (bool, string) {
			a := big.NewRat(x, 1)
			b := big.NewRat(y, 1)
			c := big.NewRat(0, 1)
			c = c.Mul(a, b)
			c = c.Quo(c, b)
			c = c.Sub(c, a)
			zero := big.NewRat(0, 1)
			return c.Cmp(zero) == 0, c.FloatString(10)
		},
		"sdk.Dec (x/y)*y": func(x, y int64) (bool, string) {
			a := NewDec(x)
			b := NewDec(y)
			c := a.Quo(b).Mul(b).Sub(a)
			return c.IsZero(), c.String()
		},
		"sdk.Dec (x*y)/y": func(x, y int64) (bool, string) {
			a := NewDec(x)
			b := NewDec(y)
			c := a.Mul(b).Quo(b).Sub(a)
			return c.IsZero(), c.String()
		},
		"float32 (x/y)*y": func(x, y int64) (bool, string) {
			a := float32(x)
			b := float32(y)
			c := a / b
			d := c * b
			e := d - a
			return e == float32(0), fmt.Sprintf("%e", e)
		},
		"float32 (x*y)/y": func(x, y int64) (bool, string) {
			a := float32(x)
			b := float32(y)
			c := a * b
			d := c / b
			e := d - a
			return e == float32(0), fmt.Sprintf("%e", e)
		},
		"float64 (x/y)*y": func(x, y int64) (bool, string) {
			a := float64(x)
			b := float64(y)
			c := a / b
			d := c * b
			e := d - a
			return e == float64(0), fmt.Sprintf("%e", e)
		},
		"float64 (x*y)/y": func(x, y int64) (bool, string) {
			a := float64(x)
			b := float64(y)
			c := a * b
			d := c / b
			e := d - a
			return e == float64(0), fmt.Sprintf("%e", e)
		},
		"decimal32 (x/y)*y":  apdQuoFirst(decimal32Context),
		"decimal32 (x*y)/y":  apdMulFirst(decimal32Context),
		"decimal64 (x/y)*y":  apdQuoFirst(decimal64Context),
		"decimal64 (x*y)/y":  apdMulFirst(decimal64Context),
		"decimal128 (x/y)*y": apdQuoFirst(decimal128Context),
		"decimal128 (x*y)/y": apdMulFirst(decimal128Context),
	}

	b.Logf("N: %d", b.N)
	for name, tc := range cases {
		b.Run(name, func(b *testing.B) {
			b.StopTimer()

			r := rand.New(rand.NewSource(0))
			xs := make([]int64, b.N)
			ys := make([]int64, b.N)
			for i := 0; i < b.N; i++ {
				xs[i] = int64(r.Int31())
				ys[i] = int64(r.Int31())
			}

			numCorrect := 0
			errs := make([]string, b.N)

			b.StartTimer()
			for i := 0; i < b.N; i++ {
				var correct bool
				correct, errs[i] = tc(xs[i], ys[i])
				if correct {
					numCorrect++
				}
			}
			b.StopTimer()

			totalErr := apd.New(0, 1)
			for i := 0; i < b.N; i++ {
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
			n.SetInt64(int64(b.N))
			_, err := decimal128Context.Quo(totalErr, totalErr, n)
			if err != nil {
				panic(err)
			}

			totalErrF, err := totalErr.Float64()
			if err != nil {
				panic(err)
			}

			b.Logf("%s Out of %d trials: %.2f%% Correct, %.2e Average Error",
				name,
				b.N,
				float64(numCorrect)/float64(b.N)*float64(100),
				totalErrF,
			)
		})
	}
}
