package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/header"
	authtypes "cosmossdk.io/x/auth/types"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	"cosmossdk.io/x/staking/testutil"
	"cosmossdk.io/x/staking/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func (s *KeeperTestSuite) TestConsKeyRotn() {
	stakingKeeper, ctx, accountKeeper, bankKeeper := s.stakingKeeper, s.ctx, s.accountKeeper, s.bankKeeper

	msgServer := stakingkeeper.NewMsgServerImpl(stakingKeeper)
	s.setValidators(6)
	validators, err := stakingKeeper.GetAllValidators(ctx)
	s.Require().NoError(err)

	s.Require().Len(validators, 6)

	existingPubkey, ok := validators[1].ConsensusPubkey.GetCachedValue().(cryptotypes.PubKey)
	s.Require().True(ok)

	bondedPool := authtypes.NewEmptyModuleAccount(types.BondedPoolName)
	accountKeeper.EXPECT().GetModuleAccount(gomock.Any(), types.BondedPoolName).Return(bondedPool).AnyTimes()
	bankKeeper.EXPECT().GetBalance(gomock.Any(), bondedPool.GetAddress(), sdk.DefaultBondDenom).Return(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000)).AnyTimes()

	testCases := []struct {
		name      string
		malleate  func() sdk.Context
		validator string
		newPubKey cryptotypes.PubKey
		isErr     bool
		errMsg    string
	}{
		{
			name: "1st iteration no error",
			malleate: func() sdk.Context {
				val, err := stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[0].GetOperator())
				s.Require().NoError(err)

				bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), sdk.AccAddress(val), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
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
			errMsg:    "decoding bech32 failed",
			newPubKey: PKs[498],
			validator: "non_existing_val",
		},
		{
			name: "limit exceeding",
			malleate: func() sdk.Context {
				val, err := stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[2].GetOperator())
				s.Require().NoError(err)
				bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), sdk.AccAddress(val), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

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
				params, err := stakingKeeper.Params.Get(ctx)
				s.Require().NoError(err)
				val, err := stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[3].GetOperator())
				s.Require().NoError(err)
				bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), sdk.AccAddress(val), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

				// 1st rotation should pass, since limit is 1
				req, err := types.NewMsgRotateConsPubKey(validators[3].GetOperator(), PKs[494])
				s.Require().NoError(err)
				_, err = msgServer.RotateConsPubKey(ctx, req)
				s.Require().NoError(err)

				// this shouldn't mature the recent rotation since unbonding period isn't reached
				s.Require().NoError(stakingKeeper.PurgeAllMaturedConsKeyRotatedKeys(ctx, ctx.BlockTime()))

				// 2nd rotation should fail since limit exceeding
				req, err = types.NewMsgRotateConsPubKey(validators[3].GetOperator(), PKs[493])
				s.Require().NoError(err)
				_, err = msgServer.RotateConsPubKey(ctx, req)
				s.Require().Error(err, "exceeding maximum consensus pubkey rotations within unbonding period")

				// This should remove the keys from queue
				// after setting the blocktime to reach the unbonding period
				newCtx := ctx.WithHeaderInfo(header.Info{Time: ctx.BlockTime().Add(params.UnbondingTime)})
				s.Require().NoError(stakingKeeper.PurgeAllMaturedConsKeyRotatedKeys(newCtx, newCtx.BlockTime()))
				return newCtx
			},
			isErr:     false,
			newPubKey: PKs[493],
			validator: validators[3].GetOperator(),
		},
		{
			name: "verify other validator rotation blocker",
			malleate: func() sdk.Context {
				params, err := stakingKeeper.Params.Get(ctx)
				s.Require().NoError(err)
				valStr4 := validators[4].GetOperator()
				valStr5 := validators[5].GetOperator()
				valAddr4, err := stakingKeeper.ValidatorAddressCodec().StringToBytes(valStr4)
				s.Require().NoError(err)

				valAddr5, err := stakingKeeper.ValidatorAddressCodec().StringToBytes(valStr5)
				s.Require().NoError(err)

				bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), sdk.AccAddress(valAddr4), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), sdk.AccAddress(valAddr5), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

				// add 2 days to the current time and add rotate key, it should allow to rotate.
				newCtx := ctx.WithHeaderInfo(header.Info{Time: ctx.BlockTime().Add(2 * 24 * time.Hour)})
				req1, err := types.NewMsgRotateConsPubKey(valStr5, PKs[491])
				s.Require().NoError(err)
				_, err = msgServer.RotateConsPubKey(newCtx, req1)
				s.Require().NoError(err)

				// 1st rotation should pass, since limit is 1
				req, err := types.NewMsgRotateConsPubKey(valStr4, PKs[490])
				s.Require().NoError(err)
				_, err = msgServer.RotateConsPubKey(ctx, req)
				s.Require().NoError(err)

				// this shouldn't mature the recent rotation since unbonding period isn't reached
				s.Require().NoError(stakingKeeper.PurgeAllMaturedConsKeyRotatedKeys(ctx, ctx.BlockTime()))

				// 2nd rotation should fail since limit exceeding
				req, err = types.NewMsgRotateConsPubKey(valStr4, PKs[489])
				s.Require().NoError(err)
				_, err = msgServer.RotateConsPubKey(ctx, req)
				s.Require().Error(err, "exceeding maximum consensus pubkey rotations within unbonding period")

				// This should remove the keys from queue
				// after setting the blocktime to reach the unbonding period,
				// but other validator which rotated with addition of 2 days shouldn't be removed, so it should stop the rotation of valStr5.
				newCtx1 := ctx.WithHeaderInfo(header.Info{Time: ctx.BlockTime().Add(params.UnbondingTime).Add(time.Hour)})
				s.Require().NoError(stakingKeeper.PurgeAllMaturedConsKeyRotatedKeys(newCtx1, newCtx1.BlockTime()))
				return newCtx1
			},
			isErr:     true,
			newPubKey: PKs[492],
			errMsg:    "exceeding maximum consensus pubkey rotations within unbonding period",
			validator: validators[5].GetOperator(),
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

				addr, err := stakingKeeper.ValidatorAddressCodec().StringToBytes(tc.validator)
				s.Require().NoError(err)

				valInfo, err := stakingKeeper.GetValidator(newCtx, addr)
				s.Require().NoError(err)
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
		selfDelegation := types.NewDelegation(val0AccAddr.String(), addrVals[i].String(), issuedShares)
		stakingKeeper.SetDelegation(ctx, selfDelegation)
		stakingKeeper.SetValidatorByConsAddr(ctx, val)
	}

	validators, err := stakingKeeper.GetAllValidators(ctx)
	s.Require().NoError(err)
	s.Require().Len(validators, n)
}
