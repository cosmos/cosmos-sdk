package vesting_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

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
				time.Now().Unix(),
				-10,
				true,
			),
			expErr:    true,
			expErrMsg: "invalid end time",
		},
		"invalid start time": {
			input: vestingtypes.NewMsgCreateVestingAccount(
				fromAddr,
				to1Addr,
				sdk.Coins{fooCoin},
				-10,
				time.Now().Unix(),
				true,
			),
			expErr:    true,
			expErrMsg: "invalid start time",
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
				time.Now().Unix()-1,
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
				time.Now().Unix()-1,
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
				time.Now().Unix()-1,
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
				time.Now().Unix()-1,
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

func TestVestingTestSuite(t *testing.T) {
	suite.Run(t, new(VestingTestSuite))
}
