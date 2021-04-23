package keeper_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestPowerReductionChange() {
	// modify power reduction
	newPowerReduction := sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(12), nil))
	params := suite.app.StakingKeeper.GetParams(suite.ctx)
	params.PowerReduction = newPowerReduction
	suite.app.StakingKeeper.SetParams(suite.ctx, params)

	// check power reduction change
	suite.Require().Equal(newPowerReduction, suite.app.StakingKeeper.PowerReduction(suite.ctx))
}

func (suite *KeeperTestSuite) TestTokensToConsensusPower() {
	suite.Require().Equal(int64(0), suite.app.StakingKeeper.TokensToConsensusPower(suite.ctx, sdk.DefaultPowerReduction.Sub(sdk.NewInt(1))))
	suite.Require().Equal(int64(1), suite.app.StakingKeeper.TokensToConsensusPower(suite.ctx, sdk.DefaultPowerReduction))
}

func (suite *KeeperTestSuite) TestTokensFromConsensusPower() {
	suite.Require().Equal(sdk.NewInt(0), suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, 0))
	suite.Require().Equal(sdk.DefaultPowerReduction, suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, 1))
}
