package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestTokensToConsensusPower() {
	s.Require().Equal(int64(0), s.stakingKeeper.TokensToConsensusPower(sdk.DefaultPowerReduction.Sub(sdk.NewInt(1))))
	s.Require().Equal(int64(1), s.stakingKeeper.TokensToConsensusPower(sdk.DefaultPowerReduction))
}

func (s *KeeperTestSuite) TestTokensFromConsensusPower() {
	s.Require().Equal(sdk.NewInt(0), s.stakingKeeper.TokensFromConsensusPower(0))
	s.Require().Equal(sdk.DefaultPowerReduction, s.stakingKeeper.TokensFromConsensusPower(1))
}
