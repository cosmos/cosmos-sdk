package log

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	gogoproto "github.com/cosmos/gogoproto/proto"
)

// MemLoggerConfig configures the in-memory compressing logger.
type MemLoggerConfig struct {
	// Interval controls how often the current buffer is compressed into a chunk.
	// If zero, defaults to 2 seconds.
	Interval time.Duration

	// MaxUncompressedBytes triggers an early compression when the current
	// uncompressed buffer grows beyond this size. If zero, no size-based
	// trigger is applied.
	MaxUncompressedBytes int

	// MaxChunks caps the number of retained compressed chunks in memory.
	// When exceeded and DropOldest is true, the oldest chunks are dropped.
	// If zero, chunks are unbounded.
	MaxChunks int

	// DropOldest defines the eviction policy when MaxChunks is exceeded.
	// If true, drops oldest chunks to make room; if false, keeps all
	// chunks (which may grow memory).
	DropOldest bool

	// GzipLevel controls compression level for gzip. If 0, uses gzip.DefaultCompression.
	GzipLevel int

	// OutputDir, when set, is used by Flush to write a timestamped
	// concatenated gzip file under this directory. If empty, Flush returns
	// an error.
	OutputDir string
}

// MemLogger implements Logger and buffers JSONL log events in memory.
// It periodically compresses the buffer into gzip chunks to limit growth,
// and can flush all chunks (concatenated gzip streams) to disk.
type MemLogger struct {
	agg *memAggregator
	ctx []any
}

// Ensure MemLogger implements the SDK Logger interface.
var _ Logger = (*MemLogger)(nil)

// NewMemLogger creates a new in-memory compressing logger with the given config.
func NewMemLogger(cfg MemLoggerConfig) Logger {
	if cfg.Interval == 0 {
		cfg.Interval = 2 * time.Second
	}
	if cfg.GzipLevel == 0 {
		// Favor speed to minimize runtime overhead.
		cfg.GzipLevel = gzip.BestSpeed
	}
	agg := newMemAggregator(cfg)
	return &MemLogger{agg: agg}
}

// Info logs a message at level info.
func (l *MemLogger) Info(msg string, keyVals ...any) { l.agg.append("info", l.ctx, msg, keyVals...) }

// Warn logs a message at level warn.
func (l *MemLogger) Warn(msg string, keyVals ...any) { l.agg.append("warn", l.ctx, msg, keyVals...) }

// Error logs a message at level error.
func (l *MemLogger) Error(msg string, keyVals ...any) { l.agg.append("error", l.ctx, msg, keyVals...) }

// Debug logs a message at level debug.
func (l *MemLogger) Debug(msg string, keyVals ...any) { l.agg.append("debug", l.ctx, msg, keyVals...) }

// With returns a child logger that adds the provided keyvals to each event.
func (l *MemLogger) With(keyVals ...any) Logger {
	// copy context defensively
	newCtx := make([]any, 0, len(l.ctx)+len(keyVals))
	newCtx = append(newCtx, l.ctx...)
	newCtx = append(newCtx, keyVals...)
	return &MemLogger{agg: l.agg, ctx: newCtx}
}

// Impl returns the underlying implementation (self).
func (l *MemLogger) Impl() any { return l }

// FlushToWriter writes all compressed chunks followed by the current in-flight
// buffer (as a final gzip chunk) to the provided writer. The output is a valid
// sequence of concatenated gzip streams.
func (l *MemLogger) FlushToWriter(w io.Writer) error { return l.agg.flushToWriter(w) }

// FlushToFile writes the compressed stream sequence to the given file path.
// If the file exists, it will be truncated.
func (l *MemLogger) FlushToFile(path string) error { return l.agg.flushToFile(path) }

// Close stops the background compressor goroutine. It does not flush.
func (l *MemLogger) Close() error { l.agg.close(); return nil }

// DumpUncompressed writes all logs as plain JSONL without compression.
func (l *MemLogger) DumpUncompressed(w io.Writer) error { return l.agg.dumpUncompressed(w) }

// Flush writes the concatenated gzip stream to a timestamped file under
// OutputDir. Returns an error if OutputDir is empty or on write failure.
func (l *MemLogger) Flush() error {
	dir := l.agg.cfg.OutputDir
	if dir == "" {
		return errors.New("memlogger: OutputDir is not set")
	}
	name := time.Now().UTC().Format("memlog-20060102T150405.000000000Z.gz")
	p := dir + string(os.PathSeparator) + name
	return l.FlushToFile(p)
}

// ---- Internal aggregator ----

type memAggregator struct {
	cfg MemLoggerConfig

	mu   sync.Mutex
	buf  bytes.Buffer // current uncompressed JSONL buffer
	chks [][]byte     // compressed gzip chunks (each is a full gzip member)

	tick   *time.Ticker
	stopCh chan struct{}
	wg     sync.WaitGroup

	// compression pipeline
	workCh chan []byte
	compWg sync.WaitGroup

	// pools to reduce allocations during compression
	gzPool  sync.Pool // *gzip.Writer
	bufPool sync.Pool // *bytes.Buffer
}

// event is encoded to a single flat JSON object similar to TM JSON logger
// output shape: top-level keys include "ts", "level", and "_msg",
// plus any contextual keyvals, without nesting.
type event map[string]any

func newMemAggregator(cfg MemLoggerConfig) *memAggregator {
	m := &memAggregator{
		cfg:     cfg,
		tick:    time.NewTicker(cfg.Interval),
		stopCh:  make(chan struct{}),
		workCh:  make(chan []byte, 2),
		gzPool:  sync.Pool{New: func() any { w, _ := gzip.NewWriterLevel(io.Discard, cfg.GzipLevel); return w }},
		bufPool: sync.Pool{New: func() any { return new(bytes.Buffer) }},
	}
	m.wg.Add(1)
	go m.run()
	m.compWg.Add(1)
	go m.compressor()
	return m
}

func (m *memAggregator) run() {
	defer m.wg.Done()
	for {
		select {
		case <-m.tick.C:
			m.enqueueCurrentBuffer()
		case <-m.stopCh:
			return
		}
	}
}

func (m *memAggregator) append(level string, ctx []any, msg string, keyvals ...any) {
	// Build flat event to mirror go-kit JSON logger format used by TMJSONLogger.
	ev := make(event, 4+len(ctx)+len(keyvals))
	ev["ts"] = time.Now().UTC()
	ev["level"] = level
	ev["_msg"] = msg

	// merge ctx + keyvals into top-level fields (pairwise)
	merged := make([]any, 0, len(ctx)+len(keyvals))
	merged = append(merged, ctx...)
	merged = append(merged, keyvals...)
	for i := 0; i < len(merged); i += 2 {
		var key string
		if i < len(merged) {
			if ks, ok := merged[i].(string); ok {
				key = ks
			} else {
				key = toString(merged[i])
			}
		}
		var val any
		if i+1 < len(merged) {
			val = normalizeValue(merged[i+1])
		} else {
			val = "<missing>"
		}
		ev[key] = val
	}

	// encode as JSONL
	b, _ := json.Marshal(ev)
	b = append(b, '\n')

	m.mu.Lock()
	_, _ = m.buf.Write(b)

	// Size-based early compression: swap buffer and enqueue to background worker.
	var toCompress []byte
	if max := m.cfg.MaxUncompressedBytes; max > 0 && m.buf.Len() >= max {
		toCompress = m.takeBufferLocked()
	}
	m.mu.Unlock()

	if len(toCompress) > 0 {
		m.enqueueData(toCompress)
	}
}

func (m *memAggregator) enqueueCurrentBuffer() {
	// Swap current buffer if any and enqueue it for compression.
	m.mu.Lock()
	if m.buf.Len() == 0 {
		m.mu.Unlock()
		return
	}
	data := m.takeBufferLocked()
	m.mu.Unlock()

	m.enqueueData(data)
}

// takeBufferLocked swaps out the current uncompressed buffer and returns its bytes.
// Caller must hold m.mu.
func (m *memAggregator) takeBufferLocked() []byte {
	data := m.buf.Bytes()
	out := make([]byte, len(data))
	copy(out, data)
	m.buf.Reset()
	return out
}

// addChunkLocked appends a compressed chunk and enforces the capacity policy.
// Caller must hold m.mu.
func (m *memAggregator) addChunkLocked(chunk []byte) {
	m.chks = append(m.chks, chunk)
	if m.cfg.MaxChunks > 0 && len(m.chks) > m.cfg.MaxChunks && m.cfg.DropOldest {
		overflow := len(m.chks) - m.cfg.MaxChunks
		m.chks = append([][]byte{}, m.chks[overflow:]...)
	}
}

func (m *memAggregator) enqueueData(data []byte) {
	// Block to preserve backpressure but outside locks; this avoids dropping logs
	// and keeps compression off the hot path of logging.
	m.workCh <- data
}

func (m *memAggregator) compressor() {
	defer m.compWg.Done()
	for data := range m.workCh {
		chunk, err := m.gzipWithPool(data)
		if err != nil {
			// Best-effort: if compression fails, re-append uncompressed to current buffer.
			m.mu.Lock()
			_, _ = m.buf.Write(data)
			m.mu.Unlock()
			continue
		}
		m.mu.Lock()
		m.addChunkLocked(chunk)
		m.mu.Unlock()
	}
}

func (m *memAggregator) gzipWithPool(in []byte) ([]byte, error) {
	// get pooled buffer and writer
	b := m.bufPool.Get().(*bytes.Buffer)
	b.Reset()
	var out []byte
	w := m.gzPool.Get().(*gzip.Writer)
	w.Reset(b)
	if _, err := w.Write(in); err != nil {
		_ = w.Close()
		m.gzPool.Put(w)
		b.Reset()
		m.bufPool.Put(b)
		return nil, err
	}
	if err := w.Close(); err != nil {
		m.gzPool.Put(w)
		b.Reset()
		m.bufPool.Put(b)
		return nil, err
	}
	m.gzPool.Put(w)
	// Copy bytes to detach from pooled buffer before putting it back.
	out = make([]byte, b.Len())
	copy(out, b.Bytes())
	b.Reset()
	m.bufPool.Put(b)
	return out, nil
}

func (m *memAggregator) flushToWriter(w io.Writer) error {
	// Compress any pending buffer into a final chunk.
	m.mu.Lock()
	var data []byte
	if m.buf.Len() > 0 {
		data = m.takeBufferLocked()
	}
	// Snapshot existing chunks.
	chunks := make([][]byte, len(m.chks))
	copy(chunks, m.chks)
	m.mu.Unlock()

	if len(data) > 0 {
		gzChunk, err := m.gzipWithPool(data)
		if err != nil {
			// If compression fails during flush, return error.
			return err
		}
		chunks = append(chunks, gzChunk)
	}

	for _, c := range chunks {
		if _, err := w.Write(c); err != nil {
			return err
		}
	}
	return nil
}

func (m *memAggregator) flushToFile(path string) error {
	if path == "" {
		return errors.New("empty path")
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return m.flushToWriter(f)
}

func (m *memAggregator) dumpUncompressed(w io.Writer) error {
	// Snapshot chunks and current buffer.
	m.mu.Lock()
	chunks := make([][]byte, len(m.chks))
	copy(chunks, m.chks)
	bufCopy := make([]byte, m.buf.Len())
	copy(bufCopy, m.buf.Bytes())
	m.mu.Unlock()

	// Expand compressed chunks.
	for _, c := range chunks {
		if err := ungzipTo(c, w); err != nil {
			return err
		}
	}
	// Write the pending uncompressed tail.
	if len(bufCopy) > 0 {
		if _, err := w.Write(bufCopy); err != nil {
			return err
		}
	}
	return nil
}

func (m *memAggregator) close() {
	// Stop periodic enqueues.
	close(m.stopCh)
	m.tick.Stop()
	m.wg.Wait()

	// Enqueue any remaining buffer before shutting down compressor.
	m.mu.Lock()
	if m.buf.Len() > 0 {
		data := m.takeBufferLocked()
		m.mu.Unlock()
		// Best effort: enqueue; if blocked, still wait — we are shutting down.
		m.workCh <- data
	} else {
		m.mu.Unlock()
	}

	// Stop compressor and wait.
	close(m.workCh)
	m.compWg.Wait()
}

// ---- helpers ----

func ungzipTo(in []byte, w io.Writer) error {
	zr, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		return err
	}
	defer zr.Close()
	if _, err := io.Copy(w, zr); err != nil {
		return err
	}
	return nil
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	case time.Time:
		return t.Format(time.RFC3339Nano)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return "<unstringable>"
		}
		return string(b)
	}
}

// normalizeValue attempts to match the behavior of the TM JSON logger by
// stringifying values that are best represented via String(), while leaving
// proto messages and JSON-marshable types as structured values.
func normalizeValue(v any) any {
	switch t := v.(type) {
	case time.Time:
		return t
	case json.Marshaler:
		// Give precedence to explicit JSON marshaling implementations.
		return t
	case gogoproto.Message:
		// Keep protobuf messages structured to preserve fields.
		return t
	case fmt.Stringer:
		// Use the string form for types like MConnection, LazySprintf, Peer, etc.
		return t.String()
	default:
		return v
	}
}
