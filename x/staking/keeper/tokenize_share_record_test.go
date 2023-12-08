package keeper_test

import (
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *KeeperTestSuite) TestGetLastTokenizeShareRecordId() {
	ctx, keeper := suite.ctx, suite.stakingKeeper
	lastTokenizeShareRecordID := keeper.GetLastTokenizeShareRecordID(ctx)
	suite.Equal(lastTokenizeShareRecordID, uint64(0))
	keeper.SetLastTokenizeShareRecordID(ctx, 100)
	lastTokenizeShareRecordID = keeper.GetLastTokenizeShareRecordID(ctx)
	suite.Equal(lastTokenizeShareRecordID, uint64(100))
}

func (suite *KeeperTestSuite) TestGetTokenizeShareRecord() {
	ctx, keeper := suite.ctx, suite.stakingKeeper
	addrs := simtestutil.CreateIncrementalAccounts(2)

	owner1, owner2 := addrs[0], addrs[1]
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
	err := keeper.AddTokenizeShareRecord(ctx, tokenizeShareRecord1)
	suite.NoError(err)
	err = keeper.AddTokenizeShareRecord(ctx, tokenizeShareRecord2)
	suite.NoError(err)
	err = keeper.AddTokenizeShareRecord(ctx, tokenizeShareRecord3)
	suite.NoError(err)

	tokenizeShareRecord, err := keeper.GetTokenizeShareRecord(ctx, 2)
	suite.NoError(err)
	suite.Equal(tokenizeShareRecord, tokenizeShareRecord3)

	tokenizeShareRecord, err = keeper.GetTokenizeShareRecordByDenom(ctx, tokenizeShareRecord2.GetShareTokenDenom())
	suite.NoError(err)
	suite.Equal(tokenizeShareRecord, tokenizeShareRecord2)

	tokenizeShareRecords := keeper.GetAllTokenizeShareRecords(ctx)
	suite.Equal(len(tokenizeShareRecords), 3)

	tokenizeShareRecords = keeper.GetTokenizeShareRecordsByOwner(ctx, owner1)
	suite.Equal(len(tokenizeShareRecords), 2)

	tokenizeShareRecords = keeper.GetTokenizeShareRecordsByOwner(ctx, owner2)
	suite.Equal(len(tokenizeShareRecords), 1)
}
