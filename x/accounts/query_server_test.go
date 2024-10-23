package accounts

import (
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

	ms := NewMsgServer(k)
	qs := NewQueryServer(k)

	// create account
	initMsg, err := implementation.PackAny(&emptypb.Empty{})
	require.NoError(t, err)

	initResp, err := ms.Init(ctx, &v1.MsgInit{
		Sender:      "sender",
		AccountType: "test",
		Message:     initMsg,
	})
	require.NoError(t, err)

	t.Run("account query", func(t *testing.T) {
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
	})

	t.Run("account number", func(t *testing.T) {
		numResp, err := qs.AccountNumber(ctx, &v1.AccountNumberRequest{Address: initResp.AccountAddress})
		require.NoError(t, err)
		require.Equal(t, 0, int(numResp.Number))
	})

	t.Run("account type", func(t *testing.T) {
		typ, err := qs.AccountType(ctx, &v1.AccountTypeRequest{Address: initResp.AccountAddress})
		require.NoError(t, err)
		require.Equal(t, "test", typ.AccountType)
	})

	emptyMsgName := implementation.MessageName(&types.Empty{})
	t.Run("schema", func(t *testing.T) {
		schemaReq := &v1.SchemaRequest{Address: initResp.AccountAddress}
		schemaResp, err := qs.Schema(ctx, schemaReq)
		require.NoError(t, err)
		require.NotNil(t, schemaResp)

		require.Equal(t, schemaResp.InitSchema.Request, emptyMsgName)
		require.Equal(t, schemaResp.InitSchema.Response, emptyMsgName)

		// check handlers
		require.Len(t, schemaResp.ExecuteHandlers, 4)
		require.Equal(t, schemaResp.ExecuteHandlers[0].Request, emptyMsgName)
		require.Equal(t, schemaResp.ExecuteHandlers[0].Response, emptyMsgName)

		require.Len(t, schemaResp.QueryHandlers, 4)
		require.Equal(t, schemaResp.QueryHandlers[1].Request, emptyMsgName)
		require.Equal(t, schemaResp.QueryHandlers[1].Response, emptyMsgName)
	})

	t.Run("init schema", func(t *testing.T) {
		r, err := qs.InitSchema(ctx, &v1.InitSchemaRequest{AccountType: "test"})
		require.NoError(t, err)
		require.NotNil(t, r.InitSchema)
		require.Equal(t, r.InitSchema.Request, emptyMsgName)
		require.Equal(t, r.InitSchema.Response, emptyMsgName)
	})
}
