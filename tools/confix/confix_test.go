package confix_test

import (
	"testing"

	"cosmossdk.io/tools/confix"
	"github.com/creachadair/tomledit"
)

func mustParseConfig(t *testing.T, path string) *tomledit.Document {
	doc, err := confix.LoadConfig(path)
	if err != nil {
		t.Fatalf("Loading config: %v", err)
	}
	return doc
}

func TestUpgrade(t *testing.T) {
}

func TestLoadConfig(t *testing.T) {
}

func TestCheckValid(t *testing.T) {
}
