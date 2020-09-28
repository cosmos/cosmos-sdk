package types

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

var (
	atom  = "atom"  // 1 (base denom unit)
	matom = "matom" // 10^-3 (milli)
	uatom = "uatom" // 10^-6 (micro)
	natom = "natom" // 10^-9 (nano)
)

type internalDenomTestSuite struct {
	suite.Suite
}

func TestInternalDenomTestSuite(t *testing.T) {
	suite.Run(t, new(internalDenomTestSuite))
}

func (s *internalDenomTestSuite) TestRegisterDenom() {
	atomUnit := OneDec() // 1 (base denom unit)

	s.Require().NoError(RegisterDenom(atom, atomUnit))
	s.Require().Error(RegisterDenom(atom, atomUnit))

	res, ok := GetDenomUnit(atom)
	s.Require().True(ok)
	s.Require().Equal(atomUnit, res)

	res, ok = GetDenomUnit(matom)
	s.Require().False(ok)
	s.Require().Equal(ZeroDec(), res)

	// reset registration
	denomUnits = map[string]Dec{}
}

func (s *internalDenomTestSuite) TestConvertCoins() {
	atomUnit := OneDec() // 1 (base denom unit)
	s.Require().NoError(RegisterDenom(atom, atomUnit))

	matomUnit := NewDecWithPrec(1, 3) // 10^-3 (milli)
	s.Require().NoError(RegisterDenom(matom, matomUnit))

	uatomUnit := NewDecWithPrec(1, 6) // 10^-6 (micro)
	s.Require().NoError(RegisterDenom(uatom, uatomUnit))

	natomUnit := NewDecWithPrec(1, 9) // 10^-9 (nano)
	s.Require().NoError(RegisterDenom(natom, natomUnit))

	testCases := []struct {
		input  Coin
		denom  string
		result Coin
		expErr bool
	}{
		{NewCoin("foo", ZeroInt()), atom, Coin{}, true},
		{NewCoin(atom, ZeroInt()), "foo", Coin{}, true},
		{NewCoin(atom, ZeroInt()), "FOO", Coin{}, true},

		{NewCoin(atom, NewInt(5)), matom, NewCoin(matom, NewInt(5000)), false},       // atom => matom
		{NewCoin(atom, NewInt(5)), uatom, NewCoin(uatom, NewInt(5000000)), false},    // atom => uatom
		{NewCoin(atom, NewInt(5)), natom, NewCoin(natom, NewInt(5000000000)), false}, // atom => natom

		{NewCoin(uatom, NewInt(5000000)), matom, NewCoin(matom, NewInt(5000)), false},       // uatom => matom
		{NewCoin(uatom, NewInt(5000000)), natom, NewCoin(natom, NewInt(5000000000)), false}, // uatom => natom
		{NewCoin(uatom, NewInt(5000000)), atom, NewCoin(atom, NewInt(5)), false},            // uatom => atom

		{NewCoin(matom, NewInt(5000)), natom, NewCoin(natom, NewInt(5000000000)), false}, // matom => natom
		{NewCoin(matom, NewInt(5000)), uatom, NewCoin(uatom, NewInt(5000000)), false},    // matom => uatom
	}

	for i, tc := range testCases {
		res, err := ConvertCoin(tc.input, tc.denom)
		s.Require().Equal(
			tc.expErr, err != nil,
			"unexpected error; tc: #%d, input: %s, denom: %s", i+1, tc.input, tc.denom,
		)
		s.Require().Equal(
			tc.result, res,
			"invalid result; tc: #%d, input: %s, denom: %s", i+1, tc.input, tc.denom,
		)
	}

	// reset registration
	denomUnits = map[string]Dec{}
}
