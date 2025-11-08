package log

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// defaultWalRollSizeBytes defines the default maximum size for a WAL segment
// before rolling to a new file. This is a minimal safeguard to prevent
// unbounded growth without adding user-facing configuration complexity.
const defaultWalRollSizeBytes int64 = 1 << 30 // 1 GiB

// walDirName is the fixed directory name under dataDir used to store WAL files.
// Keep this as a single source of truth to avoid string typos.
const walDirName = "log.wal"

// walWriter is an append-only writer for concatenated gzip members with a
// sidecar index (.idx). Each AppendCompressed{,WithMeta} writes exactly one
// complete gzip member to the current segment (.wal.gz) and then emits a
// single NDJSON line to the index containing the byte-range and basic stats.
// It buffers writes in userspace and flushes the data before the index line so
// that any reader tailing the index can safely range-read the advertised bytes.
// Durability is guaranteed when the caller invokes Sync() or Close().
type walWriter struct {
	mu sync.Mutex
	// dataDir: base data directory under which WAL files are stored.
	// Files are organized as: <dataDir>/log.wal/node-<nodeID>/<YYYY-MM-DD>/seg-XXXXXX.wal.gz
	dataDir string
	// nodeID: optional node identifier used in the WAL path. If empty, "default" is used.
	nodeID string

	// segment state
	f            *os.File      // current open segment file (.wal.gz)
	bw           *bufio.Writer // buffered writer for the segment
	idxF         *os.File      // current index sidecar (.idx)
	idxBW        *bufio.Writer // buffered writer for the index
	createdAt    time.Time     // creation time of current segment
	sizeBytes    int64         // number of bytes written to current segment
	currentPath  string        // full path of current segment
	currentIndex int           // monotonically increasing index per day
	frameSeq     uint64        // frame sequence within current segment

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
	if err := os.MkdirAll(filepath.Join(cfg.DataDir, walDirName), 0o755); err != nil {
		return nil, err
	}
	// Lazy-open: do not create a segment until the first append.
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
	if w.idxBW != nil {
		if err := w.idxBW.Flush(); err != nil {
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
	if w.idxF != nil {
		if err := w.idxF.Close(); err != nil {
			return err
		}
		w.idxF = nil
		w.idxBW = nil
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
	if w.idxBW != nil {
		if err := w.idxBW.Flush(); err != nil {
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
	// Lazy-open the first segment on first write.
	if w.f == nil {
		if err := w.rotateLocked(true); err != nil {
			return err
		}
	}
	// Current offset before write
	off := w.sizeBytes
	n, err := w.bw.Write(member)
	if err != nil {
		return err
	}
	w.sizeBytes += int64(n)
	// Flush member to make it visible to readers before emitting index
	if w.idxBW != nil {
		if err := w.bw.Flush(); err != nil {
			return err
		}
		// Write minimal index line
		if err := w.writeIndexLineLocked(off, int64(n), 0, 0, 0, 0); err != nil {
			return err
		}
	}
	if w.rollSizeBytes > 0 && w.sizeBytes >= w.rollSizeBytes {
		return w.rotateLocked(false)
	}
	return nil
}

// AppendCompressedWithMeta appends a gzip member and records sidecar index
// with provided metadata.
//
// Semantics of metadata fields:
//   - recs: number of records in the uncompressed payload. In the mem logger
//     pipeline, payloads are newline-delimited JSON (NDJSON), so recs is the
//     count of '\n' bytes in the uncompressed buffer. If unknown, pass 0.
//   - firstTS/lastTS: application timestamps in Unix nanos associated with the
//     payload (e.g., first/last log record time). If unavailable, pass 0.
//   - crc32: IEEE CRC32 of the uncompressed payload (crc32.ChecksumIEEE). If
//     unknown, pass 0. This is useful for downstream integrity checks when
//     streaming the decompressed frame.
//
// Note: the index 'off' and 'len' written by the WAL always refer to COMPRESSED
// byte ranges within the .wal.gz file; readers must first slice [off, off+len)
// and then wrap that slice with a gzip reader to obtain the uncompressed bytes.
func (w *walWriter) AppendCompressedWithMeta(member []byte, recs uint32, firstTS, lastTS int64, crc32 uint32) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.stopped {
		return errors.New("wal: closed")
	}
	if w.f == nil {
		if err := w.rotateLocked(true); err != nil {
			return err
		}
	}
	off := w.sizeBytes
	n, err := w.bw.Write(member)
	if err != nil {
		return err
	}
	w.sizeBytes += int64(n)
	// Ensure bytes are flushed before advertising via index
	if w.idxBW != nil {
		if err := w.bw.Flush(); err != nil {
			return err
		}
		if err := w.writeIndexLineLocked(off, int64(n), recs, firstTS, lastTS, crc32); err != nil {
			return err
		}
	}
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
		if w.idxBW != nil {
			if err := w.idxBW.Flush(); err != nil {
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
		if w.idxF != nil {
			if err := w.idxF.Close(); err != nil {
				return err
			}
		}
	}
	day := time.Now().UTC().Format("2006-01-02")
	dir := filepath.Join(w.dataDir, walDirName, fmt.Sprintf("node-%s", w.nodeID), day)
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
	w.frameSeq = 0
	// open index sidecar
	idxPath := filepath.Join(dir, fmt.Sprintf("seg-%06d.wal.idx", idx))
	idxF, err := os.OpenFile(idxPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	w.idxF = idxF
	w.idxBW = bufio.NewWriterSize(idxF, 64*1024)
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

// writeIndexLineLocked writes a single NDJSON line describing the appended frame.
// Caller must hold w.mu. It assumes segment data has been flushed to file.
func (w *walWriter) writeIndexLineLocked(off int64, length int64, recs uint32, firstTS, lastTS int64, crc32 uint32) error {
	w.frameSeq++
	// Build NDJSON line manually to avoid json.Marshal overhead
	// {"file":"...","frame":N,"off":N,"len":N,"recs":N,"first_ts":N,"last_ts":N,"crc32":N}\n
	file := filepath.Base(w.currentPath)
	// Allocate a small scratch buffer; index lines are short
	buf := make([]byte, 0, 200)
	buf = append(buf, '{')
	buf = append(buf, '"', 'f', 'i', 'l', 'e', '"', ':', ' ')
	buf = strconv.AppendQuote(buf, file)
	buf = append(buf, ',', ' ', '"', 'f', 'r', 'a', 'm', 'e', '"', ':', ' ')
	buf = strconv.AppendUint(buf, w.frameSeq, 10)
	buf = append(buf, ',', ' ', '"', 'o', 'f', 'f', '"', ':', ' ')
	buf = strconv.AppendUint(buf, uint64(off), 10)
	buf = append(buf, ',', ' ', '"', 'l', 'e', 'n', '"', ':', ' ')
	buf = strconv.AppendUint(buf, uint64(length), 10)
	buf = append(buf, ',', ' ', '"', 'r', 'e', 'c', 's', '"', ':', ' ')
	buf = strconv.AppendUint(buf, uint64(recs), 10)
	buf = append(buf, ',', ' ', '"', 'f', 'i', 'r', 's', 't', '_', 't', 's', '"', ':', ' ')
	buf = strconv.AppendInt(buf, firstTS, 10)
	buf = append(buf, ',', ' ', '"', 'l', 'a', 's', 't', '_', 't', 's', '"', ':', ' ')
	buf = strconv.AppendInt(buf, lastTS, 10)
	buf = append(buf, ',', ' ', '"', 'c', 'r', 'c', '3', '2', '"', ':', ' ')
	buf = strconv.AppendUint(buf, uint64(crc32), 10)
	buf = append(buf, '}', '\n')
	if _, err := w.idxBW.Write(buf); err != nil {
		return err
	}
	return w.idxBW.Flush()
}
