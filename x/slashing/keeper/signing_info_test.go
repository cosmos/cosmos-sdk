package keeper_test

import (
	"time"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
	info, found := keeper.GetValidatorSigningInfo(ctx, consAddr)
	require.True(found)
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
	keeper.Tombstone(ctx, consAddr)
	require.True(keeper.IsTombstoned(ctx, consAddr))

	// test JailUntil
	jailTime := time.Now().Add(time.Hour).UTC()
	keeper.JailUntil(ctx, consAddr, jailTime)
	sInfo, _ := keeper.GetValidatorSigningInfo(ctx, consAddr)
	require.Equal(sInfo.JailedUntil, jailTime)
}

func (s *KeeperTestSuite) TestValidatorMissedBlockBitArray() {
	ctx, keeper := s.ctx, s.slashingKeeper
	require := s.Require()

	params := testutil.TestParams()
	params.SignedBlocksWindow = 100
	require.NoError(keeper.SetParams(ctx, params))

	testCases := []struct {
		name   string
		index  int64
		missed bool
	}{
		{
			name:   "missed block with false",
			index:  50,
			missed: false,
		},
		{
			name:   "missed block with true",
			index:  51,
			missed: true,
		},
	}
	for ind, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			keeper.SetValidatorMissedBlockBitArray(ctx, consAddr, tc.index, tc.missed)
			missed := keeper.GetValidatorMissedBlockBitArray(ctx, consAddr, tc.index)

			require.Equal(missed, tc.missed)
			missedBlocks := keeper.GetValidatorMissedBlocks(ctx, consAddr)
			require.Equal(len(missedBlocks), ind+1)
			require.Equal(missedBlocks[ind].Index, tc.index)
			require.Equal(missedBlocks[ind].Missed, tc.missed)
		})
	}
}

func (s *KeeperTestSuite) TestPerformConsensusPubKeyUpdate() {
	ctx, slashingKeeper := s.ctx, s.slashingKeeper

	require := s.Require()

	pks := simtestutil.CreateTestPubKeys(500)

	oldConsAddr := sdk.ConsAddress(pks[0].Address())
	newConsAddr := sdk.ConsAddress(pks[1].Address())
	newInfo := slashingtypes.NewValidatorSigningInfo(
		oldConsAddr,
		int64(4),
		int64(3),
		time.Unix(2, 0).UTC(),
		false,
		int64(10),
	)
	slashingKeeper.SetValidatorSigningInfo(ctx, oldConsAddr, newInfo)
	slashingKeeper.SetValidatorMissedBlockBitArray(ctx, oldConsAddr, 10, true)
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
