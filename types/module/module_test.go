package module_test

import (
	"encoding/json"
	"errors"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/golang/mock/gomock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
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

	wantDefaultGenesis := map[string]json.RawMessage{"mockAppModuleBasic1": json.RawMessage(``)}

	mockAppModuleBasic1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)

	mockAppModuleBasic1.EXPECT().Name().AnyTimes().Return("mockAppModuleBasic1")
	mockAppModuleBasic1.EXPECT().DefaultGenesis(gomock.Eq(cdc)).Times(1).Return(json.RawMessage(``))
	mockAppModuleBasic1.EXPECT().ValidateGenesis(gomock.Eq(cdc), gomock.Eq(nil), gomock.Eq(wantDefaultGenesis["mockAppModuleBasic1"])).Times(1).Return(errFoo)
	mockAppModuleBasic1.EXPECT().RegisterLegacyAminoCodec(gomock.Eq(legacyAmino)).Times(1)
	mockAppModuleBasic1.EXPECT().RegisterInterfaces(gomock.Eq(interfaceRegistry)).Times(1)
	mockAppModuleBasic1.EXPECT().GetTxCmd().Times(1).Return(nil)
	mockAppModuleBasic1.EXPECT().GetQueryCmd().Times(1).Return(nil)

	mm := module.NewBasicManager(mockAppModuleBasic1)
	require.Equal(t, mm["mockAppModuleBasic1"], mockAppModuleBasic1)

	mm.RegisterLegacyAminoCodec(legacyAmino)
	mm.RegisterInterfaces(interfaceRegistry)

	require.Equal(t, wantDefaultGenesis, mm.DefaultGenesis(cdc))

	var data map[string]string
	require.Equal(t, map[string]string(nil), data)

	require.True(t, errors.Is(errFoo, mm.ValidateGenesis(cdc, nil, wantDefaultGenesis)))

	mockCmd := &cobra.Command{Use: "root"}
	mm.AddTxCommands(mockCmd)

	mm.AddQueryCommands(mockCmd)

	// validate genesis returns nil
	require.Nil(t, module.NewBasicManager().ValidateGenesis(cdc, nil, wantDefaultGenesis))
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
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mm := module.NewManager(mockAppModule1, mockAppModule2)
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.Modules))

	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	mockAppModule1.EXPECT().ExportGenesis(gomock.Eq(ctx), gomock.Eq(cdc)).AnyTimes().Return(json.RawMessage(`{"key1": "value1"}`))
	mockAppModule2.EXPECT().ExportGenesis(gomock.Eq(ctx), gomock.Eq(cdc)).AnyTimes().Return(json.RawMessage(`{"key2": "value2"}`))

	want := map[string]json.RawMessage{
		"module1": json.RawMessage(`{"key1": "value1"}`),
		"module2": json.RawMessage(`{"key2": "value2"}`),
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
<<<<<<< HEAD
=======

// Core API exclusive tests
func TestCoreAPIManager(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	module1 := mock.NewMockCoreAppModule(mockCtrl)
	module2 := MockCoreAppModule{}
	mm := module.NewManagerFromMap(map[string]appmodule.AppModule{
		"module1": module1,
		"module2": module2,
	})

	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.Modules))
	require.Equal(t, module1, mm.Modules["module1"])
	require.Equal(t, module2, mm.Modules["module2"])
}

func TestCoreAPIManager_InitGenesis(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockCoreAppModule(mockCtrl)
	mm := module.NewManagerFromMap(map[string]appmodule.AppModule{"module1": mockAppModule1})
	require.NotNil(t, mm)
	require.Equal(t, 1, len(mm.Modules))

	ctx := sdk.NewContext(nil, cmtproto.Header{}, false, log.NewNopLogger())
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	genesisData := map[string]json.RawMessage{"module1": json.RawMessage(`{"key": "value"}`)}

	// this should panic since the validator set is empty even after init genesis
	mockAppModule1.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Any()).Times(1).Return(nil)
	_, err := mm.InitGenesis(ctx, cdc, genesisData)
	require.ErrorContains(t, err, "validator set is empty after InitGenesis, please ensure at least one validator is initialized with a delegation greater than or equal to the DefaultPowerReduction")

	// TODO: add happy path test. We are not returning any validator updates, this will come with the services.
	// REF: https://github.com/cosmos/cosmos-sdk/issues/14688
}

func TestCoreAPIManager_ExportGenesis(t *testing.T) {
	mockAppModule1 := MockCoreAppModule{}
	mockAppModule2 := MockCoreAppModule{}
	mm := module.NewManagerFromMap(map[string]appmodule.AppModule{
		"module1": mockAppModule1,
		"module2": mockAppModule2,
	})
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.Modules))

	ctx := sdk.NewContext(nil, cmtproto.Header{}, false, log.NewNopLogger())
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	want := map[string]json.RawMessage{
		"module1": json.RawMessage(`{
  "someField": "someKey"
}`),
		"module2": json.RawMessage(`{
  "someField": "someKey"
}`),
	}

	res, err := mm.ExportGenesis(ctx, cdc)
	require.NoError(t, err)
	require.Equal(t, want, res)

	res, err = mm.ExportGenesisForModules(ctx, cdc, []string{})
	require.NoError(t, err)
	require.Equal(t, want, res)

	res, err = mm.ExportGenesisForModules(ctx, cdc, []string{"module1"})
	require.NoError(t, err)
	require.Equal(t, map[string]json.RawMessage{"module1": want["module1"]}, res)

	res, err = mm.ExportGenesisForModules(ctx, cdc, []string{"module2"})
	require.NoError(t, err)
	require.NotEqual(t, map[string]json.RawMessage{"module1": want["module1"]}, res)

	_, err = mm.ExportGenesisForModules(ctx, cdc, []string{"module1", "modulefoo"})
	require.Error(t, err)
}

func TestCoreAPIManagerOrderSetters(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)
	mockAppModule1 := mock.NewMockCoreAppModule(mockCtrl)
	mockAppModule2 := mock.NewMockCoreAppModule(mockCtrl)
	mockAppModule3 := mock.NewMockCoreAppModule(mockCtrl)

	mm := module.NewManagerFromMap(
		map[string]appmodule.AppModule{
			"module1": mockAppModule1,
			"module2": mockAppModule2,
			"module3": mockAppModule3,
		})
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

	require.Equal(t, []string{"module1", "module2", "module3"}, mm.OrderPrepareCheckStaters)
	mm.SetOrderPrepareCheckStaters("module3", "module2", "module1")
	require.Equal(t, []string{"module3", "module2", "module1"}, mm.OrderPrepareCheckStaters)

	require.Equal(t, []string{"module1", "module2", "module3"}, mm.OrderPrecommiters)
	mm.SetOrderPrecommiters("module3", "module2", "module1")
	require.Equal(t, []string{"module3", "module2", "module1"}, mm.OrderPrecommiters)
}

func TestCoreAPIManager_RunMigrationBeginBlock(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockCoreAppModule(mockCtrl)
	mockAppModule2 := mock.NewMockUpgradeModule(mockCtrl)
	mm := module.NewManagerFromMap(map[string]appmodule.AppModule{
		"module1": mockAppModule1,
		"module2": mockAppModule2,
	})
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.Modules))

	mockAppModule1.EXPECT().BeginBlock(gomock.Any()).Times(0)
	mockAppModule2.EXPECT().BeginBlock(gomock.Any()).Times(1).Return(nil)
	success := mm.RunMigrationBeginBlock(sdk.Context{})
	require.Equal(t, true, success)

	// test false
	require.Equal(t, false, module.NewManager().RunMigrationBeginBlock(sdk.Context{}))

	// test panic
	mockAppModule2.EXPECT().BeginBlock(gomock.Any()).Times(1).Return(errors.New("some error"))
	success = mm.RunMigrationBeginBlock(sdk.Context{})
	require.Equal(t, false, success)
}

func TestCoreAPIManager_BeginBlock(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockCoreAppModule(mockCtrl)
	mockAppModule2 := mock.NewMockCoreAppModule(mockCtrl)
	mm := module.NewManagerFromMap(map[string]appmodule.AppModule{
		"module1": mockAppModule1,
		"module2": mockAppModule2,
	})
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.Modules))

	mockAppModule1.EXPECT().BeginBlock(gomock.Any()).Times(1).Return(nil)
	mockAppModule2.EXPECT().BeginBlock(gomock.Any()).Times(1).Return(nil)
	_, err := mm.BeginBlock(sdk.Context{})
	require.NoError(t, err)

	// test panic
	mockAppModule1.EXPECT().BeginBlock(gomock.Any()).Times(1).Return(errors.New("some error"))
	_, err = mm.BeginBlock(sdk.Context{})
	require.EqualError(t, err, "some error")
}

func TestCoreAPIManager_EndBlock(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockCoreAppModule(mockCtrl)
	mockAppModule2 := mock.NewMockCoreAppModule(mockCtrl)
	mm := module.NewManagerFromMap(map[string]appmodule.AppModule{
		"module1": mockAppModule1,
		"module2": mockAppModule2,
	})
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.Modules))

	mockAppModule1.EXPECT().EndBlock(gomock.Any()).Times(1).Return(nil)
	mockAppModule2.EXPECT().EndBlock(gomock.Any()).Times(1).Return(nil)
	res, err := mm.EndBlock(sdk.Context{})
	require.NoError(t, err)
	require.Len(t, res.ValidatorUpdates, 0)

	// test panic
	mockAppModule1.EXPECT().EndBlock(gomock.Any()).Times(1).Return(errors.New("some error"))
	_, err = mm.EndBlock(sdk.Context{})
	require.EqualError(t, err, "some error")
}

func TestManager_PrepareCheckState(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockCoreAppModule(mockCtrl)
	mockAppModule2 := mock.NewMockCoreAppModule(mockCtrl)
	mm := module.NewManagerFromMap(map[string]appmodule.AppModule{
		"module1": mockAppModule1,
		"module2": mockAppModule2,
	})
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.Modules))

	mockAppModule1.EXPECT().PrepareCheckState(gomock.Any()).Times(1).Return(nil)
	mockAppModule2.EXPECT().PrepareCheckState(gomock.Any()).Times(1).Return(nil)
	err := mm.PrepareCheckState(sdk.Context{})
	require.NoError(t, err)

	mockAppModule1.EXPECT().PrepareCheckState(gomock.Any()).Times(1).Return(errors.New("some error"))
	err = mm.PrepareCheckState(sdk.Context{})
	require.EqualError(t, err, "some error")
}

func TestManager_Precommit(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockCoreAppModule(mockCtrl)
	mockAppModule2 := mock.NewMockCoreAppModule(mockCtrl)
	mm := module.NewManagerFromMap(map[string]appmodule.AppModule{
		"module1": mockAppModule1,
		"module2": mockAppModule2,
	})
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.Modules))

	mockAppModule1.EXPECT().Precommit(gomock.Any()).Times(1).Return(nil)
	mockAppModule2.EXPECT().Precommit(gomock.Any()).Times(1).Return(nil)
	err := mm.Precommit(sdk.Context{})
	require.NoError(t, err)

	mockAppModule1.EXPECT().Precommit(gomock.Any()).Times(1).Return(errors.New("some error"))
	err = mm.Precommit(sdk.Context{})
	require.EqualError(t, err, "some error")
}

// MockCoreAppModule allows us to test functions like DefaultGenesis
type MockCoreAppModule struct{}

// RegisterServices implements appmodule.HasServices
func (MockCoreAppModule) RegisterServices(reg grpc.ServiceRegistrar) error {
	// Use Auth's service definitions as a placeholder
	authtypes.RegisterQueryServer(reg, &authtypes.UnimplementedQueryServer{})
	authtypes.RegisterMsgServer(reg, &authtypes.UnimplementedMsgServer{})
	return nil
}

func (MockCoreAppModule) IsOnePerModuleType() {}
func (MockCoreAppModule) IsAppModule()        {}
func (MockCoreAppModule) DefaultGenesis(target appmodule.GenesisTarget) error {
	someFieldWriter, err := target("someField")
	if err != nil {
		return err
	}
	_, err = someFieldWriter.Write([]byte(`"someKey"`))
	if err != nil {
		return err
	}
	return someFieldWriter.Close()
}

func (MockCoreAppModule) ValidateGenesis(src appmodule.GenesisSource) error {
	rdr, err := src("someField")
	if err != nil {
		return err
	}
	data, err := io.ReadAll(rdr)
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
	_, err = wrt.Write([]byte(`"someKey"`))
	if err != nil {
		return err
	}
	return wrt.Close()
}

var (
	_ appmodule.AppModule   = MockCoreAppModule{}
	_ appmodule.HasGenesis  = MockCoreAppModule{}
	_ appmodule.HasServices = MockCoreAppModule{}
)
>>>>>>> 0c1f6fc16 (fix: Add MigrationModuleManager to handle migration of upgrade module before other modules (#16583))
