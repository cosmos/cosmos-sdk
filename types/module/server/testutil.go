package server

//
//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	"github.com/cosmos/cosmos-sdk/baseapp"
//	"github.com/cosmos/cosmos-sdk/codec"
//	"github.com/cosmos/cosmos-sdk/codec/types"
//	"github.com/cosmos/cosmos-sdk/testutil/testdata"
//	sdk "github.com/cosmos/cosmos-sdk/types"
//	"github.com/cosmos/cosmos-sdk/types/module"
//	mod "github.com/cosmos/cosmos-sdk/types/module"
//	"github.com/cosmos/cosmos-sdk/types/testutil"
//	"github.com/stretchr/testify/require"
//	abci "github.com/tendermint/tendermint/abci/types"
//	"github.com/tendermint/tendermint/libs/log"
//	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
//	dbm "github.com/tendermint/tm-db"
//	"google.golang.org/grpc"
//	"testing"
//)
//
//type FixtureFactory struct {
//	t       *testing.T
//	modules []mod.Module
//	signers []sdk.AccAddress
//	cdc     *codec.ProtoCodec
//	baseApp *baseapp.BaseApp
//}
//
//func NewFixtureFactory(t *testing.T, numSigners int) *FixtureFactory {
//	signers := makeTestAddresses(numSigners)
//	return &FixtureFactory{
//		t:       t,
//		signers: signers,
//		// cdc and baseApp are initialized here just for compatibility with legacy modules which don't use ADR 033
//		// TODO: remove once all code using this uses ADR 033 module wiring
//		cdc:     codec.NewProtoCodec(types.NewInterfaceRegistry()),
//		baseApp: baseapp.NewBaseApp("test", log.NewNopLogger(), dbm.NewMemDB(), nil),
//	}
//}
//
//func (ff *FixtureFactory) SetModules(modules []mod.Module) {
//	ff.modules = modules
//}
//
//// Codec is exposed just for compatibility of these test suites with legacy modules and can be removed when everything
//// has been migrated to ADR 033
//func (ff *FixtureFactory) Codec() *codec.ProtoCodec {
//	return ff.cdc
//}
//
//// BaseApp is exposed just for compatibility of these test suites with legacy modules and can be removed when everything
//// has been migrated to ADR 033
//func (ff *FixtureFactory) BaseApp() *baseapp.BaseApp {
//	return ff.baseApp
//}
//
//func makeTestAddresses(count int) []sdk.AccAddress {
//	addrs := make([]sdk.AccAddress, count)
//	for i := 0; i < count; i++ {
//		_, _, addrs[i] = testdata.KeyTestPubAddr()
//	}
//	return addrs
//}
//
//func (ff FixtureFactory) Setup() testutil.Fixture {
//	cdc := ff.cdc
//	registry := cdc.InterfaceRegistry()
//	baseApp := ff.baseApp
//	baseApp.MsgServiceRouter().SetInterfaceRegistry(registry)
//	baseApp.GRPCQueryRouter().SetInterfaceRegistry(registry)
//	mm := NewManager(baseApp, cdc)
//	err := mm.RegisterModules(ff.modules)
//	require.NoError(ff.t, err)
//	err = mm.CompleteInitialization()
//	require.NoError(ff.t, err)
//	err = baseApp.LoadLatestVersion()
//	require.NoError(ff.t, err)
//
//	return fixture{
//		baseApp:               baseApp,
//		router:                mm.router,
//		cdc:                   cdc,
//		initGenesisHandlers:   mm.initGenesisHandlers,
//		exportGenesisHandlers: mm.exportGenesisHandlers,
//		t:                     ff.t,
//		signers:               ff.signers,
//	}
//}
//
//type fixture struct {
//	baseApp               *baseapp.BaseApp
//	router                *router
//	cdc                   *codec.ProtoCodec
//	initGenesisHandlers   map[string]module.InitGenesisHandler
//	exportGenesisHandlers map[string]module.ExportGenesisHandler
//	t                     *testing.T
//	signers               []sdk.AccAddress
//}
//
//func (f fixture) Context() context.Context {
//	ctx := f.baseApp.NewUncachedContext(false, tmproto.Header{})
//	return ctx
//}
//
//func (f fixture) TxConn() grpc.ClientConnInterface {
//	return testKey{invokerFactory: f.router.testTxFactory(f.signers)}
//}
//
//func (f fixture) QueryConn() grpc.ClientConnInterface {
//	return testKey{invokerFactory: f.router.testQueryFactory()}
//}
//
//func (f fixture) Signers() []sdk.AccAddress {
//	return f.signers
//}
//
//func (f fixture) InitGenesis(ctx sdk.Context, genesisData map[string]json.RawMessage) (abci.ResponseInitChain, error) {
//	return initGenesis(ctx, f.cdc, genesisData, []abci.ValidatorUpdate{}, f.initGenesisHandlers)
//}
//
//func (f fixture) ExportGenesis(ctx sdk.Context) (map[string]json.RawMessage, error) {
//	return exportGenesis(ctx, f.cdc, f.exportGenesisHandlers)
//}
//
//func (f fixture) Codec() *codec.ProtoCodec {
//	return f.cdc
//}
//
//func (f fixture) Teardown() {}
//
//type testKey struct {
//	invokerFactory InvokerFactory
//}
//
//var _ grpc.ClientConnInterface = testKey{}
//
//func (t testKey) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, _ ...grpc.CallOption) error {
//	invoker, err := t.invokerFactory(CallInfo{Method: method})
//	if err != nil {
//		return err
//	}
//
//	return invoker(ctx, args, reply)
//}
//
//func (t testKey) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
//	return nil, fmt.Errorf("unsupported")
//}
