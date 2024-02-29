package runtime_test

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	counterkeeper "github.com/cosmos/cosmos-sdk/x/counter/keeper"
	countertypes "github.com/cosmos/cosmos-sdk/x/counter/types"
	"github.com/stretchr/testify/require"
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
	// countertypes.RegisterQueryServer(router, counterKeeper)

	routerService := runtime.NewMsgRouterService(storeService, router)

	resp, err := routerService.InvokeUntyped(context.Background(), &countertypes.MsgIncreaseCounter{
		Signer: "cosmos1",
		Count:  42,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	resp = &countertypes.QueryGetCountResponse{}
	err = routerService.InvokeTyped(context.Background(), &countertypes.QueryGetCountRequest{}, &countertypes.QueryGetCountResponse{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, &countertypes.QueryGetCountResponse{TotalCount: 42}, resp)
}
