package accounts_test

import (
	"context"
	"testing"

	counterv1 "cosmossdk.io/api/cosmos/accounts/examples/counter/v1"
	"cosmossdk.io/collections/colltest"
	"cosmossdk.io/core/store"
	"cosmossdk.io/x/accounts"
	"cosmossdk.io/x/accounts/examples/counter"
	"cosmossdk.io/x/accounts/examples/echo"
	"cosmossdk.io/x/accounts/tempcore/header"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestModule(t *testing.T) {
	hs, ss, ctx := accountDeps[header.Header]()

	module, err := accounts.NewAccounts(
		ss,
		hs,
		accounts.AddAccount("counter", counter.NewCounter),
		accounts.AddAccount("echo", echo.NewEcho),
	)
	require.NoError(t, err)

	sender := []byte("sender")

	counterAddr, resp, err := module.Create(ctx, "counter", sender, &counterv1.MsgInit{CounterValue: 100})
	require.NoError(t, err)

	echoAddr, _, err := module.Create(ctx, "echo", sender, &emptypb.Empty{})
	require.NoError(t, err)

	resp, err = module.Execute(ctx, sender, counterAddr, &counterv1.MsgIncreaseCounter{})
	require.NoError(t, err)
	require.NotNil(t, resp, "response is nil")
	t.Log(resp)

	resp, err = module.Query(ctx, counterAddr, &counterv1.QueryCounterRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp, "response is nil")
	t.Log(resp)

	// test comms between accounts

	resp, err = module.Execute(ctx, sender, counterAddr, &counterv1.MsgExecuteEcho{Target: echoAddr, Msg: "hello"})
	require.NoError(t, err)

	t.Log(resp)
}

func accountDeps[H header.Header]() (header.Service[H], store.KVStoreService, context.Context) {
	ss, ctx := colltest.MockStore()
	return nil, ss, ctx
}
