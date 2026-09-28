package blocklog

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"
)

// recordingLogger is a fake inner logger that remembers every call.
type recordingLogger struct {
	calls   *[]string
	verbose *bool
}

func newRecording() recordingLogger {
	return recordingLogger{calls: new([]string), verbose: new(bool)}
}

func (r recordingLogger) rec(level, msg string, kv ...any) {
	*r.calls = append(*r.calls, level+":"+msg+":"+fmt.Sprintf("%v", kv))
}
func (r recordingLogger) Info(msg string, kv ...any) { r.rec("info", msg, kv...) }
func (r recordingLogger) InfoContext(_ context.Context, m string, kv ...any) {
	r.rec("infoctx", m, kv...)
}
func (r recordingLogger) Warn(msg string, kv ...any) { r.rec("warn", msg, kv...) }
func (r recordingLogger) WarnContext(_ context.Context, m string, kv ...any) {
	r.rec("warnctx", m, kv...)
}
func (r recordingLogger) Error(msg string, kv ...any) { r.rec("error", msg, kv...) }
func (r recordingLogger) ErrorContext(_ context.Context, m string, kv ...any) {
	r.rec("errorctx", m, kv...)
}
func (r recordingLogger) Debug(msg string, kv ...any) { r.rec("debug", msg, kv...) }
func (r recordingLogger) DebugContext(_ context.Context, m string, kv ...any) {
	r.rec("debugctx", m, kv...)
}
func (r recordingLogger) With(kv ...any) log.Logger { r.rec("with", "", kv...); return r }
func (r recordingLogger) Impl() any                 { return r.calls }
func (r recordingLogger) SetVerboseMode(v bool)     { *r.verbose = v }

func newStore(t *testing.T) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	s, err := Open(dir, 3, log.NewNopLogger())
	require.NoError(t, err)
	return s, dir
}

func TestWrapPassesEveryCallThrough(t *testing.T) {
	s, _ := newStore(t)
	inner := newRecording()
	l := Wrap(inner, s)
	ctx := context.Background()

	l.Debug("d", "a", 1)
	l.DebugContext(ctx, "dc")
	l.Info("i", "b", 2)
	l.InfoContext(ctx, "ic")
	l.Warn("w")
	l.WarnContext(ctx, "wc")
	l.Error("e", "err", errors.New("boom"))
	l.ErrorContext(ctx, "ec")

	require.Equal(t, []string{
		"debug:d:[a 1]", "debugctx:dc:[]", "info:i:[b 2]", "infoctx:ic:[]",
		"warn:w:[]", "warnctx:wc:[]", "error:e:[err boom]", "errorctx:ec:[]",
	}, *inner.calls)
	require.Equal(t, inner.calls, l.Impl(), "Impl delegates to the inner logger")
}

func TestWrapCapturesDebugBelowInnerLevel(t *testing.T) {
	s, dir := newStore(t)
	var buf bytes.Buffer
	inner := log.NewLogger(&buf, log.LevelOption(zerolog.InfoLevel), log.OutputJSONOption())
	l := Wrap(inner, s)

	s.Begin(3)
	l.Debug("hidden from console", "k", "v")
	l.Info("shown")
	s.Commit()

	require.NotContains(t, buf.String(), "hidden from console")
	require.Contains(t, buf.String(), "shown")

	got := readEntries(t, dir, 3)
	require.Len(t, got, 2)
	require.Equal(t, "debug", got[0].Level)
	require.Equal(t, "hidden from console", got[0].Msg)
	require.Equal(t, map[string]string{"k": "v"}, got[0].Fields)
	require.Equal(t, "info", got[1].Level)
	_, err := time.Parse(time.RFC3339Nano, got[0].Time)
	require.NoError(t, err)
}

func TestWrapWithReturnsCapturingChild(t *testing.T) {
	s, dir := newStore(t)
	inner := newRecording()
	root := Wrap(inner, s, log.ModuleKey, "server")
	child := root.With(log.ModuleKey, "x/bank", "chain", "test")
	grandchild := child.With("height", 7)

	s.Begin(1)
	root.Info("root")
	child.Info("child")
	grandchild.Warn("grandchild")
	s.Commit()

	got := readEntries(t, dir, 1)
	require.Len(t, got, 3)
	require.Equal(t, "server", got[0].Module, "seeded attribution fields apply without touching the inner logger")
	require.Nil(t, got[0].Fields)
	require.Equal(t, "x/bank", got[1].Module, "With overrides the module")
	require.Equal(t, map[string]string{"chain": "test"}, got[1].Fields)
	require.Equal(t, "x/bank", got[2].Module)
	require.Equal(t, map[string]string{"chain": "test", "height": "7"}, got[2].Fields)
	require.Contains(t, *inner.calls, "with::[module x/bank chain test]", "With is forwarded to the inner logger")
	require.NotContains(t, *inner.calls, "with::[module server]", "seed fields are attribution only")
}

type stringer struct{}

func (stringer) String() string { return "stringer!" }

func TestWrapStringifiesFields(t *testing.T) {
	s, dir := newStore(t)
	l := Wrap(newRecording(), s)

	s.Begin(1)
	l.Info("m", "err", errors.New("bad"), "str", stringer{}, "n", 42, "nil", nil, 7, "int key", "dangling")
	s.Commit()

	got := readEntries(t, dir, 1)
	require.Len(t, got, 1)
	require.Equal(t, map[string]string{
		"err": "bad", "str": "stringer!", "n": "42", "nil": "<nil>", "7": "int key", "dangling": "<nil>",
	}, got[0].Fields)
}

func TestWrapForwardsVerboseMode(t *testing.T) {
	s, _ := newStore(t)
	inner := newRecording()
	l := Wrap(inner, s)
	vl, ok := l.(log.VerboseModeLogger)
	require.True(t, ok)
	vl.SetVerboseMode(true)
	require.True(t, *inner.verbose)

	// an inner logger without verbose support is simply ignored
	Wrap(log.NewNopLogger(), s).(log.VerboseModeLogger).SetVerboseMode(true)
}

func TestWrapExposesBlockHooks(t *testing.T) {
	s, dir := newStore(t)
	l := Wrap(newRecording(), s)
	hooks, ok := l.(interface {
		BeginBlockLog(int64)
		CommitBlockLog()
	})
	require.True(t, ok)

	hooks.BeginBlockLog(9)
	l.Info("in block")
	hooks.CommitBlockLog()
	require.Equal(t, int64(9), s.LastClosed())
	require.Len(t, readEntries(t, dir, 9), 1)
}

// fakeKeeper mirrors x/bank: it receives a logger at construction and never
// reads ctx.Logger().
type fakeKeeper struct{ logger log.Logger }

func newFakeKeeper(logger log.Logger) fakeKeeper {
	return fakeKeeper{logger: logger.With(log.ModuleKey, "x/bank")}
}

func (k fakeKeeper) MintCoins(amount string) {
	k.logger.Debug("minted coins from module account", "amount", amount)
}

func TestInjectedKeeperLoggerIsCapturedPerHeight(t *testing.T) {
	s, dir := newStore(t)
	root := Wrap(newRecording(), s, log.ModuleKey, "server")
	keeper := newFakeKeeper(root) // constructed once, before any block

	s.Begin(7)
	keeper.MintCoins("1stake")
	s.Commit()
	s.Begin(8)
	keeper.MintCoins("2stake")
	s.Commit()

	h7 := readEntries(t, dir, 7)
	require.Len(t, h7, 1)
	require.Equal(t, int64(7), h7[0].Height)
	require.Equal(t, "x/bank", h7[0].Module)
	require.Equal(t, "1stake", h7[0].Fields["amount"])

	h8 := readEntries(t, dir, 8)
	require.Len(t, h8, 1)
	require.Equal(t, int64(8), h8[0].Height)
	require.Equal(t, "2stake", h8[0].Fields["amount"])
}
