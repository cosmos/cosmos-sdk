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

	t.Run("schema caching", func(t *testing.T) {
		// Request schema once
		schemaReq := &v1.SchemaRequest{AccountType: "test"}
		schemaResp1, err := qs.Schema(ctx, schemaReq)
		require.NoError(t, err)
		require.NotNil(t, schemaResp1)

		// Request schema again
		schemaResp2, err := qs.Schema(ctx, schemaReq)
		require.NoError(t, err)
		require.NotNil(t, schemaResp2)

		// Check if both responses are the same (cached)
		require.Equal(t, schemaResp1, schemaResp2)
	})
}
