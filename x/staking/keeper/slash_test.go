package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
)

// tests Jail, Unjail
func (s *KeeperTestSuite) TestRevocation() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	valAddr := sdk.ValAddress(PKs[0].Address().Bytes())
	consAddr := sdk.ConsAddress(PKs[0].Address())
	validator := teststaking.NewValidator(s.T(), valAddr, PKs[0])

	// initial state
	keeper.SetValidator(ctx, validator)
	keeper.SetValidatorByConsAddr(ctx, validator)
	val, found := keeper.GetValidator(ctx, valAddr)
	require.True(found)
	require.False(val.IsJailed())

	// test jail
	keeper.Jail(ctx, consAddr)
	val, found = keeper.GetValidator(ctx, valAddr)
	require.True(found)
	require.True(val.IsJailed())

	// test unjail
	keeper.Unjail(ctx, consAddr)
	val, found = keeper.GetValidator(ctx, valAddr)
	require.True(found)
	require.False(val.IsJailed())
}
