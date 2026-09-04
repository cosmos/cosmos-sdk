package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
)

// These tests probe the invariant that x/feegrant maintains TWO pieces of state per grant:
//
//	FeeAllowance      (grantee, granter)            -> Grant
//	FeeAllowanceQueue (expiration, grantee, granter) -> bool
//
// GrantAllowance writes both. revokeAllowance removes both. UpdateAllowance writes only the
// first, and RemoveExpiredAllowances deletes grants purely from the queue key without ever
// re-reading the grant's own expiration. So any UpdateAllowance that changes the expiration
// leaves the queue disagreeing with the grant.

// Case A: extend an allowance's expiry. The stale queue entry at the OLD expiry causes the
// pruner to delete a grant that is still valid.
func (suite *KeeperTestSuite) TestUpdateAllowanceExtendedExpiryIsPrunedEarly() {
	granter, grantee := suite.addrs[0], suite.addrs[1]

	oldExp := suite.ctx.BlockTime().AddDate(0, 0, 1)  // +1 day
	newExp := suite.ctx.BlockTime().AddDate(0, 0, 10) // +10 days

	err := suite.feegrantKeeper.GrantAllowance(suite.ctx, granter, grantee, &feegrant.BasicAllowance{
		SpendLimit: suite.atom,
		Expiration: &oldExp,
	})
	suite.Require().NoError(err)

	// The granter extends the allowance to +10 days.
	err = suite.feegrantKeeper.UpdateAllowance(suite.ctx, granter, grantee, &feegrant.BasicAllowance{
		SpendLimit: suite.atom,
		Expiration: &newExp,
	})
	suite.Require().NoError(err)

	// Move to just after the OLD expiry. The grant is still valid for another 9 days.
	ctx := suite.ctx.WithBlockTime(oldExp.Add(time.Second))

	stored, err := suite.feegrantKeeper.GetAllowance(ctx, granter, grantee)
	suite.Require().NoError(err)
	suite.Require().NotNil(stored, "precondition: the grant exists before pruning")
	exp, err := stored.ExpiresAt()
	suite.Require().NoError(err)
	suite.Require().True(exp.Equal(newExp), "precondition: stored grant carries the NEW expiration")

	suite.Require().NoError(suite.feegrantKeeper.RemoveExpiredAllowances(ctx, 10))

	// The grant should survive: it does not expire for another 9 days.
	after, err := suite.feegrantKeeper.GetAllowance(ctx, granter, grantee)
	suite.Require().NoError(err, "grant was deleted before its expiration")
	suite.Require().NotNil(after, "grant valid until %s was pruned at %s", newExp, ctx.BlockTime())
}

// Case B: add an expiry to a grant that had none. No queue entry is ever written, so the grant
// becomes unusable at its expiration but can never be pruned - permanent state growth.
func (suite *KeeperTestSuite) TestUpdateAllowanceAddedExpiryIsNeverPruned() {
	granter, grantee := suite.addrs[2], suite.addrs[3]

	err := suite.feegrantKeeper.GrantAllowance(suite.ctx, granter, grantee, &feegrant.BasicAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewCoin("atom", sdkmath.NewInt(100))),
		Expiration: nil, // no expiration
	})
	suite.Require().NoError(err)

	exp := suite.ctx.BlockTime().AddDate(0, 0, 1)
	err = suite.feegrantKeeper.UpdateAllowance(suite.ctx, granter, grantee, &feegrant.BasicAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewCoin("atom", sdkmath.NewInt(100))),
		Expiration: &exp,
	})
	suite.Require().NoError(err)

	// Long after the expiration, pruning should reclaim it.
	ctx := suite.ctx.WithBlockTime(exp.AddDate(0, 0, 30))
	suite.Require().NoError(suite.feegrantKeeper.RemoveExpiredAllowances(ctx, 10))

	after, _ := suite.feegrantKeeper.GetAllowance(ctx, granter, grantee)
	suite.Require().Nil(after, "expired grant is unreachable by the pruner and stays in state forever")
}

// Case C: remove an expiry entirely. The stale queue entry still deletes the grant.
func (suite *KeeperTestSuite) TestUpdateAllowanceRemovedExpiryStillPruned() {
	granter, grantee := suite.addrs[4], suite.addrs[5]

	oldExp := suite.ctx.BlockTime().AddDate(0, 0, 1)
	err := suite.feegrantKeeper.GrantAllowance(suite.ctx, granter, grantee, &feegrant.BasicAllowance{
		SpendLimit: suite.atom,
		Expiration: &oldExp,
	})
	suite.Require().NoError(err)

	// Granter makes the allowance perpetual.
	err = suite.feegrantKeeper.UpdateAllowance(suite.ctx, granter, grantee, &feegrant.BasicAllowance{
		SpendLimit: suite.atom,
		Expiration: nil,
	})
	suite.Require().NoError(err)

	ctx := suite.ctx.WithBlockTime(oldExp.Add(time.Second))
	suite.Require().NoError(suite.feegrantKeeper.RemoveExpiredAllowances(ctx, 10))

	after, err := suite.feegrantKeeper.GetAllowance(ctx, granter, grantee)
	suite.Require().NoError(err, "perpetual grant was deleted at its former expiration")
	suite.Require().NotNil(after, "grant with no expiration was pruned")
}
