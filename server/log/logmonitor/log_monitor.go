package logmonitor

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// LogMonitor is responsible for monitoring logs and triggering actions based on specific patterns.
type LogMonitor struct {
	shutdownFn      func(string)
	mu              sync.Mutex
	shutdownStrings []string
}

// multiWriter is a custom io.Writer that writes to multiple destinations.
type multiWriter struct {
	writers []io.Writer
	monitor *LogMonitor
}

// NewLogMonitor creates a new LogMonitor instance.
func NewLogMonitor(cfg *Config, shutdownFn func(string)) *LogMonitor {
	return &LogMonitor{
		shutdownFn:      shutdownFn,
		shutdownStrings: cfg.ShutdownStrings,
	}
}

// InitGlobalLogMonitor initializes the log monitoring system.
func InitGlobalLogMonitor(cfg Config, shutdownFn func(string)) (io.Writer, io.Writer) {
	if !cfg.Enabled {
		return os.Stdout, os.Stderr
	}
	monitor := NewLogMonitor(&cfg, shutdownFn)
	stdout := NewMultiWriter(monitor, os.Stdout)
	stderr := NewMultiWriter(monitor, os.Stderr)
	return stdout, stderr
}

// NewMultiWriter creates a new multiWriter instance.
func NewMultiWriter(monitor *LogMonitor, writers ...io.Writer) io.Writer {
	return &multiWriter{
		writers: writers,
		monitor: monitor,
	}
}

// Write implements the io.Writer interface for multiWriter.
func (mw *multiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return
		}
	}
	_, _ = mw.monitor.Write(p)
	return
}

// Write implements the io.Writer interface for LogMonitor.
func (lm *LogMonitor) Write(p []byte) (n int, err error) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	stringLog := string(p)
	cleanLog := ansiRegex.ReplaceAllString(stringLog, "")

	for _, shutdownString := range lm.shutdownStrings {
		if strings.Contains(cleanLog, shutdownString) {
			lm.shutdownFn(fmt.Sprintf("Detected '%s' in logs", shutdownString))
			break
		}
	}
	return len(p), nil
}
