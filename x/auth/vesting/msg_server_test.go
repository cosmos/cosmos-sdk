package vesting_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtestutil "github.com/cosmos/cosmos-sdk/x/auth/vesting/testutil"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

var (
	fromAddr   = sdk.AccAddress([]byte("from1________________"))
	to1Addr    = sdk.AccAddress([]byte("to1__________________"))
	to2Addr    = sdk.AccAddress([]byte("to2__________________"))
	to3Addr    = sdk.AccAddress([]byte("to3__________________"))
	fooCoin    = sdk.NewInt64Coin("foo", 100)
	periodCoin = sdk.NewInt64Coin("foo", 20)
)

type VestingTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	accountKeeper authkeeper.AccountKeeper
	bankKeeper    *vestingtestutil.MockBankKeeper
	msgServer     vestingtypes.MsgServer
}

func (s *VestingTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(authtypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	maccPerms := map[string][]string{}

	ctrl := gomock.NewController(s.T())
	s.bankKeeper = vestingtestutil.NewMockBankKeeper(ctrl)
	s.accountKeeper = authkeeper.NewAccountKeeper(
		encCfg.Codec,
		storeService,
		authtypes.ProtoBaseAccount,
		maccPerms,
		authcodec.NewBech32Codec("cosmos"),
		"cosmos",
		authtypes.NewModuleAddress("gov").String(),
	)

	vestingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	authtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	s.msgServer = vesting.NewMsgServerImpl(s.accountKeeper, s.bankKeeper)
}

func (s *VestingTestSuite) TestCreateVestingAccount() {
	testCases := map[string]struct {
		preRun    func()
		input     *vestingtypes.MsgCreateVestingAccount
		expErr    bool
		expErrMsg string
	}{
		"empty from address": {
			input: vestingtypes.NewMsgCreateVestingAccount(
				[]byte{},
				to1Addr,
				sdk.Coins{fooCoin},
				time.Now().Unix(),
				true,
			),
			expErr:    true,
			expErrMsg: "invalid 'from' address",
		},
		"empty to address": {
			input: vestingtypes.NewMsgCreateVestingAccount(
				fromAddr,
				[]byte{},
				sdk.Coins{fooCoin},
				time.Now().Unix(),
				true,
			),
			expErr:    true,
			expErrMsg: "invalid 'to' address",
		},
		"invalid coins": {
			input: vestingtypes.NewMsgCreateVestingAccount(
				fromAddr,
				to1Addr,
				sdk.Coins{sdk.Coin{Denom: "stake", Amount: math.NewInt(-1)}},
				time.Now().Unix(),
				true,
			),
			expErr:    true,
			expErrMsg: "-1stake: invalid coins",
		},
		"invalid end time": {
			input: vestingtypes.NewMsgCreateVestingAccount(
				fromAddr,
				to1Addr,
				sdk.Coins{fooCoin},
				-10,
				true,
			),
			expErr:    true,
			expErrMsg: "invalid end time",
		},
		"create for existing account": {
			preRun: func() {
				toAcc := s.accountKeeper.NewAccountWithAddress(s.ctx, to1Addr)
				s.bankKeeper.EXPECT().IsSendEnabledCoins(gomock.Any(), fooCoin).Return(nil)
				s.accountKeeper.SetAccount(s.ctx, toAcc)
				s.bankKeeper.EXPECT().BlockedAddr(to1Addr).Return(false)
			},
			input: vestingtypes.NewMsgCreateVestingAccount(
				fromAddr,
				to1Addr,
				sdk.Coins{fooCoin},
				time.Now().Unix(),
				true,
			),
			expErr:    true,
			expErrMsg: "already exists",
		},
		"create for blocked account": {
			preRun: func() {
				s.bankKeeper.EXPECT().IsSendEnabledCoins(gomock.Any(), fooCoin).Return(nil)
				s.bankKeeper.EXPECT().BlockedAddr(to1Addr).Return(true)
			},
			input: vestingtypes.NewMsgCreateVestingAccount(
				fromAddr,
				to1Addr,
				sdk.Coins{fooCoin},
				time.Now().Unix(),
				true,
			),
			expErr:    true,
			expErrMsg: "not allowed to receive funds",
		},
		"create a valid delayed vesting account": {
			preRun: func() {
				s.bankKeeper.EXPECT().IsSendEnabledCoins(gomock.Any(), fooCoin).Return(nil)
				s.bankKeeper.EXPECT().BlockedAddr(to2Addr).Return(false)
				s.bankKeeper.EXPECT().SendCoins(gomock.Any(), fromAddr, to2Addr, sdk.Coins{fooCoin}).Return(nil)
			},
			input: vestingtypes.NewMsgCreateVestingAccount(
				fromAddr,
				to2Addr,
				sdk.Coins{fooCoin},
				time.Now().Unix(),
				true,
			),
			expErr:    false,
			expErrMsg: "",
		},
		"create a valid continuous vesting account": {
			preRun: func() {
				s.bankKeeper.EXPECT().IsSendEnabledCoins(gomock.Any(), fooCoin).Return(nil)
				s.bankKeeper.EXPECT().BlockedAddr(to3Addr).Return(false)
				s.bankKeeper.EXPECT().SendCoins(gomock.Any(), fromAddr, to3Addr, sdk.Coins{fooCoin}).Return(nil)
			},
			input: vestingtypes.NewMsgCreateVestingAccount(
				fromAddr,
				to3Addr,
				sdk.Coins{fooCoin},
				time.Now().Unix(),
				false,
			),
			expErr:    false,
			expErrMsg: "",
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			if tc.preRun != nil {
				tc.preRun()
			}
			_, err := s.msgServer.CreateVestingAccount(s.ctx, tc.input)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *VestingTestSuite) TestCreatePermanentLockedAccount() {
	testCases := map[string]struct {
		preRun    func()
		input     *vestingtypes.MsgCreatePermanentLockedAccount
		expErr    bool
		expErrMsg string
	}{
		"empty from address": {
			input: vestingtypes.NewMsgCreatePermanentLockedAccount(
				[]byte{},
				to1Addr,
				sdk.Coins{fooCoin},
			),
			expErr:    true,
			expErrMsg: "invalid 'from' address",
		},
		"empty to address": {
			input: vestingtypes.NewMsgCreatePermanentLockedAccount(
				fromAddr,
				[]byte{},
				sdk.Coins{fooCoin},
			),
			expErr:    true,
			expErrMsg: "invalid 'to' address",
		},
		"invalid coins": {
			input: vestingtypes.NewMsgCreatePermanentLockedAccount(
				fromAddr,
				to1Addr,
				sdk.Coins{sdk.Coin{Denom: "stake", Amount: math.NewInt(-1)}},
			),
			expErr:    true,
			expErrMsg: "-1stake: invalid coins",
		},
		"create for existing account": {
			preRun: func() {
				toAcc := s.accountKeeper.NewAccountWithAddress(s.ctx, to1Addr)
				s.bankKeeper.EXPECT().IsSendEnabledCoins(gomock.Any(), fooCoin).Return(nil)
				s.bankKeeper.EXPECT().BlockedAddr(to1Addr).Return(false)
				s.accountKeeper.SetAccount(s.ctx, toAcc)
			},
			input: vestingtypes.NewMsgCreatePermanentLockedAccount(
				fromAddr,
				to1Addr,
				sdk.Coins{fooCoin},
			),
			expErr:    true,
			expErrMsg: "already exists",
		},
		"create for blocked account": {
			preRun: func() {
				toAcc := s.accountKeeper.NewAccountWithAddress(s.ctx, to1Addr)
				s.bankKeeper.EXPECT().IsSendEnabledCoins(gomock.Any(), fooCoin).Return(nil)
				s.bankKeeper.EXPECT().BlockedAddr(to1Addr).Return(true)
				s.accountKeeper.SetAccount(s.ctx, toAcc)
			},
			input: vestingtypes.NewMsgCreatePermanentLockedAccount(
				fromAddr,
				to1Addr,
				sdk.Coins{fooCoin},
			),
			expErr:    true,
			expErrMsg: "not allowed to receive funds",
		},

		"create a valid permanent locked account": {
			preRun: func() {
				s.bankKeeper.EXPECT().IsSendEnabledCoins(gomock.Any(), fooCoin).Return(nil)
				s.bankKeeper.EXPECT().BlockedAddr(to2Addr).Return(false)
				s.bankKeeper.EXPECT().SendCoins(gomock.Any(), fromAddr, to2Addr, sdk.Coins{fooCoin}).Return(nil)
			},
			input: vestingtypes.NewMsgCreatePermanentLockedAccount(
				fromAddr,
				to2Addr,
				sdk.Coins{fooCoin},
			),
			expErr:    false,
			expErrMsg: "",
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			if tc.preRun != nil {
				tc.preRun()
			}

			_, err := s.msgServer.CreatePermanentLockedAccount(s.ctx, tc.input)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *VestingTestSuite) TestCreatePeriodicVestingAccount() {
	testCases := []struct {
		name      string
		preRun    func()
		input     *vestingtypes.MsgCreatePeriodicVestingAccount
		expErr    bool
		expErrMsg string
	}{
		{
			name: "empty from address",
			input: vestingtypes.NewMsgCreatePeriodicVestingAccount(
				[]byte{},
				to1Addr,
				time.Now().Unix(),
				[]vestingtypes.Period{
					{
						Length: 10,
						Amount: sdk.NewCoins(periodCoin),
					},
				},
			),
			expErr:    true,
			expErrMsg: "invalid 'from' address",
		},
		{
			name: "empty to address",
			input: vestingtypes.NewMsgCreatePeriodicVestingAccount(
				fromAddr,
				[]byte{},
				time.Now().Unix(),
				[]vestingtypes.Period{
					{
						Length: 10,
						Amount: sdk.NewCoins(periodCoin),
					},
				},
			),
			expErr:    true,
			expErrMsg: "invalid 'to' address",
		},
		{
			name: "invalid start time",
			input: vestingtypes.NewMsgCreatePeriodicVestingAccount(
				fromAddr,
				to1Addr,
				0,
				[]vestingtypes.Period{
					{
						Length: 10,
						Amount: sdk.NewCoins(periodCoin),
					},
				},
			),
			expErr:    true,
			expErrMsg: "invalid start time",
		},
		{
			name: "invalid period",
			input: vestingtypes.NewMsgCreatePeriodicVestingAccount(
				fromAddr,
				to1Addr,
				time.Now().Unix(),
				[]vestingtypes.Period{
					{
						Length: 0,
						Amount: sdk.NewCoins(periodCoin),
					},
				},
			),
			expErr:    true,
			expErrMsg: "invalid period",
		},
		{
			name: "invalid coins",
			input: vestingtypes.NewMsgCreatePeriodicVestingAccount(
				fromAddr,
				to1Addr,
				time.Now().Unix(),
				[]vestingtypes.Period{
					{
						Length: 1,
						Amount: sdk.Coins{sdk.Coin{Denom: "stake", Amount: math.NewInt(-1)}},
					},
				},
			),
			expErr:    true,
			expErrMsg: "-1stake: invalid coins",
		},
		{
			name: "create for existing account",
			preRun: func() {
				s.bankKeeper.EXPECT().BlockedAddr(to1Addr).Return(false)
				toAcc := s.accountKeeper.NewAccountWithAddress(s.ctx, to1Addr)
				s.accountKeeper.SetAccount(s.ctx, toAcc)
			},
			input: vestingtypes.NewMsgCreatePeriodicVestingAccount(
				fromAddr,
				to1Addr,
				time.Now().Unix(),
				[]vestingtypes.Period{
					{
						Length: 10,
						Amount: sdk.NewCoins(periodCoin),
					},
				},
			),
			expErr:    true,
			expErrMsg: "already exists",
		},
		{
			name: "create for blocked address",
			preRun: func() {
				s.bankKeeper.EXPECT().BlockedAddr(to2Addr).Return(true)
			},
			input: vestingtypes.NewMsgCreatePeriodicVestingAccount(
				fromAddr,
				to2Addr,
				time.Now().Unix(),
				[]vestingtypes.Period{
					{
						Length: 10,
						Amount: sdk.NewCoins(periodCoin),
					},
					{
						Length: 20,
						Amount: sdk.NewCoins(fooCoin),
					},
				},
			),
			expErr:    true,
			expErrMsg: "not allowed to receive funds",
		},
		{
			name: "create a valid periodic vesting account",
			preRun: func() {
				s.bankKeeper.EXPECT().IsSendEnabledCoins(gomock.Any(), periodCoin.Add(fooCoin)).Return(nil)
				s.bankKeeper.EXPECT().BlockedAddr(to2Addr).Return(false)
				s.bankKeeper.EXPECT().SendCoins(gomock.Any(), fromAddr, to2Addr, gomock.Any()).Return(nil)
			},
			input: vestingtypes.NewMsgCreatePeriodicVestingAccount(
				fromAddr,
				to2Addr,
				time.Now().Unix(),
				[]vestingtypes.Period{
					{
						Length: 10,
						Amount: sdk.NewCoins(periodCoin),
					},
					{
						Length: 20,
						Amount: sdk.NewCoins(fooCoin),
					},
				},
			),
			expErr:    false,
			expErrMsg: "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.preRun != nil {
				tc.preRun()
			}
			_, err := s.msgServer.CreatePeriodicVestingAccount(s.ctx, tc.input)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func TestVestingTestSuite(t *testing.T) {
	suite.Run(t, new(VestingTestSuite))
}
