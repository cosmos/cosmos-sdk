package log_test

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"cosmossdk.io/log"
)

const message = "test message"

func BenchmarkLoggers(b *testing.B) {
	b.ReportAllocs()

	type benchCase struct {
		name    string
		keyVals []any
	}

	// Just test two simple cases for the nop logger benchmarks.
	nopCases := []benchCase{
		{name: "empty key vals"},
		{name: "single string", keyVals: []any{"foo", "bar"}},
	}

	benchCases := append(nopCases, []benchCase{
		{
			name:    "single small int",
			keyVals: []any{"foo", 1},
		},
		{
			// Small numbers may be optimized, so check if an unusual/larger number performs differently.
			name:    "single largeish int",
			keyVals: []any{"foo", 123456789},
		},
		{
			name:    "single float",
			keyVals: []any{"foo", 2.71828182},
		},
		{
			name:    "single byte slice",
			keyVals: []any{"foo", []byte{0xde, 0xad, 0xbe, 0xef}},
		},
		{
			name:    "single duration",
			keyVals: []any{"foo", 10 * time.Second},
		},

		{
			name:    "two values",
			keyVals: []any{"foo", "foo", "bar", "bar"},
		},
		{
			name:    "four values",
			keyVals: []any{"foo", "foo", "bar", "bar", "baz", "baz", "quux", "quux"},
		},
		{
			name:    "eight values",
			keyVals: []any{"one", 1, "two", 2, "three", 3, "four", 4, "five", 5, "six", 6, "seven", 7, "eight", 8},
		},
	}...)

	// If running with "go test -v", print out the log messages as a sanity check.
	if testing.Verbose() {
		checkBuf := new(bytes.Buffer)
		for _, bc := range benchCases {
			checkBuf.Reset()
			logger := log.NewCustomLogger(slog.New(slog.NewJSONHandler(checkBuf, nil)))
			logger.Info(message, bc.keyVals...)

			b.Logf("slog logger output for %s: %s", bc.name, checkBuf.String())
		}
	}

	// The real logger exposed by this package,
	// writing to an io.Discard writer,
	// so that real write time is negligible.
	b.Run("slog", func(b *testing.B) {
		for _, bc := range benchCases {
			b.Run(bc.name, func(b *testing.B) {
				logger := log.NewCustomLogger(slog.New(slog.NewJSONHandler(io.Discard, nil)))

				for b.Loop() {
					logger.Info(message, bc.keyVals...)
				}
			})
		}
	})

	// The nop logger we expose in the public API.
	b.Run("specialized nop logger", func(b *testing.B) {
		for _, bc := range nopCases {
			b.Run(bc.name, func(b *testing.B) {
				logger := log.NewNopLogger()

				for b.Loop() {
					logger.Info(message, bc.keyVals...)
				}
			})
		}
	})
}

func BenchmarkLoggers_StructuredVsFields(b *testing.B) {
	b.ReportAllocs()

	errorToLog := errors.New("error")
	byteSliceToLog := []byte{0xde, 0xad, 0xbe, 0xef}

	b.Run("slog structured", func(b *testing.B) {
		logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
		for b.Loop() {
			logger.Info(message, slog.Int64("foo", 100000))
			logger.Info(message, slog.String("foo", "foo"))
			logger.Error(message,
				slog.Int64("foo", 100000),
				slog.String("bar", "foo"),
				slog.Any("other", byteSliceToLog),
				slog.Any("error", errorToLog),
			)
		}
	})

	b.Run("logger", func(b *testing.B) {
		logger := log.NewCustomLogger(slog.New(slog.NewJSONHandler(io.Discard, nil)))
		for b.Loop() {
			logger.Info(message, "foo", 100000)
			logger.Info(message, "foo", "foo")
			logger.Error(message, "foo", 100000, "bar", "foo", "other", byteSliceToLog, "error", errorToLog)
		}
	})
}
