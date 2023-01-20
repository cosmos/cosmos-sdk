package module_test

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"testing"

	"cosmossdk.io/core/appmodule"
	"github.com/golang/mock/gomock"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

var errFoo = errors.New("dummy")

func TestBasicManager(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)
	legacyAmino := codec.NewLegacyAmino()
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	expDefaultGenesis := map[string]json.RawMessage{
		"mockAppModuleBasic1": json.RawMessage(``),
		"mockCoreAppModule2":  json.RawMessage(`null`),
		"MockCoreAppModule": json.RawMessage(`{
  "someField": "asd"
}`),
	}

	mockAppModuleBasic1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModuleBasic1.EXPECT().Name().AnyTimes().Return("mockAppModuleBasic1")
	mockAppModuleBasic1.EXPECT().DefaultGenesis(gomock.Eq(cdc)).Times(1).Return(json.RawMessage(``))
	// Allow ValidateGenesis to be called any times because other module can fail before this one is called.
	mockAppModuleBasic1.EXPECT().ValidateGenesis(gomock.Eq(cdc), gomock.Eq(nil), gomock.Eq(expDefaultGenesis["mockAppModuleBasic1"])).AnyTimes().Return(nil)
	mockAppModuleBasic1.EXPECT().RegisterLegacyAminoCodec(gomock.Eq(legacyAmino)).Times(1)
	mockAppModuleBasic1.EXPECT().RegisterInterfaces(gomock.Eq(interfaceRegistry)).Times(1)
	mockAppModuleBasic1.EXPECT().GetTxCmd().Times(1).Return(nil)
	mockAppModuleBasic1.EXPECT().GetQueryCmd().Times(1).Return(nil)

	mockCoreAppModule2 := mock.NewMockCoreAppModule(mockCtrl)
	mockCoreAppModule2.EXPECT().Name().AnyTimes().Return("mockCoreAppModule2")
	mockCoreAppModule2.EXPECT().DefaultGenesis(gomock.Any()).Times(1).Return(nil)
	mockCoreAppModule2.EXPECT().ValidateGenesis(gomock.Any()).Times(1).Return(nil)
	mockCoreAppModule2.EXPECT().RegisterLegacyAminoCodec(gomock.Eq(legacyAmino)).Times(1)
	mockCoreAppModule2.EXPECT().RegisterInterfaces(gomock.Eq(interfaceRegistry)).Times(1)
	mockCoreAppModule2.EXPECT().GetTxCmd().Times(1).Return(nil)
	mockCoreAppModule2.EXPECT().GetQueryCmd().Times(1).Return(nil)

	mockCoreAppModule3 := MockCoreAppModule{}

	mm := module.NewBasicManager(mockAppModuleBasic1, mockCoreAppModule2, mockCoreAppModule3)

	require.Equal(t, mockAppModuleBasic1, mm["mockAppModuleBasic1"])
	require.Equal(t, mockCoreAppModule2, mm["mockCoreAppModule2"])
	require.Equal(t, mockCoreAppModule3, mm["MockCoreAppModule"])

	mm.RegisterLegacyAminoCodec(legacyAmino)
	mm.RegisterInterfaces(interfaceRegistry)

	require.Equal(t, expDefaultGenesis, mm.DefaultGenesis(cdc))

	var data map[string]string
	require.Equal(t, map[string]string(nil), data)

	require.ErrorIs(t, mm.ValidateGenesis(cdc, nil, expDefaultGenesis), errFoo)

	mockCmd := &cobra.Command{Use: "root"}
	mm.AddTxCommands(mockCmd)

	mm.AddQueryCommands(mockCmd)

	// validate genesis returns nil
	require.Nil(t, module.NewBasicManager().ValidateGenesis(cdc, nil, expDefaultGenesis))
}

func TestGenesisOnlyAppModule(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockModule := mock.NewMockAppModuleGenesis(mockCtrl)
	mockInvariantRegistry := mock.NewMockInvariantRegistry(mockCtrl)
	goam := module.NewGenesisOnlyAppModule(mockModule)

	// no-op
	goam.RegisterInvariants(mockInvariantRegistry)
}

func TestManagerOrderSetters(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)
	mockAppModule1 := mock.NewMockAppModule(mockCtrl)
	mockAppModule2 := mock.NewMockAppModule(mockCtrl)

	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mm := module.NewManager(mockAppModule1, mockAppModule2)
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.Modules))

	require.Equal(t, []string{"module1", "module2"}, mm.OrderInitGenesis)
	mm.SetOrderInitGenesis("module2", "module1")
	require.Equal(t, []string{"module2", "module1"}, mm.OrderInitGenesis)

	require.Equal(t, []string{"module1", "module2"}, mm.OrderExportGenesis)
	mm.SetOrderExportGenesis("module2", "module1")
	require.Equal(t, []string{"module2", "module1"}, mm.OrderExportGenesis)

	require.Equal(t, []string{"module1", "module2"}, mm.OrderBeginBlockers)
	mm.SetOrderBeginBlockers("module2", "module1")
	require.Equal(t, []string{"module2", "module1"}, mm.OrderBeginBlockers)

	require.Equal(t, []string{"module1", "module2"}, mm.OrderEndBlockers)
	mm.SetOrderEndBlockers("module2", "module1")
	require.Equal(t, []string{"module2", "module1"}, mm.OrderEndBlockers)
}

func TestManager_RegisterInvariants(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule2 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mm := module.NewManager(mockAppModule1, mockAppModule2)
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.Modules))

	// test RegisterInvariants
	mockInvariantRegistry := mock.NewMockInvariantRegistry(mockCtrl)
	mockAppModule1.EXPECT().RegisterInvariants(gomock.Eq(mockInvariantRegistry)).Times(1)
	mockAppModule2.EXPECT().RegisterInvariants(gomock.Eq(mockInvariantRegistry)).Times(1)
	mm.RegisterInvariants(mockInvariantRegistry)
}

func TestManager_RegisterQueryServices(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule2 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mm := module.NewManager(mockAppModule1, mockAppModule2)
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.Modules))

	msgRouter := mock.NewMockServer(mockCtrl)
	queryRouter := mock.NewMockServer(mockCtrl)
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	cfg := module.NewConfigurator(cdc, msgRouter, queryRouter)
	mockAppModule1.EXPECT().RegisterServices(cfg).Times(1)
	mockAppModule2.EXPECT().RegisterServices(cfg).Times(1)

	mm.RegisterServices(cfg)
}

func TestManager_InitGenesis(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule2 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mm := module.NewManager(mockAppModule1, mockAppModule2)
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.Modules))

	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	genesisData := map[string]json.RawMessage{"module1": json.RawMessage(`{"key": "value"}`)}

	// this should panic since the validator set is empty even after init genesis
	mockAppModule1.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Eq(cdc), gomock.Eq(genesisData["module1"])).Times(1).Return(nil)
	require.Panics(t, func() { mm.InitGenesis(ctx, cdc, genesisData) })

	// test panic
	genesisData = map[string]json.RawMessage{
		"module1": json.RawMessage(`{"key": "value"}`),
		"module2": json.RawMessage(`{"key": "value"}`),
	}
	mockAppModule1.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Eq(cdc), gomock.Eq(genesisData["module1"])).Times(1).Return([]abci.ValidatorUpdate{{}})
	mockAppModule2.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Eq(cdc), gomock.Eq(genesisData["module2"])).Times(1).Return([]abci.ValidatorUpdate{{}})
	require.Panics(t, func() { mm.InitGenesis(ctx, cdc, genesisData) })
}

func TestManager_ExportGenesis(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule2 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockCoreAppModule := MockCoreAppModule{}
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mm := module.NewManager(mockAppModule1, mockAppModule2, mockCoreAppModule)
	require.NotNil(t, mm)
	require.Equal(t, 3, len(mm.Modules))

	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	mockAppModule1.EXPECT().ExportGenesis(gomock.Eq(ctx), gomock.Eq(cdc)).AnyTimes().Return(json.RawMessage(`{"key1": "value1"}`))
	mockAppModule2.EXPECT().ExportGenesis(gomock.Eq(ctx), gomock.Eq(cdc)).AnyTimes().Return(json.RawMessage(`{"key2": "value2"}`))

	want := map[string]json.RawMessage{
		"module1": json.RawMessage(`{"key1": "value1"}`),
		"module2": json.RawMessage(`{"key2": "value2"}`),
		"MockCoreAppModule": json.RawMessage(`{
  "someField": "asd"
}`),
	}

	require.Equal(t, want, mm.ExportGenesis(ctx, cdc))
	require.Equal(t, want, mm.ExportGenesisForModules(ctx, cdc, []string{}))
	require.Equal(t, map[string]json.RawMessage{"module1": json.RawMessage(`{"key1": "value1"}`)}, mm.ExportGenesisForModules(ctx, cdc, []string{"module1"}))
	require.NotEqual(t, map[string]json.RawMessage{"module1": json.RawMessage(`{"key1": "value1"}`)}, mm.ExportGenesisForModules(ctx, cdc, []string{"module2"}))

	require.Panics(t, func() {
		mm.ExportGenesisForModules(ctx, cdc, []string{"module1", "modulefoo"})
	})
}

func TestManager_BeginBlock(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockBeginBlockAppModule(mockCtrl)
	mockAppModule2 := mock.NewMockBeginBlockAppModule(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mm := module.NewManager(mockAppModule1, mockAppModule2)
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.Modules))

	req := abci.RequestBeginBlock{Hash: []byte("test")}

	mockAppModule1.EXPECT().BeginBlock(gomock.Any(), gomock.Eq(req)).Times(1)
	mockAppModule2.EXPECT().BeginBlock(gomock.Any(), gomock.Eq(req)).Times(1)
	mm.BeginBlock(sdk.Context{}, req)
}

func TestManager_EndBlock(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockEndBlockAppModule(mockCtrl)
	mockAppModule2 := mock.NewMockEndBlockAppModule(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mm := module.NewManager(mockAppModule1, mockAppModule2)
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.Modules))

	req := abci.RequestEndBlock{Height: 10}

	mockAppModule1.EXPECT().EndBlock(gomock.Any(), gomock.Eq(req)).Times(1).Return([]abci.ValidatorUpdate{{}})
	mockAppModule2.EXPECT().EndBlock(gomock.Any(), gomock.Eq(req)).Times(1)
	ret := mm.EndBlock(sdk.Context{}, req)
	require.Equal(t, []abci.ValidatorUpdate{{}}, ret.ValidatorUpdates)

	// test panic
	mockAppModule1.EXPECT().EndBlock(gomock.Any(), gomock.Eq(req)).Times(1).Return([]abci.ValidatorUpdate{{}})
	mockAppModule2.EXPECT().EndBlock(gomock.Any(), gomock.Eq(req)).Times(1).Return([]abci.ValidatorUpdate{{}})
	require.Panics(t, func() { mm.EndBlock(sdk.Context{}, req) })
}

// Core API tests
func TestCoreAPIManager(t *testing.T) {
}

// MockCoreAppModule allows us to test functions like DefaultGenesis
type MockCoreAppModule struct{}

func (MockCoreAppModule) Name() string        { return "MockCoreAppModule" }
func (MockCoreAppModule) IsOnePerModuleType() {}
func (MockCoreAppModule) IsAppModule()        {}
func (MockCoreAppModule) DefaultGenesis(target appmodule.GenesisTarget) error {
	someFieldWriter, err := target("someField")
	if err != nil {
		return err
	}
	someFieldWriter.Write([]byte(`"asd"`))
	return someFieldWriter.Close()
}
func (MockCoreAppModule) ValidateGenesis(src appmodule.GenesisSource) error {
	rdr, err := src("someField")
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(rdr)
	if err != nil {
		return err
	}

	// this check will always fail, but it's just an example
	if string(data) != `"dummy validation"` {
		return errFoo
	}

	return nil
}
func (MockCoreAppModule) InitGenesis(context.Context, appmodule.GenesisSource) error { return nil }
func (MockCoreAppModule) ExportGenesis(ctx context.Context, target appmodule.GenesisTarget) error {
	wrt, err := target("someField")
	if err != nil {
		return err
	}
	wrt.Write([]byte(`"asd"`))
	return wrt.Close()
}
func (MockCoreAppModule) RegisterLegacyAminoCodec(*codec.LegacyAmino)                 {}
func (MockCoreAppModule) RegisterInterfaces(codectypes.InterfaceRegistry)             {}
func (MockCoreAppModule) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {}
func (MockCoreAppModule) GetTxCmd() *cobra.Command                                    { return nil }
func (MockCoreAppModule) GetQueryCmd() *cobra.Command                                 { return nil }

var (
	_ appmodule.AppModule   = MockCoreAppModule{}
	_ appmodule.HasGenesis  = MockCoreAppModule{}
	_ module.AppModuleBasic = MockCoreAppModule{}
)
