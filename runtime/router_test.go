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

func TestMsgRouterService(t *testing.T) {
	router := baseapp.NewMsgServiceRouter()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	router.SetInterfaceRegistry(interfaceRegistry)
	key := storetypes.NewKVStoreKey(countertypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	counterKeeper := counterkeeper.NewKeeper(runtime.NewEnvironment(storeService, log.NewNopLogger()))
	countertypes.RegisterInterfaces(interfaceRegistry)
	countertypes.RegisterMsgServer(router, counterKeeper)
	countertypes.RegisterQueryServer(router, counterKeeper)

	routerService := runtime.NewMsgRouterService(storeService, router)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))

	t.Run("invalid msg", func(t *testing.T) {
		_, err := routerService.InvokeUntyped(testCtx.Ctx, &bankv1beta1.MsgSend{})
		require.ErrorContains(t, err, "could not find response type for message /cosmos.bank.v1beta1.MsgSend")
	})

	t.Run("valid msg (proto v1)", func(t *testing.T) {
		resp, err := routerService.InvokeUntyped(testCtx.Ctx, &countertypes.MsgIncreaseCounter{
			Signer: "cosmos1",
			Count:  42,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("valid msg (proto v2)", func(t *testing.T) {
		resp, err := routerService.InvokeUntyped(testCtx.Ctx, &counterv1.MsgIncreaseCounter{
			Signer: "cosmos1",
			Count:  42,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("invalid query", func(t *testing.T) {
		err := routerService.InvokeTyped(testCtx.Ctx, &bankv1beta1.QueryBalanceRequest{}, &bankv1beta1.QueryBalanceResponse{})
		require.ErrorContains(t, err, "unknown message: /cosmos.bank.v1beta1.QueryBalanceRequest")
	})

	t.Run("valid query (proto v1)", func(t *testing.T) {
		resp := &countertypes.QueryGetCountResponse{}
		err := routerService.InvokeTyped(testCtx.Ctx, &countertypes.QueryGetCountRequest{}, resp)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, &countertypes.QueryGetCountResponse{TotalCount: 42}, resp)
	})

	t.Run("valid query (proto v2)", func(t *testing.T) {
		resp := &counterv1.QueryGetCountResponse{}
		err := routerService.InvokeTyped(testCtx.Ctx, &counterv1.QueryGetCountRequest{}, resp)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, &counterv1.QueryGetCountResponse{TotalCount: 42}, resp)
	})
}
