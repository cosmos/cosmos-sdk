//go:build system_test

package systemtests

import (
	"net/url"
	"path/filepath"
	"testing"

	"github.com/creachadair/tomledit"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/cosmos/cosmos-sdk/tools/systemtests"
)

// TestBlockLogs enables per-block log capture, produces blocks at the default
// info log level and checks that debug lines from x/bank's injected keeper
// logger are served by the node Logs REST endpoint, and that pruned heights
// are gone.
func TestBlockLogs(t *testing.T) {
	const retain = 5

	sut := systemtests.Sut
	sut.ResetChain(t)
	for i := 0; i < sut.NodesCount(); i++ {
		appTomlPath := filepath.Join(sut.NodeDir(i), "config", "app.toml")
		systemtests.EditToml(appTomlPath, func(doc *tomledit.Document) {
			setInt(doc, retain, "block-logs", "retain-blocks")
		})
	}
	sut.StartChain(t)
	sut.AwaitNBlocks(t, retain+3)

	base := sut.APIAddress() + "/cosmos/base/node/v1beta1/logs"

	// x/mint mints every block through x/bank, whose keeper logs at debug with
	// the logger it received at construction, while the node runs at info.
	raw := systemtests.GetRequest(t, base+"?start=1&level=debug&module="+url.QueryEscape("x/bank"))
	entries := gjson.GetBytes(raw, "entries").Array()
	require.NotEmpty(t, entries, string(raw))
	minHeight, maxHeight := int64(1<<62), int64(0)
	for _, e := range entries {
		require.Equal(t, "x/bank", e.Get("module").String(), e.Raw)
		require.Equal(t, "debug", e.Get("level").String(), e.Raw)
		require.Equal(t, "minted coins from module account", e.Get("msg").String(), e.Raw)
		require.NotEmpty(t, e.Get("fields.amount").String(), e.Raw)
		h := e.Get("height").Int()
		minHeight, maxHeight = min(minHeight, h), max(maxHeight, h)
	}
	require.Greater(t, maxHeight, minHeight, "entries span several heights")
	require.Less(t, maxHeight-minHeight, int64(retain), "only the retention window is kept")
	require.Equal(t, maxHeight-retain+1, minHeight, "exactly the retention window is kept")

	// the module filter excludes untagged entries and the level filter is a minimum
	raw = systemtests.GetRequest(t, base+"?start=1&level=error&module="+url.QueryEscape("x/bank"))
	require.Empty(t, gjson.GetBytes(raw, "entries").Array(), string(raw))

	// pruned height: file is gone and the range is clipped, so nothing comes back
	raw = systemtests.GetRequest(t, base+"?start=1&end=1&level=debug")
	require.Empty(t, gjson.GetBytes(raw, "entries").Array(), string(raw))

	// pagination cursor round trip
	raw = systemtests.GetRequest(t, base+"?start=1&level=debug&pagination.limit=1")
	require.Len(t, gjson.GetBytes(raw, "entries").Array(), 1, string(raw))
	first := gjson.GetBytes(raw, "entries.0.time").String()
	nextKey := gjson.GetBytes(raw, "pagination.next_key").String()
	require.NotEmpty(t, nextKey, string(raw))
	raw = systemtests.GetRequest(t, base+"?start=1&level=debug&pagination.limit=1&pagination.key="+url.QueryEscape(nextKey))
	require.Len(t, gjson.GetBytes(raw, "entries").Array(), 1, string(raw))
	require.NotEqual(t, first, gjson.GetBytes(raw, "entries.0.time").String(), "cursor advanced")

	// invalid requests
	for _, q := range []string{"?start=1&end=999999", "?level=bogus", "?pagination.offset=3", "?pagination.reverse=true"} {
		systemtests.GetRequestWithHeaders(t, base+q, nil, 400)
	}
}
