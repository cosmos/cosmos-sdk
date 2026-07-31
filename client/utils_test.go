package client_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
)

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

// pageFlags returns a flag set carrying the standard pagination flags with --page-key set.
func pageFlags(t *testing.T, pageKey string) *pflag.FlagSet {
	t.Helper()
	cmd := &cobra.Command{}
	flags.AddPaginationFlagsToCmd(cmd, "things")
	require.NoError(t, cmd.Flags().Set(flags.FlagPageKey, pageKey))
	return cmd.Flags()
}

// TestReadPageRequestPageKeyRoundTrip checks that the next_key a query response prints
// can be handed straight back to --page-key and yields the original raw key bytes.
func TestReadPageRequestPageKeyRoundTrip(t *testing.T) {
	rawKey := []byte{0xff, 0xfe, 0x01, 0x7e, '/', '+'}

	out, err := codec.ProtoMarshalJSON(&query.PageResponse{NextKey: rawKey}, nil)
	require.NoError(t, err)
	var resp struct {
		NextKey string `json:"next_key"`
	}
	require.NoError(t, json.Unmarshal(out, &resp))
	require.NotEmpty(t, resp.NextKey)

	pageReq, err := client.ReadPageRequest(pageFlags(t, resp.NextKey))
	require.NoError(t, err)
	require.Equal(t, rawKey, pageReq.Key)
}

func TestReadPageRequestPageKeyEmpty(t *testing.T) {
	pageReq, err := client.ReadPageRequest(pageFlags(t, ""))
	require.NoError(t, err)
	require.Empty(t, pageReq.Key)
}

func TestReadPageRequestPageKeyInvalidBase64(t *testing.T) {
	pageReq, err := client.ReadPageRequest(pageFlags(t, "not base64!"))
	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.ErrorContains(t, err, "invalid --page-key")
	require.ErrorContains(t, err, "illegal base64 data at input byte")
	require.Nil(t, pageReq)
}

// TestReadPageRequestPageKeyNullLiteral checks that the literal "null" a JSON/YAML
// response prints as the last page's empty next_key is rejected instead of being
// base64-decoded to garbage bytes.
func TestReadPageRequestPageKeyNullLiteral(t *testing.T) {
	for _, pageKey := range []string{"null", "NULL", "Null"} {
		pageReq, err := client.ReadPageRequest(pageFlags(t, pageKey))
		require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest, pageKey)
		require.ErrorContains(t, err, "last page", pageKey)
		require.Nil(t, pageReq, pageKey)
	}
}

// TestReadPageRequestDeprecatedHelpersNoDoubleDecode checks that callers still going
// through the deprecated helpers get the same result: the helpers no longer decode, so
// ReadPageRequest's own decoding is not applied to already-decoded bytes.
func TestReadPageRequestDeprecatedHelpersNoDoubleDecode(t *testing.T) {
	rawKey := []byte{0xde, 0xad, 0xbe, 0xef}
	encoded := base64.StdEncoding.EncodeToString(rawKey)

	fs, err := client.FlagSetWithPageKeyDecoded(pageFlags(t, encoded))
	require.NoError(t, err)
	pageReq, err := client.ReadPageRequest(fs)
	require.NoError(t, err)
	require.Equal(t, rawKey, pageReq.Key)

	pageReq, err = client.ReadPageRequest(client.MustFlagSetWithPageKeyDecoded(pageFlags(t, encoded)))
	require.NoError(t, err)
	require.Equal(t, rawKey, pageReq.Key)

	// MustFlagSetWithPageKeyDecoded panicked on invalid base64 before it became a no-op.
	invalid := pageFlags(t, "not base64!")
	require.NotPanics(t, func() { client.MustFlagSetWithPageKeyDecoded(invalid) })
}
