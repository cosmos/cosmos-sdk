package node

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/server/blocklog"
	"github.com/cosmos/cosmos-sdk/types/query"
)

func TestLogsDisabled(t *testing.T) {
	srv := NewLogsServer(nil)
	_, err := srv.Logs(context.Background(), &LogsRequest{})
	require.Equal(t, codes.FailedPrecondition, status.Code(err))
	require.Contains(t, err.Error(), "retain-blocks")
}

func TestLogsRejectsUnsupportedPagination(t *testing.T) {
	store, err := blocklog.Open(t.TempDir(), 3, log.NewNopLogger())
	require.NoError(t, err)
	srv := NewLogsServer(store)

	for name, p := range map[string]*query.PageRequest{
		"offset":      {Offset: 1},
		"count_total": {CountTotal: true},
		"reverse":     {Reverse: true},
	} {
		_, err := srv.Logs(context.Background(), &LogsRequest{Pagination: p})
		require.Equal(t, codes.InvalidArgument, status.Code(err), name)
	}
}

func TestLogsMapsInvalidArgument(t *testing.T) {
	store, err := blocklog.Open(t.TempDir(), 3, log.NewNopLogger())
	require.NoError(t, err)
	store.Begin(1)
	store.Commit()

	_, err = NewLogsServer(store).Logs(context.Background(), &LogsRequest{Level: "loud"})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestLogsReturnsEntriesAndNextKey(t *testing.T) {
	store, err := blocklog.Open(t.TempDir(), 3, log.NewNopLogger())
	require.NoError(t, err)
	wrapped := blocklog.Wrap(log.NewNopLogger(), store).With(log.ModuleKey, "x/bank")
	store.Begin(1)
	wrapped.Info("first", "amount", 1)
	wrapped.Info("second")
	wrapped.Debug("third")
	store.Commit()

	srv := NewLogsServer(store)
	res, err := srv.Logs(context.Background(), &LogsRequest{Pagination: &query.PageRequest{Limit: 1}})
	require.NoError(t, err)
	require.Len(t, res.Entries, 1)
	e := res.Entries[0]
	require.Equal(t, int64(1), e.Height)
	require.Equal(t, "info", e.Level)
	require.Equal(t, "x/bank", e.Module)
	require.Equal(t, "first", e.Msg)
	require.Equal(t, map[string]string{"amount": "1"}, e.Fields)
	require.NotEmpty(t, e.Time)
	require.NotNil(t, res.Pagination)
	require.NotEmpty(t, res.Pagination.NextKey)

	res, err = srv.Logs(context.Background(), &LogsRequest{Pagination: &query.PageRequest{Key: res.Pagination.NextKey, Limit: 1}})
	require.NoError(t, err)
	require.Equal(t, "second", res.Entries[0].Msg)
	require.Empty(t, res.Pagination.NextKey, "third is debug and filtered out")
}
