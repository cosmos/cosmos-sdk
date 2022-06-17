package keeper_test

import (
	"time"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func (suite *KeeperTestSuite) TestGetSetValidatorSigningInfo() {
	ctx := suite.ctx

	addrDels := suite.addrDels
	info, found := suite.slashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[2]))
	suite.Require().False(found)
	newInfo := types.NewValidatorSigningInfo(
		sdk.ConsAddress(addrDels[2]),
		int64(4),
		int64(3),
		time.Unix(2, 0),
		false,
		int64(10),
	)
	suite.slashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[2]), newInfo)
	info, found = suite.slashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[2]))
	suite.Require().True(found)
	suite.Require().Equal(info.StartHeight, int64(4))
	suite.Require().Equal(info.IndexOffset, int64(3))
	suite.Require().Equal(info.JailedUntil, time.Unix(2, 0).UTC())
	suite.Require().Equal(info.MissedBlocksCounter, int64(10))
}

func (suite *KeeperTestSuite) TestGetSetValidatorMissedBlockBitArray() {
	ctx := suite.ctx
	addrDels := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, ctx, 1, suite.stakingKeeper.TokensFromConsensusPower(ctx, 200))

	missed := suite.slashingKeeper.GetValidatorMissedBlockBitArray(ctx, sdk.ConsAddress(addrDels[0]), 0)
	suite.Require().False(missed) // treat empty key as not missed
	suite.slashingKeeper.SetValidatorMissedBlockBitArray(ctx, sdk.ConsAddress(addrDels[0]), 0, true)
	missed = suite.slashingKeeper.GetValidatorMissedBlockBitArray(ctx, sdk.ConsAddress(addrDels[0]), 0)
	suite.Require().True(missed) // now should be missed
}

func (suite *KeeperTestSuite) TestTombstoned() {
	ctx := suite.ctx
	addrDels := suite.addrDels

	suite.Require().Panics(func() { suite.slashingKeeper.Tombstone(ctx, sdk.ConsAddress(addrDels[4])) })
	suite.Require().False(suite.slashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(addrDels[4])))

	newInfo := types.NewValidatorSigningInfo(
		sdk.ConsAddress(addrDels[4]),
		int64(4),
		int64(3),
		time.Unix(2, 0),
		false,
		int64(10),
	)
	suite.slashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[4]), newInfo)

	suite.Require().False(suite.slashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(addrDels[4])))
	suite.slashingKeeper.Tombstone(ctx, sdk.ConsAddress(addrDels[4]))
	suite.Require().True(suite.slashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(addrDels[4])))
	suite.Require().Panics(func() { suite.slashingKeeper.Tombstone(ctx, sdk.ConsAddress(addrDels[4])) })
}

func (suite *KeeperTestSuite) TestJailUntil() {
	ctx := suite.ctx
	addrDels := suite.addrDels

	suite.Require().Panics(func() { suite.slashingKeeper.JailUntil(ctx, sdk.ConsAddress(addrDels[3]), time.Now()) })

	newInfo := types.NewValidatorSigningInfo(
		sdk.ConsAddress(addrDels[3]),
		int64(4),
		int64(3),
		time.Unix(2, 0),
		false,
		int64(10),
	)
	suite.slashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[3]), newInfo)
	suite.slashingKeeper.JailUntil(ctx, sdk.ConsAddress(addrDels[3]), time.Unix(253402300799, 0).UTC())

	info, ok := suite.slashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[3]))
	suite.Require().True(ok)
	suite.Require().Equal(time.Unix(253402300799, 0).UTC(), info.JailedUntil)
}
