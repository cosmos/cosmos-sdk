package baseapp_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGRPCGatewayRouter(t *testing.T) {
	qr := baseapp.NewGRPCQueryRouter()
	interfaceRegistry := testdata.NewTestInterfaceRegistry()
	qr.SetInterfaceRegistry(interfaceRegistry)
	testdata.RegisterQueryServer(qr, testdata.QueryImpl{})
	helper := &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: qr,
		Ctx:             sdk.Context{}.WithContext(context.Background()),
	}
	client := testdata.NewQueryClient(helper)

	res, err := client.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, "hello", res.Message)

	require.Panics(t, func() {
		_, _ = client.Echo(context.Background(), nil)
	})

	res2, err := client.SayHello(context.Background(), &testdata.SayHelloRequest{Name: "Foo"})
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, "Hello Foo!", res2.Greeting)

	spot := &testdata.Dog{Name: "Spot", Size_: "big"}
	any, err := types.NewAnyWithValue(spot)
	require.NoError(t, err)
	res3, err := client.TestAny(context.Background(), &testdata.TestAnyRequest{AnyAnimal: any})
	require.NoError(t, err)
	require.NotNil(t, res3)
	require.Equal(t, spot, res3.HasAnimal.Animal.GetCachedValue())
}

func TestRegisterQueryServiceTwice(t *testing.T) {
	// Setup baseapp.
	db := dbm.NewMemDB()
	encCfg := simapp.MakeTestEncodingConfig()
	app := baseapp.NewBaseApp("test", log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, encCfg.TxConfig.TxDecoder())
	app.SetInterfaceRegistry(encCfg.InterfaceRegistry)
	testdata.RegisterInterfaces(encCfg.InterfaceRegistry)

	// First time registering service shouldn't panic.
	require.NotPanics(t, func() {
		testdata.RegisterQueryServer(
			app.GRPCQueryRouter(),
			testdata.QueryImpl{},
		)
	})

	// Second time should panic.
	require.Panics(t, func() {
		testdata.RegisterQueryServer(
			app.GRPCQueryRouter(),
			testdata.QueryImpl{},
		)
	})
}
