package module_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

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

	mockAppModuleBasic1 := mock.NewMockAppModuleBasic(mockCtrl)

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

	mockAppModule1 := mock.NewMockAppModule(mockCtrl)
	mockAppModule2 := mock.NewMockAppModule(mockCtrl)
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

	mockAppModule1 := mock.NewMockAppModule(mockCtrl)
	mockAppModule2 := mock.NewMockAppModule(mockCtrl)
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

	mockAppModule1 := mock.NewMockAppModule(mockCtrl)
	mockAppModule2 := mock.NewMockAppModule(mockCtrl)
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

	mockAppModule1 := mock.NewMockAppModule(mockCtrl)
	mockAppModule2 := mock.NewMockAppModule(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return("module1")
	mockAppModule2.EXPECT().Name().Times(2).Return("module2")
	mm := module.NewManager(mockAppModule1, mockAppModule2)
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.Modules))

	ctx := sdk.Context{}
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	mockAppModule1.EXPECT().ExportGenesis(gomock.Eq(ctx), gomock.Eq(cdc)).AnyTimes().Return(json.RawMessage(`{"key1": "value1"}`))
	mockAppModule2.EXPECT().ExportGenesis(gomock.Eq(ctx), gomock.Eq(cdc)).AnyTimes().Return(json.RawMessage(`{"key2": "value2"}`))

	want := map[string]json.RawMessage{
		"module1": json.RawMessage(`{"key1": "value1"}`),
		"module2": json.RawMessage(`{"key2": "value2"}`),
	}

	actual, err := mm.ExportGenesis(ctx, cdc)
	require.NoError(t, err)
	require.Equal(t, want, actual)

	actual, err = mm.ExportGenesisForModules(ctx, cdc, []string{})
	require.NoError(t, err)
	require.Equal(t, want, actual)

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

func TestModule_CreateExportFile(t *testing.T) {
	tmp := t.TempDir()
	mod := "test"
	index := 0

	f1, err := module.CreateExportFile(tmp, mod, index)
	require.NoError(t, err)
	defer f1.Close()

	fname := filepath.Join(filepath.Clean(tmp), fmt.Sprintf("genesis_%s_%d.bin", mod, index))
	require.Equal(t, fname, f1.Name())

	n, err := f1.WriteString("123")
	require.NoError(t, err)
	require.Equal(t, len("123"), n)

	// if we create the same export file again, the original will be truncated.
	f2, err := module.CreateExportFile(tmp, mod, index)
	require.NoError(t, err)
	require.Equal(t, f1.Name(), f2.Name())
	defer f2.Close()

	fs, err := f2.Stat()
	require.NoError(t, err)
	require.Equal(t, int64(0), fs.Size())
}

func TestModule_OpenModuleStateFile(t *testing.T) {
	tmp := t.TempDir()
	mod := "test"
	index := 0

	fp1, err := module.CreateExportFile(tmp, mod, index)
	require.NoError(t, err)
	defer fp1.Close()

	fp2, err := module.OpenModuleStateFile(tmp, mod, index)
	require.NoError(t, err)
	defer fp2.Close()

	fp1Stat, err := fp1.Stat()
	require.NoError(t, err)

	fp2Stat, err := fp2.Stat()
	require.NoError(t, err)

	require.Equal(t, fp1Stat, fp2Stat)

	// should failed to file request file
	_, err = module.OpenModuleStateFile(tmp, mod, index+1)
	require.ErrorContains(t, err, "failed to open file")
}

func TestManager_FileWrite(t *testing.T) {
	tmp := t.TempDir()
	mod := "test"

	// write empty state to file, will still create a file
	err := module.FileWrite(tmp, mod, []byte{})
	require.NoError(t, err)

	fp, err := module.OpenModuleStateFile(tmp, mod, 0)
	require.NoError(t, err)
	defer fp.Close()

	fs, err := fp.Stat()
	require.NoError(t, err)
	require.Equal(t, int64(0), fs.Size())

	// write bytes with maximum state chunk size, should only write 1 file
	bz := make([]byte, module.StateChunkSize)
	err = module.FileWrite(tmp, mod, bz)
	require.NoError(t, err)

	fp0, err := module.OpenModuleStateFile(tmp, mod, 0)
	require.NoError(t, err)
	defer fp0.Close()

	var buf bytes.Buffer
	n, err := buf.ReadFrom(fp0)
	require.NoError(t, err)
	require.Equal(t, int64(module.StateChunkSize), n)
	require.True(t, bytes.Equal(bz, buf.Bytes()))

	// write bytes larger than maximum state chunk size, should create multiple files
	bz = append(bz, []byte{1}...)
	err = module.FileWrite(tmp, mod, bz)
	require.NoError(t, err)

	// open the first file, read the content, and verify
	fp0, err = module.OpenModuleStateFile(tmp, mod, 0)
	require.NoError(t, err)
	defer fp0.Close()

	buf.Reset()
	n, err = buf.ReadFrom(fp0)
	require.NoError(t, err)
	require.Equal(t, int64(module.StateChunkSize), n)
	require.True(t, bytes.Equal(bz[:module.StateChunkSize], buf.Bytes()))

	// open the second file, read the content, and verify
	fp1, err := module.OpenModuleStateFile(tmp, mod, 1)
	require.NoError(t, err)
	defer fp1.Close()

	buf.Reset()
	n, err = buf.ReadFrom(fp1)
	require.NoError(t, err)
	require.Equal(t, int64(1), n)
	require.True(t, bytes.Equal(bz[module.StateChunkSize:], buf.Bytes()))
}

func TestManager_FileRead(t *testing.T) {
	tmp := t.TempDir()
	mod := "test"
	bz := make([]byte, module.StateChunkSize+1)
	bz[module.StateChunkSize] = byte(1)

	err := module.FileWrite(tmp, mod, bz)
	require.NoError(t, err)

	bzRead, err := module.FileRead(tmp, mod)
	require.NoError(t, err)
	require.True(t, bytes.Equal(bz, bzRead))
}

func TestManager_InitGenesisWithPath(t *testing.T) {
	tmp := t.TempDir()
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mod1 := "module1"
	mod2 := "module2"
	mockAppModule1 := mock.NewMockAppModule(mockCtrl)
	mockAppModule2 := mock.NewMockAppModule(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return(mod1)
	mockAppModule2.EXPECT().Name().Times(2).Return(mod2)
	mm := module.NewManager(mockAppModule1, mockAppModule2)
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.Modules))

	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	appGenesisState := map[string]json.RawMessage{
		mod1: json.RawMessage(`{"key": "value1"}`),
		mod2: json.RawMessage(`{"key": "value2"}`),
	}
	vs := []abci.ValidatorUpdate{{}}

	mockAppModule1.EXPECT().InitGenesis(
		gomock.Eq(ctx), gomock.Eq(cdc), gomock.Eq(appGenesisState[mod1])).Times(1).Return(vs)
	mockAppModule2.EXPECT().InitGenesis(
		gomock.Eq(ctx), gomock.Eq(cdc), gomock.Eq(appGenesisState[mod2])).Times(1).Return([]abci.ValidatorUpdate{})

	// we assume the genesis state has been exported to the module folders
	err := module.FileWrite(filepath.Join(tmp, mod1), mod1, appGenesisState[mod1])
	require.NoError(t, err)

	err = module.FileWrite(filepath.Join(tmp, mod2), mod2, appGenesisState[mod2])
	require.NoError(t, err)

	// set the file import path
	mm.SetGenesisPath(tmp)
	var res abci.ResponseInitChain
	require.NotPanics(t, func() {
		res = mm.InitGenesis(ctx, cdc, nil)
	})

	// check the final import status
	require.Equal(t, res, abci.ResponseInitChain{Validators: vs})
}

func TestManager_ExportGenesisWithPath(t *testing.T) {
	tmp := t.TempDir()
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mod1 := "module1"
	mod2 := "module2"
	mockAppModule1 := mock.NewMockAppModule(mockCtrl)
	mockAppModule2 := mock.NewMockAppModule(mockCtrl)
	mockAppModule1.EXPECT().Name().Times(2).Return(mod1)
	mockAppModule2.EXPECT().Name().Times(2).Return(mod2)
	mm := module.NewManager(mockAppModule1, mockAppModule2)
	require.NotNil(t, mm)
	require.Equal(t, 2, len(mm.Modules))

	ctx := sdk.Context{}
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	appGenesisState := map[string]json.RawMessage{
		mod1: json.RawMessage(`{"key": "value1"}`),
		mod2: json.RawMessage(`{"key": "value2"}`),
	}

	// set the export state in each mock modules
	mockAppModule1.EXPECT().ExportGenesis(gomock.Eq(ctx), gomock.Eq(cdc)).Times(1).Return(appGenesisState[mod1])
	mockAppModule2.EXPECT().ExportGenesis(gomock.Eq(ctx), gomock.Eq(cdc)).Times(1).Return(appGenesisState[mod2])

	// assign the export path
	mm.SetGenesisPath(tmp)

	// run actual genesis state export
	actual, err := mm.ExportGenesis(ctx, cdc)
	require.NoError(t, err)
	require.Equal(t, make(map[string]json.RawMessage), actual)

	// check the state has been exported to the correct file path and verify the data
	bz, err := module.FileRead(filepath.Join(tmp, mod1), mod1)
	require.NoError(t, err)
	require.Equal(t, appGenesisState[mod1], json.RawMessage(bz))

	bz, err = module.FileRead(filepath.Join(tmp, mod2), mod2)
	require.NoError(t, err)
	require.Equal(t, appGenesisState[mod2], json.RawMessage(bz))
}
