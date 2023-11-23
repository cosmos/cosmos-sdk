package math

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type uintTestSuite struct {
	suite.Suite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(uintTestSuite))
}

func (s *uintTestSuite) SetupSuite() {
	s.T().Parallel()
}

func (s *uintTestSuite) TestUintPanics() {
	// Max Uint = 1.15e+77
	// Min Uint = 0
	u1 := NewUint(0)
	u2 := OneUint()

	s.Require().Equal(uint64(0), u1.Uint64())
	s.Require().Equal(uint64(1), u2.Uint64())

	s.Require().Panics(func() { NewUintFromBigInt(big.NewInt(-5)) })
	s.Require().Panics(func() { NewUintFromString("-1") })
	s.Require().NotPanics(func() {
		s.Require().True(NewUintFromString("0").Equal(ZeroUint()))
		s.Require().True(NewUintFromString("5").Equal(NewUint(5)))
	})

	// Overflow check
	s.Require().True(u1.Add(u1).Equal(ZeroUint()))
	s.Require().True(u1.Add(OneUint()).Equal(OneUint()))
	s.Require().Equal(uint64(0), u1.Uint64())
	s.Require().Equal(uint64(1), OneUint().Uint64())
	s.Require().Panics(func() { u1.SubUint64(2) })
	s.Require().True(u1.SubUint64(0).Equal(ZeroUint()))
	s.Require().True(u2.Add(OneUint()).Sub(OneUint()).Equal(OneUint()))    // i2 == 1
	s.Require().True(u2.Add(OneUint()).Mul(NewUint(5)).Equal(NewUint(10))) // i2 == 10
	s.Require().True(NewUint(7).Quo(NewUint(2)).Equal(NewUint(3)))
	s.Require().True(NewUint(0).Quo(NewUint(2)).Equal(ZeroUint()))
	s.Require().True(NewUint(5).MulUint64(4).Equal(NewUint(20)))
	s.Require().True(NewUint(5).MulUint64(0).Equal(ZeroUint()))

	uintmax := NewUintFromBigInt(new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1)))
	uintmin := ZeroUint()

	// divs by zero
	s.Require().Panics(func() { OneUint().Mul(ZeroUint().SubUint64(uint64(1))) })
	s.Require().Panics(func() { OneUint().QuoUint64(0) })
	s.Require().Panics(func() { OneUint().Quo(ZeroUint()) })
	s.Require().Panics(func() { ZeroUint().QuoUint64(0) })
	s.Require().Panics(func() { OneUint().Quo(ZeroUint().Sub(OneUint())) })
	s.Require().Panics(func() { uintmax.Add(OneUint()) })
	s.Require().Panics(func() { uintmax.Incr() })
	s.Require().Panics(func() { uintmin.Sub(OneUint()) })
	s.Require().Panics(func() { uintmin.Decr() })

	s.Require().NotPanics(func() { Uint{}.BigInt() })

	s.Require().Equal(uint64(0), MinUint(ZeroUint(), OneUint()).Uint64())
	s.Require().Equal(uint64(1), MaxUint(ZeroUint(), OneUint()).Uint64())

	// comparison ops
	s.Require().True(
		OneUint().GT(ZeroUint()),
	)
	s.Require().False(
		OneUint().LT(ZeroUint()),
	)
	s.Require().True(
		OneUint().GTE(ZeroUint()),
	)
	s.Require().False(
		OneUint().LTE(ZeroUint()),
	)

	s.Require().False(ZeroUint().GT(OneUint()))
	s.Require().True(ZeroUint().LT(OneUint()))
	s.Require().False(ZeroUint().GTE(OneUint()))
	s.Require().True(ZeroUint().LTE(OneUint()))
}

func (s *uintTestSuite) TestIsNil() {
	s.Require().False(OneUint().IsNil())
	s.Require().True(Uint{}.IsNil())
}

func (s *uintTestSuite) TestConvertToBigIntMutativeForUint() {
	r := big.NewInt(30)
	i := NewUintFromBigInt(r)

	// Compare value of BigInt & BigIntMut
	s.Require().Equal(i.BigInt(), i.BigIntMut())

	// Modify BigIntMut() pointer and ensure i.BigIntMut() & i.BigInt() change
	p1 := i.BigIntMut()
	p1.SetInt64(40)
	s.Require().Equal(big.NewInt(40), i.BigIntMut())
	s.Require().Equal(big.NewInt(40), i.BigInt())

	// Modify big.Int() pointer and ensure i.BigIntMut() & i.BigInt() don't change
	p2 := i.BigInt()
	p2.SetInt64(50)
	s.Require().NotEqual(big.NewInt(50), i.BigIntMut())
	s.Require().NotEqual(big.NewInt(50), i.BigInt())
}

func (s *uintTestSuite) TestArithUint() {
	for d := 0; d < 1000; d++ {
		n1 := uint64(rand.Uint32())
		u1 := NewUint(n1)
		n2 := uint64(rand.Uint32())
		u2 := NewUint(n2)

		cases := []struct {
			ures Uint
			nres uint64
		}{
			{u1.Add(u2), n1 + n2},
			{u1.Mul(u2), n1 * n2},
			{u1.Quo(u2), n1 / n2},
			{u1.AddUint64(n2), n1 + n2},
			{u1.MulUint64(n2), n1 * n2},
			{u1.QuoUint64(n2), n1 / n2},
			{MinUint(u1, u2), minuint(n1, n2)},
			{MaxUint(u1, u2), maxuint(n1, n2)},
			{u1.Incr(), n1 + 1},
		}

		for tcnum, tc := range cases {
			s.Require().Equal(tc.nres, tc.ures.Uint64(), "Uint arithmetic operation does not match with uint64 operation. tc #%d", tcnum)
		}

		if n2 > n1 {
			n1, n2 = n2, n1
			u1, u2 = NewUint(n1), NewUint(n2)
		}

		subs := []struct {
			ures Uint
			nres uint64
		}{
			{u1.Sub(u2), n1 - n2},
			{u1.SubUint64(n2), n1 - n2},
			{u1.Decr(), n1 - 1},
		}

		for tcnum, tc := range subs {
			s.Require().Equal(tc.nres, tc.ures.Uint64(), "Uint subtraction does not match with uint64 operation. tc #%d", tcnum)
		}
	}
}

func (s *uintTestSuite) TestCompUint() {
	for d := 0; d < 10000; d++ {
		n1 := rand.Uint64()
		i1 := NewUint(n1)
		n2 := rand.Uint64()
		i2 := NewUint(n2)

		cases := []struct {
			ires bool
			nres bool
		}{
			{i1.Equal(i2), n1 == n2},
			{i1.GT(i2), n1 > n2},
			{i1.LT(i2), n1 < n2},
			{i1.GTE(i2), !i1.LT(i2)},
			{!i1.GTE(i2), i1.LT(i2)},
			{i1.LTE(i2), n1 <= n2},
			{i2.LTE(i1), n2 <= n1},
		}

		for tcnum, tc := range cases {
			s.Require().Equal(tc.nres, tc.ires, "Uint comparison operation does not match with uint64 operation. tc #%d", tcnum)
		}
	}
}

func (s *uintTestSuite) TestImmutabilityAllUint() {
	ops := []func(*Uint){
		func(i *Uint) { _ = i.Add(NewUint(rand.Uint64())) },
		func(i *Uint) { _ = i.Sub(NewUint(rand.Uint64() % i.Uint64())) },
		func(i *Uint) { _ = i.Mul(randuint()) },
		func(i *Uint) { _ = i.Quo(randuint()) },
		func(i *Uint) { _ = i.AddUint64(rand.Uint64()) },
		func(i *Uint) { _ = i.SubUint64(rand.Uint64() % i.Uint64()) },
		func(i *Uint) { _ = i.MulUint64(rand.Uint64()) },
		func(i *Uint) { _ = i.QuoUint64(rand.Uint64()) },
		func(i *Uint) { _ = i.IsZero() },
		func(i *Uint) { _ = i.Equal(randuint()) },
		func(i *Uint) { _ = i.GT(randuint()) },
		func(i *Uint) { _ = i.GTE(randuint()) },
		func(i *Uint) { _ = i.LT(randuint()) },
		func(i *Uint) { _ = i.LTE(randuint()) },
		func(i *Uint) { _ = i.String() },
		func(i *Uint) { _ = i.Incr() },
		func(i *Uint) {
			if i.IsZero() {
				return
			}

			_ = i.Decr()
		},
	}

	for i := 0; i < 1000; i++ {
		n := rand.Uint64()
		ni := NewUint(n)

		for opnum, op := range ops {
			op(&ni)

			s.Require().Equal(n, ni.Uint64(), "Uint is modified by operation. #%d", opnum)
			s.Require().Equal(NewUint(n), ni, "Uint is modified by operation. #%d", opnum)
		}
	}
}

func (s *uintTestSuite) TestSafeSub() {
	testCases := []struct {
		x, y     Uint
		expected uint64
		panic    bool
	}{
		{NewUint(0), NewUint(0), 0, false},
		{NewUint(10), NewUint(5), 5, false},
		{NewUint(5), NewUint(10), 5, true},
		{NewUint(math.MaxUint64), NewUint(0), math.MaxUint64, false},
	}

	for i, tc := range testCases {
		tc := tc
		if tc.panic {
			s.Require().Panics(func() { tc.x.Sub(tc.y) })
			continue
		}
		s.Require().Equal(
			tc.expected, tc.x.Sub(tc.y).Uint64(),
			"invalid subtraction result; x: %s, y: %s, tc: #%d", tc.x, tc.y, i,
		)
	}
}

func (s *uintTestSuite) TestParseUint() {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    Uint
		wantErr bool
	}{
		{"malformed", args{"malformed"}, Uint{}, true},
		{"empty", args{""}, Uint{}, true},
		{"positive", args{"50"}, NewUint(uint64(50)), false},
		{"negative", args{"-1"}, Uint{}, true},
		{"zero", args{"0"}, ZeroUint(), false},
	}
	for _, tt := range tests {
		got, err := ParseUint(tt.args.s)
		if tt.wantErr {
			s.Require().Error(err)
			continue
		}
		s.Require().NoError(err)
		s.Require().True(got.Equal(tt.want))
	}
}

func (s *uintTestSuite) TestNewUintFromBigInt() {
	r := big.NewInt(42)
	i := NewUintFromBigInt(r)
	s.Require().Equal(r, i.BigInt())

	// modify r and ensure i doesn't change
	r = r.SetInt64(100)
	s.Require().NotEqual(r, i.BigInt())
}

func randuint() Uint {
	return NewUint(rand.Uint64())
}

func (s *uintTestSuite) TestRelativePow() {
	tests := []struct {
		args []Uint
		want Uint
	}{
		{[]Uint{ZeroUint(), ZeroUint(), OneUint()}, OneUint()},
		{[]Uint{ZeroUint(), ZeroUint(), NewUint(10)}, NewUint(1)},
		{[]Uint{ZeroUint(), OneUint(), NewUint(10)}, ZeroUint()},
		{[]Uint{NewUint(10), NewUint(2), OneUint()}, NewUint(100)},
		{[]Uint{NewUint(210), NewUint(2), NewUint(100)}, NewUint(441)},
		{[]Uint{NewUint(2100), NewUint(2), NewUint(1000)}, NewUint(4410)},
		{[]Uint{NewUint(1000000001547125958), NewUint(600), NewUint(1000000000000000000)}, NewUint(1000000928276004850)},
	}
	for i, tc := range tests {
		res := RelativePow(tc.args[0], tc.args[1], tc.args[2])
		s.Require().Equal(tc.want, res, "unexpected result for test case %d, input: %v, got: %v", i, tc.args, res)
	}
}

func minuint(i1, i2 uint64) uint64 {
	if i1 < i2 {
		return i1
	}
	return i2
}

func maxuint(i1, i2 uint64) uint64 {
	if i1 > i2 {
		return i1
	}
	return i2
}

func TestRoundTripMarshalToUint(t *testing.T) {
	values := []uint64{
		0,
		1,
		1 << 10,
		1<<10 - 3,
		1<<63 - 1,
		1<<32 - 7,
		1<<22 - 8,
		math.MaxUint64,
	}

	for _, value := range values {
		value := value
		t.Run(fmt.Sprintf("%d", value), func(t *testing.T) {
			t.Parallel()

			var scratch [20]byte
			uv := NewUint(value)
			n, err := uv.MarshalTo(scratch[:])
			if err != nil {
				t.Fatal(err)
			}
			rt := new(Uint)
			if err := rt.Unmarshal(scratch[:n]); err != nil {
				t.Fatal(err)
			}
			if !rt.Equal(uv) {
				t.Fatalf("roundtrip=%q != original=%q", rt, uv)
			}
		})
	}
}

func TestWeakUnmarshalNegativeSign(t *testing.T) {
	neg10, _ := new(big.Int).SetString("-10", 0)
	blob, err := neg10.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	ui := new(Uint)
	err = ui.Unmarshal(blob)
	if err == nil {
		t.Fatal("Failed to catch the negative value")
	}
	if errStr := err.Error(); !strings.Contains(errStr, "non-positive") {
		t.Fatalf("negative value not reported, got instead %q", errStr)
	}
}

func TestWeakUnmarshalOverflow(t *testing.T) {
	exp := new(big.Int).SetUint64(256)
	pos10, _ := new(big.Int).SetString("10", 0)
	exp10Pow256 := new(big.Int).Exp(pos10, exp, nil)
	blob, err := exp10Pow256.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	ui := new(Uint)
	err = ui.Unmarshal(blob)
	if err == nil {
		t.Fatal("Failed to catch the overflowed value")
	}
	if errStr := err.Error(); !strings.Contains(errStr, "out of range") {
		t.Fatalf("out of range value not reported, got instead %q", errStr)
	}
}
