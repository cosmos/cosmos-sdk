package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type stakingTestSuite struct {
	suite.Suite
}

func TestStakingTestSuite(t *testing.T) {
	suite.Run(t, new(stakingTestSuite))
}

func (s *stakingTestSuite) SetupSuite() {
	s.T().Parallel()
}

func (s *stakingTestSuite) TestBondStatus() {
	s.Require().False(sdk.Unbonded.Equal(sdk.Bonded))
	s.Require().False(sdk.Unbonded.Equal(sdk.Unbonding))
	s.Require().False(sdk.Bonded.Equal(sdk.Unbonding))
	s.Require().Panicsf(func() { sdk.BondStatus(0).String() }, "invalid bond status") // nolint:govet
	s.Require().Equal(sdk.BondStatusUnbonded, sdk.Unbonded.String())
	s.Require().Equal(sdk.BondStatusBonded, sdk.Bonded.String())
	s.Require().Equal(sdk.BondStatusUnbonding, sdk.Unbonding.String())
}

func (s *stakingTestSuite) TestTokensToConsensusPower() {
	s.Require().Equal(int64(0), sdk.TokensToConsensusPower(sdk.NewInt(999_999)))
	s.Require().Equal(int64(1), sdk.TokensToConsensusPower(sdk.NewInt(1_000_000)))
}
