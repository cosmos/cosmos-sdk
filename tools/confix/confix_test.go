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

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		path, want string
		force      bool
	}{
		{path: "testdata/non-app.toml", want: ""},
		{path: "testdata/unknown-app.toml", want: ""},
		{path: "testdata/v45-app.toml", want: "0.45"},
		{path: "testdata/v46-app.toml", want: "0.46"},
		{path: "testdata/v47-app.toml", want: "0.47"},
		{path: "testdata/unknown-app.toml", want: "next", force: true},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			got := confix.ExtractVersion(mustParseConfig(t, test.path), test.force)
			if got != test.want {
				t.Errorf("Wrong version: got %q, want %q", got, test.want)
			}
		})
	}
}

func TestApplyFixes(t *testing.T) {
	ctx := context.Background()

	t.Run("Unknown", func(t *testing.T) {
		err := confix.ApplyFixes(ctx, mustParseConfig(t, "testdata/unknown-app.toml"), false)
		assert.ErrorContains(t, err, "unknown SDK version")
	})
	t.Run("Unknown Overrride", func(t *testing.T) {
		err := confix.ApplyFixes(ctx, mustParseConfig(t, "testdata/unknown-app.toml"), true)
		assert.NilError(t, err)
	})
	t.Run("OK", func(t *testing.T) {
		doc := mustParseConfig(t, "testdata/v45-config.toml")
		err := confix.ApplyFixes(ctx, doc, false)
		assert.NilError(t, err)

		// TODO
	})
}
