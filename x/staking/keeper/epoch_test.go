package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *KeeperTestSuite) TestEpochSaveLoad() {
	app, ctx, vals := suite.app, suite.ctx, suite.vals
	delAddr := suite.addrs[0]
	valAddr := vals[0].GetOperator()

	valTokens := sdk.TokensFromConsensusPower(100)
	validCoin := sdk.NewCoin(sdk.DefaultBondDenom, valTokens)

	originMsg := types.NewMsgDelegate(delAddr, valAddr, validCoin)

	epochNumber := int64(0)
	app.StakingKeeper.QueueMsgForEpoch(ctx, epochNumber, originMsg)
	nextActionID := app.StakingKeeper.GetNewActionID(ctx)
	suite.Require().Greater(nextActionID, uint64(1), "nextActionID should be greater than 1")

	actionID := nextActionID - 1
	savedMsg := app.StakingKeeper.GetEpochAction(ctx, epochNumber, actionID)

	suite.Require().Equal(savedMsg, originMsg, "savedMsg should be equal to originMsg")
}
