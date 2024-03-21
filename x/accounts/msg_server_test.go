package accounts

import (
	"context"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/internal/implementation"
	v1 "cosmossdk.io/x/accounts/v1"
	banktypes "cosmossdk.io/x/bank/types"
)

func TestMsgServer(t *testing.T) {
	k, ctx := newKeeper(t, accountstd.AddAccount("test", NewTestAccount))
	k.queryRouter = mockQuery(func(ctx context.Context, req, resp implementation.ProtoMsg) error {
		_, ok := req.(*banktypes.QueryBalanceRequest)
		require.True(t, ok)
		proto.Merge(resp.(proto.Message), &banktypes.QueryBalanceResponse{})
		return nil
	})

	s := NewMsgServer(k)

	// create
	initMsg, err := implementation.PackAny(&emptypb.Empty{})
	require.NoError(t, err)

	initResp, err := s.Init(ctx, &v1.MsgInit{
		Sender:      "sender",
		AccountType: "test",
		Message:     initMsg,
	})
	require.NoError(t, err)
	require.NotNil(t, initResp)

	// execute
	executeMsg := &wrapperspb.StringValue{
		Value: "10",
	}
	executeMsgAny, err := implementation.PackAny(executeMsg)
	require.NoError(t, err)

	execResp, err := s.Execute(ctx, &v1.MsgExecute{
		Sender:  "sender",
		Target:  initResp.AccountAddress,
		Message: executeMsgAny,
	})
	require.NoError(t, err)
	require.NotNil(t, execResp)
}
