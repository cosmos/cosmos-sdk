package keeper_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	PKS     = simtestutil.CreateTestPubKeys(3)
	Addr    = sdk.AccAddress(PKS[0].Address())
	ValAddr = sdk.ValAddress(Addr)
)

func (s *KeeperTestSuite) execExpectCalls() {
	s.accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()
	s.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), Addr, stakingtypes.NotBondedPoolName, gomock.Any()).AnyTimes()
}

func (s *KeeperTestSuite) TestMsgCreateValidator() {
	ctx, msgServer := s.ctx, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	pk1 := ed25519.GenPrivKey().PubKey()
	require.NotNil(pk1)

	pubkey, err := codectypes.NewAnyWithValue(pk1)
	require.NoError(err)

	testCases := []struct {
		name      string
		input     *stakingtypes.MsgCreateValidator
		expErr    bool
		expErrMsg string
	}{
		{
			name: "empty description",
			input: &stakingtypes.MsgCreateValidator{
				Description: stakingtypes.Description{},
				Commission: stakingtypes.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin("stake", 10000),
			},
			expErr:    true,
			expErrMsg: "empty description",
		},
		{
			name: "invalid validator address",
			input: &stakingtypes.MsgCreateValidator{
				Description: stakingtypes.Description{
					Moniker: "NewValidator",
				},
				Commission: stakingtypes.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  sdk.AccAddress([]byte("invalid")).String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin("stake", 10000),
			},
			expErr:    true,
			expErrMsg: "invalid validator address",
		},
		{
			name: "empty validator pubkey",
			input: &stakingtypes.MsgCreateValidator{
				Description: stakingtypes.Description{
					Moniker: "NewValidator",
				},
				Commission: stakingtypes.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            nil,
				Value:             sdk.NewInt64Coin("stake", 10000),
			},
			expErr:    true,
			expErrMsg: "empty validator public key",
		},
		{
			name: "empty delegation amount",
			input: &stakingtypes.MsgCreateValidator{
				Description: stakingtypes.Description{
					Moniker: "NewValidator",
				},
				Commission: stakingtypes.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin("stake", 0),
			},
			expErr:    true,
			expErrMsg: "invalid delegation amount",
		},
		{
			name: "nil delegation amount",
			input: &stakingtypes.MsgCreateValidator{
				Description: stakingtypes.Description{
					Moniker: "NewValidator",
				},
				Commission: stakingtypes.CommissionRates{
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
			input: &stakingtypes.MsgCreateValidator{
				Description: stakingtypes.Description{
					Moniker: "NewValidator",
				},
				Commission: stakingtypes.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(0),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin("stake", 10000),
			},
			expErr:    true,
			expErrMsg: "minimum self delegation must be a positive integer",
		},
		{
			name: "negative minimum self delegation",
			input: &stakingtypes.MsgCreateValidator{
				Description: stakingtypes.Description{
					Moniker: "NewValidator",
				},
				Commission: stakingtypes.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(-1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin("stake", 10000),
			},
			expErr:    true,
			expErrMsg: "minimum self delegation must be a positive integer",
		},
		{
			name: "delegation less than minimum self delegation",
			input: &stakingtypes.MsgCreateValidator{
				Description: stakingtypes.Description{
					Moniker: "NewValidator",
				},
				Commission: stakingtypes.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(100),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin("stake", 10),
			},
			expErr:    true,
			expErrMsg: "validator's self delegation must be greater than their minimum self delegation",
		},
		{
			name: "valid msg",
			input: &stakingtypes.MsgCreateValidator{
				Description: stakingtypes.Description{
					Moniker:         "NewValidator",
					Identity:        "xyz",
					Website:         "xyz.com",
					SecurityContact: "xyz@gmail.com",
					Details:         "details",
				},
				Commission: stakingtypes.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin("stake", 10000),
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
	newCtx := ctx.WithBlockTime(ctx.BlockTime().AddDate(0, 0, 1))

	pk := ed25519.GenPrivKey().PubKey()
	require.NotNil(pk)

	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	msg, err := stakingtypes.NewMsgCreateValidator(ValAddr.String(), pk, sdk.NewCoin("stake", math.NewInt(10)), stakingtypes.Description{Moniker: "NewVal"}, comm, math.OneInt())
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
		input     *stakingtypes.MsgEditValidator
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid validator",
			ctx:  newCtx,
			input: &stakingtypes.MsgEditValidator{
				Description: stakingtypes.Description{
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
			input: &stakingtypes.MsgEditValidator{
				Description:       stakingtypes.Description{},
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
			input: &stakingtypes.MsgEditValidator{
				Description: stakingtypes.Description{
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
			input: &stakingtypes.MsgEditValidator{
				Description: stakingtypes.Description{
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
			input: &stakingtypes.MsgEditValidator{
				Description: stakingtypes.Description{
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
			input: &stakingtypes.MsgEditValidator{
				Description: stakingtypes.Description{
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
			input: &stakingtypes.MsgEditValidator{
				Description: stakingtypes.Description{
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
			input: &stakingtypes.MsgEditValidator{
				Description: stakingtypes.Description{
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
			input: &stakingtypes.MsgEditValidator{
				Description: stakingtypes.Description{
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

	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))

	msg, err := stakingtypes.NewMsgCreateValidator(ValAddr.String(), pk, sdk.NewCoin("stake", math.NewInt(10)), stakingtypes.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(err)

	res, err := msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	require.NotNil(res)

	testCases := []struct {
		name      string
		input     *stakingtypes.MsgDelegate
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid validator",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: sdk.AccAddress([]byte("invalid")).String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid validator address",
		},
		{
			name: "empty delegator",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: "",
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: empty address string is not allowed",
		},
		{
			name: "invalid delegator",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: "invalid",
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: decoding bech32 failed",
		},
		{
			name: "validator does not exist",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: sdk.ValAddress([]byte("val")).String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
		{
			name: "zero amount",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(0))},
			},
			expErr:    true,
			expErrMsg: "invalid delegation amount",
		},
		{
			name: "negative amount",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(-1))},
			},
			expErr:    true,
			expErrMsg: "invalid delegation amount",
		},
		{
			name: "invalid BondDenom",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: "test", Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid coin denomination",
		},
		{
			name: "valid msg",
			input: &stakingtypes.MsgDelegate{
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

	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	amt := sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))}

	msg, err := stakingtypes.NewMsgCreateValidator(srcValAddr.String(), pk, amt, stakingtypes.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(err)
	res, err := msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	require.NotNil(res)
	s.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), addr2, stakingtypes.NotBondedPoolName, gomock.Any()).AnyTimes()

	msg, err = stakingtypes.NewMsgCreateValidator(dstValAddr.String(), dstPk, amt, stakingtypes.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(err)

	res, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	require.NotNil(res)

	shares := math.LegacyNewDec(100)
	del := stakingtypes.NewDelegation(Addr.String(), srcValAddr.String(), shares)
	require.NoError(keeper.SetDelegation(ctx, del))
	_, err = keeper.Delegations.Get(ctx, collections.Join(Addr, srcValAddr))
	require.NoError(err)

	testCases := []struct {
		name      string
		input     *stakingtypes.MsgBeginRedelegate
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid source validator",
			input: &stakingtypes.MsgBeginRedelegate{
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
			input: &stakingtypes.MsgBeginRedelegate{
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
			input: &stakingtypes.MsgBeginRedelegate{
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
			input: &stakingtypes.MsgBeginRedelegate{
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
			input: &stakingtypes.MsgBeginRedelegate{
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
			input: &stakingtypes.MsgBeginRedelegate{
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
			input: &stakingtypes.MsgBeginRedelegate{
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
			input: &stakingtypes.MsgBeginRedelegate{
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
			input: &stakingtypes.MsgBeginRedelegate{
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
			input: &stakingtypes.MsgBeginRedelegate{
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

	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	amt := sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))}

	msg, err := stakingtypes.NewMsgCreateValidator(ValAddr.String(), pk, amt, stakingtypes.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(err)
	res, err := msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	require.NotNil(res)

	shares := math.LegacyNewDec(100)
	del := stakingtypes.NewDelegation(Addr.String(), ValAddr.String(), shares)
	require.NoError(keeper.SetDelegation(ctx, del))
	_, err = keeper.Delegations.Get(ctx, collections.Join(Addr, ValAddr))
	require.NoError(err)

	testCases := []struct {
		name      string
		input     *stakingtypes.MsgUndelegate
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid validator",
			input: &stakingtypes.MsgUndelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: sdk.AccAddress([]byte("invalid")).String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "invalid validator address",
		},
		{
			name: "empty delegator",
			input: &stakingtypes.MsgUndelegate{
				DelegatorAddress: "",
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: shares.RoundInt()},
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: empty address string is not allowed",
		},
		{
			name: "invalid delegator",
			input: &stakingtypes.MsgUndelegate{
				DelegatorAddress: "invalid",
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: shares.RoundInt()},
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: decoding bech32 failed",
		},
		{
			name: "validator does not exist",
			input: &stakingtypes.MsgUndelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: sdk.ValAddress([]byte("invalid")).String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
		{
			name: "amount greater than delegated shares amount",
			input: &stakingtypes.MsgUndelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(101)),
			},
			expErr:    true,
			expErrMsg: "invalid shares amount",
		},
		{
			name: "zero amount",
			input: &stakingtypes.MsgUndelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(0)),
			},
			expErr:    true,
			expErrMsg: "invalid shares amount",
		},
		{
			name: "invalid coin denom",
			input: &stakingtypes.MsgUndelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin("test", shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "invalid coin denomination",
		},
		{
			name: "valid msg",
			input: &stakingtypes.MsgUndelegate{
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

	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	amt := sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))}

	s.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), Addr, stakingtypes.NotBondedPoolName, gomock.Any()).AnyTimes()

	msg, err := stakingtypes.NewMsgCreateValidator(ValAddr.String(), pk, amt, stakingtypes.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(err)
	res, err := msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	require.NotNil(res)

	shares := math.LegacyNewDec(100)
	del := stakingtypes.NewDelegation(Addr.String(), ValAddr.String(), shares)
	require.NoError(keeper.SetDelegation(ctx, del))
	resDel, err := keeper.Delegations.Get(ctx, collections.Join(Addr, ValAddr))
	require.NoError(err)
	require.Equal(del, resDel)

	ubd := stakingtypes.NewUnbondingDelegation(Addr, ValAddr, 10, ctx.BlockTime().Add(time.Minute*10), shares.RoundInt(), 0, keeper.ValidatorAddressCodec(), ak.AddressCodec())
	require.NoError(keeper.SetUnbondingDelegation(ctx, ubd))
	resUnbond, err := keeper.GetUnbondingDelegation(ctx, Addr, ValAddr)
	require.NoError(err)
	require.Equal(ubd, resUnbond)

	testCases := []struct {
		name      string
		input     *stakingtypes.MsgCancelUnbondingDelegation
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid validator",
			input: &stakingtypes.MsgCancelUnbondingDelegation{
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
			input: &stakingtypes.MsgCancelUnbondingDelegation{
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
			input: &stakingtypes.MsgCancelUnbondingDelegation{
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
			input: &stakingtypes.MsgCancelUnbondingDelegation{
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
			input: &stakingtypes.MsgCancelUnbondingDelegation{
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
			input: &stakingtypes.MsgCancelUnbondingDelegation{
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
			input: &stakingtypes.MsgCancelUnbondingDelegation{
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
			input: &stakingtypes.MsgCancelUnbondingDelegation{
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
			input: &stakingtypes.MsgCancelUnbondingDelegation{
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
			input: &stakingtypes.MsgCancelUnbondingDelegation{
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
		input     *stakingtypes.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "valid params",
			input: &stakingtypes.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params:    stakingtypes.DefaultParams(),
			},
			expErr: false,
		},
		{
			name: "invalid authority",
			input: &stakingtypes.MsgUpdateParams{
				Authority: "invalid",
				Params:    stakingtypes.DefaultParams(),
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "negative commission rate",
			input: &stakingtypes.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params: stakingtypes.Params{
					MinCommissionRate: math.LegacyNewDec(-10),
					UnbondingTime:     stakingtypes.DefaultUnbondingTime,
					MaxValidators:     stakingtypes.DefaultMaxValidators,
					MaxEntries:        stakingtypes.DefaultMaxEntries,
					HistoricalEntries: stakingtypes.DefaultHistoricalEntries,
					BondDenom:         stakingtypes.BondStatusBonded,
				},
			},
			expErr:    true,
			expErrMsg: "minimum commission rate cannot be negative",
		},
		{
			name: "commission rate cannot be bigger than 100",
			input: &stakingtypes.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params: stakingtypes.Params{
					MinCommissionRate: math.LegacyNewDec(2),
					UnbondingTime:     stakingtypes.DefaultUnbondingTime,
					MaxValidators:     stakingtypes.DefaultMaxValidators,
					MaxEntries:        stakingtypes.DefaultMaxEntries,
					HistoricalEntries: stakingtypes.DefaultHistoricalEntries,
					BondDenom:         stakingtypes.BondStatusBonded,
				},
			},
			expErr:    true,
			expErrMsg: "minimum commission rate cannot be greater than 100%",
		},
		{
			name: "invalid bond denom",
			input: &stakingtypes.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params: stakingtypes.Params{
					MinCommissionRate: stakingtypes.DefaultMinCommissionRate,
					UnbondingTime:     stakingtypes.DefaultUnbondingTime,
					MaxValidators:     stakingtypes.DefaultMaxValidators,
					MaxEntries:        stakingtypes.DefaultMaxEntries,
					HistoricalEntries: stakingtypes.DefaultHistoricalEntries,
					BondDenom:         "",
				},
			},
			expErr:    true,
			expErrMsg: "bond denom cannot be blank",
		},
		{
			name: "max validators must be positive",
			input: &stakingtypes.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params: stakingtypes.Params{
					MinCommissionRate: stakingtypes.DefaultMinCommissionRate,
					UnbondingTime:     stakingtypes.DefaultUnbondingTime,
					MaxValidators:     0,
					MaxEntries:        stakingtypes.DefaultMaxEntries,
					HistoricalEntries: stakingtypes.DefaultHistoricalEntries,
					BondDenom:         stakingtypes.BondStatusBonded,
				},
			},
			expErr:    true,
			expErrMsg: "max validators must be positive",
		},
		{
			name: "max entries most be positive",
			input: &stakingtypes.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params: stakingtypes.Params{
					MinCommissionRate: stakingtypes.DefaultMinCommissionRate,
					UnbondingTime:     stakingtypes.DefaultUnbondingTime,
					MaxValidators:     stakingtypes.DefaultMaxValidators,
					MaxEntries:        0,
					HistoricalEntries: stakingtypes.DefaultHistoricalEntries,
					BondDenom:         stakingtypes.BondStatusBonded,
				},
			},
			expErr:    true,
			expErrMsg: "max entries must be positive",
		},
		{
			name: "negative unbounding time",
			input: &stakingtypes.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params: stakingtypes.Params{
					UnbondingTime:     time.Hour * 24 * 7 * 3 * -1,
					MaxEntries:        stakingtypes.DefaultMaxEntries,
					MaxValidators:     stakingtypes.DefaultMaxValidators,
					HistoricalEntries: stakingtypes.DefaultHistoricalEntries,
					MinCommissionRate: stakingtypes.DefaultMinCommissionRate,
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
