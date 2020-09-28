package types

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"
)

type internalIntTestSuite struct {
	suite.Suite
}

func TestInternalIntTestSuite(t *testing.T) {
	suite.Run(t, new(internalIntTestSuite))
}

func (s *internalIntTestSuite) TestEncodingRandom() {
	for i := 0; i < 1000; i++ {
		n := rand.Int63()
		ni := NewInt(n)
		var ri Int

		str, err := ni.Marshal()
		s.Require().Nil(err)
		err = (&ri).Unmarshal(str)
		s.Require().Nil(err)

		s.Require().Equal(ni, ri, "binary mismatch; tc #%d, expected %s, actual %s", i, ni.String(), ri.String())
		s.Require().True(ni.i != ri.i, "pointer addresses are equal; tc #%d", i)

		bz, err := ni.MarshalJSON()
		s.Require().Nil(err)
		err = (&ri).UnmarshalJSON(bz)
		s.Require().Nil(err)

		s.Require().Equal(ni, ri, "json mismatch; tc #%d, expected %s, actual %s", i, ni.String(), ri.String())
		s.Require().True(ni.i != ri.i, "pointer addresses are equal; tc #%d", i)
	}

	for i := 0; i < 1000; i++ {
		n := rand.Uint64()
		ni := NewUint(n)
		var ri Uint

		str, err := ni.Marshal()
		s.Require().Nil(err)
		err = (&ri).Unmarshal(str)
		s.Require().Nil(err)

		s.Require().Equal(ni, ri, "binary mismatch; tc #%d, expected %s, actual %s", i, ni.String(), ri.String())
		s.Require().True(ni.i != ri.i, "pointer addresses are equal; tc #%d", i)

		bz, err := ni.MarshalJSON()
		s.Require().Nil(err)
		err = (&ri).UnmarshalJSON(bz)
		s.Require().Nil(err)

		s.Require().Equal(ni, ri, "json mismatch; tc #%d, expected %s, actual %s", i, ni.String(), ri.String())
		s.Require().True(ni.i != ri.i, "pointer addresses are equal; tc #%d", i)
	}
}

func (s *internalIntTestSuite) TestSerializationOverflow() {
	bx, _ := new(big.Int).SetString("91888242871839275229946405745257275988696311157297823662689937894645226298583", 10)
	x := Int{bx}
	y := new(Int)

	bz, err := x.Marshal()
	s.Require().NoError(err)

	// require deserialization to fail due to overflow
	s.Require().Error(y.Unmarshal(bz))

	// require JSON deserialization to fail due to overflow
	bz, err = x.MarshalJSON()
	s.Require().NoError(err)

	s.Require().Error(y.UnmarshalJSON(bz))
}

func (s *internalIntTestSuite) TestImmutabilityArithInt() {
	size := 500

	ops := []intop{
		intarith(Int.Add, (*big.Int).Add),
		intarith(Int.Sub, (*big.Int).Sub),
		intarith(Int.Mul, (*big.Int).Mul),
		intarith(Int.Quo, (*big.Int).Quo),
		intarithraw(Int.AddRaw, (*big.Int).Add),
		intarithraw(Int.SubRaw, (*big.Int).Sub),
		intarithraw(Int.MulRaw, (*big.Int).Mul),
		intarithraw(Int.QuoRaw, (*big.Int).Quo),
	}

	for i := 0; i < 100; i++ {
		uis := make([]Int, size)
		bis := make([]*big.Int, size)

		n := rand.Int63()
		ui := NewInt(n)
		bi := new(big.Int).SetInt64(n)

		for j := 0; j < size; j++ {
			op := ops[rand.Intn(len(ops))]
			uis[j], bis[j] = op(ui, bi)
		}

		for j := 0; j < size; j++ {
			s.Require().Equal(0, bis[j].Cmp(uis[j].BigInt()), "Int is different from *big.Int. tc #%d, Int %s, *big.Int %s", j, uis[j].String(), bis[j].String())
			s.Require().Equal(NewIntFromBigInt(bis[j]), uis[j], "Int is different from *big.Int. tc #%d, Int %s, *big.Int %s", j, uis[j].String(), bis[j].String())
			s.Require().True(uis[j].i != bis[j], "Pointer addresses are equal. tc #%d, Int %s, *big.Int %s", j, uis[j].String(), bis[j].String())
		}
	}
}

type intop func(Int, *big.Int) (Int, *big.Int)

func intarith(uifn func(Int, Int) Int, bifn func(*big.Int, *big.Int, *big.Int) *big.Int) intop {
	return func(ui Int, bi *big.Int) (Int, *big.Int) {
		r := rand.Int63()
		br := new(big.Int).SetInt64(r)
		return uifn(ui, NewInt(r)), bifn(new(big.Int), bi, br)
	}
}

func intarithraw(uifn func(Int, int64) Int, bifn func(*big.Int, *big.Int, *big.Int) *big.Int) intop {
	return func(ui Int, bi *big.Int) (Int, *big.Int) {
		r := rand.Int63()
		br := new(big.Int).SetInt64(r)
		return uifn(ui, r), bifn(new(big.Int), bi, br)
	}
}
