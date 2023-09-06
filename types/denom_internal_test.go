package types

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
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
	atomUnit := math.LegacyOneDec() // 1 (base denom unit)

	s.Require().NoError(RegisterDenom(atom, atomUnit))
	s.Require().Error(RegisterDenom(atom, atomUnit))

	res, ok := GetDenomUnit(atom)
	s.Require().True(ok)
	s.Require().Equal(atomUnit, res)

	res, ok = GetDenomUnit(matom)
	s.Require().False(ok)
	s.Require().Equal(math.LegacyZeroDec(), res)

	err := SetBaseDenom(atom)
	s.Require().NoError(err)

	res, ok = GetDenomUnit(atom)
	s.Require().True(ok)
	s.Require().Equal(atomUnit, res)

	// reset registration
	baseDenom = ""
	denomUnits = map[string]math.LegacyDec{}
}

func (s *internalDenomTestSuite) TestConvertCoins() {
	atomUnit := math.LegacyOneDec() // 1 (base denom unit)
	s.Require().NoError(RegisterDenom(atom, atomUnit))

	matomUnit := math.LegacyNewDecWithPrec(1, 3) // 10^-3 (milli)
	s.Require().NoError(RegisterDenom(matom, matomUnit))

	uatomUnit := math.LegacyNewDecWithPrec(1, 6) // 10^-6 (micro)
	s.Require().NoError(RegisterDenom(uatom, uatomUnit))

	natomUnit := math.LegacyNewDecWithPrec(1, 9) // 10^-9 (nano)
	s.Require().NoError(RegisterDenom(natom, natomUnit))

	res, err := GetBaseDenom()
	s.Require().NoError(err)
	s.Require().Equal(res, natom)
	s.Require().Equal(NormalizeCoin(NewCoin(uatom, math.NewInt(1))), NewCoin(natom, math.NewInt(1000)))
	s.Require().Equal(NormalizeCoin(NewCoin(matom, math.NewInt(1))), NewCoin(natom, math.NewInt(1000000)))
	s.Require().Equal(NormalizeCoin(NewCoin(atom, math.NewInt(1))), NewCoin(natom, math.NewInt(1000000000)))

	coins, err := ParseCoinsNormalized("1atom,1matom,1uatom")
	s.Require().NoError(err)
	s.Require().Equal(coins, Coins{
		Coin{natom, math.NewInt(1000000000)},
		Coin{natom, math.NewInt(1000000)},
		Coin{natom, math.NewInt(1000)},
	})

	testCases := []struct {
		input  Coin
		denom  string
		result Coin
		expErr bool
	}{
		{NewCoin("foo", math.ZeroInt()), atom, Coin{}, true},
		{NewCoin(atom, math.ZeroInt()), "foo", Coin{}, true},
		{NewCoin(atom, math.ZeroInt()), "FOO", Coin{}, true},

		{NewCoin(atom, math.NewInt(5)), matom, NewCoin(matom, math.NewInt(5000)), false},       // atom => matom
		{NewCoin(atom, math.NewInt(5)), uatom, NewCoin(uatom, math.NewInt(5000000)), false},    // atom => uatom
		{NewCoin(atom, math.NewInt(5)), natom, NewCoin(natom, math.NewInt(5000000000)), false}, // atom => natom

		{NewCoin(uatom, math.NewInt(5000000)), matom, NewCoin(matom, math.NewInt(5000)), false},       // uatom => matom
		{NewCoin(uatom, math.NewInt(5000000)), natom, NewCoin(natom, math.NewInt(5000000000)), false}, // uatom => natom
		{NewCoin(uatom, math.NewInt(5000000)), atom, NewCoin(atom, math.NewInt(5)), false},            // uatom => atom

		{NewCoin(matom, math.NewInt(5000)), natom, NewCoin(natom, math.NewInt(5000000000)), false}, // matom => natom
		{NewCoin(matom, math.NewInt(5000)), uatom, NewCoin(uatom, math.NewInt(5000000)), false},    // matom => uatom
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
	baseDenom = ""
	denomUnits = map[string]math.LegacyDec{}
}

func (s *internalDenomTestSuite) TestConvertDecCoins() {
	atomUnit := math.LegacyOneDec() // 1 (base denom unit)
	s.Require().NoError(RegisterDenom(atom, atomUnit))

	matomUnit := math.LegacyNewDecWithPrec(1, 3) // 10^-3 (milli)
	s.Require().NoError(RegisterDenom(matom, matomUnit))

	uatomUnit := math.LegacyNewDecWithPrec(1, 6) // 10^-6 (micro)
	s.Require().NoError(RegisterDenom(uatom, uatomUnit))

	natomUnit := math.LegacyNewDecWithPrec(1, 9) // 10^-9 (nano)
	s.Require().NoError(RegisterDenom(natom, natomUnit))

	res, err := GetBaseDenom()
	s.Require().NoError(err)
	s.Require().Equal(res, natom)
	s.Require().Equal(NormalizeDecCoin(NewDecCoin(uatom, math.NewInt(1))), NewDecCoin(natom, math.NewInt(1000)))
	s.Require().Equal(NormalizeDecCoin(NewDecCoin(matom, math.NewInt(1))), NewDecCoin(natom, math.NewInt(1000000)))
	s.Require().Equal(NormalizeDecCoin(NewDecCoin(atom, math.NewInt(1))), NewDecCoin(natom, math.NewInt(1000000000)))

	coins, err := ParseCoinsNormalized("0.1atom,0.1matom,0.1uatom")
	s.Require().NoError(err)
	s.Require().Equal(coins, Coins{
		Coin{natom, math.NewInt(100000000)},
		Coin{natom, math.NewInt(100000)},
		Coin{natom, math.NewInt(100)},
	})

	testCases := []struct {
		input  DecCoin
		denom  string
		result DecCoin
		expErr bool
	}{
		{NewDecCoin("foo", math.ZeroInt()), atom, DecCoin{}, true},
		{NewDecCoin(atom, math.ZeroInt()), "foo", DecCoin{}, true},
		{NewDecCoin(atom, math.ZeroInt()), "FOO", DecCoin{}, true},

		// 0.5atom
		{NewDecCoinFromDec(atom, math.LegacyNewDecWithPrec(5, 1)), matom, NewDecCoin(matom, math.NewInt(500)), false},       // atom => matom
		{NewDecCoinFromDec(atom, math.LegacyNewDecWithPrec(5, 1)), uatom, NewDecCoin(uatom, math.NewInt(500000)), false},    // atom => uatom
		{NewDecCoinFromDec(atom, math.LegacyNewDecWithPrec(5, 1)), natom, NewDecCoin(natom, math.NewInt(500000000)), false}, // atom => natom

		{NewDecCoin(uatom, math.NewInt(5000000)), matom, NewDecCoin(matom, math.NewInt(5000)), false},       // uatom => matom
		{NewDecCoin(uatom, math.NewInt(5000000)), natom, NewDecCoin(natom, math.NewInt(5000000000)), false}, // uatom => natom
		{NewDecCoin(uatom, math.NewInt(5000000)), atom, NewDecCoin(atom, math.NewInt(5)), false},            // uatom => atom

		{NewDecCoin(matom, math.NewInt(5000)), natom, NewDecCoin(natom, math.NewInt(5000000000)), false}, // matom => natom
		{NewDecCoin(matom, math.NewInt(5000)), uatom, NewDecCoin(uatom, math.NewInt(5000000)), false},    // matom => uatom
	}

	for i, tc := range testCases {
		res, err := ConvertDecCoin(tc.input, tc.denom)
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
	baseDenom = ""
	denomUnits = map[string]math.LegacyDec{}
}

func (s *internalDenomTestSuite) TestDecOperationOrder() {
	dec, err := math.LegacyNewDecFromStr("11")
	s.Require().NoError(err)
	s.Require().NoError(RegisterDenom("unit1", dec))
	dec, err = math.LegacyNewDecFromStr("100000011")
	s.Require().NoError(err)
	s.Require().NoError(RegisterDenom("unit2", dec))

	coin, err := ConvertCoin(NewCoin("unit1", math.NewInt(100000011)), "unit2")
	s.Require().NoError(err)
	s.Require().Equal(coin, NewCoin("unit2", math.NewInt(11)))

	// reset registration
	baseDenom = ""
	denomUnits = map[string]math.LegacyDec{}
}

func (s *internalDenomTestSuite) TestSetBaseDenomError() {
	err := SetBaseDenom(atom)
	s.Require().Error(err)

	// reset registration
	baseDenom = ""
	denomUnits = map[string]math.LegacyDec{}
}
