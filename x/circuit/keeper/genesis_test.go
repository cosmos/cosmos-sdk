package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/circuit"
	"cosmossdk.io/x/circuit/keeper"
	"cosmossdk.io/x/circuit/types"

	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

type GenesisTestSuite struct {
	suite.Suite

	ctx       context.Context
	keeper    keeper.Keeper
	cdc       *codec.ProtoCodec
	addrBytes []byte
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (s *GenesisTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(types.ModuleName)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, circuit.AppModule{})

	sdkCtx := testCtx.Ctx
	s.ctx = sdkCtx
	s.cdc = codec.NewProtoCodec(encCfg.InterfaceRegistry)
	authority := authtypes.NewModuleAddress("gov")
	ac := addresscodec.NewBech32Codec("cosmos")

	bz, err := ac.StringToBytes(authority.String())
	s.Require().NoError(err)
	s.addrBytes = bz

	s.keeper = keeper.NewKeeper(runtime.NewEnvironment(runtime.NewKVStoreService(key), log.NewNopLogger()), s.cdc, authority.String(), ac)
}

func (s *GenesisTestSuite) TestInitExportGenesis() {
	perms := types.Permissions{
		Level:         3,
		LimitTypeUrls: []string{"test"},
	}
	err := s.keeper.Permissions.Set(s.ctx, s.addrBytes, perms)
	s.Require().NoError(err)

	var accounts []*types.GenesisAccountPermissions
	genAccsPerms := types.GenesisAccountPermissions{
		Address:     sdk.AccAddress(s.addrBytes).String(),
		Permissions: &perms,
	}
	accounts = append(accounts, &genAccsPerms)

	url := "test_url"

	genesisState := &types.GenesisState{
		AccountPermissions: accounts,
		DisabledTypeUrls:   []string{url},
	}

	err = s.keeper.InitGenesis(s.ctx, genesisState)
	s.Require().NoError(err)

	exported, err := s.keeper.ExportGenesis(s.ctx)
	s.Require().NoError(err)
	bz, err := s.cdc.MarshalJSON(exported)
	s.Require().NoError(err)

	var exportedGenesisState types.GenesisState
	err = s.cdc.UnmarshalJSON(bz, &exportedGenesisState)
	s.Require().NoError(err)

	s.Require().Equal(genesisState.AccountPermissions, exportedGenesisState.AccountPermissions)
	s.Require().Equal(genesisState.DisabledTypeUrls, exportedGenesisState.DisabledTypeUrls)
}
