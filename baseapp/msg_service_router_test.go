package baseapp_test

import (
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
)

func TestRegisterMsgService(t *testing.T) {
	// Setup baseapp.
	var (
		appBuilder *runtime.AppBuilder
		registry   codectypes.InterfaceRegistry
	)
	err := depinject.Inject(
		depinject.Configs(
			baseapptestutil.MakeMinimalConfig(),
			depinject.Supply(log.NewTestLogger(t)),
		), &appBuilder, &registry)
	require.NoError(t, err)
	app := appBuilder.Build(dbm.NewMemDB(), nil)

	require.Panics(t, func() {
		testdata.RegisterMsgServer(
			app.MsgServiceRouter(),
			testdata.MsgServerImpl{},
		)
	})

	// Register testdata Msg services, and rerun `RegisterMsgService`.
	testdata.RegisterInterfaces(registry)

	require.NotPanics(t, func() {
		testdata.RegisterMsgServer(
			app.MsgServiceRouter(),
			testdata.MsgServerImpl{},
		)
	})
}

func TestRegisterMsgServiceTwice(t *testing.T) {
	// Setup baseapp.
	var (
		appBuilder *runtime.AppBuilder
		registry   codectypes.InterfaceRegistry
	)
	err := depinject.Inject(
		depinject.Configs(
			baseapptestutil.MakeMinimalConfig(),
			depinject.Supply(log.NewTestLogger(t)),
		), &appBuilder, &registry)
	require.NoError(t, err)
	db := dbm.NewMemDB()
	app := appBuilder.Build(db, nil)
	testdata.RegisterInterfaces(registry)

	// First time registering service shouldn't panic.
	require.NotPanics(t, func() {
		testdata.RegisterMsgServer(
			app.MsgServiceRouter(),
			testdata.MsgServerImpl{},
		)
	})

	// Second time should panic.
	require.Panics(t, func() {
		testdata.RegisterMsgServer(
			app.MsgServiceRouter(),
			testdata.MsgServerImpl{},
		)
	})
}

func TestHybridHandlerByMsgName(t *testing.T) {
	// Setup baseapp and router.
	var (
		appBuilder *runtime.AppBuilder
		registry   codectypes.InterfaceRegistry
	)
	err := depinject.Inject(
		depinject.Configs(
			baseapptestutil.MakeMinimalConfig(),
			depinject.Supply(log.NewTestLogger(t)),
		), &appBuilder, &registry)
	require.NoError(t, err)
	db := dbm.NewMemDB()
	app := appBuilder.Build(db, nil)
	testdata.RegisterInterfaces(registry)

	testdata.RegisterMsgServer(
		app.MsgServiceRouter(),
		testdata.MsgServerImpl{},
	)

	handler := app.MsgServiceRouter().HybridHandlerByMsgName("testpb.MsgCreateDog")

	require.NotNil(t, handler)
	require.NoError(t, app.Init())
	ctx := app.NewContext(true)
	resp := new(testdata.MsgCreateDogResponse)
	err = handler(ctx, &testdata.MsgCreateDog{
		Dog:   &testdata.Dog{Name: "Spot"},
		Owner: "me",
	}, resp)
	require.NoError(t, err)
	require.Equal(t, resp.Name, "Spot")
}
