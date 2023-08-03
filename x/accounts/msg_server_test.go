package accounts_test

import (
	"testing"

	counterv1 "cosmossdk.io/api/cosmos/accounts/examples/counter/v1"
	accountsv1 "cosmossdk.io/api/cosmos/accounts/v1"
	"cosmossdk.io/x/accounts"
	"cosmossdk.io/x/accounts/examples/counter"
	"cosmossdk.io/x/accounts/examples/echo"
	"cosmossdk.io/x/accounts/tempcore/header"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestMsgServer(t *testing.T) {
	addressCodec, headerService, storeService, ctx := accountDeps[header.Header]()

	module, err := accounts.NewAccounts(
		addressCodec,
		storeService,
		headerService,
		accounts.AddAccount("counter", counter.NewCounter),
		accounts.AddAccount("echo", echo.NewEcho),
	)
	require.NoError(t, err)

	msgServer := module.MsgServer()

	createResp, err := msgServer.Create(ctx, &accountsv1.MsgCreate{
		Creator:     "frojdi",
		AccountType: "counter",
		Message:     []byte(`{"counter_value": 100}`),
	})
	require.NoError(t, err)

	t.Log(createResp)

	increaseResp, err := msgServer.Execute(ctx, &accountsv1.MsgExecute{
		Sender:  "frojdi",
		Target:  createResp.Address,
		Message: messageToAnyJSON(t, &counterv1.MsgIncreaseCounter{}),
	})

	require.NoError(t, err)
	require.NotNil(t, increaseResp, "response is nil")
	t.Log(increaseResp)
}

func messageToAnyJSON(t *testing.T, msg proto.Message) []byte {
	anyMsg, err := anypb.New(&counterv1.MsgIncreaseCounter{})
	require.NoError(t, err)

	anyMsgJSON, err := protojson.Marshal(anyMsg)
	require.NoError(t, err)
	return anyMsgJSON
}
