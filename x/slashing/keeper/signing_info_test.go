package keeper_test

import (
	"time"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"

	"github.com/cosmos/cosmos-sdk/x/slashing/testutil"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

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
	keeper.SetValidatorSigningInfo(ctx, consAddr, signingInfo)

	require.True(keeper.HasValidatorSigningInfo(ctx, consAddr))
	info, err := keeper.GetValidatorSigningInfo(ctx, consAddr)
	require.NoError(err)
	require.Equal(info.StartHeight, ctx.BlockHeight())
	require.Equal(info.IndexOffset, int64(3))
	require.Equal(info.JailedUntil, time.Unix(2, 0).UTC())
	require.Equal(info.MissedBlocksCounter, int64(10))

	var signingInfos []slashingtypes.ValidatorSigningInfo

	keeper.IterateValidatorSigningInfos(ctx, func(consAddr sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
		signingInfos = append(signingInfos, info)
		return false
	})

	require.Equal(signingInfos[0].Address, signingInfo.Address)

	// test Tombstone
	err = keeper.Tombstone(ctx, consAddr)
	require.NoError(err)
	require.True(keeper.IsTombstoned(ctx, consAddr))

	// test JailUntil
	jailTime := time.Now().Add(time.Hour).UTC()
	keeper.JailUntil(ctx, consAddr, jailTime)
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

func (s *KeeperTestSuite) TestPerformConsensusPubKeyUpdate() {
	ctx, slashingKeeper := s.ctx, s.slashingKeeper

	require := s.Require()

	pks := simtestutil.CreateTestPubKeys(500)

	oldConsAddr := sdk.ConsAddress(pks[0].Address())
	newConsAddr := sdk.ConsAddress(pks[1].Address())
	s.stakingKeeper.EXPECT().SetMappedConskey(gomock.Any(), oldConsAddr, newConsAddr).Return().AnyTimes()

	newInfo := slashingtypes.NewValidatorSigningInfo(
		oldConsAddr,
		int64(4),
		int64(3),
		time.Unix(2, 0).UTC(),
		false,
		int64(10),
	)
	slashingKeeper.SetValidatorSigningInfo(ctx, oldConsAddr, newInfo)
	slashingKeeper.SetMissedBlockBitmapValue(ctx, oldConsAddr, 10, true)
	err := slashingKeeper.PerformConsensusPubKeyUpdate(ctx, pks[0], pks[1])
	require.NoError(err)

	// check pubkey relation is set properly
	savedPubKey, err := slashingKeeper.GetPubkey(ctx, newConsAddr.Bytes())
	require.NoError(err)
	require.Equal(savedPubKey, pks[1])

	// check validator SigningInfo is set properly to new consensus pubkey
	signingInfo, found := slashingKeeper.GetValidatorSigningInfo(ctx, newConsAddr)
	require.True(found)
	require.Equal(signingInfo, newInfo)

	// check missed blocks array is removed on old consensus pubkey
	missedBlocks := slashingKeeper.GetValidatorMissedBlocks(ctx, oldConsAddr)
	require.Len(missedBlocks, 0)

	// check missed blocks are set correctly for new pubkey
	missedBlocks = slashingKeeper.GetValidatorMissedBlocks(ctx, newConsAddr)
	require.Len(missedBlocks, 1)
}
