package simapp

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil/statecheck"
	"github.com/cosmos/cosmos-sdk/types/module"
)

var updateSchemaSnapshot = flag.Bool("update-schema-snapshot", false, "rewrite testdata/collections_schema.json from the current app")

const (
	schemaSnapshotPath       = "testdata/collections_schema.json"
	regenerateSchemaSnapshot = "go test -run TestCollectionsSchemaSnapshot -update-schema-snapshot ."
)

// moduleSchemaSnapshot is the store layout of one module: one line per
// collection with its name, prefix, key codec and value codec, and the
// module's ConsensusVersion.
type moduleSchemaSnapshot struct {
	ConsensusVersion uint64   `json:"consensus_version"`
	Collections      []string `json:"collections"`
}

// TestCollectionsSchemaSnapshot fails when the collections schema of a module
// differs from the golden snapshot, so every change to a collection prefix or
// codec shows up in review. If existing entries still decode with the new
// codecs, as with an added or renamed collection, regenerate the snapshot
// with
//
//	go test -run TestCollectionsSchemaSnapshot -update-schema-snapshot .
//
// Otherwise bump ConsensusVersion, register a migration that rewrites the
// existing entries, and regenerate the snapshot.
func TestCollectionsSchemaSnapshot(t *testing.T) {
	app := Setup(t, false)
	current := collectionsSchemaSnapshot(t, app)

	if *updateSchemaSnapshot {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")
		require.NoError(t, enc.Encode(current))
		require.NoError(t, os.WriteFile(filepath.Clean(schemaSnapshotPath), buf.Bytes(), 0o600))
		return
	}

	bz, err := os.ReadFile(schemaSnapshotPath)
	require.NoError(t, err, "create the snapshot with %s", regenerateSchemaSnapshot)
	var golden map[string]moduleSchemaSnapshot
	require.NoError(t, json.Unmarshal(bz, &golden))

	for name, cur := range current {
		old, ok := golden[name]
		if !ok {
			t.Errorf("module %s is missing from %s; regenerate with %s", name, schemaSnapshotPath, regenerateSchemaSnapshot)
			continue
		}
		if slices.Equal(old.Collections, cur.Collections) && old.ConsensusVersion == cur.ConsensusVersion {
			continue
		}
		if old.ConsensusVersion == cur.ConsensusVersion {
			t.Errorf("collections schema of module %s changed at ConsensusVersion %d\nsnapshot: %q\ncurrent:  %q\n"+
				"If existing entries still decode with the new codecs, regenerate with\n\t%s\n"+
				"Otherwise bump ConsensusVersion and register a migration that rewrites them.",
				name, cur.ConsensusVersion, old.Collections, cur.Collections, regenerateSchemaSnapshot)
			continue
		}
		t.Errorf("module %s moved from ConsensusVersion %d to %d; regenerate with %s",
			name, old.ConsensusVersion, cur.ConsensusVersion, regenerateSchemaSnapshot)
	}
	for name := range golden {
		if _, ok := current[name]; !ok {
			t.Errorf("module %s is in %s but no longer in the app; regenerate with %s", name, schemaSnapshotPath, regenerateSchemaSnapshot)
		}
	}
}

func collectionsSchemaSnapshot(t *testing.T, app *SimApp) map[string]moduleSchemaSnapshot {
	t.Helper()
	out := make(map[string]moduleSchemaSnapshot)
	for _, m := range app.CollectionsModules() {
		mod, ok := app.ModuleManager.Modules[m.Name].(module.HasConsensusVersion)
		require.True(t, ok, "module %s has no ConsensusVersion", m.Name)

		var colls []string
		for _, c := range m.Schema.ListCollections() {
			keyCodec, valueCodec := statecheck.CodecTypes(c)
			colls = append(colls, fmt.Sprintf("%s prefix=%X key=%s value=%s key_codec=%s value_codec=%s",
				c.GetName(), c.GetPrefix(), statecheck.KeyType(c), c.ValueCodec().ValueType(), keyCodec, valueCodec))
		}
		out[m.Name] = moduleSchemaSnapshot{ConsensusVersion: mod.ConsensusVersion(), Collections: colls}
	}
	return out
}
