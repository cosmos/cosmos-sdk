package module_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v2"
	abci "github.com/cometbft/cometbft/v2/abci/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var errFoo = errors.New("dummy")

func (MockCoreAppModule) GetQueryCmd() *cobra.Command {
	return &cobra.Command{
		Use: "foo",
	}
}

func TestBasicManager(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)
	legacyAmino := codec.NewLegacyAmino()
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Test with a legacy module, a mock core module that doesn't return anything,
	// and a core module defined in this file
	expDefaultGenesis := map[string]json.RawMessage{
		"mockAppModuleBasic1": json.RawMessage(``),
		"mockCoreAppModule2":  json.RawMessage(`null`),
		"mockCoreAppModule3": json.RawMessage(`{
  "someField": "someKey"
}`),
	}

	// legacy module
	mockAppModuleBasic1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModuleBasic1.EXPECT().Name().AnyTimes().Return("mockAppModuleBasic1")
	mockAppModuleBasic1.EXPECT().DefaultGenesis(gomock.Eq(cdc)).Times(1).Return(json.RawMessage(``))
	// Allow ValidateGenesis to be called any times because other module can fail before this one is called.
	mockAppModuleBasic1.EXPECT().ValidateGenesis(gomock.Eq(cdc), gomock.Eq(nil), gomock.Eq(expDefaultGenesis["mockAppModuleBasic1"])).AnyTimes().Return(nil)
	mockAppModuleBasic1.EXPECT().RegisterLegacyAminoCodec(gomock.Eq(legacyAmino)).Times(1)
	mockAppModuleBasic1.EXPECT().RegisterInterfaces(gomock.Eq(interfaceRegistry)).Times(1)

	// mock core API module
	mockCoreAppModule2 := mock.NewMockCoreAppModule(mockCtrl)
	mockCoreAppModule2.EXPECT().DefaultGenesis(gomock.Any()).AnyTimes().Return(nil)
	mockCoreAppModule2.EXPECT().ValidateGenesis(gomock.Any()).AnyTimes().Return(nil)
	mockAppModule2 := module.CoreAppModuleBasicAdaptor("mockCoreAppModule2", mockCoreAppModule2)

	// mock core API module (but all methods are implemented)
	mockCoreAppModule3 := module.CoreAppModuleBasicAdaptor("mockCoreAppModule3", MockCoreAppModule{})

	mm := module.NewBasicManager(mockAppModuleBasic1, mockAppModule2, mockCoreAppModule3)

	require.Equal(t, mockAppModuleBasic1, mm["mockAppModuleBasic1"])
	require.Equal(t, mockAppModule2, mm["mockCoreAppModule2"])
	require.Equal(t, mockCoreAppModule3, mm["mockCoreAppModule3"])

	mm.RegisterLegacyAminoCodec(legacyAmino)
	mm.RegisterInterfaces(interfaceRegistry)

	require.Equal(t, expDefaultGenesis, mm.DefaultGenesis(cdc))

	var data map[string]string
	require.Equal(t, map[string]string(nil), data)

	require.ErrorIs(t, mm.ValidateGenesis(cdc, nil, expDefaultGenesis), errFoo)

	mockCmd := &cobra.Command{Use: "root"}
	mm.AddTxCommands(mockCmd)
	mm.AddQueryCommands(mockCmd)
	require.Equal(t, 1, len(mockCmd.Commands()))

	// validate genesis returns nil
	require.Nil(t, module.NewBasicManager().ValidateGenesis(cdc, nil, expDefaultGenesis))
}

func TestAssertNoForgottenModules(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)
	mockAppModule1 := mock.NewMockHasABCIEndBlock(mockCtrl)
	mockAppModule3 := mock.NewMockCoreAppModule(mockCtrl)

	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mm := module.NewManager(
		mockAppModule1,
		module.CoreAppModuleBasicAdaptor("module3", mockAppModule3),
	)
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.Modules))

	require.Equal(t, []string{"module1", "module3"}, mm.OrderInitGenesis)
	require.PanicsWithValue(t, "all modules must be defined when setting SetOrderInitGenesis, missing: [module3]", func() {
		mm.SetOrderInitGenesis("module1")
	})

	require.Equal(t, []string{"module1", "module3"}, mm.OrderExportGenesis)
	require.PanicsWithValue(t, "all modules must be defined when setting SetOrderExportGenesis, missing: [module3]", func() {
		mm.SetOrderExportGenesis("module1")
	})

	require.Equal(t, []string{"module1", "module3"}, mm.OrderEndBlockers)
	require.PanicsWithValue(t, "all modules must be defined when setting SetOrderEndBlockers, missing: [module1]", func() {
		mm.SetOrderEndBlockers("module3")
	})
}

func TestManagerOrderSetters(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)
	mockAppModule1 := mock.NewMockAppModule(mockCtrl)
	mockAppModule2 := mock.NewMockAppModule(mockCtrl)
	mockAppModule3 := mock.NewMockCoreAppModule(mockCtrl)

	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mm := module.NewManager(mockAppModule1, mockAppModule2, module.CoreAppModuleBasicAdaptor("module3", mockAppModule3))
	require.NotNil(t, mm)
	require.Equal(t, 3, len(mm.Modules))

	require.Equal(t, []string{"module1", "module2", "module3"}, mm.OrderInitGenesis)
	mm.SetOrderInitGenesis("module2", "module1", "module3")
	require.Equal(t, []string{"module2", "module1", "module3"}, mm.OrderInitGenesis)

	require.Equal(t, []string{"module1", "module2", "module3"}, mm.OrderExportGenesis)
	mm.SetOrderExportGenesis("module2", "module1", "module3")
	require.Equal(t, []string{"module2", "module1", "module3"}, mm.OrderExportGenesis)

	require.Equal(t, []string{}, mm.OrderPreBlockers)
	mm.SetOrderPreBlockers("module2", "module1", "module3")
	require.Equal(t, []string{"module2", "module1", "module3"}, mm.OrderPreBlockers)

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

func TestManager_RegisterQueryServices(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule2 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule3 := MockCoreAppModule{}
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	// TODO: This is not working for Core API modules yet
	mm := module.NewManager(mockAppModule1, mockAppModule2, module.CoreAppModuleBasicAdaptor("mockAppModule3", mockAppModule3))
	require.NotNil(t, mm)
	require.Equal(t, 3, len(mm.Modules))

	msgRouter := mock.NewMockServer(mockCtrl)
	msgRouter.EXPECT().RegisterService(gomock.Any(), gomock.Any()).Times(1)
	queryRouter := mock.NewMockServer(mockCtrl)
	queryRouter.EXPECT().RegisterService(gomock.Any(), gomock.Any()).Times(1)

	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	cfg := module.NewConfigurator(cdc, msgRouter, queryRouter)
	mockAppModule1.EXPECT().RegisterServices(cfg).Times(1)
	mockAppModule2.EXPECT().RegisterServices(cfg).Times(1)

	require.NotPanics(t, func() { _ = mm.RegisterServices(cfg) })
}

func TestManager_InitGenesis(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule2 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule3 := mock.NewMockCoreAppModule(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(4).Return("module2")
	mm := module.NewManager(mockAppModule1, mockAppModule2, module.CoreAppModuleBasicAdaptor("module3", mockAppModule3))
	require.NotNil(t, mm)
	require.Equal(t, 3, len(mm.Modules))

	ctx := sdk.NewContext(nil, cmtproto.Header{}, false, log.NewNopLogger())
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	genesisData := map[string]json.RawMessage{"module1": json.RawMessage(`{"key": "value"}`)}

	// this should panic since the validator set is empty even after init genesis
	mockAppModule1.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Eq(cdc), gomock.Eq(genesisData["module1"])).Times(1)
	_, err := mm.InitGenesis(ctx, cdc, genesisData)
	require.ErrorContains(t, err, "validator set is empty after InitGenesis")

	// test panic
	genesisData = map[string]json.RawMessage{
		"module1": json.RawMessage(`{"key": "value"}`),
		"module2": json.RawMessage(`{"key": "value"}`),
		"module3": json.RawMessage(`{"key": "value"}`),
	}

	mockAppModuleABCI1 := mock.NewMockAppModuleWithAllExtensionsABCI(mockCtrl)
	mockAppModuleABCI2 := mock.NewMockAppModuleWithAllExtensionsABCI(mockCtrl)
	mockAppModuleABCI1.EXPECT().Name().Times(4).Return("module1")
	mockAppModuleABCI2.EXPECT().Name().Times(2).Return("module2")
	mmABCI := module.NewManager(mockAppModuleABCI1, mockAppModuleABCI2)
	// panic because more than one module returns validator set updates
	mockAppModuleABCI1.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Eq(cdc), gomock.Eq(genesisData["module1"])).Times(1).Return([]abci.ValidatorUpdate{{}})
	mockAppModuleABCI2.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Eq(cdc), gomock.Eq(genesisData["module2"])).Times(1).Return([]abci.ValidatorUpdate{{}})
	_, err = mmABCI.InitGenesis(ctx, cdc, genesisData)
	require.ErrorContains(t, err, "validator InitGenesis updates already set by a previous module")

	// happy path

	mm2 := module.NewManager(mockAppModuleABCI1, mockAppModule2, module.CoreAppModuleBasicAdaptor("module3", mockAppModule3))
	mockAppModuleABCI1.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Eq(cdc), gomock.Eq(genesisData["module1"])).Times(1).Return([]abci.ValidatorUpdate{{}})
	mockAppModule2.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Eq(cdc), gomock.Eq(genesisData["module2"])).Times(1)
	mockAppModule3.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Any()).Times(1).Return(nil)
	_, err = mm2.InitGenesis(ctx, cdc, genesisData)
	require.NoError(t, err)
}

func TestManager_ExportGenesis(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule2 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockCoreAppModule := MockCoreAppModule{}
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mm := module.NewManager(mockAppModule1, mockAppModule2, module.CoreAppModuleBasicAdaptor("mockCoreAppModule", mockCoreAppModule))
	require.NotNil(t, mm)
	require.Equal(t, 3, len(mm.Modules))

	ctx := sdk.NewContext(nil, cmtproto.Header{}, false, log.NewNopLogger())
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	mockAppModule1.EXPECT().ExportGenesis(gomock.Eq(ctx), gomock.Eq(cdc)).AnyTimes().Return(json.RawMessage(`{"key1": "value1"}`))
	mockAppModule2.EXPECT().ExportGenesis(gomock.Eq(ctx), gomock.Eq(cdc)).AnyTimes().Return(json.RawMessage(`{"key2": "value2"}`))

	want := map[string]json.RawMessage{
		"module1": json.RawMessage(`{"key1": "value1"}`),
		"module2": json.RawMessage(`{"key2": "value2"}`),
		"mockCoreAppModule": json.RawMessage(`{
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
	require.Equal(t, map[string]json.RawMessage{"module1": json.RawMessage(`{"key1": "value1"}`)}, res)

	res, err = mm.ExportGenesisForModules(ctx, cdc, []string{"module2"})
	require.NoError(t, err)
	require.NotEqual(t, map[string]json.RawMessage{"module1": json.RawMessage(`{"key1": "value1"}`)}, res)

	_, err = mm.ExportGenesisForModules(ctx, cdc, []string{"module1", "modulefoo"})
	require.Error(t, err)
}

func TestManager_EndBlock(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockHasABCIEndBlock(mockCtrl)
	mockAppModule2 := mock.NewMockHasABCIEndBlock(mockCtrl)
	mockAppModule3 := mock.NewMockAppModule(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mockAppModule3.EXPECT().Name().Times(2).Return("module3")
	mm := module.NewManager(mockAppModule1, mockAppModule2, mockAppModule3)
	require.NotNil(t, mm)
	require.Equal(t, 3, len(mm.Modules))

	mockAppModule1.EXPECT().EndBlock(gomock.Any()).Times(1).Return([]abci.ValidatorUpdate{{}}, nil)
	mockAppModule2.EXPECT().EndBlock(gomock.Any()).Times(1)
	ret, err := mm.EndBlock(sdk.Context{})
	require.NoError(t, err)
	require.Equal(t, []abci.ValidatorUpdate{{}}, ret.ValidatorUpdates)

	// test panic
	mockAppModule1.EXPECT().EndBlock(gomock.Any()).Times(1).Return([]abci.ValidatorUpdate{{}}, nil)
	mockAppModule2.EXPECT().EndBlock(gomock.Any()).Times(1).Return([]abci.ValidatorUpdate{{}}, nil)
	_, err = mm.EndBlock(sdk.Context{})
	require.Error(t, err)
}

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

	require.Equal(t, []string{}, mm.OrderPreBlockers)
	mm.SetOrderPreBlockers("module2", "module1", "module3")
	require.Equal(t, []string{"module2", "module1", "module3"}, mm.OrderPreBlockers)

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

func TestCoreAPIManager_PreBlock(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAppModule1 := mock.NewMockCoreAppModuleWithPreBlock(mockCtrl)
	mm := module.NewManagerFromMap(map[string]appmodule.AppModule{
		"module1": mockAppModule1,
		"module2": mock.NewMockCoreAppModule(mockCtrl),
	})
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.Modules))
	require.Equal(t, 1, len(mm.OrderPreBlockers))

	mockAppModule1.EXPECT().PreBlock(gomock.Any()).Times(1).Return(&sdk.ResponsePreBlock{
		ConsensusParamsChanged: true,
	}, nil)
	res, err := mm.PreBlock(sdk.Context{})
	require.NoError(t, err)
	require.True(t, res.ConsensusParamsChanged)

	// test false
	mockAppModule1.EXPECT().PreBlock(gomock.Any()).Times(1).Return(&sdk.ResponsePreBlock{
		ConsensusParamsChanged: false,
	}, nil)
	res, err = mm.PreBlock(sdk.Context{})
	require.NoError(t, err)
	require.False(t, res.ConsensusParamsChanged)

	// test error
	mockAppModule1.EXPECT().PreBlock(gomock.Any()).Times(1).Return(nil, errors.New("some error"))
	_, err = mm.PreBlock(sdk.Context{})
	require.EqualError(t, err, "some error")
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
	_, _ = someFieldWriter.Write([]byte(`"someKey"`))
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
	_, _ = wrt.Write([]byte(`"someKey"`))
	return wrt.Close()
}

var (
	_ appmodule.AppModule   = MockCoreAppModule{}
	_ appmodule.HasGenesis  = MockCoreAppModule{}
	_ appmodule.HasServices = MockCoreAppModule{}
)
