package baseapp_test

import (
	"context"
	"sync"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	testdata_pulsar "github.com/cosmos/cosmos-sdk/testutil/testdata/testpb"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestGRPCQueryRouter tests the basic functionality of the gRPC query router
// including message handling, echo requests, and Any type support.
func TestGRPCQueryRouter(t *testing.T) {
	// Setup the gRPC query router with test data
	qr := baseapp.NewGRPCQueryRouter()
	interfaceRegistry := testdata.NewTestInterfaceRegistry()
	qr.SetInterfaceRegistry(interfaceRegistry)
	testdata_pulsar.RegisterQueryServer(qr, testdata_pulsar.QueryImpl{})
	helper := &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: qr,
		Ctx:             sdk.Context{}.WithContext(context.Background()),
	}
	client := testdata.NewQueryClient(helper)

	// Test echo functionality with a message
	res, err := client.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, "hello", res.Message)

	// Test echo functionality with nil request
	res, err = client.Echo(context.Background(), nil)
	require.Nil(t, err)
	require.Empty(t, res.Message)

	// Test SayHello functionality
	res2, err := client.SayHello(context.Background(), &testdata.SayHelloRequest{Name: "Foo"})
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, "Hello Foo!", res2.Greeting)

	// Test Any type handling with a Dog object
	spot := &testdata.Dog{Name: "Spot", Size_: "big"}
	any, err := types.NewAnyWithValue(spot)
	require.NoError(t, err)
	res3, err := client.TestAny(context.Background(), &testdata.TestAnyRequest{AnyAnimal: any})
	require.NoError(t, err)
	require.NotNil(t, res3)
	require.Equal(t, spot, res3.HasAnimal.Animal.GetCachedValue())
}

// TestGRPCRouterHybridHandlers tests the hybrid handler functionality
// that can handle both protov2 and gogoproto messages.
func TestGRPCRouterHybridHandlers(t *testing.T) {
	// Helper function to test router behavior with different message types
	assertRouterBehaviour := func(helper *baseapp.QueryServiceTestHelper) {
		// Test getting the handler by request name
		handlers := helper.HybridHandlerByRequestName("testpb.EchoRequest")
		require.NotNil(t, handlers)
		require.Len(t, handlers, 1)
		handler := handlers[0]
		// Test sending a protov2 message - should work and return a protov2 message
		v2Resp := new(testdata_pulsar.EchoResponse)
		err := handler(helper.Ctx, &testdata_pulsar.EchoRequest{Message: "hello"}, v2Resp)
		require.Nil(t, err)
		require.Equal(t, "hello", v2Resp.Message)
		// Test sending a gogoproto message - should work and return a gogoproto message
		gogoResp := new(testdata.EchoResponse)
		err = handler(helper.Ctx, &testdata.EchoRequest{Message: "hello"}, gogoResp)
		require.NoError(t, err)
		require.Equal(t, "hello", gogoResp.Message)
	}

	// Test with protov2 server implementation
	t.Run("protov2 server", func(t *testing.T) {
		qr := baseapp.NewGRPCQueryRouter()
		interfaceRegistry := testdata.NewTestInterfaceRegistry()
		qr.SetInterfaceRegistry(interfaceRegistry)
		testdata_pulsar.RegisterQueryServer(qr, testdata_pulsar.QueryImpl{})
		helper := &baseapp.QueryServiceTestHelper{
			GRPCQueryRouter: qr,
			Ctx:             sdk.Context{}.WithContext(context.Background()),
		}
		assertRouterBehaviour(helper)
	})

	// Test with gogoproto server implementation
	t.Run("gogoproto server", func(t *testing.T) {
		qr := baseapp.NewGRPCQueryRouter()
		interfaceRegistry := testdata.NewTestInterfaceRegistry()
		qr.SetInterfaceRegistry(interfaceRegistry)
		testdata.RegisterQueryServer(qr, testdata.QueryImpl{})
		helper := &baseapp.QueryServiceTestHelper{
			GRPCQueryRouter: qr,
			Ctx:             sdk.Context{}.WithContext(context.Background()),
		}
		assertRouterBehaviour(helper)
	})
}

// TestRegisterQueryServiceTwice tests that registering the same query service
// twice should panic on the second registration.
func TestRegisterQueryServiceTwice(t *testing.T) {
	// Setup baseapp with dependency injection
	var appBuilder *runtime.AppBuilder
	err := depinject.Inject(
		depinject.Configs(
			makeMinimalConfig(),
			depinject.Supply(log.NewTestLogger(t)),
		),
		&appBuilder)
	require.NoError(t, err)
	db := dbm.NewMemDB()
	app := appBuilder.Build(db, nil)

	// First registration should succeed without panicking
	require.NotPanics(t, func() {
		testdata.RegisterQueryServer(
			app.GRPCQueryRouter(),
			testdata.QueryImpl{},
		)
	})

	// Second registration should panic due to duplicate service
	require.Panics(t, func() {
		testdata.RegisterQueryServer(
			app.GRPCQueryRouter(),
			testdata.QueryImpl{},
		)
	})
}

// TestQueryDataRaces_sameConnectionToSameHandler tests that we don't have data races
// per https://github.com/cosmos/cosmos-sdk/issues/10324
// but with the same client connection being used concurrently.
func TestQueryDataRaces_sameConnectionToSameHandler(t *testing.T) {
	// Use a shared helper instance to test concurrent access to the same connection
	var mu sync.Mutex
	var helper *baseapp.QueryServiceTestHelper
	makeClientConn := func(qr *baseapp.GRPCQueryRouter) *baseapp.QueryServiceTestHelper {
		mu.Lock()
		defer mu.Unlock()

		if helper == nil {
			helper = &baseapp.QueryServiceTestHelper{
				GRPCQueryRouter: qr,
				Ctx:             sdk.Context{}.WithContext(context.Background()),
			}
		}
		return helper
	}
	testQueryDataRacesSameHandler(t, makeClientConn)
}

// TestQueryDataRaces_uniqueConnectionsToSameHandler tests that we don't have data races
// per https://github.com/cosmos/cosmos-sdk/issues/10324
// but with unique client connections requesting from the same handler concurrently.
func TestQueryDataRaces_uniqueConnectionsToSameHandler(t *testing.T) {
	// Create a new helper instance for every call to test unique connections
	testQueryDataRacesSameHandler(t, func(qr *baseapp.GRPCQueryRouter) *baseapp.QueryServiceTestHelper {
		return &baseapp.QueryServiceTestHelper{
			GRPCQueryRouter: qr,
			Ctx:             sdk.Context{}.WithContext(context.Background()),
		}
	})
}

// testQueryDataRacesSameHandler is a helper function that tests for data races
// when multiple goroutines access the same query handler concurrently.
func testQueryDataRacesSameHandler(t *testing.T, makeClientConn func(*baseapp.GRPCQueryRouter) *baseapp.QueryServiceTestHelper) {
	t.Helper()
	t.Parallel()

	// Setup the query router for testing
	qr := baseapp.NewGRPCQueryRouter()
	interfaceRegistry := testdata.NewTestInterfaceRegistry()
	qr.SetInterfaceRegistry(interfaceRegistry)
	testdata.RegisterQueryServer(qr, testdata.QueryImpl{})

	// The goal is to invoke the router concurrently and check for any data races.
	// 0. Run with: go test -race
	// 1. Synchronize all 1,000 goroutines to wait and query at the same time.
	// 2. Once the greenlight is given, perform queries through the router.
	var wg sync.WaitGroup
	defer wg.Wait()

	// Setup synchronization channels for concurrent testing
	greenlight := make(chan bool)
	n := 1000
	ready := make(chan bool, n)
	go func() {
		// Wait for all goroutines to be ready
		for range n {
			<-ready
		}
		// Signal all goroutines to start simultaneously
		close(greenlight)
	}()

	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Signal readiness and wait for the green light to start
			ready <- true
			<-greenlight

			// Perform various query operations to test for data races
			client := testdata.NewQueryClient(makeClientConn(qr))
			res, err := client.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
			require.Nil(t, err)
			require.NotNil(t, res)
			require.Equal(t, "hello", res.Message)

			res, err = client.Echo(context.Background(), nil)
			require.Nil(t, err)
			require.Empty(t, res.Message)

			res2, err := client.SayHello(context.Background(), &testdata.SayHelloRequest{Name: "Foo"})
			require.Nil(t, err)
			require.NotNil(t, res)
			require.Equal(t, "Hello Foo!", res2.Greeting)

			// Test Any type handling
			spot := &testdata.Dog{Name: "Spot", Size_: "big"}
			any, err := types.NewAnyWithValue(spot)
			require.NoError(t, err)
			res3, err := client.TestAny(context.Background(), &testdata.TestAnyRequest{AnyAnimal: any})
			require.NoError(t, err)
			require.NotNil(t, res3)
			require.Equal(t, spot, res3.HasAnimal.Animal.GetCachedValue())
		}()
	}
}
