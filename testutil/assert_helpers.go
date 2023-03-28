package testutil

import "testing"

func AssertPanics(t *testing.T, f func()) {
	t.Helper()
	panicked := false
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
	}()
	f()
	if !panicked {
		t.Errorf("should panic")
	}
}

func AssertNotPanics(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("should not panic: %v", r)
		}
	}()
	f()
}
