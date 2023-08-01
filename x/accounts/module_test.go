package accounts_test

import (
	"context"
	"testing"

	"cosmossdk.io/collections/colltest"
	"cosmossdk.io/core/store"
	"cosmossdk.io/x/accounts"
	"cosmossdk.io/x/accounts/examples/counter"
	counterv1 "cosmossdk.io/x/accounts/examples/counter/v1"
	"cosmossdk.io/x/accounts/examples/echo"
	"cosmossdk.io/x/accounts/tempcore/header"
	"github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"
)

func TestModule(t *testing.T) {
	hs, ss, ctx := accountDeps[header.Header]()

	module, err := accounts.NewAccounts[header.Header](
		ss,
		hs,
		accounts.AddAccount("counter", counter.NewCounter),
		accounts.AddAccount("echo", echo.NewEcho),
	)
	require.NoError(t, err)

	sender := []byte("sender")

	counterAddr, resp, err := module.Create(ctx, "counter", sender, &counterv1.MsgInit{CounterValue: 100})
	require.NoError(t, err)

	echoAddr, _, err := module.Create(ctx, "echo", sender, &types.Empty{})
	require.NoError(t, err)

	resp, err = module.Execute(ctx, sender, counterAddr, &counterv1.MsgIncreaseCounter{})
	require.NoError(t, err)
	require.NotNil(t, resp, "response is nil")
	t.Log(resp.String())

	resp, err = module.Query(ctx, counterAddr, &counterv1.QueryCounterRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp, "response is nil")
	t.Log(resp.String())

	// test comms between accounts

	resp, err = module.Execute(ctx, sender, counterAddr, &counterv1.MsgExecuteEcho{Target: echoAddr, Msg: "hello"})
	require.NoError(t, err)

	t.Log(resp.String())
}

func accountDeps[H header.Header]() (header.Service[H], store.KVStoreService, context.Context) {
	ss, ctx := colltest.MockStore()
	return nil, ss, ctx
}
