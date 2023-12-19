package keeper_test

import (
	"github.com/golang/mock/gomock"

	"cosmossdk.io/collections"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	"cosmossdk.io/x/staking/testutil"
	"cosmossdk.io/x/staking/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestConsPubKeyRotationHistory() {
	stakingKeeper, ctx := s.stakingKeeper, s.ctx

	_, addrVals := createValAddrs(2)

	// create a validator with a self-delegation
	val := testutil.NewValidator(s.T(), addrVals[0], PKs[0])
	valTokens := stakingKeeper.TokensFromConsensusPower(ctx, 10)
	val, issuedShares := val.AddTokensFromDel(valTokens)
	s.Require().Equal(valTokens, issuedShares.RoundInt())

	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), types.NotBondedPoolName, types.BondedPoolName, gomock.Any())
	_ = stakingkeeper.TestingUpdateValidator(stakingKeeper, ctx, val, true)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := types.NewDelegation(val0AccAddr.String(), addrVals[0].String(), issuedShares)

	stakingKeeper.SetDelegation(ctx, selfDelegation)

	validators, err := stakingKeeper.GetAllValidators(ctx)
	s.Require().NoError(err)
	s.Require().Len(validators, 1)

	validator := validators[0]
	valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	s.Require().NoError(err)

	historyObjects, err := stakingKeeper.GetValidatorConsPubKeyRotationHistory(ctx, valAddr)
	s.Require().NoError(err)
	s.Require().Len(historyObjects, 0)

	newConsPub, err := codectypes.NewAnyWithValue(PKs[1])
	s.Require().NoError(err)

	newConsPub2, err := codectypes.NewAnyWithValue(PKs[2])
	s.Require().NoError(err)

	params, err := stakingKeeper.Params.Get(ctx)
	s.Require().NoError(err)

	height := uint64(ctx.BlockHeight())
	stakingKeeper.RotationHistory.Set(ctx, collections.Join(valAddr.Bytes(), height), types.ConsPubKeyRotationHistory{
		OperatorAddress: valAddr,
		OldConsPubkey:   validator.ConsensusPubkey,
		NewConsPubkey:   newConsPub,
		Height:          height,
		Fee:             params.KeyRotationFee,
	})

	historyObjects, err = stakingKeeper.GetValidatorConsPubKeyRotationHistory(ctx, valAddr)
	s.Require().NoError(err)
	s.Require().Len(historyObjects, 1)

	historyObjects, err = stakingKeeper.GetBlockConsPubKeyRotationHistory(ctx)
	s.Require().NoError(err)
	s.Require().Len(historyObjects, 1)

	err = stakingKeeper.RotationHistory.Set(ctx, collections.Join(valAddr.Bytes(), height+1), types.ConsPubKeyRotationHistory{
		OperatorAddress: valAddr,
		OldConsPubkey:   newConsPub,
		NewConsPubkey:   newConsPub2,
		Height:          height + 1,
		Fee:             params.KeyRotationFee,
	})
	s.Require().NoError(err)

	historyObjects1, err := stakingKeeper.GetValidatorConsPubKeyRotationHistory(ctx, valAddr)
	s.Require().NoError(err)
	s.Require().Len(historyObjects1, 2)

	historyObjects, err = stakingKeeper.GetBlockConsPubKeyRotationHistory(ctx)
	s.Require().NoError(err)

	s.Require().Len(historyObjects, 1)
}

func (s *KeeperTestSuite) setValidators(n int) {
	stakingKeeper, ctx := s.stakingKeeper, s.ctx

	_, addrVals := createValAddrs(n)

	for i := 0; i < n; i++ {
		val := testutil.NewValidator(s.T(), addrVals[i], PKs[i])
		valTokens := stakingKeeper.TokensFromConsensusPower(ctx, 10)
		val, issuedShares := val.AddTokensFromDel(valTokens)
		s.Require().Equal(valTokens, issuedShares.RoundInt())

		s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), types.NotBondedPoolName, types.BondedPoolName, gomock.Any())
		_ = stakingkeeper.TestingUpdateValidator(stakingKeeper, ctx, val, true)
		val0AccAddr := sdk.AccAddress(addrVals[i].Bytes())
		selfDelegation := types.NewDelegation(val0AccAddr.String(), addrVals[i].String(), issuedShares)
		stakingKeeper.SetDelegation(ctx, selfDelegation)
		stakingKeeper.SetValidatorByConsAddr(ctx, val)
	}

	validators, err := stakingKeeper.GetAllValidators(ctx)
	s.Require().NoError(err)
	s.Require().Len(validators, n)
}
