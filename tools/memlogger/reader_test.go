package memlogger

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sdklog "github.com/cosmos/cosmos-sdk/log"
)

// Test end-to-end: write via MemLogger (WAL with framed gzip + .idx),
// then read via Reader and validate offsets and decompressed content.
func TestAgentReader_ConsumeFrames(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Create a mem logger configured to write WAL under <root>/data/...
	l, err := sdklog.NewMemLogger(sdklog.MemLoggerConfig{
		Interval:         50 * time.Millisecond,
		MemoryLimitBytes: 0, // time-based flush
		GzipLevel:        gzip.BestSpeed,
		P2pNodeId:        "agenttest",
		OutputDir:        root,
	})
	if err != nil {
		t.Fatalf("NewMemLogger: %v", err)
	}
	defer l.Close()

	// Produce some logs over a few ticks.
	const batches = 3
	const perBatch = 10
	for b := 0; b < batches; b++ {
		for i := 0; i < perBatch; i++ {
			// produce a structured event
			l.Info("evt", "batch", b, "i", i)
		}
		time.Sleep(60 * time.Millisecond)
	}

	// Force a durable flush to WAL.
	if err := l.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	// Locate the index file.
	idxPath := findLatestIdx(t, root, "agenttest")

	// Open the agent Reader on the index and iterate frames.
	var rdr Reader
	if err := rdr.OpenIndex(filepath.Dir(idxPath), filepath.Base(idxPath)); err != nil {
		t.Fatalf("open idx: %v", err)
	}
	defer rdr.Close()

	// Consume all available frames without blocking.
	totalLines := 0
	for {
		fm, rc, err := rdr.NextFrame()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("next frame: %v", err)
		}
		// Decompress and count lines; also verify JSONL shape quickly.
		zr, err := gzip.NewReader(rc)
		if err != nil {
			t.Fatalf("gzip reader: %v", err)
		}
		br := bufio.NewReader(zr)
		var lines int
		for {
			line, err := br.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Fatalf("read line: %v", err)
			}
			// basic JSON sanity check
			var m map[string]any
			if e := json.Unmarshal(bytes.TrimSpace(line), &m); e != nil {
				t.Fatalf("json: %v\n%s", e, string(line))
			}
			lines++
		}
		zr.Close()
		rc.Close()
		if fm.Recs != 0 && int(fm.Recs) != lines {
			t.Fatalf("recs mismatch: idx=%d counted=%d", fm.Recs, lines)
		}
		totalLines += lines
	}
	if totalLines != batches*perBatch {
		t.Fatalf("total lines mismatch: got=%d want=%d", totalLines, batches*perBatch)
	}
}

func findLatestIdx(t *testing.T, root, node string) string {
	t.Helper()
	base := filepath.Join(root, "data", "log.wal", "node-"+node)
	days, err := os.ReadDir(base)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	var latestDay string
	for _, d := range days {
		if d.IsDir() && len(d.Name()) == len("2006-01-02") && strings.Count(d.Name(), "-") == 2 {
			if d.Name() > latestDay {
				latestDay = d.Name()
			}
		}
	}
	if latestDay == "" {
		t.Fatalf("no day dir")
	}
	dayDir := filepath.Join(base, latestDay)
	ents, err := os.ReadDir(dayDir)
	if err != nil {
		t.Fatalf("readdir day: %v", err)
	}
	var latest string
	for _, e := range ents {
		if strings.HasSuffix(e.Name(), ".wal.idx") && e.Name() > latest {
			latest = e.Name()
		}
	}
	if latest == "" {
		t.Fatalf("no idx in day dir")
	}
	return filepath.Join(dayDir, latest)
}
