package indexer

import "testing"

func TestRegister(t *testing.T) {
	Register("test", func(params InitParams) (InitResult, error) {
		return InitResult{}, nil
	})

	if Lookup("test") == nil {
		t.Fatalf("expected to find indexer")
	}

	if Lookup("test2") != nil {
		t.Fatalf("expected not to find indexer")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected to panic")
		}
	}()
	Register("test", func(params InitParams) (InitResult, error) {
		return InitResult{}, nil
	})
}
