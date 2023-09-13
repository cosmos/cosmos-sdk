package keeper_test

import (
	"time"

	"github.com/golang/mock/gomock"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/testutil"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func (s *KeeperTestSuite) TestExportAndInitGenesis() {
	ctx, keeper := s.ctx, s.slashingKeeper
	require := s.Require()
	err := keeper.Params.Set(ctx, testutil.TestParams())
	s.Require().NoError(err)
	consAddr1 := sdk.ConsAddress(([]byte("addr1_______________")))
	consAddr2 := sdk.ConsAddress(([]byte("addr2_______________")))

	info1 := types.NewValidatorSigningInfo(consAddr1, int64(4), int64(3),
		time.Now().UTC().Add(100000000000), false, int64(10))
	info2 := types.NewValidatorSigningInfo(consAddr2, int64(5), int64(4),
		time.Now().UTC().Add(10000000000), false, int64(10))

	s.Require().NoError(keeper.ValidatorSigningInfo.Set(ctx, consAddr1, info1))
	s.Require().NoError(keeper.ValidatorSigningInfo.Set(ctx, consAddr2, info2))
	genesisState := keeper.ExportGenesis(ctx)

	require.Equal(genesisState.Params, testutil.TestParams())
	require.Len(genesisState.SigningInfos, 2)
	require.Equal(genesisState.SigningInfos[0].ValidatorSigningInfo, info1)

	// Tombstone validators after genesis shouldn't effect genesis state
	err = keeper.Tombstone(ctx, consAddr1)
	require.NoError(err)
	err = keeper.Tombstone(ctx, consAddr2)
	require.NoError(err)

	ok := keeper.IsTombstoned(ctx, consAddr1)
	require.True(ok)

	newInfo1, _ := keeper.ValidatorSigningInfo.Get(ctx, consAddr1)
	require.NotEqual(info1, newInfo1)

	// Initialize genesis with genesis state before tombstone
	s.stakingKeeper.EXPECT().IterateValidators(ctx, gomock.Any()).Return(nil)
	keeper.InitGenesis(ctx, s.stakingKeeper, genesisState)

	// Validator isTombstoned should return false as GenesisState is initialized
	ok = keeper.IsTombstoned(ctx, consAddr1)
	require.False(ok)

	newInfo1, _ = keeper.ValidatorSigningInfo.Get(ctx, consAddr1)
	newInfo2, _ := keeper.ValidatorSigningInfo.Get(ctx, consAddr2)
	require.Equal(info1, newInfo1)
	require.Equal(info2, newInfo2)
}
