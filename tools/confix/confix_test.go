package confix_test

import (
	"context"
	"testing"

	"github.com/creachadair/tomledit"
	"gotest.tools/v3/assert"

	"cosmossdk.io/tools/confix"
)

func mustParseConfig(t *testing.T, path string) *tomledit.Document {
	doc, err := confix.LoadConfig(path)
	if err != nil {
		t.Fatalf("Loading config: %v", err)
	}
	return doc
}

func TestApplyFixes(t *testing.T) {
	ctx := context.Background()

	t.Run("Unknown", func(t *testing.T) {
		err := confix.ApplyFixes(ctx, mustParseConfig(t, "testdata/unknown-app.toml"))
		assert.ErrorContains(t, err, "unknown SDK version")
	})
	t.Run("Unknown Overrride", func(t *testing.T) {
		err := confix.ApplyFixes(ctx, mustParseConfig(t, "testdata/unknown-app.toml"))
		assert.NilError(t, err)
	})
	t.Run("OK", func(t *testing.T) {
		doc := mustParseConfig(t, "testdata/v45-config.toml")
		err := confix.ApplyFixes(ctx, doc)
		assert.NilError(t, err)

		// TODO
	})
}
