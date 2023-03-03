package log

import "github.com/rs/zerolog"

// TestingT is the interface required for logging in tests.
// It is a subset of testing.T to avoid a direct dependency on the testing package.
type TestingT zerolog.TestingLog

// NewTestLogger returns a logger that calls t.Log to write entries.
//
// If the logs may help debug a test failure,
// you may want to use NewTestLogger(t) in your test.
// Otherwise, use NewNopLogger().
func NewTestLogger(t TestingT) Logger {
	cw := zerolog.NewConsoleWriter()
	cw.Out = zerolog.TestWriter{
		T: t,
		// Normally one would use zerolog.ConsoleTestWriter
		// to set the option on NewConsoleWriter,
		// but the zerolog source for that is hardcoded to Frame=6.
		// With Frame=6, all source locations are printed as "logger.go",
		// but Frame=7 prints correct source locations.
		Frame: 7,
	}
	return NewCustomLogger(zerolog.New(cw))
}
