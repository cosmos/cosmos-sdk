package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *KeeperTestSuite) TestAfterValidatorBonded() {
	ctx, keeper := s.ctx, s.slashingKeeper
	require := s.Require()

	valAddr := sdk.ValAddress(consAddr.Bytes())
	s.Require().NoError(keeper.Hooks().AfterValidatorBonded(ctx, consAddr, valAddr))

	_, err := keeper.GetValidatorSigningInfo(ctx, consAddr)
	require.NoError(err)
}

func (s *KeeperTestSuite) TestAfterValidatorCreatedOrRemoved() {
	ctx, keeper := s.ctx, s.slashingKeeper
	require := s.Require()

	_, pubKey, addr := testdata.KeyTestPubAddr()
	valAddr := sdk.ValAddress(addr)

	validator, err := stakingtypes.NewValidator(sdk.ValAddress(addr).String(), pubKey, stakingtypes.Description{})
	require.NoError(err)

	s.stakingKeeper.EXPECT().Validator(ctx, valAddr).Return(validator, nil)
	err = keeper.Hooks().AfterValidatorCreated(ctx, valAddr)
	require.NoError(err)

	ePubKey, err := keeper.GetPubkey(ctx, addr.Bytes())
	require.NoError(err)
	require.Equal(ePubKey, pubKey)

	err = keeper.Hooks().AfterValidatorRemoved(ctx, sdk.ConsAddress(addr), nil)
	require.NoError(err)

	_, err = keeper.GetPubkey(ctx, addr.Bytes())
	require.Error(err)
}

func (s *KeeperTestSuite) TestAfterValidatorConsKeyUpdated() {
	ctx, keeper := s.ctx, s.slashingKeeper
	require := s.Require()

	oldPk := ed25519.GenPrivKey().PubKey()
	newPk := ed25519.GenPrivKey().PubKey()
	oldConsAddr := sdk.ConsAddress(oldPk.Address())
	newConsAddr := sdk.ConsAddress(newPk.Address())
	valAddr := sdk.ValAddress(oldPk.Address())

	validator, err := stakingtypes.NewValidator(valAddr.String(), newPk, stakingtypes.Description{})
	require.NoError(err)

	expectedInfo := slashingtypes.NewValidatorSigningInfo(
		oldConsAddr,
		10,
		3,
		time.Unix(100, 0),
		true,
		1,
	)
	require.NoError(keeper.SetValidatorSigningInfo(ctx, oldConsAddr, expectedInfo))
	require.NoError(keeper.SetMissedBlockBitmapValue(ctx, oldConsAddr, 2, true))
	require.NoError(keeper.AddPubkey(ctx, oldPk))

	s.stakingKeeper.EXPECT().Validator(ctx, valAddr).Return(validator, nil)
	require.NoError(keeper.Hooks().AfterValidatorConsKeyUpdated(ctx, oldConsAddr, newConsAddr, valAddr))

	_, err = keeper.GetValidatorSigningInfo(ctx, oldConsAddr)
	require.Error(err)

	gotInfo, err := keeper.GetValidatorSigningInfo(ctx, newConsAddr)
	require.NoError(err)
	require.Equal(newConsAddr.String(), gotInfo.Address)
	require.Equal(expectedInfo.StartHeight, gotInfo.StartHeight)
	require.Equal(expectedInfo.IndexOffset, gotInfo.IndexOffset)
	require.True(expectedInfo.JailedUntil.Equal(gotInfo.JailedUntil))
	require.Equal(expectedInfo.Tombstoned, gotInfo.Tombstoned)
	require.Equal(expectedInfo.MissedBlocksCounter, gotInfo.MissedBlocksCounter)

	missed, err := keeper.GetMissedBlockBitmapValue(ctx, newConsAddr, 2)
	require.NoError(err)
	require.True(missed)
	missed, err = keeper.GetMissedBlockBitmapValue(ctx, oldConsAddr, 2)
	require.NoError(err)
	require.False(missed)

	_, err = keeper.GetPubkey(ctx, oldConsAddr.Bytes())
	require.Error(err)
	gotPk, err := keeper.GetPubkey(ctx, newConsAddr.Bytes())
	require.NoError(err)
	require.Equal(newPk, gotPk)
}
