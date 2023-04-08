package types_test

import sdk "github.com/cosmos/cosmos-sdk/types"

func (s *coinTestSuite) TestMapCoinAdd() {
	cases := []struct {
		inputOne sdk.Coins
		inputTwo sdk.Coins
	}{
		{sdk.Coins{s.ca1, s.cm1}, sdk.Coins{s.ca1, s.cm1}},
		{sdk.Coins{s.ca0, s.cm1}, sdk.Coins{s.ca0, s.cm0}},
		{sdk.Coins{s.ca2}, sdk.Coins{s.cm0}},
		{sdk.Coins{s.ca1}, sdk.Coins{s.ca1, s.cm2}},
		{sdk.Coins{s.ca0, s.cm0}, sdk.Coins{s.ca0, s.cm0}},
	}

	for tcIndex, tc := range cases {
		expected := tc.inputOne.Add(tc.inputTwo...)
		m := sdk.NewMapCoins(tc.inputOne)
		m.Add(tc.inputTwo...)
		res := m.ToCoins()
		s.Require().True(res.IsValid())
		s.Require().Equal(expected, res, "sum of coins is incorrect, tc #%d", tcIndex)
	}

}
