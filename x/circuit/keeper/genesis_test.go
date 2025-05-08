package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/circuit"
	"github.com/cosmos/cosmos-sdk/x/circuit/keeper"
	"github.com/cosmos/cosmos-sdk/x/circuit/types"
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
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(circuit.AppModuleBasic{})

	sdkCtx := testCtx.Ctx
	s.ctx = sdkCtx
	s.cdc = codec.NewProtoCodec(encCfg.InterfaceRegistry)
	authority := authtypes.NewModuleAddress("gov")
	ac := addresscodec.NewBech32Codec("cosmos")

	bz, err := ac.StringToBytes(authority.String())
	s.Require().NoError(err)
	s.addrBytes = bz

	s.keeper = keeper.NewKeeper(s.cdc, runtime.NewKVStoreService(key), authority.String(), ac)
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

	s.keeper.InitGenesis(s.ctx, genesisState)

	exported := s.keeper.ExportGenesis(s.ctx)
	bz, err := s.cdc.MarshalJSON(exported)
	s.Require().NoError(err)

	var exportedGenesisState types.GenesisState
	err = s.cdc.UnmarshalJSON(bz, &exportedGenesisState)
	s.Require().NoError(err)

	s.Require().Equal(genesisState.AccountPermissions, exportedGenesisState.AccountPermissions)
	s.Require().Equal(genesisState.DisabledTypeUrls, exportedGenesisState.DisabledTypeUrls)
}
