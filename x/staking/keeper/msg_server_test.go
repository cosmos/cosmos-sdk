package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/golang/mock/gomock"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	authtypes "cosmossdk.io/x/auth/types"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	"cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	PKS     = simtestutil.CreateTestPubKeys(3)
	Addr    = sdk.AccAddress(PKS[0].Address())
	ValAddr = sdk.ValAddress(Addr)
)

func (s *KeeperTestSuite) execExpectCalls() {
	s.accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()
	s.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), Addr, types.NotBondedPoolName, gomock.Any()).AnyTimes()
}

func (s *KeeperTestSuite) TestMsgCreateValidator() {
	ctx, msgServer := s.ctx, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	pk1 := ed25519.GenPrivKey().PubKey()
	require.NotNil(pk1)

	pubkey, err := codectypes.NewAnyWithValue(pk1)
	require.NoError(err)

	var ed25519pk cryptotypes.PubKey = &ed25519.PubKey{Key: []byte{1, 2, 3, 4, 5, 6}}
	pubkeyInvalidLen, err := codectypes.NewAnyWithValue(ed25519pk)
	require.NoError(err)

	ctx = ctx.WithConsensusParams(cmtproto.ConsensusParams{
		Validator: &cmtproto.ValidatorParams{
			PubKeyTypes: []string{sdk.PubKeyEd25519Type},
		},
	})

	testCases := []struct {
		name      string
		input     *types.MsgCreateValidator
		expErr    bool
		expErrMsg string
	}{
		{
			name: "empty description",
			input: &types.MsgCreateValidator{
				Description: types.Description{},
				Commission: types.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000),
			},
			expErr:    true,
			expErrMsg: "empty description",
		},
		{
			name: "invalid validator address",
			input: &types.MsgCreateValidator{
				Description: types.Description{
					Moniker: "NewValidator",
				},
				Commission: types.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  sdk.AccAddress([]byte("invalid")).String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000),
			},
			expErr:    true,
			expErrMsg: "invalid validator address",
		},
		{
			name: "empty validator pubkey",
			input: &types.MsgCreateValidator{
				Description: types.Description{
					Moniker: "NewValidator",
				},
				Commission: types.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            nil,
				Value:             sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000),
			},
			expErr:    true,
			expErrMsg: "empty validator public key",
		},
		{
			name: "validator pubkey len is invalid",
			input: &types.MsgCreateValidator{
				Description: types.Description{
					Moniker: "NewValidator",
				},
				Commission: types.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkeyInvalidLen,
				Value:             sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000),
			},
			expErr:    true,
			expErrMsg: "consensus pubkey len is invalid",
		},
		{
			name: "empty delegation amount",
			input: &types.MsgCreateValidator{
				Description: types.Description{
					Moniker: "NewValidator",
				},
				Commission: types.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin(sdk.DefaultBondDenom, 0),
			},
			expErr:    true,
			expErrMsg: "invalid delegation amount",
		},
		{
			name: "nil delegation amount",
			input: &types.MsgCreateValidator{
				Description: types.Description{
					Moniker: "NewValidator",
				},
				Commission: types.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.Coin{},
			},
			expErr:    true,
			expErrMsg: "invalid delegation amount",
		},
		{
			name: "zero minimum self delegation",
			input: &types.MsgCreateValidator{
				Description: types.Description{
					Moniker: "NewValidator",
				},
				Commission: types.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(0),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000),
			},
			expErr:    true,
			expErrMsg: "minimum self delegation must be a positive integer",
		},
		{
			name: "negative minimum self delegation",
			input: &types.MsgCreateValidator{
				Description: types.Description{
					Moniker: "NewValidator",
				},
				Commission: types.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(-1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000),
			},
			expErr:    true,
			expErrMsg: "minimum self delegation must be a positive integer",
		},
		{
			name: "delegation less than minimum self delegation",
			input: &types.MsgCreateValidator{
				Description: types.Description{
					Moniker: "NewValidator",
				},
				Commission: types.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(100),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin(sdk.DefaultBondDenom, 10),
			},
			expErr:    true,
			expErrMsg: "validator's self delegation must be greater than their minimum self delegation",
		},
		{
			name: "valid msg",
			input: &types.MsgCreateValidator{
				Description: types.Description{
					Moniker:         "NewValidator",
					Identity:        "xyz",
					Website:         "xyz.com",
					SecurityContact: "xyz@gmail.com",
					Details:         "details",
				},
				Commission: types.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000),
			},
			expErr: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := msgServer.CreateValidator(ctx, tc.input)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgEditValidator() {
	ctx, msgServer := s.ctx, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	// create new context with updated block time
	newCtx := ctx.WithHeaderInfo(header.Info{Time: ctx.HeaderInfo().Time.AddDate(0, 0, 1)})
	pk := ed25519.GenPrivKey().PubKey()
	require.NotNil(pk)

	comm := types.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	msg, err := types.NewMsgCreateValidator(ValAddr.String(), pk, sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10)), types.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(err)

	res, err := msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	require.NotNil(res)

	newRate := math.LegacyZeroDec()
	invalidRate := math.LegacyNewDec(2)

	lowSelfDel := math.OneInt()
	highSelfDel := math.NewInt(100)
	negSelfDel := math.NewInt(-1)
	newSelfDel := math.NewInt(3)

	testCases := []struct {
		name      string
		ctx       sdk.Context
		input     *types.MsgEditValidator
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid validator",
			ctx:  newCtx,
			input: &types.MsgEditValidator{
				Description: types.Description{
					Moniker: "TestValidator",
				},
				ValidatorAddress:  sdk.AccAddress([]byte("invalid")).String(),
				CommissionRate:    &newRate,
				MinSelfDelegation: &newSelfDel,
			},
			expErr:    true,
			expErrMsg: "invalid validator address",
		},
		{
			name: "empty description",
			ctx:  newCtx,
			input: &types.MsgEditValidator{
				Description:       types.Description{},
				ValidatorAddress:  ValAddr.String(),
				CommissionRate:    &newRate,
				MinSelfDelegation: &newSelfDel,
			},
			expErr:    true,
			expErrMsg: "empty description",
		},
		{
			name: "negative self delegation",
			ctx:  newCtx,
			input: &types.MsgEditValidator{
				Description: types.Description{
					Moniker: "TestValidator",
				},
				ValidatorAddress:  ValAddr.String(),
				CommissionRate:    &newRate,
				MinSelfDelegation: &negSelfDel,
			},
			expErr:    true,
			expErrMsg: "minimum self delegation must be a positive integer",
		},
		{
			name: "invalid commission rate",
			ctx:  newCtx,
			input: &types.MsgEditValidator{
				Description: types.Description{
					Moniker: "TestValidator",
				},
				ValidatorAddress:  ValAddr.String(),
				CommissionRate:    &invalidRate,
				MinSelfDelegation: &newSelfDel,
			},
			expErr:    true,
			expErrMsg: "commission rate must be between 0 and 1 (inclusive)",
		},
		{
			name: "validator does not exist",
			ctx:  newCtx,
			input: &types.MsgEditValidator{
				Description: types.Description{
					Moniker: "TestValidator",
				},
				ValidatorAddress:  sdk.ValAddress([]byte("val")).String(),
				CommissionRate:    &newRate,
				MinSelfDelegation: &newSelfDel,
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
		{
			name: "change commmission rate in <24hrs",
			ctx:  ctx,
			input: &types.MsgEditValidator{
				Description: types.Description{
					Moniker: "TestValidator",
				},
				ValidatorAddress:  ValAddr.String(),
				CommissionRate:    &newRate,
				MinSelfDelegation: &newSelfDel,
			},
			expErr:    true,
			expErrMsg: "commission cannot be changed more than once in 24h",
		},
		{
			name: "minimum self delegation cannot decrease",
			ctx:  newCtx,
			input: &types.MsgEditValidator{
				Description: types.Description{
					Moniker: "TestValidator",
				},
				ValidatorAddress:  ValAddr.String(),
				CommissionRate:    &newRate,
				MinSelfDelegation: &lowSelfDel,
			},
			expErr:    true,
			expErrMsg: "minimum self delegation cannot be decrease",
		},
		{
			name: "validator self-delegation must be greater than min self delegation",
			ctx:  newCtx,
			input: &types.MsgEditValidator{
				Description: types.Description{
					Moniker: "TestValidator",
				},
				ValidatorAddress:  ValAddr.String(),
				CommissionRate:    &newRate,
				MinSelfDelegation: &highSelfDel,
			},
			expErr:    true,
			expErrMsg: "validator's self delegation must be greater than their minimum self delegation",
		},
		{
			name: "valid msg",
			ctx:  newCtx,
			input: &types.MsgEditValidator{
				Description: types.Description{
					Moniker:         "TestValidator",
					Identity:        "abc",
					Website:         "abc.com",
					SecurityContact: "abc@gmail.com",
					Details:         "newDetails",
				},
				ValidatorAddress:  ValAddr.String(),
				CommissionRate:    &newRate,
				MinSelfDelegation: &newSelfDel,
			},
			expErr: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := msgServer.EditValidator(tc.ctx, tc.input)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgDelegate() {
	ctx, keeper, msgServer := s.ctx, s.stakingKeeper, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	pk := ed25519.GenPrivKey().PubKey()
	require.NotNil(pk)

	comm := types.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))

	msg, err := types.NewMsgCreateValidator(ValAddr.String(), pk, sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10)), types.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(err)

	res, err := msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	require.NotNil(res)

	testCases := []struct {
		name      string
		input     *types.MsgDelegate
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid validator",
			input: &types.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: sdk.AccAddress([]byte("invalid")).String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid validator address",
		},
		{
			name: "empty delegator",
			input: &types.MsgDelegate{
				DelegatorAddress: "",
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: empty address string is not allowed",
		},
		{
			name: "invalid delegator",
			input: &types.MsgDelegate{
				DelegatorAddress: "invalid",
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: decoding bech32 failed",
		},
		{
			name: "validator does not exist",
			input: &types.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: sdk.ValAddress([]byte("val")).String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
		{
			name: "zero amount",
			input: &types.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(0))},
			},
			expErr:    true,
			expErrMsg: "invalid delegation amount",
		},
		{
			name: "negative amount",
			input: &types.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(-1))},
			},
			expErr:    true,
			expErrMsg: "invalid delegation amount",
		},
		{
			name: "invalid BondDenom",
			input: &types.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: "test", Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid coin denomination",
		},
		{
			name: "valid msg",
			input: &types.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := msgServer.Delegate(ctx, tc.input)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgBeginRedelegate() {
	ctx, keeper, msgServer := s.ctx, s.stakingKeeper, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	srcValAddr := ValAddr
	addr2 := sdk.AccAddress(PKS[1].Address())
	dstValAddr := sdk.ValAddress(addr2)

	pk := ed25519.GenPrivKey().PubKey()
	require.NotNil(pk)
	dstPk := ed25519.GenPrivKey().PubKey()
	require.NotNil(dstPk)

	comm := types.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	amt := sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))}

	msg, err := types.NewMsgCreateValidator(srcValAddr.String(), pk, amt, types.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(err)
	res, err := msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	require.NotNil(res)
	s.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), addr2, types.NotBondedPoolName, gomock.Any()).AnyTimes()

	msg, err = types.NewMsgCreateValidator(dstValAddr.String(), dstPk, amt, types.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(err)

	res, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	require.NotNil(res)

	shares := math.LegacyNewDec(100)
	del := types.NewDelegation(Addr.String(), srcValAddr.String(), shares)
	require.NoError(keeper.SetDelegation(ctx, del))
	_, err = keeper.Delegations.Get(ctx, collections.Join(Addr, srcValAddr))
	require.NoError(err)

	testCases := []struct {
		name      string
		input     *types.MsgBeginRedelegate
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid source validator",
			input: &types.MsgBeginRedelegate{
				DelegatorAddress:    Addr.String(),
				ValidatorSrcAddress: sdk.AccAddress([]byte("invalid")).String(),
				ValidatorDstAddress: dstValAddr.String(),
				Amount:              sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "invalid source validator address",
		},
		{
			name: "empty delegator",
			input: &types.MsgBeginRedelegate{
				DelegatorAddress:    "",
				ValidatorSrcAddress: srcValAddr.String(),
				ValidatorDstAddress: dstValAddr.String(),
				Amount:              sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: empty address string is not allowed",
		},
		{
			name: "invalid delegator",
			input: &types.MsgBeginRedelegate{
				DelegatorAddress:    "invalid",
				ValidatorSrcAddress: srcValAddr.String(),
				ValidatorDstAddress: dstValAddr.String(),
				Amount:              sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name: "invalid destination validator",
			input: &types.MsgBeginRedelegate{
				DelegatorAddress:    Addr.String(),
				ValidatorSrcAddress: srcValAddr.String(),
				ValidatorDstAddress: sdk.AccAddress([]byte("invalid")).String(),
				Amount:              sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "invalid destination validator address",
		},
		{
			name: "validator does not exist",
			input: &types.MsgBeginRedelegate{
				DelegatorAddress:    Addr.String(),
				ValidatorSrcAddress: sdk.ValAddress([]byte("invalid")).String(),
				ValidatorDstAddress: dstValAddr.String(),
				Amount:              sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
		{
			name: "self redelegation",
			input: &types.MsgBeginRedelegate{
				DelegatorAddress:    Addr.String(),
				ValidatorSrcAddress: srcValAddr.String(),
				ValidatorDstAddress: srcValAddr.String(),
				Amount:              sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "cannot redelegate to the same validator",
		},
		{
			name: "amount greater than delegated shares amount",
			input: &types.MsgBeginRedelegate{
				DelegatorAddress:    Addr.String(),
				ValidatorSrcAddress: srcValAddr.String(),
				ValidatorDstAddress: dstValAddr.String(),
				Amount:              sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(101)),
			},
			expErr:    true,
			expErrMsg: "invalid shares amount",
		},
		{
			name: "zero amount",
			input: &types.MsgBeginRedelegate{
				DelegatorAddress:    Addr.String(),
				ValidatorSrcAddress: srcValAddr.String(),
				ValidatorDstAddress: dstValAddr.String(),
				Amount:              sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(0)),
			},
			expErr:    true,
			expErrMsg: "invalid shares amount",
		},
		{
			name: "invalid coin denom",
			input: &types.MsgBeginRedelegate{
				DelegatorAddress:    Addr.String(),
				ValidatorSrcAddress: srcValAddr.String(),
				ValidatorDstAddress: dstValAddr.String(),
				Amount:              sdk.NewCoin("test", shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "invalid coin denomination",
		},
		{
			name: "valid msg",
			input: &types.MsgBeginRedelegate{
				DelegatorAddress:    Addr.String(),
				ValidatorSrcAddress: srcValAddr.String(),
				ValidatorDstAddress: dstValAddr.String(),
				Amount:              sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := msgServer.BeginRedelegate(ctx, tc.input)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgUndelegate() {
	ctx, keeper, msgServer := s.ctx, s.stakingKeeper, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	pk := ed25519.GenPrivKey().PubKey()
	require.NotNil(pk)

	comm := types.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	amt := sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))}

	msg, err := types.NewMsgCreateValidator(ValAddr.String(), pk, amt, types.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(err)
	res, err := msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	require.NotNil(res)

	shares := math.LegacyNewDec(100)
	del := types.NewDelegation(Addr.String(), ValAddr.String(), shares)
	require.NoError(keeper.SetDelegation(ctx, del))
	_, err = keeper.Delegations.Get(ctx, collections.Join(Addr, ValAddr))
	require.NoError(err)

	testCases := []struct {
		name      string
		input     *types.MsgUndelegate
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid validator",
			input: &types.MsgUndelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: sdk.AccAddress([]byte("invalid")).String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "invalid validator address",
		},
		{
			name: "empty delegator",
			input: &types.MsgUndelegate{
				DelegatorAddress: "",
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: shares.RoundInt()},
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: empty address string is not allowed",
		},
		{
			name: "invalid delegator",
			input: &types.MsgUndelegate{
				DelegatorAddress: "invalid",
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: shares.RoundInt()},
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: decoding bech32 failed",
		},
		{
			name: "validator does not exist",
			input: &types.MsgUndelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: sdk.ValAddress([]byte("invalid")).String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
		{
			name: "amount greater than delegated shares amount",
			input: &types.MsgUndelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(101)),
			},
			expErr:    true,
			expErrMsg: "invalid shares amount",
		},
		{
			name: "zero amount",
			input: &types.MsgUndelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(0)),
			},
			expErr:    true,
			expErrMsg: "invalid shares amount",
		},
		{
			name: "invalid coin denom",
			input: &types.MsgUndelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin("test", shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "invalid coin denomination",
		},
		{
			name: "valid msg",
			input: &types.MsgUndelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := msgServer.Undelegate(ctx, tc.input)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgCancelUnbondingDelegation() {
	ctx, keeper, msgServer, ak := s.ctx, s.stakingKeeper, s.msgServer, s.accountKeeper
	require := s.Require()

	pk := ed25519.GenPrivKey().PubKey()
	require.NotNil(pk)

	comm := types.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	amt := sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))}

	s.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), Addr, types.NotBondedPoolName, gomock.Any()).AnyTimes()

	msg, err := types.NewMsgCreateValidator(ValAddr.String(), pk, amt, types.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(err)
	res, err := msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	require.NotNil(res)

	shares := math.LegacyNewDec(100)
	del := types.NewDelegation(Addr.String(), ValAddr.String(), shares)
	require.NoError(keeper.SetDelegation(ctx, del))
	resDel, err := keeper.Delegations.Get(ctx, collections.Join(Addr, ValAddr))
	require.NoError(err)
	require.Equal(del, resDel)

	ubd := types.NewUnbondingDelegation(Addr, ValAddr, 10, ctx.HeaderInfo().Time.Add(time.Minute*10), shares.RoundInt(), 0, keeper.ValidatorAddressCodec(), ak.AddressCodec())
	require.NoError(keeper.SetUnbondingDelegation(ctx, ubd))
	resUnbond, err := keeper.GetUnbondingDelegation(ctx, Addr, ValAddr)
	require.NoError(err)
	require.Equal(ubd, resUnbond)

	testCases := []struct {
		name      string
		input     *types.MsgCancelUnbondingDelegation
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid validator",
			input: &types.MsgCancelUnbondingDelegation{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: sdk.AccAddress([]byte("invalid")).String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
				CreationHeight:   10,
			},
			expErr:    true,
			expErrMsg: "invalid validator address",
		},
		{
			name: "empty delegator",
			input: &types.MsgCancelUnbondingDelegation{
				DelegatorAddress: "",
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
				CreationHeight:   10,
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: empty address string is not allowed",
		},
		{
			name: "invalid delegator",
			input: &types.MsgCancelUnbondingDelegation{
				DelegatorAddress: "invalid",
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
				CreationHeight:   10,
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: decoding bech32 failed",
		},
		{
			name: "entry not found at height",
			input: &types.MsgCancelUnbondingDelegation{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
				CreationHeight:   11,
			},
			expErr:    true,
			expErrMsg: "unbonding delegation entry is not found at block height",
		},
		{
			name: "invalid height",
			input: &types.MsgCancelUnbondingDelegation{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
				CreationHeight:   -1,
			},
			expErr:    true,
			expErrMsg: "invalid height",
		},
		{
			name: "invalid coin",
			input: &types.MsgCancelUnbondingDelegation{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin("test", shares.RoundInt()),
				CreationHeight:   10,
			},
			expErr:    true,
			expErrMsg: "invalid coin denomination",
		},
		{
			name: "validator does not exist",
			input: &types.MsgCancelUnbondingDelegation{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: sdk.ValAddress([]byte("invalid")).String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
				CreationHeight:   10,
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
		{
			name: "amount is greater than balance",
			input: &types.MsgCancelUnbondingDelegation{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(101)),
				CreationHeight:   10,
			},
			expErr:    true,
			expErrMsg: "amount is greater than the unbonding delegation entry balance",
		},
		{
			name: "zero amount",
			input: &types.MsgCancelUnbondingDelegation{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(0)),
				CreationHeight:   10,
			},
			expErr:    true,
			expErrMsg: "invalid amount",
		},
		{
			name: "valid msg",
			input: &types.MsgCancelUnbondingDelegation{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
				CreationHeight:   10,
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := msgServer.CancelUnbondingDelegation(ctx, tc.input)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgUpdateParams() {
	ctx, keeper, msgServer := s.ctx, s.stakingKeeper, s.msgServer
	require := s.Require()

	testCases := []struct {
		name      string
		input     *types.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "valid params",
			input: &types.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params:    types.DefaultParams(),
			},
			expErr: false,
		},
		{
			name: "invalid authority",
			input: &types.MsgUpdateParams{
				Authority: "invalid",
				Params:    types.DefaultParams(),
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "negative commission rate",
			input: &types.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params: types.Params{
					MinCommissionRate: math.LegacyNewDec(-10),
					UnbondingTime:     types.DefaultUnbondingTime,
					MaxValidators:     types.DefaultMaxValidators,
					MaxEntries:        types.DefaultMaxEntries,
					HistoricalEntries: types.DefaultHistoricalEntries,
					BondDenom:         types.BondStatusBonded,
				},
			},
			expErr:    true,
			expErrMsg: "minimum commission rate cannot be negative",
		},
		{
			name: "commission rate cannot be bigger than 100",
			input: &types.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params: types.Params{
					MinCommissionRate: math.LegacyNewDec(2),
					UnbondingTime:     types.DefaultUnbondingTime,
					MaxValidators:     types.DefaultMaxValidators,
					MaxEntries:        types.DefaultMaxEntries,
					HistoricalEntries: types.DefaultHistoricalEntries,
					BondDenom:         types.BondStatusBonded,
				},
			},
			expErr:    true,
			expErrMsg: "minimum commission rate cannot be greater than 100%",
		},
		{
			name: "invalid bond denom",
			input: &types.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params: types.Params{
					MinCommissionRate: types.DefaultMinCommissionRate,
					UnbondingTime:     types.DefaultUnbondingTime,
					MaxValidators:     types.DefaultMaxValidators,
					MaxEntries:        types.DefaultMaxEntries,
					HistoricalEntries: types.DefaultHistoricalEntries,
					BondDenom:         "",
				},
			},
			expErr:    true,
			expErrMsg: "bond denom cannot be blank",
		},
		{
			name: "max validators must be positive",
			input: &types.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params: types.Params{
					MinCommissionRate: types.DefaultMinCommissionRate,
					UnbondingTime:     types.DefaultUnbondingTime,
					MaxValidators:     0,
					MaxEntries:        types.DefaultMaxEntries,
					HistoricalEntries: types.DefaultHistoricalEntries,
					BondDenom:         types.BondStatusBonded,
				},
			},
			expErr:    true,
			expErrMsg: "max validators must be positive",
		},
		{
			name: "max entries most be positive",
			input: &types.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params: types.Params{
					MinCommissionRate: types.DefaultMinCommissionRate,
					UnbondingTime:     types.DefaultUnbondingTime,
					MaxValidators:     types.DefaultMaxValidators,
					MaxEntries:        0,
					HistoricalEntries: types.DefaultHistoricalEntries,
					BondDenom:         types.BondStatusBonded,
				},
			},
			expErr:    true,
			expErrMsg: "max entries must be positive",
		},
		{
			name: "negative unbounding time",
			input: &types.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params: types.Params{
					UnbondingTime:     time.Hour * 24 * 7 * 3 * -1,
					MaxEntries:        types.DefaultMaxEntries,
					MaxValidators:     types.DefaultMaxValidators,
					HistoricalEntries: types.DefaultHistoricalEntries,
					MinCommissionRate: types.DefaultMinCommissionRate,
					BondDenom:         "denom",
				},
			},
			expErr:    true,
			expErrMsg: "unbonding time must be positive",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := msgServer.UpdateParams(ctx, tc.input)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
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
