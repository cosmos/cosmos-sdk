package client_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/query"
)

func TestReadPageRequest(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		expected    *query.PageRequest
		expectedErr string
	}{
		{
			"defaults",
			nil,
			&query.PageRequest{Key: []byte{}, Limit: 100},
			"",
		},
		{
			"page sets offset",
			[]string{"--page=3", "--limit=10"},
			&query.PageRequest{Key: []byte{}, Offset: 20, Limit: 10},
			"",
		},
		{
			"page-key alone",
			[]string{"--page-key=abc"},
			&query.PageRequest{Key: []byte("abc"), Limit: 100},
			"",
		},
		{
			"page and offset",
			[]string{"--page=2", "--offset=10"},
			nil,
			"page and offset cannot be used together",
		},
		{
			"page and page-key",
			[]string{"--page=2", "--page-key=abc"},
			nil,
			"page and page-key cannot be used together",
		},
		{
			"page 1 and page-key is allowed",
			[]string{"--page=1", "--page-key=abc"},
			&query.PageRequest{Key: []byte("abc"), Limit: 100},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			flags.AddPaginationFlagsToCmd(cmd, "things")
			require.NoError(t, cmd.Flags().Parse(tc.args))

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if tc.expectedErr != "" {
				require.ErrorContains(t, err, tc.expectedErr)
				require.Nil(t, pageReq)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, pageReq)
		})
	}
}

func TestPaginate(t *testing.T) {
	testCases := []struct {
		name                           string
		numObjs, page, limit, defLimit int
		expectedStart, expectedEnd     int
	}{
		{
			"all objects in a single page",
			100, 1, 100, 100,
			0, 100,
		},
		{
			"page one of three",
			75, 1, 25, 100,
			0, 25,
		},
		{
			"page two of three",
			75, 2, 25, 100,
			25, 50,
		},
		{
			"page three of three",
			75, 3, 25, 100,
			50, 75,
		},
		{
			"end is greater than total number of objects",
			75, 2, 50, 100,
			50, 75,
		},
		{
			"fallback to default limit",
			75, 5, 0, 10,
			40, 50,
		},
		{
			"invalid start page",
			75, 4, 25, 100,
			-1, -1,
		},
		{
			"invalid zero start page",
			75, 0, 25, 100,
			-1, -1,
		},
		{
			"invalid negative start page",
			75, -1, 25, 100,
			-1, -1,
		},
		{
			"invalid default limit",
			75, 2, 0, -10,
			-1, -1,
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			start, end := client.Paginate(tc.numObjs, tc.page, tc.limit, tc.defLimit)
			require.Equal(t, tc.expectedStart, start, "invalid result; test case #%d", i)
			require.Equal(t, tc.expectedEnd, end, "invalid result; test case #%d", i)
		})
	}
}
