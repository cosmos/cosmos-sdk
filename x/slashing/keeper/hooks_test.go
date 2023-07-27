package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *KeeperTestSuite) TestAfterValidatorBonded() {
	ctx, keeper := s.ctx, s.slashingKeeper
	require := s.Require()

	valAddr := sdk.ValAddress(consAddr.Bytes())
	err := keeper.Hooks().AfterValidatorBonded(ctx, consAddr, valAddr)
	require.NoError(err)
	_, err = keeper.ValidatorSigningInfo.Get(ctx, consAddr)
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
