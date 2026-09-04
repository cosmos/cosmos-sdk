package blocklog

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"
)

// seed writes heights 1..n, each with the given entries, into a store with the
// given retention and returns it.
func seed(t *testing.T, retain uint64, n int64, perHeight ...Entry) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	s, err := Open(dir, retain, log.NewNopLogger())
	require.NoError(t, err)
	for h := int64(1); h <= n; h++ {
		s.Begin(h)
		for _, e := range perHeight {
			s.Write(e)
		}
		s.Commit()
	}
	return s, dir
}

var mixed = []Entry{
	{Level: "debug", Msg: "d", Module: "x/bank"},
	{Level: "info", Msg: "i", Module: "x/bank"},
	{Level: "warn", Msg: "w", Module: "x/mint"},
	{Level: "error", Msg: "e"},
}

func msgs(entries []Entry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Msg
	}
	return out
}

func TestQueryLevelIsInclusiveMinimum(t *testing.T) {
	s, _ := seed(t, 10, 1, mixed...)
	for level, want := range map[string][]string{
		"debug": {"d", "i", "w", "e"},
		"":      {"i", "w", "e"}, // default info
		"info":  {"i", "w", "e"},
		"warn":  {"w", "e"},
		"error": {"e"},
	} {
		res, err := s.Query(Request{Level: level})
		require.NoError(t, err, level)
		require.Equal(t, want, msgs(res.Entries), "level=%q", level)
	}
	_, err := s.Query(Request{Level: "trace"})
	require.ErrorIs(t, err, ErrInvalidArgument)
}

func TestQueryModuleExactMatchExcludesUntagged(t *testing.T) {
	s, _ := seed(t, 10, 1, mixed...)
	res, err := s.Query(Request{Level: "debug", Module: "x/bank"})
	require.NoError(t, err)
	require.Equal(t, []string{"d", "i"}, msgs(res.Entries))

	res, err = s.Query(Request{Level: "debug", Module: "x/"})
	require.NoError(t, err)
	require.Empty(t, res.Entries)
}

func TestQueryHeightDefaultsAndErrors(t *testing.T) {
	s, _ := seed(t, 10, 5, Entry{Level: "info", Msg: "m"})

	res, err := s.Query(Request{})
	require.NoError(t, err)
	require.Len(t, res.Entries, 1)
	require.Equal(t, int64(5), res.Entries[0].Height, "start and end default to last closed height")

	res, err = s.Query(Request{Start: 2, End: 4})
	require.NoError(t, err)
	require.Equal(t, []int64{2, 3, 4}, heights(res.Entries))

	res, err = s.Query(Request{Start: 3})
	require.NoError(t, err)
	require.Equal(t, []int64{3, 4, 5}, heights(res.Entries), "end defaults to last closed")

	for name, req := range map[string]Request{
		"start > end":                 {Start: 4, End: 3},
		"end beyond closed":           {Start: 1, End: 6},
		"negative start":              {Start: -1, End: 2},
		"end=0 keeps start>end check": {Start: 6},
	} {
		_, err := s.Query(req)
		require.ErrorIs(t, err, ErrInvalidArgument, name)
	}
}

func TestQueryClipsStartToRetentionWindow(t *testing.T) {
	s, _ := seed(t, 3, 10, Entry{Level: "info", Msg: "m"})

	res, err := s.Query(Request{Start: 1, End: 10})
	require.NoError(t, err)
	require.Equal(t, []int64{8, 9, 10}, heights(res.Entries))

	// range entirely outside the window: empty, not an error
	res, err = s.Query(Request{Start: 1, End: 2})
	require.NoError(t, err)
	require.Empty(t, res.Entries)
	require.Nil(t, res.NextKey)
}

func TestQueryNoBlocksClosedYet(t *testing.T) {
	s, err := Open(t.TempDir(), 3, log.NewNopLogger())
	require.NoError(t, err)
	_, err = s.Query(Request{})
	require.ErrorIs(t, err, ErrInvalidArgument)
}

func TestQueryCursorWalksRangeWithoutGapsOrDuplicates(t *testing.T) {
	s, _ := seed(t, 10, 3, mixed...) // 3 heights * 3 entries >= info = 9

	var all []Entry
	var key []byte
	pages := 0
	for {
		res, err := s.Query(Request{Start: 1, End: 3, Key: key, Limit: 2})
		require.NoError(t, err)
		pages++
		require.LessOrEqual(t, len(res.Entries), 2)
		all = append(all, res.Entries...)
		if res.NextKey == nil {
			break
		}
		key = res.NextKey
		require.Less(t, pages, 20, "runaway pagination")
	}
	require.Equal(t, 5, pages, "9 entries in pages of 2, last page has 1 and no next key")
	require.Equal(t, []int64{1, 1, 1, 2, 2, 2, 3, 3, 3}, heights(all))
	require.Equal(t, []string{"i", "w", "e", "i", "w", "e", "i", "w", "e"}, msgs(all))
}

func TestQueryNextKeyOnlyWhenMoreMatches(t *testing.T) {
	s, _ := seed(t, 10, 2, mixed...)
	// exactly 2 matching entries in the range and limit 2: no next key
	res, err := s.Query(Request{Start: 1, End: 2, Level: "error", Limit: 2})
	require.NoError(t, err)
	require.Len(t, res.Entries, 2)
	require.Nil(t, res.NextKey)
}

func TestQueryRejectsBadKeys(t *testing.T) {
	s, _ := seed(t, 10, 3, mixed...)
	_, err := s.Query(Request{Start: 1, End: 3, Key: []byte("short")})
	require.ErrorIs(t, err, ErrInvalidArgument)

	res, err := s.Query(Request{Start: 3, End: 3, Limit: 1})
	require.NoError(t, err)
	require.NotNil(t, res.NextKey)
	_, err = s.Query(Request{Start: 1, End: 2, Key: res.NextKey})
	require.ErrorIs(t, err, ErrInvalidArgument, "key height outside the requested range")
}

func TestQuerySkipsMalformedLines(t *testing.T) {
	s, dir := seed(t, 10, 1, Entry{Level: "info", Msg: "good"})
	f, err := os.OpenFile(filepath.Join(dir, "1.jsonl"), os.O_APPEND|os.O_WRONLY, 0o644)
	require.NoError(t, err)
	_, err = f.WriteString(`{"height":1,"level":"info","msg":"trunc`)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	res, err := s.Query(Request{})
	require.NoError(t, err)
	require.Equal(t, []string{"good"}, msgs(res.Entries))
}

func TestQuerySkipsMissingFiles(t *testing.T) {
	s, dir := seed(t, 10, 3, Entry{Level: "info", Msg: "m"})
	require.NoError(t, os.Remove(filepath.Join(dir, "2.jsonl")))
	res, err := s.Query(Request{Start: 1, End: 3})
	require.NoError(t, err)
	require.Equal(t, []int64{1, 3}, heights(res.Entries))
}

func TestQueryLimitDefaultsAndClamps(t *testing.T) {
	many := make([]Entry, 0, MaxLimit+50)
	for range MaxLimit + 50 {
		many = append(many, Entry{Level: "info", Msg: "m"})
	}
	s, _ := seed(t, 10, 1, many...)

	res, err := s.Query(Request{})
	require.NoError(t, err)
	require.Len(t, res.Entries, DefaultLimit)

	res, err = s.Query(Request{Limit: MaxLimit + 10})
	require.NoError(t, err)
	require.Len(t, res.Entries, MaxLimit)
	require.NotNil(t, res.NextKey)
}

func heights(entries []Entry) []int64 {
	out := make([]int64, len(entries))
	for i, e := range entries {
		out[i] = e.Height
	}
	return out
}
