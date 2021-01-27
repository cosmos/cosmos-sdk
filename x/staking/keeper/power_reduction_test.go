package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestPowerReductionChange(t *testing.T) {
	// TODO: Implement!
}

func (suite *KeeperTestSuite) TestTokensToConsensusPower() {
	suite.Require().Equal(int64(0), suite.app.StakingKeeper.TokensToConsensusPower(suite.ctx, sdk.NewInt(999_999)))
	suite.Require().Equal(int64(1), suite.app.StakingKeeper.TokensToConsensusPower(suite.ctx, sdk.NewInt(1_000_000)))
}

func (suite *KeeperTestSuite) TestTokensFromConsensusPower() {
	suite.Require().Equal(sdk.NewInt(0), suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, 0))
	suite.Require().Equal(sdk.NewInt(1_000_000), suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, 1))
}
