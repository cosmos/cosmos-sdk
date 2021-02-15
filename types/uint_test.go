package types_test

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	u1 := sdk.NewUint(0)
	u2 := sdk.OneUint()

	s.Require().Equal(uint64(0), u1.Uint64())
	s.Require().Equal(uint64(1), u2.Uint64())

	s.Require().Panics(func() { sdk.NewUintFromBigInt(big.NewInt(-5)) })
	s.Require().Panics(func() { sdk.NewUintFromString("-1") })
	s.Require().NotPanics(func() {
		s.Require().True(sdk.NewUintFromString("0").Equal(sdk.ZeroUint()))
		s.Require().True(sdk.NewUintFromString("5").Equal(sdk.NewUint(5)))
	})

	// Overflow check
	s.Require().True(u1.Add(u1).Equal(sdk.ZeroUint()))
	s.Require().True(u1.Add(sdk.OneUint()).Equal(sdk.OneUint()))
	s.Require().Equal(uint64(0), u1.Uint64())
	s.Require().Equal(uint64(1), sdk.OneUint().Uint64())
	s.Require().Panics(func() { u1.SubUint64(2) })
	s.Require().True(u1.SubUint64(0).Equal(sdk.ZeroUint()))
	s.Require().True(u2.Add(sdk.OneUint()).Sub(sdk.OneUint()).Equal(sdk.OneUint()))    // i2 == 1
	s.Require().True(u2.Add(sdk.OneUint()).Mul(sdk.NewUint(5)).Equal(sdk.NewUint(10))) // i2 == 10
	s.Require().True(sdk.NewUint(7).Quo(sdk.NewUint(2)).Equal(sdk.NewUint(3)))
	s.Require().True(sdk.NewUint(0).Quo(sdk.NewUint(2)).Equal(sdk.ZeroUint()))
	s.Require().True(sdk.NewUint(5).MulUint64(4).Equal(sdk.NewUint(20)))
	s.Require().True(sdk.NewUint(5).MulUint64(0).Equal(sdk.ZeroUint()))

	uintmax := sdk.NewUintFromBigInt(new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1)))
	uintmin := sdk.ZeroUint()

	// divs by zero
	s.Require().Panics(func() { sdk.OneUint().Mul(sdk.ZeroUint().SubUint64(uint64(1))) })
	s.Require().Panics(func() { sdk.OneUint().QuoUint64(0) })
	s.Require().Panics(func() { sdk.OneUint().Quo(sdk.ZeroUint()) })
	s.Require().Panics(func() { sdk.ZeroUint().QuoUint64(0) })
	s.Require().Panics(func() { sdk.OneUint().Quo(sdk.ZeroUint().Sub(sdk.OneUint())) })
	s.Require().Panics(func() { uintmax.Add(sdk.OneUint()) })
	s.Require().Panics(func() { uintmax.Incr() })
	s.Require().Panics(func() { uintmin.Sub(sdk.OneUint()) })
	s.Require().Panics(func() { uintmin.Decr() })

	s.Require().Equal(uint64(0), sdk.MinUint(sdk.ZeroUint(), sdk.OneUint()).Uint64())
	s.Require().Equal(uint64(1), sdk.MaxUint(sdk.ZeroUint(), sdk.OneUint()).Uint64())

	// comparison ops
	s.Require().True(
		sdk.OneUint().GT(sdk.ZeroUint()),
	)
	s.Require().False(
		sdk.OneUint().LT(sdk.ZeroUint()),
	)
	s.Require().True(
		sdk.OneUint().GTE(sdk.ZeroUint()),
	)
	s.Require().False(
		sdk.OneUint().LTE(sdk.ZeroUint()),
	)

	s.Require().False(sdk.ZeroUint().GT(sdk.OneUint()))
	s.Require().True(sdk.ZeroUint().LT(sdk.OneUint()))
	s.Require().False(sdk.ZeroUint().GTE(sdk.OneUint()))
	s.Require().True(sdk.ZeroUint().LTE(sdk.OneUint()))
}

func (s *uintTestSuite) TestArithUint() {
	for d := 0; d < 1000; d++ {
		n1 := uint64(rand.Uint32())
		u1 := sdk.NewUint(n1)
		n2 := uint64(rand.Uint32())
		u2 := sdk.NewUint(n2)

		cases := []struct {
			ures sdk.Uint
			nres uint64
		}{
			{u1.Add(u2), n1 + n2},
			{u1.Mul(u2), n1 * n2},
			{u1.Quo(u2), n1 / n2},
			{u1.AddUint64(n2), n1 + n2},
			{u1.MulUint64(n2), n1 * n2},
			{u1.QuoUint64(n2), n1 / n2},
			{sdk.MinUint(u1, u2), minuint(n1, n2)},
			{sdk.MaxUint(u1, u2), maxuint(n1, n2)},
			{u1.Incr(), n1 + 1},
		}

		for tcnum, tc := range cases {
			s.Require().Equal(tc.nres, tc.ures.Uint64(), "Uint arithmetic operation does not match with uint64 operation. tc #%d", tcnum)
		}

		if n2 > n1 {
			n1, n2 = n2, n1
			u1, u2 = sdk.NewUint(n1), sdk.NewUint(n2)
		}

		subs := []struct {
			ures sdk.Uint
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
		i1 := sdk.NewUint(n1)
		n2 := rand.Uint64()
		i2 := sdk.NewUint(n2)

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
	ops := []func(*sdk.Uint){
		func(i *sdk.Uint) { _ = i.Add(sdk.NewUint(rand.Uint64())) },
		func(i *sdk.Uint) { _ = i.Sub(sdk.NewUint(rand.Uint64() % i.Uint64())) },
		func(i *sdk.Uint) { _ = i.Mul(randuint()) },
		func(i *sdk.Uint) { _ = i.Quo(randuint()) },
		func(i *sdk.Uint) { _ = i.AddUint64(rand.Uint64()) },
		func(i *sdk.Uint) { _ = i.SubUint64(rand.Uint64() % i.Uint64()) },
		func(i *sdk.Uint) { _ = i.MulUint64(rand.Uint64()) },
		func(i *sdk.Uint) { _ = i.QuoUint64(rand.Uint64()) },
		func(i *sdk.Uint) { _ = i.IsZero() },
		func(i *sdk.Uint) { _ = i.Equal(randuint()) },
		func(i *sdk.Uint) { _ = i.GT(randuint()) },
		func(i *sdk.Uint) { _ = i.GTE(randuint()) },
		func(i *sdk.Uint) { _ = i.LT(randuint()) },
		func(i *sdk.Uint) { _ = i.LTE(randuint()) },
		func(i *sdk.Uint) { _ = i.String() },
		func(i *sdk.Uint) { _ = i.Incr() },
		func(i *sdk.Uint) {
			if i.IsZero() {
				return
			}

			_ = i.Decr()
		},
	}

	for i := 0; i < 1000; i++ {
		n := rand.Uint64()
		ni := sdk.NewUint(n)

		for opnum, op := range ops {
			op(&ni)

			s.Require().Equal(n, ni.Uint64(), "Uint is modified by operation. #%d", opnum)
			s.Require().Equal(sdk.NewUint(n), ni, "Uint is modified by operation. #%d", opnum)
		}
	}
}

func (s *uintTestSuite) TestSafeSub() {
	testCases := []struct {
		x, y     sdk.Uint
		expected uint64
		panic    bool
	}{
		{sdk.NewUint(0), sdk.NewUint(0), 0, false},
		{sdk.NewUint(10), sdk.NewUint(5), 5, false},
		{sdk.NewUint(5), sdk.NewUint(10), 5, true},
		{sdk.NewUint(math.MaxUint64), sdk.NewUint(0), math.MaxUint64, false},
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
		want    sdk.Uint
		wantErr bool
	}{
		{"malformed", args{"malformed"}, sdk.Uint{}, true},
		{"empty", args{""}, sdk.Uint{}, true},
		{"positive", args{"50"}, sdk.NewUint(uint64(50)), false},
		{"negative", args{"-1"}, sdk.Uint{}, true},
		{"zero", args{"0"}, sdk.ZeroUint(), false},
	}
	for _, tt := range tests {
		got, err := sdk.ParseUint(tt.args.s)
		if tt.wantErr {
			s.Require().Error(err)
			continue
		}
		s.Require().NoError(err)
		s.Require().True(got.Equal(tt.want))
	}
}

func randuint() sdk.Uint {
	return sdk.NewUint(rand.Uint64())
}

func (s *uintTestSuite) TestRelativePow() {
	tests := []struct {
		args []sdk.Uint
		want sdk.Uint
	}{
		{[]sdk.Uint{sdk.ZeroUint(), sdk.ZeroUint(), sdk.OneUint()}, sdk.OneUint()},
		{[]sdk.Uint{sdk.ZeroUint(), sdk.ZeroUint(), sdk.NewUint(10)}, sdk.NewUint(10)},
		{[]sdk.Uint{sdk.ZeroUint(), sdk.OneUint(), sdk.NewUint(10)}, sdk.ZeroUint()},
		{[]sdk.Uint{sdk.NewUint(10), sdk.NewUint(2), sdk.OneUint()}, sdk.NewUint(100)},
		{[]sdk.Uint{sdk.NewUint(210), sdk.NewUint(2), sdk.NewUint(100)}, sdk.NewUint(441)},
		{[]sdk.Uint{sdk.NewUint(2100), sdk.NewUint(2), sdk.NewUint(1000)}, sdk.NewUint(4410)},
		{[]sdk.Uint{sdk.NewUint(1000000001547125958), sdk.NewUint(600), sdk.NewUint(1000000000000000000)}, sdk.NewUint(1000000928276004850)},
	}
	for i, tc := range tests {
		res := sdk.RelativePow(tc.args[0], tc.args[1], tc.args[2])
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
	var values = []uint64{
		0,
		1,
		1 << 10,
		1<<10 - 3,
		1<<63 - 1,
		1<<32 - 7,
		1<<22 - 8,
	}

	for _, value := range values {
		value := value
		t.Run(fmt.Sprintf("%d", value), func(t *testing.T) {
			t.Parallel()

			var scratch [20]byte
			uv := sdk.NewUint(value)
			n, err := uv.MarshalTo(scratch[:])
			if err != nil {
				t.Fatal(err)
			}
			rt := new(sdk.Uint)
			if err := rt.Unmarshal(scratch[:n]); err != nil {
				t.Fatal(err)
			}
			if !rt.Equal(uv) {
				t.Fatalf("roundtrip=%q != original=%q", rt, uv)
			}
		})
	}
}
