package accounts

import (
	"context"
	"testing"

	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/internal/implementation"
	v1 "cosmossdk.io/x/accounts/v1"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"
)

func TestQueryServer(t *testing.T) {
	k, ctx := newKeeper(t, accountstd.AddAccount("test", NewTestAccount))
	k.queryRouter = mockQuery(func(ctx context.Context, req, resp implementation.ProtoMsg) error {
		return nil
	})

	ms := NewMsgServer(k)
	qs := NewQueryServer(k)

	// create account
	initMsg, err := implementation.PackAny(&gogotypes.Empty{})
	require.NoError(t, err)

	initResp, err := ms.Init(ctx, &v1.MsgInit{
		Sender:      "sender",
		AccountType: "test",
		Message:     initMsg,
	})
	require.NoError(t, err)

	t.Run("account query", func(t *testing.T) {
		// query
		req := &gogotypes.UInt64Value{Value: 10}
		anypbReq, err := implementation.PackAny(req)
		require.NoError(t, err)

		queryResp, err := qs.AccountQuery(ctx, &v1.AccountQueryRequest{
			Target:  initResp.AccountAddress,
			Request: anypbReq,
		})
		require.NoError(t, err)

		resp, err := implementation.UnpackAnyRaw(queryResp.Response)
		require.NoError(t, err)
		require.Equal(t, "10", resp.(*gogotypes.StringValue).Value)
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
}
