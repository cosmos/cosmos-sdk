package keeper_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/mint"
	"cosmossdk.io/x/mint/keeper"
	minttestutil "cosmossdk.io/x/mint/testutil"
	"cosmossdk.io/x/mint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

var minterAcc = authtypes.NewEmptyModuleAccount(types.ModuleName, authtypes.Minter)

type GenesisTestSuite struct {
	suite.Suite

	sdkCtx        sdk.Context
	keeper        keeper.Keeper
	cdc           codec.BinaryCodec
	accountKeeper types.AccountKeeper
	key           *storetypes.KVStoreKey
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (s *GenesisTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModule{})

	// gomock initializations
	ctrl := gomock.NewController(s.T())
	s.cdc = codec.NewProtoCodec(encCfg.InterfaceRegistry)
	s.sdkCtx = testCtx.Ctx
	s.key = key

	stakingKeeper := minttestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := minttestutil.NewMockAccountKeeper(ctrl)
	bankKeeper := minttestutil.NewMockBankKeeper(ctrl)
	s.accountKeeper = accountKeeper
	accountKeeper.EXPECT().GetModuleAddress(minterAcc.Name).Return(minterAcc.GetAddress())
	accountKeeper.EXPECT().GetModuleAccount(s.sdkCtx, minterAcc.Name).Return(minterAcc)

	s.keeper = keeper.NewKeeper(s.cdc, runtime.NewEnvironment(runtime.NewKVStoreService(key), log.NewNopLogger()), stakingKeeper, accountKeeper, bankKeeper, "", "")
}

func (s *GenesisTestSuite) TestImportExportGenesis() {
	genesisState := types.DefaultGenesisState()
	genesisState.Minter = types.NewMinter(math.LegacyNewDecWithPrec(20, 2), math.LegacyNewDec(1))
	genesisState.Params = types.NewParams(
		"testDenom",
		math.LegacyNewDecWithPrec(15, 2),
		math.LegacyNewDecWithPrec(22, 2),
		math.LegacyNewDecWithPrec(9, 2),
		math.LegacyNewDecWithPrec(69, 2),
		uint64(60*60*8766/5),
	)

	s.keeper.InitGenesis(s.sdkCtx, s.accountKeeper, genesisState)

	minter, err := s.keeper.Minter.Get(s.sdkCtx)
	s.Require().Equal(genesisState.Minter, minter)
	s.Require().NoError(err)

	invalidCtx := testutil.DefaultContextWithDB(s.T(), s.key, storetypes.NewTransientStoreKey("transient_test"))
	_, err = s.keeper.Minter.Get(invalidCtx.Ctx)
	s.Require().ErrorIs(err, collections.ErrNotFound)

	params, err := s.keeper.Params.Get(s.sdkCtx)
	s.Require().Equal(genesisState.Params, params)
	s.Require().NoError(err)

	genesisState2 := s.keeper.ExportGenesis(s.sdkCtx)
	s.Require().Equal(genesisState, genesisState2)
}
