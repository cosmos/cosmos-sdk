package server_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/cosmos/cosmos-sdk/server"
)

func TestVariousLevels(t *testing.T) {
	testCases := []struct {
		name    string
		allowed server.Option
		want    string
	}{
		{
			"AllowAll",
			server.AllowAll(),
			strings.Join([]string{
				`{"level":"debug","this is":"debug log","message":"here"}`,
				`{"level":"info","this is":"info log","message":"here"}`,
				`{"level":"error","this is":"error log","message":"here"}`,
			}, "\n"),
		},
		{
			"AllowDebug",
			server.AllowDebug(),
			strings.Join([]string{
				`{"level":"debug","this is":"debug log","message":"here"}`,
				`{"level":"info","this is":"info log","message":"here"}`,
				`{"level":"error","this is":"error log","message":"here"}`,
			}, "\n"),
		},
		{
			"AllowInfo",
			server.AllowInfo(),
			strings.Join([]string{
				`{"level":"info","this is":"info log","message":"here"}`,
				`{"level":"error","this is":"error log","message":"here"}`,
			}, "\n"),
		},
		{
			"AllowError",
			server.AllowError(),
			strings.Join([]string{
				`{"level":"error","this is":"error log","message":"here"}`,
			}, "\n"),
		},
		{
			"AllowNone",
			server.AllowNone(),
			``,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := server.NewFilter(server.ZeroLogWrapper{
				Logger: zerolog.New(&buf).With().Logger(),
			}, tc.allowed)

			logger.Debug("here", "this is", "debug log")
			logger.Info("here", "this is", "info log")
			logger.Error("here", "this is", "error log")

			if want, have := tc.want, strings.TrimSpace(buf.String()); want != have {
				t.Errorf("\nwant:\n%s\nhave:\n%s", want, have)
			}
		})
	}
}

func TestLevelContext(t *testing.T) {
	var buf bytes.Buffer
	zeroLogger := server.ZeroLogWrapper{
		Logger: zerolog.New(&buf).Level(zerolog.InfoLevel).With().Logger(),
	}

	logger := server.NewFilter(zeroLogger, server.AllowError())

	logger.Error("foo", "bar", "baz")

	want := `{"level":"error","bar":"baz","message":"foo"}`
	have := strings.TrimSpace(buf.String())
	if want != have {
		t.Errorf("\nwant '%s'\nhave '%s'", want, have)
	}

	buf.Reset()
	logger.Info("foo", "bar", "baz")
	if want, have := ``, strings.TrimSpace(buf.String()); want != have {
		t.Errorf("\nwant '%s'\nhave '%s'", want, have)
	}
}

func TestVariousAllowWith(t *testing.T) {
	var buf bytes.Buffer

	logger := server.ZeroLogWrapper{
		Logger: zerolog.New(&buf).With().Logger(),
	}

	logger1 := server.NewFilter(logger, server.AllowError(), server.AllowInfoWith("context", "value"))
	logger1.With("context", "value").Info("foo", "bar", "baz")

	want := `{"level":"info","context":"value","bar":"baz","message":"foo"}`
	have := strings.TrimSpace(buf.String())
	if want != have {
		t.Errorf("\nwant '%s'\nhave '%s'", want, have)
	}

	buf.Reset()

	logger2 := server.NewFilter(
		logger,
		server.AllowError(),
		server.AllowInfoWith("context", "value"),
		server.AllowNoneWith("user", "Sam"),
	)

	logger2.With("context", "value", "user", "Sam").Info("foo", "bar", "baz")
	if want, have := ``, strings.TrimSpace(buf.String()); want != have {
		t.Errorf("\nwant '%s'\nhave '%s'", want, have)
	}

	buf.Reset()

	logger3 := server.NewFilter(
		logger,
		server.AllowError(),
		server.AllowInfoWith("context", "value"),
		server.AllowNoneWith("user", "Sam"),
	)

	logger3.With("user", "Sam").With("context", "value").Info("foo", "bar", "baz")

	want = `{"level":"info","user":"Sam","context":"value","bar":"baz","message":"foo"}`
	have = strings.TrimSpace(buf.String())
	if want != have {
		t.Errorf("\nwant '%s'\nhave '%s'", want, have)
	}
}
