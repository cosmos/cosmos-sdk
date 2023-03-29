package log

import (
	"encoding/json"
	"fmt"
	"io"
)

func NewFilterWriter(parent io.Writer, filter FilterFunc) io.Writer {
	return &FilterWriter{parent, filter}
}

// FilterWriter is a writer that filters out all key/value pairs that do not match the filter.
type FilterWriter struct {
	parent io.Writer
	filter FilterFunc
}

func (fw *FilterWriter) Write(p []byte) (n int, err error) {
	if fw.filter == nil {
		return fw.parent.Write(p)
	}

	event := make(map[string]interface{})
	if err := json.Unmarshal(p, &event); err != nil {
		return 0, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	level, ok := event["level"].(string)
	if !ok {
		return 0, fmt.Errorf("failed to get level from event")
	}

	// only filter module keys
	module, ok := event[ModuleKey].(string)
	if ok {
		if fw.filter(module, level) {
			return len(p), nil
		}
	}

	return fw.parent.Write(p)
}
