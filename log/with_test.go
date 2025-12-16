package log

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
)

// this test ensures that when the With and WithContext methods are called,
// that the log wrapper is properly copied with all of its associated options
// otherwise, verbose mode will fail
func TestLoggerWith(t *testing.T) {
	logger := zerolog.New(&bytes.Buffer{})
	regularLevel := zerolog.WarnLevel
	verboseLevel := zerolog.InfoLevel
	filterWriter := &filterWriter{}
	wrapper := zeroLogWrapper{
		Logger:       &logger,
		regularLevel: regularLevel,
		verboseLevel: verboseLevel,
		filterWriter: filterWriter,
	}

	wrapper2 := wrapper.With("x", "y").(zeroLogWrapper)
	if wrapper2.filterWriter != filterWriter {
		t.Fatalf("expected filterWriter to be copied, but it was not")
	}
	if wrapper2.regularLevel != regularLevel {
		t.Fatalf("expected regularLevel to be copied, but it was not")
	}
	if wrapper2.verboseLevel != verboseLevel {
		t.Fatalf("expected verboseLevel to be copied, but it was not")
	}

	wrapper3 := wrapper.WithContext("a", "b").(zeroLogWrapper)
	if wrapper3.filterWriter != filterWriter {
		t.Fatalf("expected filterWriter to be copied, but it was not")
	}
	if wrapper3.regularLevel != regularLevel {
		t.Fatalf("expected regularLevel to be copied, but it was not")
	}
	if wrapper3.verboseLevel != verboseLevel {
		t.Fatalf("expected verboseLevel to be copied, but it was not")
	}
}
