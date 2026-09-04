package blocklog

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"
)

func readEntries(t *testing.T, dir string, height int64) []Entry {
	t.Helper()
	f, err := os.Open(filepath.Join(dir, strconv.FormatInt(height, 10)+".jsonl"))
	if os.IsNotExist(err) {
		return nil
	}
	require.NoError(t, err)
	defer f.Close()

	var out []Entry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var e Entry
		require.NoError(t, json.Unmarshal(sc.Bytes(), &e), "line: %s", sc.Text())
		out = append(out, e)
	}
	return out
}

func fileExists(t *testing.T, dir string, height int64) bool {
	t.Helper()
	_, err := os.Stat(filepath.Join(dir, strconv.FormatInt(height, 10)+".jsonl"))
	return err == nil
}

func TestStoreOpenCreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "block-logs")
	_, err := Open(dir, 3, log.NewNopLogger())
	require.NoError(t, err)
	st, err := os.Stat(dir)
	require.NoError(t, err)
	require.True(t, st.IsDir())
}

func TestStoreWritesOneJSONLinePerEntry(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir, 3, log.NewNopLogger())
	require.NoError(t, err)

	s.Begin(5)
	s.Write(Entry{Level: "info", Msg: "one"})
	s.Write(Entry{Level: "debug", Msg: "two", Module: "x/bank", Fields: map[string]string{"k": "v"}})
	s.Commit()

	got := readEntries(t, dir, 5)
	require.Len(t, got, 2)
	require.Equal(t, int64(5), got[0].Height, "store stamps the height")
	require.Equal(t, "one", got[0].Msg)
	require.Equal(t, "x/bank", got[1].Module)
	require.Equal(t, map[string]string{"k": "v"}, got[1].Fields)
	require.Equal(t, int64(5), s.LastClosed())
}

func TestStoreDropsWritesOutsideWindow(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir, 3, log.NewNopLogger())
	require.NoError(t, err)

	s.Write(Entry{Msg: "before"})
	s.Commit() // no-op without Begin
	require.Equal(t, int64(0), s.LastClosed())

	s.Begin(1)
	s.Write(Entry{Msg: "inside"})
	s.Commit()
	s.Write(Entry{Msg: "after"})

	got := readEntries(t, dir, 1)
	require.Len(t, got, 1)
	require.Equal(t, "inside", got[0].Msg)
	entries, _ := os.ReadDir(dir)
	require.Len(t, entries, 1)
}

func TestStoreCommitPrunesHeightMinusRetain(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir, 2, log.NewNopLogger())
	require.NoError(t, err)

	for h := int64(1); h <= 4; h++ {
		s.Begin(h)
		s.Write(Entry{Msg: "m"})
		s.Commit()
	}
	require.False(t, fileExists(t, dir, 1))
	require.False(t, fileExists(t, dir, 2))
	require.True(t, fileExists(t, dir, 3))
	require.True(t, fileExists(t, dir, 4))
}

func TestStoreFirstCommitSweepsStaleFiles(t *testing.T) {
	dir := t.TempDir()
	// files left behind by a previous run with a larger retention
	for _, h := range []int64{1, 2, 3, 7, 9} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, strconv.FormatInt(h, 10)+".jsonl"), []byte("{}\n"), 0o600))
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "junk.txt"), []byte("x"), 0o600))

	s, err := Open(dir, 3, log.NewNopLogger())
	require.NoError(t, err)
	s.Begin(10)
	s.Commit()

	// cutoff = 10-3 = 7: everything <= 7 goes, 9 and 10 stay, unrelated files untouched
	for _, h := range []int64{1, 2, 3, 7} {
		require.False(t, fileExists(t, dir, h), "height %d should be swept", h)
	}
	require.True(t, fileExists(t, dir, 9))
	require.True(t, fileExists(t, dir, 10))
	_, err = os.Stat(filepath.Join(dir, "junk.txt"))
	require.NoError(t, err)
}

func TestStoreBeginTruncatesExistingFile(t *testing.T) {
	dir := t.TempDir()
	// partial write from a crashed run that CometBFT is now replaying
	require.NoError(t, os.WriteFile(filepath.Join(dir, "4.jsonl"), []byte(`{"msg":"stale"}`+"\n"+`{"msg":"partial`), 0o600))

	s, err := Open(dir, 3, log.NewNopLogger())
	require.NoError(t, err)
	s.Begin(4)
	s.Write(Entry{Msg: "fresh"})
	s.Commit()

	got := readEntries(t, dir, 4)
	require.Len(t, got, 1)
	require.Equal(t, "fresh", got[0].Msg)
}

func TestStoreBeginWithoutCommitReopensSameHeight(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir, 3, log.NewNopLogger())
	require.NoError(t, err)

	s.Begin(2)
	s.Write(Entry{Msg: "first attempt"})
	// FinalizeBlock failed, CometBFT retries the same height
	s.Begin(2)
	s.Write(Entry{Msg: "second attempt"})
	s.Commit()

	got := readEntries(t, dir, 2)
	require.Len(t, got, 1)
	require.Equal(t, "second attempt", got[0].Msg)
}

// countingLogger counts Error calls; everything else is a no-op.
type countingLogger struct {
	log.Logger
	errors *int
}

func (c countingLogger) Error(string, ...any) { *c.errors++ }

func TestStoreErrorsAreLoggedOnceAndNeverReenterTheWrapper(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "block-logs")
	inner := countingLogger{Logger: log.NewNopLogger(), errors: new(int)}
	s, err := Open(dir, 3, inner)
	require.NoError(t, err)
	// route the store's own error logging through the capturing wrapper, as a
	// careless caller might; it must not deadlock on the store lock
	s.logger = Wrap(inner, s)

	require.NoError(t, os.RemoveAll(dir)) // every Begin now fails
	done := make(chan struct{})
	go func() {
		defer close(done)
		s.Begin(1)
		s.Begin(2)
		s.Begin(3)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("store deadlocked logging its own error")
	}
	require.Equal(t, 1, *inner.errors, "the same failure is logged once, not once per block")
}
