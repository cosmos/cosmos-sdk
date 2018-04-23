package types

import (
	"github.com/tendermint/tmlibs/log"
)

// MemLogger logs to memory
type MemLogger interface {
	log.Logger
	Logs() []LogEntry
}

// LogEntry is an entry in a log
type LogEntry struct {
	Level   string
	Message string
	Keyvals []interface{}
}

type memLogger struct {
	entries *[]LogEntry
}

func (l memLogger) Debug(msg string, keyvals ...interface{}) {
	*l.entries = append(*l.entries, LogEntry{"debug", msg, keyvals})
}

func (l memLogger) Info(msg string, keyvals ...interface{}) {
	*l.entries = append(*l.entries, LogEntry{"info", msg, keyvals})
}

func (l memLogger) Error(msg string, keyvals ...interface{}) {
	*l.entries = append(*l.entries, LogEntry{"error", msg, keyvals})
}

func (l memLogger) With(keyvals ...interface{}) log.Logger {
	panic("not implemented")
}

func (l memLogger) Logs() []LogEntry {
	return *l.entries
}

func NewMemLogger() MemLogger {
	entries := make([]LogEntry, 0)
	return &memLogger{
		entries: &entries,
	}
}
