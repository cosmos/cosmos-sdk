package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *KeeperTestSuite) TestGetLastTokenizeShareRecordId() {
	app, ctx := suite.app, suite.ctx
	lastTokenizeShareRecordID := app.StakingKeeper.GetLastTokenizeShareRecordID(ctx)
	suite.Equal(lastTokenizeShareRecordID, uint64(0))
	app.StakingKeeper.SetLastTokenizeShareRecordID(ctx, 100)
	lastTokenizeShareRecordID = app.StakingKeeper.GetLastTokenizeShareRecordID(ctx)
	suite.Equal(lastTokenizeShareRecordID, uint64(100))
}

func (suite *KeeperTestSuite) TestGetTokenizeShareRecord() {
	app, ctx := suite.app, suite.ctx
	owner1, owner2 := suite.addrs[0], suite.addrs[1]

	tokenizeShareRecord1 := types.TokenizeShareRecord{
		Id:            0,
		Owner:         owner1.String(),
		ModuleAccount: "test-module-account-1",
		Validator:     "test-validator",
	}
	tokenizeShareRecord2 := types.TokenizeShareRecord{
		Id:            1,
		Owner:         owner2.String(),
		ModuleAccount: "test-module-account-2",
		Validator:     "test-validator",
	}
	tokenizeShareRecord3 := types.TokenizeShareRecord{
		Id:            2,
		Owner:         owner1.String(),
		ModuleAccount: "test-module-account-3",
		Validator:     "test-validator",
	}
	err := app.StakingKeeper.AddTokenizeShareRecord(ctx, tokenizeShareRecord1)
	suite.NoError(err)
	err = app.StakingKeeper.AddTokenizeShareRecord(ctx, tokenizeShareRecord2)
	suite.NoError(err)
	err = app.StakingKeeper.AddTokenizeShareRecord(ctx, tokenizeShareRecord3)
	suite.NoError(err)

	tokenizeShareRecord, err := app.StakingKeeper.GetTokenizeShareRecord(ctx, 2)
	suite.NoError(err)
	suite.Equal(tokenizeShareRecord, tokenizeShareRecord3)

	tokenizeShareRecord, err = app.StakingKeeper.GetTokenizeShareRecordByDenom(ctx, tokenizeShareRecord2.GetShareTokenDenom())
	suite.NoError(err)
	suite.Equal(tokenizeShareRecord, tokenizeShareRecord2)

	tokenizeShareRecords := app.StakingKeeper.GetAllTokenizeShareRecords(ctx)
	suite.Equal(len(tokenizeShareRecords), 3)

	tokenizeShareRecords = app.StakingKeeper.GetTokenizeShareRecordsByOwner(ctx, owner1)
	suite.Equal(len(tokenizeShareRecords), 2)

	tokenizeShareRecords = app.StakingKeeper.GetTokenizeShareRecordsByOwner(ctx, owner2)
	suite.Equal(len(tokenizeShareRecords), 1)
}
