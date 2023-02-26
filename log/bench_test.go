package log_test

import (
	"bytes"
	"io"
	"testing"
	"time"

	"cosmossdk.io/log"
	"github.com/rs/zerolog"
)

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
			// Small numbers may be optimized, so check if an unusual/larger number performs different.
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

	const message = "test message"

	// If running with "go test -v", print out the log messages as a sanity check.
	if testing.Verbose() {
		checkBuf := new(bytes.Buffer)
		for _, bc := range benchCases {
			checkBuf.Reset()
			logger := log.ZeroLogWrapper{Logger: zerolog.New(checkBuf)}
			logger.Info(message, bc.keyVals...)

			b.Logf("zero logger output for %s: %s", bc.name, checkBuf.String())
		}
	}

	// The real logger exposed by this package,
	// writing to an io.Discard writer,
	// so that real write time is negligible.
	b.Run("zerolog", func(b *testing.B) {
		for _, bc := range benchCases {
			bc := bc
			b.Run(bc.name, func(b *testing.B) {
				logger := log.ZeroLogWrapper{Logger: zerolog.New(io.Discard)}

				for i := 0; i < b.N; i++ {
					logger.Info(message, bc.keyVals...)
				}
			})
		}
	})

	// zerolog offers a no-op writer.
	// It appears to be slower than our custom NopLogger,
	// so include it in the nop benchmarks as a point of reference.
	b.Run("zerolog nop", func(b *testing.B) {
		for _, bc := range nopCases {
			bc := bc
			b.Run(bc.name, func(b *testing.B) {
				logger := log.ZeroLogWrapper{Logger: zerolog.Nop()}

				for i := 0; i < b.N; i++ {
					logger.Info(message, bc.keyVals...)
				}
			})
		}
	})

	// The nop logger we use in tests,
	// also useful as a reference for how expensive zerolog is.
	b.Run("nop logger", func(b *testing.B) {
		for _, bc := range nopCases {
			bc := bc
			b.Run(bc.name, func(b *testing.B) {
				logger := log.NewNopLogger()

				for i := 0; i < b.N; i++ {
					logger.Info(message, bc.keyVals...)
				}
			})
		}
	})
}
