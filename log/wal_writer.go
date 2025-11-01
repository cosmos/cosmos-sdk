package log

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// defaultWalRollSizeBytes defines the default maximum size for a WAL segment
// before rolling to a new file. This is a minimal safeguard to prevent
// unbounded growth without adding user-facing configuration complexity.
const defaultWalRollSizeBytes int64 = 1 << 30 // 1 GiB

// walWriter is a minimal append-only writer for concatenated gzip members.
// It buffers writes in userspace and only touches the disk on buffer flushes.
// Durability is guaranteed when the caller invokes Sync() or Close().
type walWriter struct {
	mu sync.Mutex
	// dataDir: base data directory under which WAL files are stored.
	// Files are organized as: <dataDir>/wal/node-<nodeID>/<YYYY-MM-DD>/seg-XXXXXX.wal.gz
	dataDir string
	// nodeID: optional node identifier used in the WAL path. If empty, "default" is used.
	nodeID string

	// segment state
	f            *os.File      // current open segment file
	bw           *bufio.Writer // buffered writer for the segment
	createdAt    time.Time     // creation time of current segment
	sizeBytes    int64         // number of bytes written to current segment
	currentPath  string        // full path of current segment
	currentIndex int           // monotonically increasing index per day

	stopped bool // set to true after Close(); further writes are rejected

	// bufSize controls the internal bufio.Writer size for batching.
	bufSize int

	// rollSizeBytes triggers a segment rotation when the current segment size
	// reaches or exceeds this threshold. If zero, rolling is disabled.
	rollSizeBytes int64
}

// walWriterConfig holds the minimal configuration needed to open a WAL.
// If a field is zero-valued, newWalWriter applies a sensible default.
type walWriterConfig struct {
	// DataDir: base data directory. If empty, defaults to current directory (".").
	DataDir string
	// NodeID: optional node identifier for path names. If empty, "default" is used.
	NodeID string
	// BufSize: size in bytes for internal buffering. If zero, defaults to 1 MiB.
	BufSize int
}

// newWalWriter creates a WAL writer. It creates the target directory structure
// if missing. No background goroutines are started and no fsync policy is
// enforced by default; callers should call Sync() when they need durability
// (e.g., on shutdown).
func newWalWriter(cfg walWriterConfig) (*walWriter, error) {
	if cfg.DataDir == "" {
		cfg.DataDir = "."
	}
	if cfg.NodeID == "" {
		cfg.NodeID = "default"
	}
	w := &walWriter{
		dataDir: cfg.DataDir,
		nodeID:  cfg.NodeID,
	}
	if cfg.BufSize <= 0 {
		w.bufSize = 1 << 20 // 1 MiB default
	} else {
		w.bufSize = cfg.BufSize
	}
	// Apply a sane default roll size to avoid unbounded segment growth.
	w.rollSizeBytes = defaultWalRollSizeBytes
	// Ensure base wal directory exists.
	if err := os.MkdirAll(filepath.Join(cfg.DataDir, "wal"), 0o755); err != nil {
		return nil, err
	}
	if err := w.rotateLocked(true); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *walWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.stopped {
		return nil
	}
	w.stopped = true
	if w.bw != nil {
		if err := w.bw.Flush(); err != nil {
			return err
		}
	}
	if w.f != nil {
		if err := walSync(w.f); err != nil {
			return err
		}
		if err := w.f.Close(); err != nil {
			return err
		}
		// Make subsequent Sync() a no-op and avoid using a closed file.
		w.bw = nil
		w.f = nil
	}
	return nil
}

func (w *walWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.bw != nil {
		if err := w.bw.Flush(); err != nil {
			return err
		}
	}
	if w.f != nil {
		return walSync(w.f)
	}
	return nil
}

// AppendCompressed appends a gzip member (already compressed) to the WAL.
func (w *walWriter) AppendCompressed(member []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.stopped {
		return errors.New("wal: closed")
	}
	n, err := w.bw.Write(member)
	if err != nil {
		return err
	}
	w.sizeBytes += int64(n)
	if w.rollSizeBytes > 0 && w.sizeBytes >= w.rollSizeBytes {
		return w.rotateLocked(false)
	}
	return nil
}

func (w *walWriter) rotateLocked(first bool) error {
	if !first {
		if w.bw != nil {
			if err := w.bw.Flush(); err != nil {
				return err
			}
		}
		if w.f != nil {
			if err := walSync(w.f); err != nil {
				return err
			}
			if err := w.f.Close(); err != nil {
				return err
			}
		}
	}
	day := time.Now().UTC().Format("2006-01-02")
	dir := filepath.Join(w.dataDir, "wal", fmt.Sprintf("node-%s", w.nodeID), day)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	idx, err := nextSegmentIndex(dir)
	if err != nil {
		return err
	}
	w.currentIndex = idx
	path := filepath.Join(dir, fmt.Sprintf("seg-%06d.wal.gz", idx))
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	w.f = f
	w.bw = bufio.NewWriterSize(f, w.bufSize)
	w.createdAt = time.Now()
	w.sizeBytes = 0
	w.currentPath = path
	return nil
}

func nextSegmentIndex(dir string) (int, error) {
	d, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	max := 0
	for _, e := range d {
		name := e.Name()
		var n int
		_, err := fmt.Sscanf(name, "seg-%06d.wal.gz", &n)
		if err == nil && n > max {
			max = n
		}
	}
	return max + 1, nil
}
