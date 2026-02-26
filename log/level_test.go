package log_test

import (
	"testing"

	"cosmossdk.io/log"
)

func TestParseLogLevel(t *testing.T) {
	_, err := log.ParseLogLevel("")
	if err == nil {
		t.Errorf("expected error for empty log level, got nil")
	}

	level := "consensus:foo,mempool:debug,*:error"
	_, err = log.ParseLogLevel(level)
	if err == nil {
		t.Errorf("expected error for invalid log level foo in log level list [consensus:foo mempool:debug *:error], got nil")
	}

	level = "consensus:debug,mempool:debug,*:error"
	filter, err := log.ParseLogLevel(level)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if filter == nil {
		t.Fatalf("expected non-nil filter, got nil")
	}

	if filter("consensus", "debug") {
		t.Errorf("expected filter to return false for consensus:debug")
	}
	if filter("consensus", "info") {
		t.Errorf("expected filter to return false for consensus:info")
	}
	if filter("consensus", "error") {
		t.Errorf("expected filter to return false for consensus:error")
	}
	if filter("mempool", "debug") {
		t.Errorf("expected filter to return false for mempool:debug")
	}
	if filter("mempool", "info") {
		t.Errorf("expected filter to return false for mempool:info")
	}
	if filter("mempool", "error") {
		t.Errorf("expected filter to return false for mempool:error")
	}
	if filter("state", "error") {
		t.Errorf("expected filter to return false for state:error")
	}
	if filter("server", "panic") {
		t.Errorf("expected filter to return false for server:panic")
	}

	if !filter("server", "debug") {
		t.Errorf("expected filter to return true for server:debug")
	}
	if !filter("state", "debug") {
		t.Errorf("expected filter to return true for state:debug")
	}
	if !filter("state", "info") {
		t.Errorf("expected filter to return true for state:info")
	}

	level = "error"
	filter, err = log.ParseLogLevel(level)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if filter == nil {
		t.Fatalf("expected non-nil filter, got nil")
	}

	if filter("state", "error") {
		t.Errorf("expected filter to return false for state:error")
	}
	if filter("consensus", "error") {
		t.Errorf("expected filter to return false for consensus:error")
	}

	if !filter("consensus", "debug") {
		t.Errorf("expected filter to return true for consensus:debug")
	}
	if !filter("consensus", "info") {
		t.Errorf("expected filter to return true for consensus:info")
	}
	if !filter("state", "debug") {
		t.Errorf("expected filter to return true for state:debug")
	}
}
