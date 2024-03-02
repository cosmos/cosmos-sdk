package runtime_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	counterv1 "cosmossdk.io/api/cosmos/counter/v1"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	counterkeeper "github.com/cosmos/cosmos-sdk/x/counter/keeper"
	countertypes "github.com/cosmos/cosmos-sdk/x/counter/types"
)

func TestRouterService(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	msgRouter := baseapp.NewMsgServiceRouter()
	msgRouter.SetInterfaceRegistry(interfaceRegistry)
	queryRouter := baseapp.NewGRPCQueryRouter()
	queryRouter.SetInterfaceRegistry(interfaceRegistry)
	key := storetypes.NewKVStoreKey(countertypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	counterKeeper := counterkeeper.NewKeeper(runtime.NewEnvironment(storeService, log.NewNopLogger()))
	countertypes.RegisterInterfaces(interfaceRegistry)
	countertypes.RegisterMsgServer(msgRouter, counterKeeper)
	countertypes.RegisterQueryServer(queryRouter, counterKeeper)

	routerService := runtime.NewRouterService(storeService, queryRouter, msgRouter)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))

	// Messages

	t.Run("invalid msg", func(t *testing.T) {
		_, err := routerService.MessageRouterService().InvokeUntyped(testCtx.Ctx, &bankv1beta1.MsgSend{})
		require.ErrorContains(t, err, "could not find response type for message cosmos.bank.v1beta1.MsgSend")
	})

	t.Run("invoke untyped: valid msg (proto v1)", func(t *testing.T) {
		resp, err := routerService.MessageRouterService().InvokeUntyped(testCtx.Ctx, &countertypes.MsgIncreaseCounter{
			Signer: "cosmos1",
			Count:  42,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("invoke typed: valid msg (proto v1)", func(t *testing.T) {
		resp := &countertypes.MsgIncreaseCountResponse{}
		err := routerService.MessageRouterService().InvokeTyped(testCtx.Ctx, &countertypes.MsgIncreaseCounter{
			Signer: "cosmos1",
			Count:  42,
		}, resp)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("invoke typed: valid msg (proto v2)", func(t *testing.T) {
		resp := &counterv1.MsgIncreaseCountResponse{}
		err := routerService.MessageRouterService().InvokeTyped(testCtx.Ctx, &counterv1.MsgIncreaseCounter{
			Signer: "cosmos1",
			Count:  42,
		}, resp)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	// Queries

	t.Run("invalid query", func(t *testing.T) {
		err := routerService.QueryRouterService().InvokeTyped(testCtx.Ctx, &bankv1beta1.QueryBalanceRequest{}, &bankv1beta1.QueryBalanceResponse{})
		require.ErrorContains(t, err, "unknown request: cosmos.bank.v1beta1.QueryBalanceRequest")
	})

	t.Run("invoke typed: valid query (proto v1)", func(t *testing.T) {
		_ = counterKeeper.CountStore.Set(testCtx.Ctx, 42)

		resp := &countertypes.QueryGetCountResponse{}
		err := routerService.QueryRouterService().InvokeTyped(testCtx.Ctx, &countertypes.QueryGetCountRequest{}, resp)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, int64(42), resp.TotalCount)
	})

	t.Run("invoke typed: valid query (proto v2)", func(t *testing.T) {
		_ = counterKeeper.CountStore.Set(testCtx.Ctx, 42)

		resp := &counterv1.QueryGetCountResponse{}
		err := routerService.QueryRouterService().InvokeTyped(testCtx.Ctx, &counterv1.QueryGetCountRequest{}, resp)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, int64(42), resp.TotalCount)
	})

	t.Run("invoke untyped: valid query (proto v1)", func(t *testing.T) {
		_ = counterKeeper.CountStore.Set(testCtx.Ctx, 42)

		resp, err := routerService.QueryRouterService().InvokeUntyped(testCtx.Ctx, &countertypes.QueryGetCountRequest{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		respVal, ok := resp.(*countertypes.QueryGetCountResponse)
		require.True(t, ok)
		require.Equal(t, int64(42), respVal.TotalCount)
	})
}
