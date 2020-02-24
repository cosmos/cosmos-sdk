package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
)

func (suite *KeeperTestSuite) TestParams() {
	ctx := suite.ctx.WithIsCheckTx(false)
	suite.Equal(types.DefaultParams(), suite.app.EvidenceKeeper.GetParams(ctx))
	suite.Equal(types.DefaultMaxEvidenceAge, suite.app.EvidenceKeeper.MaxEvidenceAge(ctx))
}
