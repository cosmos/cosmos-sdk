package accounts

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"cosmossdk.io/x/accounts/accountstd"
	v1 "cosmossdk.io/x/accounts/v1"
)

func TestQueryServer(t *testing.T) {
	k, ctx := newKeeper(t, accountstd.AddAccount("test", NewTestAccount))
	k.queryModuleFunc = func(ctx context.Context, req, resp protoiface.MessageV1) error {
		return nil
	}

	ms := NewMsgServer(k)
	qs := NewQueryServer(k)

	// create
	initMsg, err := proto.Marshal(&emptypb.Empty{})
	require.NoError(t, err)

	initResp, err := ms.Init(ctx, &v1.MsgInit{
		Sender:      "sender",
		AccountType: "test",
		Message:     initMsg,
	})
	require.NoError(t, err)

	// query
	req := &wrapperspb.UInt64Value{Value: 10}
	anypbReq, err := anypb.New(req)
	require.NoError(t, err)

	anypbReqBytes, err := proto.Marshal(anypbReq)
	require.NoError(t, err)

	queryResp, err := qs.AccountQuery(ctx, &v1.AccountQueryRequest{
		Target:  initResp.AccountAddress,
		Request: anypbReqBytes,
	})
	require.NoError(t, err)

	respAnyPB := &anypb.Any{}
	err = proto.Unmarshal(queryResp.Response, respAnyPB)
	require.NoError(t, err)

	resp, err := respAnyPB.UnmarshalNew()
	require.NoError(t, err)

	require.Equal(t, "10", resp.(*wrapperspb.StringValue).Value)
}
