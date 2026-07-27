package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/testutil"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func (s *KeeperTestSuite) TestMoveValidatorSigningInfo() {
	ctx, keeper := s.ctx, s.slashingKeeper
	require := s.Require()

	oldAddr := sdk.ConsAddress([]byte("old_addr____________"))
	newAddr := sdk.ConsAddress([]byte("new_addr____________"))

	info := slashingtypes.NewValidatorSigningInfo(oldAddr, 10, 3, time.Unix(100, 0), true, 7)
	require.NoError(keeper.SetValidatorSigningInfo(ctx, oldAddr, info))

	require.NoError(keeper.MoveValidatorSigningInfo(ctx, oldAddr, newAddr))

	// old address no longer has signing info
	require.False(keeper.HasValidatorSigningInfo(ctx, oldAddr))

	// new address has the migrated info, with its Address field rewritten
	got, err := keeper.GetValidatorSigningInfo(ctx, newAddr)
	require.NoError(err)
	require.Equal(newAddr.String(), got.Address)
	require.Equal(info.StartHeight, got.StartHeight)
	require.Equal(info.IndexOffset, got.IndexOffset)
	require.True(info.JailedUntil.Equal(got.JailedUntil))
	require.Equal(info.Tombstoned, got.Tombstoned)
	require.Equal(info.MissedBlocksCounter, got.MissedBlocksCounter)
}

func (s *KeeperTestSuite) TestMoveValidatorSigningInfo_NoInfo() {
	ctx, keeper := s.ctx, s.slashingKeeper
	require := s.Require()

	oldAddr := sdk.ConsAddress([]byte("old_addr____________"))
	newAddr := sdk.ConsAddress([]byte("new_addr____________"))

	// validators can rotate before they have ever bonded: there is no signing
	// info to move and this must not error or create an entry.
	require.NoError(keeper.MoveValidatorSigningInfo(ctx, oldAddr, newAddr))
	require.False(keeper.HasValidatorSigningInfo(ctx, newAddr))
}

func (s *KeeperTestSuite) TestMoveMissedBlockBitmap() {
	ctx, keeper := s.ctx, s.slashingKeeper
	require := s.Require()

	require.NoError(keeper.SetParams(ctx, testutil.TestParams()))

	oldAddr := sdk.ConsAddress([]byte("old_addr____________"))
	newAddr := sdk.ConsAddress([]byte("new_addr____________"))

	// set bits in two different chunks (chunk size is 1024 bits) so the move
	// is exercised across multiple chunk keys.
	indices := []int64{5, 1500}
	for _, idx := range indices {
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, oldAddr, idx, true))
	}

	require.NoError(keeper.MoveMissedBlockBitmap(ctx, oldAddr, newAddr))

	for _, idx := range indices {
		// new address carries the missed bits
		missed, err := keeper.GetMissedBlockBitmapValue(ctx, newAddr, idx)
		require.NoError(err)
		require.True(missed)

		// old address bits are cleared
		missed, err = keeper.GetMissedBlockBitmapValue(ctx, oldAddr, idx)
		require.NoError(err)
		require.False(missed)
	}
}

func (s *KeeperTestSuite) TestMoveMissedBlockBitmap_Empty() {
	ctx, keeper := s.ctx, s.slashingKeeper
	require := s.Require()

	oldAddr := sdk.ConsAddress([]byte("old_addr____________"))
	newAddr := sdk.ConsAddress([]byte("new_addr____________"))

	// no bitmap chunks exist for oldAddr; the move is a no-op.
	require.NoError(keeper.MoveMissedBlockBitmap(ctx, oldAddr, newAddr))
}

func (s *KeeperTestSuite) TestValidatorSigningInfo() {
	ctx, keeper := s.ctx, s.slashingKeeper
	require := s.Require()

	signingInfo := slashingtypes.NewValidatorSigningInfo(
		consAddr,
		ctx.BlockHeight(),
		int64(3),
		time.Unix(2, 0),
		false,
		int64(10),
	)

	// set the validator signing information
	require.NoError(keeper.SetValidatorSigningInfo(ctx, consAddr, signingInfo))

	require.True(keeper.HasValidatorSigningInfo(ctx, consAddr))
	info, err := keeper.GetValidatorSigningInfo(ctx, consAddr)
	require.NoError(err)
	require.Equal(info.StartHeight, ctx.BlockHeight())
	require.Equal(info.IndexOffset, int64(3))
	require.Equal(info.JailedUntil, time.Unix(2, 0).UTC())
	require.Equal(info.MissedBlocksCounter, int64(10))

	var signingInfos []slashingtypes.ValidatorSigningInfo

	require.NoError(keeper.IterateValidatorSigningInfos(ctx, func(consAddr sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
		signingInfos = append(signingInfos, info)
		return false
	}))

	require.Equal(signingInfos[0].Address, signingInfo.Address)

	// test Tombstone
	err = keeper.Tombstone(ctx, consAddr)
	require.NoError(err)
	require.True(keeper.IsTombstoned(ctx, consAddr))

	// test JailUntil
	jailTime := time.Now().Add(time.Hour).UTC()
	require.NoError(keeper.JailUntil(ctx, consAddr, jailTime))
	sInfo, _ := keeper.GetValidatorSigningInfo(ctx, consAddr)
	require.Equal(sInfo.JailedUntil, jailTime)
}

func (s *KeeperTestSuite) TestValidatorMissedBlockBitmap_SmallWindow() {
	ctx, keeper := s.ctx, s.slashingKeeper
	require := s.Require()

	for _, window := range []int64{100, 32_000} {
		params := testutil.TestParams()
		params.SignedBlocksWindow = window
		require.NoError(keeper.SetParams(ctx, params))

		// validator misses all blocks in the window
		var valIdxOffset int64
		for valIdxOffset < params.SignedBlocksWindow {
			idx := valIdxOffset % params.SignedBlocksWindow
			err := keeper.SetMissedBlockBitmapValue(ctx, consAddr, idx, true)
			require.NoError(err)

			missed, err := keeper.GetMissedBlockBitmapValue(ctx, consAddr, idx)
			require.NoError(err)
			require.True(missed)

			valIdxOffset++
		}

		// validator should have missed all blocks
		missedBlocks, err := keeper.GetValidatorMissedBlocks(ctx, consAddr)
		require.NoError(err)
		require.Len(missedBlocks, int(params.SignedBlocksWindow))

		// sign next block, which rolls the missed block bitmap
		idx := valIdxOffset % params.SignedBlocksWindow
		err = keeper.SetMissedBlockBitmapValue(ctx, consAddr, idx, false)
		require.NoError(err)

		missed, err := keeper.GetMissedBlockBitmapValue(ctx, consAddr, idx)
		require.NoError(err)
		require.False(missed)

		// validator should have missed all blocks except the last one
		missedBlocks, err = keeper.GetValidatorMissedBlocks(ctx, consAddr)
		require.NoError(err)
		require.Len(missedBlocks, int(params.SignedBlocksWindow)-1)
	}
}
