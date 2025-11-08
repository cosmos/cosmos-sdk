package memlogger

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Writer writes concatenated gzip members and a sidecar index for frame boundaries.
// It is safe for single-writer usage. External synchronization is required if used from multiple goroutines.
type Writer struct {
	mu     sync.Mutex
	dir    string
	prefix string

	// rotation
	maxBytes    int64
	maxInterval time.Duration
	gzipLevelV  int

	// current file state
	gzFile   *os.File
	idxFile  *os.File
	bw       *bufio.Writer // for idx
	fileName string
	openedAt time.Time
	curSize  int64 // bytes written to gz
	roll     uint64
	frameSeq uint64
}

// Options configures Writer.
type Options struct {
	Dir         string        // directory to place files
	Prefix      string        // file name prefix
	MaxBytes    int64         // rotate when gz bytes exceed this (0 = disabled)
	MaxInterval time.Duration // rotate after duration since open (0 = disabled)
	GzipLevel   int           // gzip level (use gzip.BestSpeed for low CPU). 0 => default
}

// FrameMeta describes a written frame (also used for idx JSON lines).
//
// Field semantics:
//   - File: segment (.wal.gz) base name containing the frame.
//   - Frame: 1-based sequence number within the segment.
//   - Off/Len: byte offset and length of the COMPRESSED gzip member in File.
//     Consumers must range-read [off, off+len) and then use gzip to decompress.
//   - Recs: number of records in the UNCOMPRESSED payload. In the mem logger
//     pipeline, payloads are NDJSON; this is typically the count of '\n'.
//   - FirstTS/LastTS: application timestamps in Unix nanos (optional).
//   - CRC32: IEEE CRC32 (crc32.ChecksumIEEE) of the UNCOMPRESSED payload (optional).
type FrameMeta struct {
	File    string `json:"file"`
	Frame   uint64 `json:"frame"`
	Off     uint64 `json:"off"`
	Len     uint64 `json:"len"`
	Recs    uint32 `json:"recs"`
	FirstTS int64  `json:"first_ts"`
	LastTS  int64  `json:"last_ts"`
	CRC32   uint32 `json:"crc32"`
}

var (
	errClosed = errors.New("memlogger: writer closed")
)

// New creates a new Writer. It will lazily open the first file on first frame write.
func New(opts Options) (*Writer, error) {
	if opts.Dir == "" {
		return nil, fmt.Errorf("Dir required")
	}
	if opts.Prefix == "" {
		opts.Prefix = "wal"
	}
	return &Writer{
		dir:         opts.Dir,
		prefix:      opts.Prefix,
		maxBytes:    opts.MaxBytes,
		maxInterval: opts.MaxInterval,
		gzipLevelV:  opts.GzipLevel,
	}, nil
}

// Close closes any open files. Safe to call multiple times.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	var err error
	if w.bw != nil {
		if e := w.bw.Flush(); err == nil {
			err = e
		}
	}
	if w.idxFile != nil {
		if e := w.idxFile.Close(); err == nil {
			err = e
		}
		w.idxFile = nil
	}
	if w.gzFile != nil {
		if e := w.gzFile.Close(); err == nil {
			err = e
		}
		w.gzFile = nil
	}
	return err
}

// WriteFrame compresses payload as a standalone gzip member, appends it to the current .gz file,
// and writes a corresponding index line. It returns the FrameMeta for the written frame.
// reccount can be 0 if unknown; firstTS/lastTS should be unix nanos if available.
func (w *Writer) WriteFrame(payload []byte, recCount uint32, firstTS, lastTS int64) (FrameMeta, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Ensure files are open and rotation is respected
	if err := w.rotateIfNeededLocked(int64(len(payload))); err != nil {
		return FrameMeta{}, err
	}
	if w.gzFile == nil || w.idxFile == nil {
		if err := w.openNewLocked(); err != nil {
			return FrameMeta{}, err
		}
	}

	// Compress into a buffer to form a complete gzip member
	var buf bytes.Buffer
	var gz *gzip.Writer
	if wlevel := gzip.DefaultCompression; true {
		// choose level from options if provided
		if wlevelOpt := w.gzipLevel(); wlevelOpt != gzip.DefaultCompression {
			wlevel = wlevelOpt
		}
		gw, err := gzip.NewWriterLevel(&buf, wlevel)
		if err != nil {
			return FrameMeta{}, err
		}
		gz = gw
	}
	// Optional: annotate member name/comment for debugging
	gz.Name = fmt.Sprintf("frame-%d", w.frameSeq+1)
	if _, err := gz.Write(payload); err != nil {
		_ = gz.Close()
		return FrameMeta{}, err
	}
	if err := gz.Close(); err != nil {
		return FrameMeta{}, err
	}
	member := buf.Bytes()

	// Prepare metadata
	off := w.curSize
	length := int64(len(member))
	crc := crc32.ChecksumIEEE(payload)
	fm := FrameMeta{
		File:    w.fileName,
		Frame:   w.frameSeq + 1,
		Off:     uint64(off),
		Len:     uint64(length),
		Recs:    recCount,
		FirstTS: firstTS,
		LastTS:  lastTS,
		CRC32:   crc,
	}

	// Append member to gz file
	if _, err := w.gzFile.Write(member); err != nil {
		return FrameMeta{}, err
	}
	w.curSize += length
	w.frameSeq++

	// Append index line (NDJSON) and flush the buffered writer (cheap I/O)
	line, err := json.Marshal(fm)
	if err != nil {
		return FrameMeta{}, err
	}
	if _, err := w.bw.Write(line); err != nil {
		return FrameMeta{}, err
	}
	if err := w.bw.WriteByte('\n'); err != nil {
		return FrameMeta{}, err
	}
	// Keep idx durable enough without heavy fsync; caller can fsync on rotation if needed.
	if err := w.bw.Flush(); err != nil {
		return FrameMeta{}, err
	}

	return fm, nil
}

// rotateIfNeededLocked checks rotation conditions; if needed, closes current files and opens new pair.
func (w *Writer) rotateIfNeededLocked(incoming int64) error {
	// If no files yet, nothing to rotate.
	if w.gzFile == nil {
		return nil
	}
	should := false
	if w.maxBytes > 0 && w.curSize+incoming > w.maxBytes {
		should = true
	}
	if !should && w.maxInterval > 0 && time.Since(w.openedAt) >= w.maxInterval {
		should = true
	}
	if !should {
		return nil
	}
	// Close current pair
	if w.bw != nil {
		_ = w.bw.Flush()
	}
	if w.idxFile != nil {
		_ = w.idxFile.Sync()
		_ = w.idxFile.Close()
		w.idxFile = nil
		w.bw = nil
	}
	if w.gzFile != nil {
		_ = w.gzFile.Sync()
		_ = w.gzFile.Close()
		w.gzFile = nil
	}
	// Open new
	return w.openNewLocked()
}

func (w *Writer) openNewLocked() error {
	now := time.Now().UTC()
	ts := now.Format("20060102T150405Z")
	w.roll++
	base := fmt.Sprintf("%s-%s-%03d", w.prefix, ts, w.roll)
	gzPath := filepath.Join(w.dir, base+".gz")
	idxPath := filepath.Join(w.dir, base+".idx")

	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return err
	}
	gz, err := os.OpenFile(gzPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	idx, err := os.OpenFile(idxPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		_ = gz.Close()
		return err
	}
	w.gzFile = gz
	w.idxFile = idx
	w.bw = bufio.NewWriterSize(idx, 64*1024)
	w.fileName = filepath.Base(gzPath)
	w.curSize = 0
	w.frameSeq = 0
	w.openedAt = now
	return nil
}

// CurrentFile returns the current gz file name (base) and bytes written.
func (w *Writer) CurrentFile() (string, int64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.fileName, w.curSize
}

func (w *Writer) gzipLevel() int {
	if w.gzipLevelV == 0 {
		return gzip.DefaultCompression
	}
	return w.gzipLevelV
}
