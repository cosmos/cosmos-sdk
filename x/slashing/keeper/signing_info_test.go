package keeper_test

import (
	"time"

	"github.com/golang/mock/gomock"

	"cosmossdk.io/x/slashing/testutil"
	slashingtypes "cosmossdk.io/x/slashing/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestValidatorSigningInfo() {
	ctx, keeper := s.ctx, s.slashingKeeper
	require := s.Require()

	consStr, err := s.stakingKeeper.ConsensusAddressCodec().BytesToString(consAddr)
	require.NoError(err)

	signingInfo := slashingtypes.NewValidatorSigningInfo(
		consStr,
		ctx.BlockHeight(),
		int64(3),
		time.Unix(2, 0),
		false,
		int64(10),
	)

	// set the validator signing information
	require.NoError(keeper.ValidatorSigningInfo.Set(ctx, consAddr, signingInfo))
	require.True(keeper.HasValidatorSigningInfo(ctx, consAddr))
	info, err := keeper.ValidatorSigningInfo.Get(ctx, consAddr)
	require.NoError(err)
	require.Equal(info.StartHeight, ctx.BlockHeight())
	require.Equal(info.IndexOffset, int64(3))
	require.Equal(info.JailedUntil, time.Unix(2, 0).UTC())
	require.Equal(info.MissedBlocksCounter, int64(10))

	var signingInfos []slashingtypes.ValidatorSigningInfo

	err = keeper.ValidatorSigningInfo.Walk(ctx, nil, func(consAddr sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool, err error) {
		signingInfos = append(signingInfos, info)
		return false, nil
	})
	require.NoError(err)
	require.Equal(signingInfos[0].Address, signingInfo.Address)

	// test Tombstone
	err = keeper.Tombstone(ctx, consAddr)
	require.NoError(err)
	require.True(keeper.IsTombstoned(ctx, consAddr))

	// test JailUntil
	jailTime := time.Now().Add(time.Hour).UTC()
	require.NoError(keeper.JailUntil(ctx, consAddr, jailTime))
	sInfo, _ := keeper.ValidatorSigningInfo.Get(ctx, consAddr)
	require.Equal(sInfo.JailedUntil, jailTime)
}

func (s *KeeperTestSuite) TestValidatorMissedBlockBitmap_SmallWindow() {
	ctx, keeper := s.ctx, s.slashingKeeper
	require := s.Require()

	for _, window := range []int64{100, 32_000} {
		params := testutil.TestParams()
		params.SignedBlocksWindow = window
		require.NoError(keeper.Params.Set(ctx, params))

		s.stakingKeeper.EXPECT().ValidatorIdentifier(gomock.Any(), consAddr).Return(consAddr, nil).AnyTimes()

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

		// if the validator rotated it's key there will be different consKeys and a mapping will be added in the state.
		consAddr1 := sdk.ConsAddress(sdk.AccAddress([]byte("addr1_______________")))
		s.stakingKeeper.EXPECT().ValidatorIdentifier(gomock.Any(), consAddr1).Return(consAddr, nil).AnyTimes()

		missedBlocks, err = keeper.GetValidatorMissedBlocks(ctx, consAddr1)
		require.NoError(err)
		require.Len(missedBlocks, int(params.SignedBlocksWindow)-1)
	}
}
