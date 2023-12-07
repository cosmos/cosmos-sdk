package accounts

import (
	"context"
	"testing"

	"github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/internal/implementation"
	v1 "cosmossdk.io/x/accounts/v1"
)

func TestQueryServer(t *testing.T) {
	k, ctx := newKeeper(t, accountstd.AddAccount("test", NewTestAccount))
	k.queryRouter = mockQuery(func(ctx context.Context, req, resp implementation.ProtoMsg) error {
		return nil
	})

	ms := NewMsgServer(k)
	qs := NewQueryServer(k)

	// create
	initMsg, err := implementation.PackAny(&emptypb.Empty{})
	require.NoError(t, err)

	initResp, err := ms.Init(ctx, &v1.MsgInit{
		Sender:      "sender",
		AccountType: "test",
		Message:     initMsg,
	})
	require.NoError(t, err)

	// query
	req := &wrapperspb.UInt64Value{Value: 10}
	anypbReq, err := implementation.PackAny(req)
	require.NoError(t, err)

	queryResp, err := qs.AccountQuery(ctx, &v1.AccountQueryRequest{
		Target:  initResp.AccountAddress,
		Request: anypbReq,
	})
	require.NoError(t, err)

	resp, err := implementation.UnpackAnyRaw(queryResp.Response)
	require.NoError(t, err)

	require.Equal(t, "10", resp.(*types.StringValue).Value)
}
