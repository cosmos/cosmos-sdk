package tx_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

// TestGetTxsEvent_RejectsDeprecatedPagination ensures that a request which
// populates the deprecated `pagination` field is rejected with InvalidArgument
// rather than silently translated. The top-level page / limit / order_by
// fields supersede `pagination`; keeping a compatibility shim would force the
// SDK to maintain a translation layer indefinitely. See #25886.
func TestGetTxsEvent_RejectsDeprecatedPagination(t *testing.T) {
	srv := tx.NewTxServer(client.Context{}, nil, codectypes.NewInterfaceRegistry())

	cases := []struct {
		name string
		pag  *query.PageRequest
	}{
		{"limit set", &query.PageRequest{Limit: 100}},
		{"offset set", &query.PageRequest{Offset: 50}},
		{"key set", &query.PageRequest{Key: []byte("k")}},
		{"reverse set", &query.PageRequest{Reverse: true}},
		{"count_total set", &query.PageRequest{CountTotal: true}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := srv.GetTxsEvent(context.Background(), &txtypes.GetTxsEventRequest{
				Query:      "tx.height=1",
				Pagination: tc.pag,
			})
			require.Error(t, err)
			st, ok := status.FromError(err)
			require.True(t, ok, "expected gRPC status error, got %T", err)
			require.Equal(t, codes.InvalidArgument, st.Code())
			require.Contains(t, st.Message(), "pagination is deprecated")
		})
	}
}

// TestGetTxsEvent_EmptyPaginationAccepted ensures that an unset or fully-zero
// Pagination field does not trigger the rejection above, so clients that
// auto-construct request structs with a zero-value Pagination still work.
func TestGetTxsEvent_EmptyPaginationAccepted(t *testing.T) {
	srv := tx.NewTxServer(client.Context{}, nil, codectypes.NewInterfaceRegistry())

	// Nil pagination — must not trip the InvalidArgument guard. The call
	// itself will fail later trying to reach a real node (clientCtx has no
	// client), but the error must NOT be the InvalidArgument from the guard.
	_, err := srv.GetTxsEvent(context.Background(), &txtypes.GetTxsEventRequest{
		Query: "tx.height=1",
	})
	if err != nil {
		st, _ := status.FromError(err)
		require.NotEqual(t, codes.InvalidArgument, st.Code(),
			"nil pagination must not be rejected: %v", err)
	}

	// Zero-value pagination — same expectation.
	_, err = srv.GetTxsEvent(context.Background(), &txtypes.GetTxsEventRequest{
		Query:      "tx.height=1",
		Pagination: &query.PageRequest{},
	})
	if err != nil {
		st, _ := status.FromError(err)
		require.NotEqual(t, codes.InvalidArgument, st.Code(),
			"zero pagination must not be rejected: %v", err)
	}
}
