package types_test

import (
	"math/big"
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type intTestSuite struct {
	suite.Suite
}

func TestIntTestSuite(t *testing.T) {
	suite.Run(t, new(intTestSuite))
}

func (s *intTestSuite) SetupSuite() {
	s.T().Parallel()
}

func (s *intTestSuite) TestFromInt64() {
	for n := 0; n < 20; n++ {
		r := rand.Int63()
		s.Require().Equal(r, sdk.NewInt(r).Int64())
	}
}

func (s *intTestSuite) TestFromUint64() {
	for n := 0; n < 20; n++ {
		r := rand.Uint64()
		s.Require().True(sdk.NewIntFromUint64(r).IsUint64())
		s.Require().Equal(r, sdk.NewIntFromUint64(r).Uint64())
	}
}

func (s *intTestSuite) TestIntPanic() {
	// Max Int = 2^255-1 = 5.789e+76
	// Min Int = -(2^255-1) = -5.789e+76
	s.Require().NotPanics(func() { sdk.NewIntWithDecimal(1, 76) })
	i1 := sdk.NewIntWithDecimal(1, 76)
	s.Require().NotPanics(func() { sdk.NewIntWithDecimal(2, 76) })
	i2 := sdk.NewIntWithDecimal(2, 76)
	s.Require().NotPanics(func() { sdk.NewIntWithDecimal(3, 76) })
	i3 := sdk.NewIntWithDecimal(3, 76)

	s.Require().Panics(func() { sdk.NewIntWithDecimal(6, 76) })
	s.Require().Panics(func() { sdk.NewIntWithDecimal(9, 80) })

	// Overflow check
	s.Require().NotPanics(func() { i1.Add(i1) })
	s.Require().NotPanics(func() { i2.Add(i2) })
	s.Require().Panics(func() { i3.Add(i3) })

	s.Require().NotPanics(func() { i1.Sub(i1.Neg()) })
	s.Require().NotPanics(func() { i2.Sub(i2.Neg()) })
	s.Require().Panics(func() { i3.Sub(i3.Neg()) })

	s.Require().Panics(func() { i1.Mul(i1) })
	s.Require().Panics(func() { i2.Mul(i2) })
	s.Require().Panics(func() { i3.Mul(i3) })

	s.Require().Panics(func() { i1.Neg().Mul(i1.Neg()) })
	s.Require().Panics(func() { i2.Neg().Mul(i2.Neg()) })
	s.Require().Panics(func() { i3.Neg().Mul(i3.Neg()) })

	// Underflow check
	i3n := i3.Neg()
	s.Require().NotPanics(func() { i3n.Sub(i1) })
	s.Require().NotPanics(func() { i3n.Sub(i2) })
	s.Require().Panics(func() { i3n.Sub(i3) })

	s.Require().NotPanics(func() { i3n.Add(i1.Neg()) })
	s.Require().NotPanics(func() { i3n.Add(i2.Neg()) })
	s.Require().Panics(func() { i3n.Add(i3.Neg()) })

	s.Require().Panics(func() { i1.Mul(i1.Neg()) })
	s.Require().Panics(func() { i2.Mul(i2.Neg()) })
	s.Require().Panics(func() { i3.Mul(i3.Neg()) })

	// Bound check
	intmax := sdk.NewIntFromBigInt(new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(255), nil), big.NewInt(1)))
	intmin := intmax.Neg()
	s.Require().NotPanics(func() { intmax.Add(sdk.ZeroInt()) })
	s.Require().NotPanics(func() { intmin.Sub(sdk.ZeroInt()) })
	s.Require().Panics(func() { intmax.Add(sdk.OneInt()) })
	s.Require().Panics(func() { intmin.Sub(sdk.OneInt()) })

	// Division-by-zero check
	s.Require().Panics(func() { i1.Quo(sdk.NewInt(0)) })

	s.Require().NotPanics(func() { sdk.Int{}.BigInt() })
}

// Tests below uses randomness
// Since we are using *big.Int as underlying value
// and (U/)Int is immutable value(see TestImmutability(U/)Int)
// it is safe to use randomness in the tests
func (s *intTestSuite) TestIdentInt() {
	for d := 0; d < 1000; d++ {
		n := rand.Int63()
		i := sdk.NewInt(n)

		ifromstr, ok := sdk.NewIntFromString(strconv.FormatInt(n, 10))
		s.Require().True(ok)

		cases := []int64{
			i.Int64(),
			i.BigInt().Int64(),
			ifromstr.Int64(),
			sdk.NewIntFromBigInt(big.NewInt(n)).Int64(),
			sdk.NewIntWithDecimal(n, 0).Int64(),
		}

		for tcnum, tc := range cases {
			s.Require().Equal(n, tc, "Int is modified during conversion. tc #%d", tcnum)
		}
	}
}

func minint(i1, i2 int64) int64 {
	if i1 < i2 {
		return i1
	}
	return i2
}

func maxint(i1, i2 int64) int64 {
	if i1 > i2 {
		return i1
	}
	return i2
}

func (s *intTestSuite) TestArithInt() {
	for d := 0; d < 1000; d++ {
		n1 := int64(rand.Int31())
		i1 := sdk.NewInt(n1)
		n2 := int64(rand.Int31())
		i2 := sdk.NewInt(n2)

		cases := []struct {
			ires sdk.Int
			nres int64
		}{
			{i1.Add(i2), n1 + n2},
			{i1.Sub(i2), n1 - n2},
			{i1.Mul(i2), n1 * n2},
			{i1.Quo(i2), n1 / n2},
			{i1.AddRaw(n2), n1 + n2},
			{i1.SubRaw(n2), n1 - n2},
			{i1.MulRaw(n2), n1 * n2},
			{i1.QuoRaw(n2), n1 / n2},
			{sdk.MinInt(i1, i2), minint(n1, n2)},
			{sdk.MaxInt(i1, i2), maxint(n1, n2)},
			{i1.Neg(), -n1},
		}

		for tcnum, tc := range cases {
			s.Require().Equal(tc.nres, tc.ires.Int64(), "Int arithmetic operation does not match with int64 operation. tc #%d", tcnum)
		}
	}

}

func (s *intTestSuite) TestCompInt() {
	for d := 0; d < 1000; d++ {
		n1 := int64(rand.Int31())
		i1 := sdk.NewInt(n1)
		n2 := int64(rand.Int31())
		i2 := sdk.NewInt(n2)

		cases := []struct {
			ires bool
			nres bool
		}{
			{i1.Equal(i2), n1 == n2},
			{i1.GT(i2), n1 > n2},
			{i1.LT(i2), n1 < n2},
			{i1.LTE(i2), n1 <= n2},
		}

		for tcnum, tc := range cases {
			s.Require().Equal(tc.nres, tc.ires, "Int comparison operation does not match with int64 operation. tc #%d", tcnum)
		}
	}
}

func randint() sdk.Int {
	return sdk.NewInt(rand.Int63())
}

func (s *intTestSuite) TestImmutabilityAllInt() {
	ops := []func(*sdk.Int){
		func(i *sdk.Int) { _ = i.Add(randint()) },
		func(i *sdk.Int) { _ = i.Sub(randint()) },
		func(i *sdk.Int) { _ = i.Mul(randint()) },
		func(i *sdk.Int) { _ = i.Quo(randint()) },
		func(i *sdk.Int) { _ = i.AddRaw(rand.Int63()) },
		func(i *sdk.Int) { _ = i.SubRaw(rand.Int63()) },
		func(i *sdk.Int) { _ = i.MulRaw(rand.Int63()) },
		func(i *sdk.Int) { _ = i.QuoRaw(rand.Int63()) },
		func(i *sdk.Int) { _ = i.Neg() },
		func(i *sdk.Int) { _ = i.IsZero() },
		func(i *sdk.Int) { _ = i.Sign() },
		func(i *sdk.Int) { _ = i.Equal(randint()) },
		func(i *sdk.Int) { _ = i.GT(randint()) },
		func(i *sdk.Int) { _ = i.LT(randint()) },
		func(i *sdk.Int) { _ = i.String() },
	}

	for i := 0; i < 1000; i++ {
		n := rand.Int63()
		ni := sdk.NewInt(n)

		for opnum, op := range ops {
			op(&ni)

			s.Require().Equal(n, ni.Int64(), "Int is modified by operation. tc #%d", opnum)
			s.Require().Equal(sdk.NewInt(n), ni, "Int is modified by operation. tc #%d", opnum)
		}
	}
}

func (s *intTestSuite) TestEncodingTableInt() {
	var i sdk.Int

	cases := []struct {
		i      sdk.Int
		jsonBz []byte
		rawBz  []byte
	}{
		{
			sdk.NewInt(0),
			[]byte("\"0\""),
			[]byte{0x30},
		},
		{
			sdk.NewInt(100),
			[]byte("\"100\""),
			[]byte{0x31, 0x30, 0x30},
		},
		{
			sdk.NewInt(-100),
			[]byte("\"-100\""),
			[]byte{0x2d, 0x31, 0x30, 0x30},
		},
		{
			sdk.NewInt(51842),
			[]byte("\"51842\""),
			[]byte{0x35, 0x31, 0x38, 0x34, 0x32},
		},
		{
			sdk.NewInt(-51842),
			[]byte("\"-51842\""),
			[]byte{0x2d, 0x35, 0x31, 0x38, 0x34, 0x32},
		},
		{
			sdk.NewInt(19513368),
			[]byte("\"19513368\""),
			[]byte{0x31, 0x39, 0x35, 0x31, 0x33, 0x33, 0x36, 0x38},
		},
		{
			sdk.NewInt(-19513368),
			[]byte("\"-19513368\""),
			[]byte{0x2d, 0x31, 0x39, 0x35, 0x31, 0x33, 0x33, 0x36, 0x38},
		},
		{
			sdk.NewInt(999999999999),
			[]byte("\"999999999999\""),
			[]byte{0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39},
		},
		{
			sdk.NewInt(-999999999999),
			[]byte("\"-999999999999\""),
			[]byte{0x2d, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39},
		},
	}

	for tcnum, tc := range cases {
		bz, err := tc.i.MarshalJSON()
		s.Require().Nil(err, "Error marshaling Int. tc #%d, err %s", tcnum, err)
		s.Require().Equal(tc.jsonBz, bz, "Marshaled value is different from exported. tc #%d", tcnum)

		err = (&i).UnmarshalJSON(bz)
		s.Require().Nil(err, "Error unmarshaling Int. tc #%d, err %s", tcnum, err)
		s.Require().Equal(tc.i, i, "Unmarshaled value is different from exported. tc #%d", tcnum)

		bz, err = tc.i.Marshal()
		s.Require().Nil(err, "Error marshaling Int. tc #%d, err %s", tcnum, err)
		s.Require().Equal(tc.rawBz, bz, "Marshaled value is different from exported. tc #%d", tcnum)

		err = (&i).Unmarshal(bz)
		s.Require().Nil(err, "Error unmarshaling Int. tc #%d, err %s", tcnum, err)
		s.Require().Equal(tc.i, i, "Unmarshaled value is different from exported. tc #%d", tcnum)
	}
}

func (s *intTestSuite) TestEncodingTableUint() {
	var i sdk.Uint

	cases := []struct {
		i      sdk.Uint
		jsonBz []byte
		rawBz  []byte
	}{
		{
			sdk.NewUint(0),
			[]byte("\"0\""),
			[]byte{0x30},
		},
		{
			sdk.NewUint(100),
			[]byte("\"100\""),
			[]byte{0x31, 0x30, 0x30},
		},
		{
			sdk.NewUint(51842),
			[]byte("\"51842\""),
			[]byte{0x35, 0x31, 0x38, 0x34, 0x32},
		},
		{
			sdk.NewUint(19513368),
			[]byte("\"19513368\""),
			[]byte{0x31, 0x39, 0x35, 0x31, 0x33, 0x33, 0x36, 0x38},
		},
		{
			sdk.NewUint(999999999999),
			[]byte("\"999999999999\""),
			[]byte{0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39},
		},
	}

	for tcnum, tc := range cases {
		bz, err := tc.i.MarshalJSON()
		s.Require().Nil(err, "Error marshaling Int. tc #%d, err %s", tcnum, err)
		s.Require().Equal(tc.jsonBz, bz, "Marshaled value is different from exported. tc #%d", tcnum)

		err = (&i).UnmarshalJSON(bz)
		s.Require().Nil(err, "Error unmarshaling Int. tc #%d, err %s", tcnum, err)
		s.Require().Equal(tc.i, i, "Unmarshaled value is different from exported. tc #%d", tcnum)

		bz, err = tc.i.Marshal()
		s.Require().Nil(err, "Error marshaling Int. tc #%d, err %s", tcnum, err)
		s.Require().Equal(tc.rawBz, bz, "Marshaled value is different from exported. tc #%d", tcnum)

		err = (&i).Unmarshal(bz)
		s.Require().Nil(err, "Error unmarshaling Int. tc #%d, err %s", tcnum, err)
		s.Require().Equal(tc.i, i, "Unmarshaled value is different from exported. tc #%d", tcnum)
	}
}

func (s *intTestSuite) TestIntMod() {
	tests := []struct {
		name      string
		x         int64
		y         int64
		ret       int64
		wantPanic bool
	}{
		{"3 % 10", 3, 10, 3, false},
		{"10 % 3", 10, 3, 1, false},
		{"4 % 2", 4, 2, 0, false},
		{"2 % 0", 2, 0, 0, true},
	}

	for _, tt := range tests {
		if tt.wantPanic {
			s.Require().Panics(func() { sdk.NewInt(tt.x).Mod(sdk.NewInt(tt.y)) })
			s.Require().Panics(func() { sdk.NewInt(tt.x).ModRaw(tt.y) })
			return
		}
		s.Require().True(sdk.NewInt(tt.x).Mod(sdk.NewInt(tt.y)).Equal(sdk.NewInt(tt.ret)))
		s.Require().True(sdk.NewInt(tt.x).ModRaw(tt.y).Equal(sdk.NewInt(tt.ret)))
	}
}

func (s *intTestSuite) TestIntEq() {
	_, resp, _, _, _ := sdk.IntEq(s.T(), sdk.ZeroInt(), sdk.ZeroInt())
	s.Require().True(resp)
	_, resp, _, _, _ = sdk.IntEq(s.T(), sdk.OneInt(), sdk.ZeroInt())
	s.Require().False(resp)
}
