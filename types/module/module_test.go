package module_test

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/tests/mocks"
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

	wantDefaultGenesis := map[string]json.RawMessage{
		"mockAppModuleBasic1": json.RawMessage(``),
		"module2":             json.RawMessage(`null`),
		"module3": json.RawMessage(`{
  "someField": "someValue"
}`),
	}

	mockAppModuleBasic1 := mocks.NewMockAppModuleBasic(mockCtrl)

	mockAppModuleBasic1.EXPECT().Name().AnyTimes().Return("mockAppModuleBasic1")
	mockAppModuleBasic1.EXPECT().DefaultGenesis(gomock.Eq(cdc)).Times(1).Return(json.RawMessage(``))
	mockAppModuleBasic1.EXPECT().ValidateGenesis(gomock.Eq(cdc), gomock.Eq(nil), gomock.Eq(wantDefaultGenesis["mockAppModuleBasic1"])).AnyTimes().Return(nil)
	mockAppModuleBasic1.EXPECT().RegisterRESTRoutes(gomock.Eq(client.Context{}), gomock.Eq(&mux.Router{})).Times(1)
	mockAppModuleBasic1.EXPECT().RegisterLegacyAminoCodec(gomock.Eq(legacyAmino)).Times(1)
	mockAppModuleBasic1.EXPECT().RegisterInterfaces(gomock.Eq(interfaceRegistry)).Times(1)
	mockAppModuleBasic1.EXPECT().GetTxCmd().Times(1).Return(nil)
	mockAppModuleBasic1.EXPECT().GetQueryCmd().Times(1).Return(nil)

	mockCoreAppModule2 := mocks.NewMockCoreAppModule(mockCtrl)
	mockCoreAppModule2.EXPECT().DefaultGenesis(gomock.Any()).Return(nil)
	mockCoreAppModule2.EXPECT().ValidateGenesis(gomock.Any()).Return(nil)

	mockCoreAppModule3 := MockCoreAppModule{}

	mm := module.NewBasicManager(
		mockAppModuleBasic1,
		module.UseCoreAPIModule("module2", mockCoreAppModule2),
		module.UseCoreAPIModule("module3", mockCoreAppModule3),
	)
	require.Equal(t, mm["mockAppModuleBasic1"], mockAppModuleBasic1)

	mm.RegisterLegacyAminoCodec(legacyAmino)
	mm.RegisterInterfaces(interfaceRegistry)

	require.Equal(t, wantDefaultGenesis, mm.DefaultGenesis(cdc))

	var data map[string]string
	require.Equal(t, map[string]string(nil), data)

	require.True(t, errors.Is(errFoo, mm.ValidateGenesis(cdc, nil, wantDefaultGenesis)))

	mm.RegisterRESTRoutes(client.Context{}, &mux.Router{})

	mockCmd := &cobra.Command{Use: "root"}
	mm.AddTxCommands(mockCmd)

	mm.AddQueryCommands(mockCmd)

	// validate genesis returns nil
	require.Nil(t, module.NewBasicManager().ValidateGenesis(cdc, nil, wantDefaultGenesis))
}

func TestGenesisOnlyAppModule(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockModule := mocks.NewMockAppModuleGenesis(mockCtrl)
	mockInvariantRegistry := mocks.NewMockInvariantRegistry(mockCtrl)
	goam := module.NewGenesisOnlyAppModule(mockModule)

	require.True(t, goam.Route().Empty())
	require.Empty(t, goam.QuerierRoute())
	require.Nil(t, goam.LegacyQuerierHandler(nil))

	// no-op
	goam.RegisterInvariants(mockInvariantRegistry)
}

func TestManagerOrderSetters(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)
	mockAppModule1 := mocks.NewMockAppModule(mockCtrl)
	mockAppModule2 := mocks.NewMockAppModule(mockCtrl)
	mockAppModule3 := mocks.NewMockCoreAppModule(mockCtrl)

	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mm := module.NewManager(mockAppModule1, mockAppModule2, module.UseCoreAPIModule("module3", mockAppModule3))
	require.NotNil(t, mm)
	require.Equal(t, 3, len(mm.Modules))

	require.Equal(t, []string{"module1", "module2", "module3"}, mm.OrderInitGenesis)
	mm.SetOrderInitGenesis("module2", "module1", "module3")
	require.Equal(t, []string{"module2", "module1", "module3"}, mm.OrderInitGenesis)

	require.Equal(t, []string{"module1", "module2", "module3"}, mm.OrderExportGenesis)
	mm.SetOrderExportGenesis("module2", "module1", "module3")
	require.Equal(t, []string{"module2", "module1", "module3"}, mm.OrderExportGenesis)

	require.Equal(t, []string{"module1", "module2", "module3"}, mm.OrderBeginBlockers)
	mm.SetOrderBeginBlockers("module2", "module1", "module3")
	require.Equal(t, []string{"module2", "module1", "module3"}, mm.OrderBeginBlockers)

	require.Equal(t, []string{"module1", "module2", "module3"}, mm.OrderEndBlockers)
	mm.SetOrderEndBlockers("module2", "module1", "module3")
	require.Equal(t, []string{"module2", "module1", "module3"}, mm.OrderEndBlockers)
}

func TestManager_RegisterInvariants(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mocks.NewMockAppModule(mockCtrl)
	mockAppModule2 := mocks.NewMockAppModule(mockCtrl)
	mockAppModule3 := mocks.NewMockCoreAppModule(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mm := module.NewManager(mockAppModule1, mockAppModule2, module.UseCoreAPIModule("module3", mockAppModule3))
	require.NotNil(t, mm)
	require.Equal(t, 3, len(mm.Modules))

	// test RegisterInvariants
	mockInvariantRegistry := mocks.NewMockInvariantRegistry(mockCtrl)
	mockAppModule1.EXPECT().RegisterInvariants(gomock.Eq(mockInvariantRegistry)).Times(1)
	mockAppModule2.EXPECT().RegisterInvariants(gomock.Eq(mockInvariantRegistry)).Times(1)
	mm.RegisterInvariants(mockInvariantRegistry)
}

func TestManager_RegisterRoutes(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mocks.NewMockAppModule(mockCtrl)
	mockAppModule2 := mocks.NewMockAppModule(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mockAppModule3 := mocks.NewMockCoreAppModule(mockCtrl)
	mm := module.NewManager(mockAppModule1, mockAppModule2, module.UseCoreAPIModule("module3", mockAppModule3))
	require.NotNil(t, mm)
	require.Equal(t, 3, len(mm.Modules))

	router := mocks.NewMockRouter(mockCtrl)
	noopHandler := sdk.Handler(func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) { return nil, nil })
	route1 := sdk.NewRoute("route1", noopHandler)
	route2 := sdk.NewRoute("", noopHandler)
	mockAppModule1.EXPECT().Route().Times(1).Return(route1)
	mockAppModule2.EXPECT().Route().Times(1).Return(route2)
	router.EXPECT().AddRoute(gomock.Any()).Times(1) // Use of Any due to limitations to compare Functions as the sdk.Handler

	queryRouter := mocks.NewMockQueryRouter(mockCtrl)
	mockAppModule1.EXPECT().QuerierRoute().Times(1).Return("querierRoute1")
	mockAppModule2.EXPECT().QuerierRoute().Times(1).Return("")
	handler3 := sdk.Querier(nil)
	amino := codec.NewLegacyAmino()
	mockAppModule1.EXPECT().LegacyQuerierHandler(amino).Times(1).Return(handler3)
	queryRouter.EXPECT().AddRoute(gomock.Eq("querierRoute1"), gomock.Eq(handler3)).Times(1)

	mm.RegisterRoutes(router, queryRouter, amino)
}

func TestManager_RegisterQueryServices(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mocks.NewMockAppModule(mockCtrl)
	mockAppModule2 := mocks.NewMockAppModule(mockCtrl)
	mockAppModule3 := mocks.NewMockCoreAppModule(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mm := module.NewManager(mockAppModule1, mockAppModule2, module.UseCoreAPIModule("module3", mockAppModule3))
	require.NotNil(t, mm)
	require.Equal(t, 3, len(mm.Modules))

	msgRouter := mocks.NewMockServer(mockCtrl)
	queryRouter := mocks.NewMockServer(mockCtrl)
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

	mockAppModule1 := mocks.NewMockAppModule(mockCtrl)
	mockAppModule2 := mocks.NewMockAppModule(mockCtrl)
	mockAppModule3 := mocks.NewMockCoreAppModule(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mm := module.NewManager(mockAppModule1, mockAppModule2, module.UseCoreAPIModule("module3", mockAppModule3))
	require.NotNil(t, mm)
	require.Equal(t, 3, len(mm.Modules))

	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	genesisData := map[string]json.RawMessage{"module1": json.RawMessage(`{"key": "value"}`)}

	mockAppModule1.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Eq(cdc), gomock.Eq(genesisData["module1"])).Times(1).Return(nil)
	require.Equal(t, abci.ResponseInitChain{Validators: []abci.ValidatorUpdate(nil)}, mm.InitGenesis(ctx, cdc, genesisData))

	// test panic
	genesisData = map[string]json.RawMessage{
		"module1": json.RawMessage(`{"key": "value"}`),
		"module2": json.RawMessage(`{"key": "value"}`),
		"module3": json.RawMessage(`{
  "someField": "someValue"
}`),
	}
	mockAppModule1.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Eq(cdc), gomock.Eq(genesisData["module1"])).Times(1).Return([]abci.ValidatorUpdate{{}})
	mockAppModule2.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Eq(cdc), gomock.Eq(genesisData["module2"])).Times(1).Return([]abci.ValidatorUpdate{{}})
	require.Panics(t, func() {
		mm.InitGenesis(ctx, cdc, genesisData)
	})

	// happy path, InitGenesis gets called for all modules
	mockAppModule1.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Eq(cdc), gomock.Eq(genesisData["module1"])).Times(1).Return([]abci.ValidatorUpdate{{}})
	mockAppModule2.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Eq(cdc), gomock.Eq(genesisData["module2"])).Times(1).Return(nil)
	mockAppModule3.EXPECT().InitGenesis(gomock.Any(), gomock.Any()).Times(1).Return(nil)
	mm.InitGenesis(ctx, cdc, genesisData)
}

func TestManager_ExportGenesis(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mocks.NewMockAppModule(mockCtrl)
	mockAppModule2 := mocks.NewMockAppModule(mockCtrl)
	mockAppModule3 := mocks.NewMockCoreAppModule(mockCtrl)
	mockAppModule4 := MockCoreAppModule{}
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mm := module.NewManager(
		mockAppModule1,
		mockAppModule2,
		module.UseCoreAPIModule("module3", mockAppModule3),
		module.UseCoreAPIModule("module4", mockAppModule4),
	)
	require.NotNil(t, mm)
	require.Equal(t, 4, len(mm.Modules))

	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	mockAppModule1.EXPECT().ExportGenesis(gomock.Eq(ctx), gomock.Eq(cdc)).Times(1).Return(json.RawMessage(`{"key1": "value1"}`))
	mockAppModule2.EXPECT().ExportGenesis(gomock.Eq(ctx), gomock.Eq(cdc)).Times(1).Return(json.RawMessage(`{"key2": "value2"}`))
	mockAppModule3.EXPECT().ExportGenesis(gomock.Any(), gomock.Any()).Times(1).Return(nil)

	want := map[string]json.RawMessage{
		"module1": json.RawMessage(`{"key1": "value1"}`),
		"module2": json.RawMessage(`{"key2": "value2"}`),
		"module3": json.RawMessage(`null`),
		"module4": json.RawMessage(`{
  "someField": "someValue"
}`),
	}
	require.Equal(t, want, mm.ExportGenesis(ctx, cdc))
}

func TestManager_BeginBlock(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mocks.NewMockBeginBlockAppModule(mockCtrl)
	mockAppModule2 := mocks.NewMockBeginBlockAppModule(mockCtrl)
	mockAppModule3 := mocks.NewMockCoreAppModule(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mm := module.NewManager(
		mockAppModule1,
		mockAppModule2,
		module.UseCoreAPIModule("module3", mockAppModule3),
	)
	require.NotNil(t, mm)
	require.Equal(t, 3, len(mm.Modules))

	req := abci.RequestBeginBlock{Hash: []byte("test")}
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())

	mockAppModule1.EXPECT().BeginBlock(gomock.Any(), gomock.Eq(req)).Times(1)
	mockAppModule2.EXPECT().BeginBlock(gomock.Any(), gomock.Eq(req)).Times(1)
	mockAppModule3.EXPECT().BeginBlock(gomock.Any()).Times(1)
	mm.BeginBlock(ctx, req)
}

func TestManager_EndBlock(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mocks.NewMockEndBlockAppModule(mockCtrl)
	mockAppModule2 := mocks.NewMockEndBlockAppModule(mockCtrl)
	mockAppModule3 := mocks.NewMockCoreAppModule(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mm := module.NewManager(
		mockAppModule1,
		mockAppModule2,
		module.UseCoreAPIModule("module3", mockAppModule3),
	)
	require.NotNil(t, mm)
	require.Equal(t, 3, len(mm.Modules))

	req := abci.RequestEndBlock{Height: 10}
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())

	mockAppModule1.EXPECT().EndBlock(gomock.Any(), gomock.Eq(req)).Times(1).Return([]abci.ValidatorUpdate{{}})
	mockAppModule2.EXPECT().EndBlock(gomock.Any(), gomock.Eq(req)).Times(1)
	mockAppModule3.EXPECT().EndBlock(gomock.Any()).Times(1)
	ret := mm.EndBlock(ctx, req)
	require.Equal(t, []abci.ValidatorUpdate{{}}, ret.ValidatorUpdates)

	// test panic
	mockAppModule1.EXPECT().EndBlock(gomock.Any(), gomock.Eq(req)).Times(1).Return([]abci.ValidatorUpdate{{}})
	mockAppModule2.EXPECT().EndBlock(gomock.Any(), gomock.Eq(req)).Times(1).Return([]abci.ValidatorUpdate{{}})
	require.Panics(t, func() { mm.EndBlock(ctx, req) })
}

// MockCoreAppModule allows us to test functions like DefaultGenesis
type MockCoreAppModule struct{}

func (MockCoreAppModule) IsOnePerModuleType() {}
func (MockCoreAppModule) IsAppModule()        {}
func (MockCoreAppModule) DefaultGenesis(target appmodule.GenesisTarget) error {
	someFieldWriter, err := target("someField")
	if err != nil {
		return err
	}
	someFieldWriter.Write([]byte(`"someValue"`))
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
	wrt.Write([]byte(`"someValue"`))
	return wrt.Close()
}

var (
	_ appmodule.AppModule  = MockCoreAppModule{}
	_ appmodule.HasGenesis = MockCoreAppModule{}
)
