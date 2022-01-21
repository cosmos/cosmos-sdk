package types_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

var powerReduction = sdk.NewIntFromUint64(1000000)

type stakingTestSuite struct {
	suite.Suite
}

func TestStakingTestSuite(t *testing.T) {
	suite.Run(t, new(stakingTestSuite))
}

func (s *stakingTestSuite) SetupSuite() {
	s.T().Parallel()
}

func (s *stakingTestSuite) TestTokensToConsensusPower() {

	s.Require().Equal(int64(0), types.TokensToConsensusPower(sdk.NewInt(999_999), powerReduction))
	s.Require().Equal(int64(1), types.TokensToConsensusPower(sdk.NewInt(1_000_000), powerReduction))
	s.Require().Equal(int64(0), types.TokensToConsensusPower(sdk.NewIntFromBigInt(big.NewInt(math.MaxInt64).Add(big.NewInt(math.MaxInt64), big.NewInt(1))), sdk.NewIntFromUint64(1)))
	s.Require().Equal(int64(0), types.TokensToConsensusPower(sdk.NewInt(1_000_000), sdk.NewIntFromUint64(0)))
}
