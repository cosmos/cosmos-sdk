// Package blocklog captures the application log lines emitted while a block is
// executed and stores them as one JSONL file per height, so a node can answer
// "what did the application log during blocks N..M" without a log aggregator.
//
// Capture is a pure passthrough on top of the wrapped log.Logger: nothing is
// read from or written to application state and no error is ever surfaced to
// the caller, so enabling it cannot affect consensus.
package blocklog

// Entry is one captured log line.
type Entry struct {
	Height int64  `json:"height"`
	Time   string `json:"time"` // wall clock, RFC3339Nano
	Level  string `json:"level"`
	Module string `json:"module,omitempty"`
	Msg    string `json:"msg"`
	// Fields holds the remaining key/value pairs, values stringified.
	Fields map[string]string `json:"fields,omitempty"`
}
