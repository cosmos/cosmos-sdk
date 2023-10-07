package keeper_test

import (
	"encoding/json"
	"testing"
	"time"

	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type corpa struct {
	ConsAddr         sdk.ConsAddress `json:"ca"`
	InfractionHeight int64           `json:"ifh"`
	Power            int64           `json:"pow"`
	SlashFactor      math.LegacyDec  `json:"sf"`
}

func FuzzKeeperSlash(f *testing.F) {
	if testing.Short() {
		f.Skip("In -short mode")
	}

	s := new(KeeperTestSuite)
	s.f = f
	s.SetupTest()

	ctx, keeper := s.ctx, s.stakingKeeper

	// 1. Generate the corpus.
	pk0 := PKs[0]
	pk0Addr := PKs[0].Address()
	consAddr1 := sdk.ConsAddress(pk0Addr)
	validator := stakingtestutil.NewValidator(f, sdk.ValAddress(pk0Addr.Bytes()), pk0)
	if err := keeper.SetValidator(ctx, validator); err != nil {
		f.Fatal(err)
	}
	if err := keeper.SetValidatorByConsAddr(ctx, validator); err != nil {
		f.Fatal(err)
	}

	legDec1 := math.LegacyOneDec()
	frac1 := math.LegacyNewDecWithPrec(5, 1)
	consAddr2 := sdk.ConsAddress(PKs[0].Address())
	blkHeight1 := int64(100)

	corpus := []*corpa{
		{consAddr1, -2, 10, frac1},
		{consAddr1, blkHeight1, 10, frac1},
		{consAddr1, 10, 10, frac1},
		{consAddr1, 9, 10, frac1},
		{consAddr2, 9, 10, legDec1},
	}

	for _, corpa := range corpus {
		bz, err := json.Marshal(corpa)
		if err != nil {
			// TODO: Log this.
			continue
		}
		f.Add(bz)
	}

	// 2. Now fuzz it.
	f.Fuzz(func(t *testing.T, input []byte) {
		cop := new(corpa)
		if err := json.Unmarshal(input, cop); err != nil {
			// Inalid data so we can't do anything.
			return
		}
		if cop.SlashFactor.IsNil() {
			return
		}

		_, _ = keeper.Slash(s.ctx, cop.ConsAddr, cop.InfractionHeight, cop.Power, cop.SlashFactor)
	})
}

var (
	bondedAcc    = authtypes.NewEmptyModuleAccount(stakingtypes.BondedPoolName)
	notBondedAcc = authtypes.NewEmptyModuleAccount(stakingtypes.NotBondedPoolName)
	PKs          = simtestutil.CreateTestPubKeys(500)
)

type KeeperTestSuite struct {
	suite.Suite

	f             *testing.F
	ctx           sdk.Context
	stakingKeeper *stakingkeeper.Keeper
	bankKeeper    *stakingtestutil.MockBankKeeper
	accountKeeper *stakingtestutil.MockAccountKeeper
	queryClient   stakingtypes.QueryClient
	msgServer     stakingtypes.MsgServer
	key           *storetypes.KVStoreKey
	cdc           codec.Codec
}

func (s *KeeperTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	s.key = key
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(s.f, key, storetypes.NewTransientStoreKey("transient_test"))
	s.key = key
	ctx := testCtx.Ctx.WithHeaderInfo(header.Info{Time: time.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()
	s.cdc = encCfg.Codec

	ctrl := gomock.NewController(s.f)
	accountKeeper := stakingtestutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress())
	accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress())
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	bankKeeper := stakingtestutil.NewMockBankKeeper(ctrl)

	keeper := stakingkeeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		authtypes.NewModuleAddress(stakingtypes.GovModuleName).String(),
		address.NewBech32Codec("cosmosvaloper"),
		address.NewBech32Codec("cosmosvalcons"),
	)
	if err := keeper.Params.Set(ctx, stakingtypes.DefaultParams()); err != nil {
		s.f.Fatal(err)
	}

	s.ctx = ctx
	s.stakingKeeper = keeper
	s.bankKeeper = bankKeeper
	s.accountKeeper = accountKeeper

	stakingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	stakingtypes.RegisterQueryServer(queryHelper, stakingkeeper.Querier{Keeper: keeper})
	s.queryClient = stakingtypes.NewQueryClient(queryHelper)
	s.msgServer = stakingkeeper.NewMsgServerImpl(keeper)
}
