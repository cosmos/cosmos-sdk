package keeper_test

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/golang/mock/gomock"
)

func (s *KeeperTestSuite) TestConsPubKeyRotationHistory() {
	stakingKeeper, ctx := s.stakingKeeper, s.ctx

	_, addrVals := createValAddrs(2)

	// create a validator with a self-delegation
	val := testutil.NewValidator(s.T(), addrVals[0], PKs[0])
	valTokens := stakingKeeper.TokensFromConsensusPower(ctx, 10)
	val, issuedShares := val.AddTokensFromDel(valTokens)
	s.Require().Equal(valTokens, issuedShares.RoundInt())

	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	_ = stakingkeeper.TestingUpdateValidator(stakingKeeper, ctx, val, true)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := stakingtypes.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	stakingKeeper.SetDelegation(ctx, selfDelegation)

	validators := stakingKeeper.GetAllValidators(ctx)
	s.Require().Len(validators, 1)

	validator := validators[0]
	valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	s.Require().NoError(err)

	historyObjects := stakingKeeper.GetValidatorConsPubKeyRotationHistory(ctx, valAddr)
	s.Require().Len(historyObjects, 0)

	newConsPub, err := codectypes.NewAnyWithValue(PKs[1])
	s.Require().NoError(err)

	newConsPub2, err := codectypes.NewAnyWithValue(PKs[2])
	s.Require().NoError(err)

	stakingKeeper.SetConsPubKeyRotationHistory(ctx, stakingtypes.ConsPubKeyRotationHistory{
		OperatorAddress: valAddr.String(),
		OldConsPubkey:   validator.ConsensusPubkey,
		NewConsPubkey:   newConsPub,
		Height:          uint64(ctx.BlockHeight()),
	})

	historyObjects = stakingKeeper.GetValidatorConsPubKeyRotationHistory(ctx, valAddr)
	s.Require().Len(historyObjects, 1)

	historyObjects = stakingKeeper.GetBlockConsPubKeyRotationHistory(ctx, ctx.BlockHeight())
	s.Require().Len(historyObjects, 1)

	stakingKeeper.SetConsPubKeyRotationHistory(ctx, stakingtypes.ConsPubKeyRotationHistory{
		OperatorAddress: valAddr.String(),
		OldConsPubkey:   newConsPub,
		NewConsPubkey:   newConsPub2,
		Height:          uint64(ctx.BlockHeight()) + 1,
	})

	historyObjects = stakingKeeper.GetValidatorConsPubKeyRotationHistory(ctx, valAddr)
	s.Require().Len(historyObjects, 2)

	historyObjects = stakingKeeper.GetBlockConsPubKeyRotationHistory(ctx, ctx.BlockHeight())
	s.Require().Len(historyObjects, 1)
}
