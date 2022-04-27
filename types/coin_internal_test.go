package types

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "github.com/cosmos/cosmos-sdk/math"
)

func TestCoinTestSuite(t *testing.T) {
	suite.Run(t, new(coinInternalSuite))
}

type coinInternalSuite struct {
	suite.Suite
}

func (s *coinInternalSuite) TestIsSorted() {
	v := sdkmath.NewInt(1)
	cases := []struct {
		coins    Coins
		expected bool
	}{
		{Coins{}, true},
		{Coins{{"1", v}}, true},
		{Coins{{"1", v}, {"1", v}}, true},
		{Coins{{"1", v}, {"2", v}}, true},
		{Coins{{"1", v}, {"2", v}, {"2", v}}, true},

		{Coins{{"1", v}, {"0", v}}, false},
		{Coins{{"1", v}, {"0", v}, {"2", v}}, false},
		{Coins{{"1", v}, {"1", v}, {"0", v}}, false},
	}
	assert := s.Assert()
	for i, tc := range cases {
		assert.Equal(tc.expected, tc.coins.isSorted(), "testcase %d failed", i)
	}
}
