package iavl

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

type MultiTreeDescription struct {
	Version          uint64                              `json:"version"`
	Trees            map[string]internal.TreeDescription `json:"tree_descriptions"`
	LastPruneVersion uint64                              `json:"last_prune_version"`
}

type TreeDescription = internal.TreeDescription
type ChangesetDescription = internal.ChangesetDescription
type CheckpointInfo = internal.CheckpointInfo

func RenderHTML(w io.Writer, desc MultiTreeDescription) error {
	return descTemplate.Execute(w, desc)
}

// TODO: add a config flag to enable/disable the debug server instead of starting by default
func (db *CommitMultiTree) startDebugServer() {
	ln, err := net.Listen("tcp", "127.0.0.1:63789")
	if err != nil {
		logger.Error("failed to start IAVL debug server", "error", err)
		return
	}
	fmt.Printf("IAVL debug server started at http://%s\n", ln.Addr().String())
	logger.Info("IAVL debug server started", "addr", "http://"+ln.Addr().String())

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		desc := db.Describe()
		if err := RenderHTML(w, desc); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	go http.Serve(ln, mux)
}

func formatBytes(b int) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func rootRetained(cp CheckpointInfo) bool {
	rootID := cp.RootID
	if rootID.IsEmpty() {
		return true // empty tree is trivially "retained"
	}
	if rootID.Checkpoint() != cp.Checkpoint {
		return false // root is from a different checkpoint, can't confirm locally
	}
	var nsi internal.NodeSetInfo
	if rootID.IsLeaf() {
		nsi = cp.Leaves
	} else {
		nsi = cp.Branches
	}
	idx := rootID.Index()
	return idx >= nsi.StartIndex && idx <= nsi.EndIndex
}

func nodeCount(count any, structSize int) string {
	var n int
	switch v := count.(type) {
	case int:
		n = v
	case uint32:
		n = int(v)
	default:
		return fmt.Sprintf("%v", count)
	}
	return fmt.Sprintf("%d (%s)", n, formatBytes(n*structSize))
}

func toJSON(v any) template.JS {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	return template.JS(b)
}

var descTemplate = template.Must(template.New("desc").Funcs(template.FuncMap{
	"formatBytes":  formatBytes,
	"rootRetained": rootRetained,
	"nodeCount":    nodeCount,
	"toJSON":       toJSON,
	"sizeLeaf":     func() int { return int(internal.SizeLeaf) },
	"sizeBranch":   func() int { return int(internal.SizeBranch) },
}).Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>IAVL Tree Inspector</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: ui-monospace, "Cascadia Code", Menlo, monospace; font-size: 13px; background: #0d1117; color: #c9d1d9; padding: 20px; }
  h1 { font-size: 16px; color: #58a6ff; margin-bottom: 12px; }
  h2 { font-size: 14px; color: #79c0ff; margin: 16px 0 8px; }
  .meta { color: #8b949e; margin-bottom: 16px; }
  .meta span { color: #c9d1d9; }
  .tree { background: #161b22; border: 1px solid #30363d; border-radius: 6px; padding: 16px; margin-bottom: 16px; }
  .tree-header { display: flex; gap: 16px; align-items: baseline; margin-bottom: 12px; }
  .tree-name { font-size: 14px; color: #79c0ff; font-weight: bold; }
  .tree-meta { color: #8b949e; font-size: 12px; }
  .tree-meta span { color: #c9d1d9; }
  table { border-collapse: collapse; width: 100%; margin-bottom: 8px; }
  th { text-align: left; color: #8b949e; font-weight: normal; font-size: 11px; text-transform: uppercase; letter-spacing: 0.5px; padding: 4px 10px; border-bottom: 1px solid #30363d; }
  td { padding: 4px 10px; border-bottom: 1px solid #21262d; }
  .num { text-align: right; font-variant-numeric: tabular-nums; }
  .dim { color: #484f58; }
  .tag { display: inline-block; font-size: 10px; padding: 1px 5px; border-radius: 3px; }
  .tag-compacted { background: #1f3a1f; color: #56d364; }
  .tag-incomplete { background: #3d1f1f; color: #f85149; }
  details { margin-top: 4px; }
  summary { cursor: pointer; color: #8b949e; font-size: 12px; }
  summary:hover { color: #c9d1d9; }
  .cp-table { margin-top: 6px; font-size: 12px; }
  .cp-table th { font-size: 10px; }
  .cp-table td { padding: 3px 8px; }
  .empty { color: #484f58; font-style: italic; }
</style>
</head>
<body>
<h1>IAVL Tree Inspector</h1>
<div class="meta">
  Version: <span>{{.Version}}</span> &middot;
  Last Prune Version: <span>{{.LastPruneVersion}}</span>
</div>

{{range $name, $tree := .Trees}}
<div class="tree">
  <div class="tree-header">
    <span class="tree-name">{{$name}}</span>
    <span class="tree-meta">
      version <span>{{$tree.Version}}</span> &middot;
      root <span>{{$tree.RootID}}</span> &middot;
      checkpoint <span>{{$tree.LatestCheckpoint}}</span> (saved: <span>{{$tree.LatestSavedCheckpoint}}</span>) &middot;
      checkpoint version <span>{{$tree.LatestCheckpointVersion}}</span> &middot;
      size <span>{{formatBytes $tree.TotalBytes}}</span>
    </span>
  </div>

  {{if $tree.Changesets}}
  <table>
    <tr>
      <th>Versions</th>
      <th class="num">Leaves</th>
      <th class="num">Branches</th>
      <th class="num">WAL</th>
      <th class="num">KV</th>
      <th class="num">Total</th>
      <th class="num">Checkpoints</th>
    </tr>
    {{range $tree.Changesets}}
    <tr>
      <td>
        {{.StartVersion}}{{if .EndVersion}}&ndash;{{.EndVersion}}{{end}}
        {{if .Incomplete}} <span class="tag tag-incomplete">closed</span>
        {{else if .CompactedAt}} <span class="tag tag-compacted">compacted @{{.CompactedAt}}</span>
        {{end}}
      </td>
      <td class="num">{{if .Incomplete}}<span class="dim">&mdash;</span>{{else}}{{nodeCount .TotalLeaves (sizeLeaf)}}{{end}}</td>
      <td class="num">{{if .Incomplete}}<span class="dim">&mdash;</span>{{else}}{{nodeCount .TotalBranches (sizeBranch)}}{{end}}</td>
      <td class="num">{{if .Incomplete}}<span class="dim">&mdash;</span>{{else}}{{formatBytes .WALSize}}{{end}}</td>
      <td class="num">{{if .Incomplete}}<span class="dim">&mdash;</span>{{else}}{{formatBytes .KVLogSize}}{{end}}</td>
      <td class="num">{{if .Incomplete}}<span class="dim">&mdash;</span>{{else}}{{formatBytes .TotalBytes}}{{end}}</td>
      <td class="num">
        {{if .Incomplete}}<span class="dim">&mdash;</span>
        {{else if not .Checkpoints}}0
        {{else}}
        <details>
          <summary>{{len .Checkpoints}}</summary>
          <table class="cp-table">
            <tr>
              <th>CP</th>
              <th>Version</th>
              <th>Root</th>
              <th class="num">Leaves</th>
              <th>Leaf Idx</th>
              <th class="num">Branches</th>
              <th>Branch Idx</th>
            </tr>
            {{range .Checkpoints}}
            <tr>
              <td>{{.Checkpoint}}</td>
              <td>{{.Version}}</td>
              <td>{{if .RootID.IsEmpty}}<span class="empty">empty</span>{{else}}{{.RootID}} {{if rootRetained .}}&#x2705;{{else}}&#x274C;{{end}}{{end}}</td>
              <td class="num">{{nodeCount .Leaves.Count (sizeLeaf)}}</td>
              <td>{{.Leaves.StartIndex}}..{{.Leaves.EndIndex}}</td>
              <td class="num">{{nodeCount .Branches.Count (sizeBranch)}}</td>
              <td>{{.Branches.StartIndex}}..{{.Branches.EndIndex}}</td>
            </tr>
            {{end}}
          </table>
        </details>
        {{end}}
      </td>
    </tr>
    {{end}}
  </table>
  {{else}}
  <div class="empty">No changesets</div>
  {{end}}
</div>
{{end}}

<script type="application/json" id="raw-data">{{toJSON .}}</script>
</body>
</html>
`))
