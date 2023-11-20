package accounts

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	"cosmossdk.io/x/accounts/accountstd"
	v1 "cosmossdk.io/x/accounts/v1"
)

func TestMsgServer(t *testing.T) {
	k, ctx := newKeeper(t, accountstd.AddAccount("test", NewTestAccount))
	k.queryRouter = mockQuery(func(ctx context.Context, req, resp proto.Message) error {
		_, ok := req.(*bankv1beta1.QueryBalanceRequest)
		require.True(t, ok)
		proto.Merge(resp, &bankv1beta1.QueryBalanceResponse{})
		return nil
	})

	s := NewMsgServer(k)

	// create
	initMsg, err := proto.Marshal(&emptypb.Empty{})
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
	executeMsgAny, err := anypb.New(executeMsg)
	require.NoError(t, err)

	executeMsgBytes, err := proto.Marshal(executeMsgAny)
	require.NoError(t, err)

	execResp, err := s.Execute(ctx, &v1.MsgExecute{
		Sender:  "sender",
		Target:  initResp.AccountAddress,
		Message: executeMsgBytes,
	})
	require.NoError(t, err)
	require.NotNil(t, execResp)
}
