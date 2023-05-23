package keeper_test

import (
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
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

	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), types.NotBondedPoolName, types.BondedPoolName, gomock.Any())
	_ = stakingkeeper.TestingUpdateValidator(stakingKeeper, ctx, val, true)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := types.NewDelegation(val0AccAddr, addrVals[0], issuedShares)

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

	stakingKeeper.SetConsPubKeyRotationHistory(ctx,
		valAddr,
		validator.ConsensusPubkey,
		newConsPub,
		uint64(ctx.BlockHeight()),
		stakingKeeper.KeyRotationFee(ctx),
	)

	historyObjects = stakingKeeper.GetValidatorConsPubKeyRotationHistory(ctx, valAddr)
	s.Require().Len(historyObjects, 1)

	historyObjects = stakingKeeper.GetBlockConsPubKeyRotationHistory(ctx, ctx.BlockHeight())
	s.Require().Len(historyObjects, 1)

	stakingKeeper.SetConsPubKeyRotationHistory(ctx,
		valAddr,
		newConsPub,
		newConsPub2,
		uint64(ctx.BlockHeight())+1,
		stakingKeeper.KeyRotationFee(ctx),
	)

	historyObjects = stakingKeeper.GetValidatorConsPubKeyRotationHistory(ctx, valAddr)
	s.Require().Len(historyObjects, 2)

	historyObjects = stakingKeeper.GetBlockConsPubKeyRotationHistory(ctx, ctx.BlockHeight())
	s.Require().Len(historyObjects, 1)
}

func (s *KeeperTestSuite) TestConsKeyRotn() {
	stakingKeeper, ctx, accountKeeper, bankKeeper := s.stakingKeeper, s.ctx, s.accountKeeper, s.bankKeeper

	msgServer := stakingkeeper.NewMsgServerImpl(stakingKeeper)
	s.setValidators(4)
	validators := stakingKeeper.GetAllValidators(ctx)
	s.Require().Len(validators, 4)

	existingPubkey, ok := validators[1].ConsensusPubkey.GetCachedValue().(cryptotypes.PubKey)
	s.Require().True(ok)

	bondedPool := authtypes.NewEmptyModuleAccount(types.BondedPoolName)
	accountKeeper.EXPECT().GetModuleAccount(gomock.Any(), types.BondedPoolName).Return(bondedPool).AnyTimes()
	bankKeeper.EXPECT().GetBalance(gomock.Any(), bondedPool.GetAddress(), sdk.DefaultBondDenom).Return(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000))).AnyTimes()

	testCases := []struct {
		name      string
		malleate  func() sdk.Context
		validator sdk.ValAddress
		newPubKey cryptotypes.PubKey
		isErr     bool
		errMsg    string
	}{
		{
			name: "1st iteration no error",
			malleate: func() sdk.Context {
				bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), sdk.AccAddress(validators[0].GetOperator()), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return ctx
			},
			isErr:     false,
			errMsg:    "",
			newPubKey: PKs[499],
			validator: validators[0].GetOperator(),
		},
		{
			name:      "pubkey already associated with another validator",
			malleate:  func() sdk.Context { return ctx },
			isErr:     true,
			errMsg:    "consensus pubkey is already used for a validator",
			newPubKey: existingPubkey,
			validator: validators[0].GetOperator(),
		},
		{
			name:      "non existing validator",
			malleate:  func() sdk.Context { return ctx },
			isErr:     true,
			errMsg:    "validator does not exist",
			newPubKey: PKs[498],
			validator: sdk.ValAddress("non_existing_val"),
		},
		{
			name: "limit exceeding",
			malleate: func() sdk.Context {
				bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), sdk.AccAddress(validators[2].GetOperator()), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

				req, err := types.NewMsgRotateConsPubKey(validators[2].GetOperator(), PKs[495])
				s.Require().NoError(err)
				_, err = msgServer.RotateConsPubKey(ctx, req)
				s.Require().NoError(err)

				return ctx
			},
			isErr:     true,
			errMsg:    "exceeding maximum consensus pubkey rotations within unbonding period",
			newPubKey: PKs[494],
			validator: validators[2].GetOperator(),
		},
		{
			name: "limit exceeding, but it should rotate after unbonding period",
			malleate: func() sdk.Context {
				params := stakingKeeper.GetParams(ctx)
				bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), sdk.AccAddress(validators[3].GetOperator()), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

				// 1st rotation should pass, since limit is 1
				req, err := types.NewMsgRotateConsPubKey(validators[3].GetOperator(), PKs[494])
				s.Require().NoError(err)
				_, err = msgServer.RotateConsPubKey(ctx, req)
				s.Require().NoError(err)

				// this shouldn't mature the recent rotation since unbonding period isn't reached
				stakingKeeper.UpdateAllMaturedConsKeyRotatedKeys(ctx, ctx.BlockHeader().Time)

				// 2nd rotation should fail since limit exceeding
				req, err = types.NewMsgRotateConsPubKey(validators[3].GetOperator(), PKs[493])
				s.Require().NoError(err)
				_, err = msgServer.RotateConsPubKey(ctx, req)
				s.Require().Error(err, "exceeding maximum consensus pubkey rotations within unbonding period")

				// This should remove the keys from queue
				// after setting the blocktime to reach the unbonding period
				newCtx := ctx.WithBlockTime(ctx.BlockHeader().Time.Add(params.UnbondingTime))
				stakingKeeper.UpdateAllMaturedConsKeyRotatedKeys(newCtx, newCtx.BlockHeader().Time)
				return newCtx
			},
			isErr:     false,
			newPubKey: PKs[493],
			validator: validators[3].GetOperator(),
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			newCtx := tc.malleate()

			req, err := types.NewMsgRotateConsPubKey(tc.validator, tc.newPubKey)
			s.Require().NoError(err)

			_, err = msgServer.RotateConsPubKey(newCtx, req)
			if tc.isErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errMsg)
			} else {
				s.Require().NoError(err)
				_, err = stakingKeeper.EndBlocker(newCtx)
				s.Require().NoError(err)

				valInfo, found := stakingKeeper.GetValidator(newCtx, tc.validator)
				s.Require().True(found)
				s.Require().Equal(valInfo.ConsensusPubkey, req.NewPubkey)
			}
		})
	}
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
		selfDelegation := types.NewDelegation(val0AccAddr, addrVals[i], issuedShares)
		stakingKeeper.SetDelegation(ctx, selfDelegation)
		stakingKeeper.SetValidatorByConsAddr(ctx, val)
	}

	validators := stakingKeeper.GetAllValidators(ctx)
	s.Require().Len(validators, n)
}
