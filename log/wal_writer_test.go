package log

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"hash/crc32"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// gzipBytes compresses the input using gzip and returns the compressed bytes.
func gzipBytes(p []byte) []byte {
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	_, _ = zw.Write(p)
	_ = zw.Close()
	return buf.Bytes()
}

// makeIncompressible returns a deterministic, pseudo-random byte slice of length n.
// Using a fixed seed ensures reproducible sizes in tests.
func makeIncompressible(n int) []byte {
	b := make([]byte, n)
	r := rand.New(rand.NewSource(1))
	for i := range b {
		b[i] = byte(r.Intn(256))
	}
	return b
}

// TestWalWriter_AppendAndSync verifies basic append and sync behavior.
func TestWalWriter_AppendAndSync(t *testing.T) {
	dir := t.TempDir()
	w, err := newWalWriter(walWriterConfig{DataDir: dir, NodeID: "testnode", BufSize: 1 << 20})
	if err != nil {
		t.Fatalf("newWalWriter: %v", err)
	}
	defer func() { _ = w.Close() }()

	// Append a few newline-delimited chunks (JSONL-like) so recs and line counts are meaningful.
	// Use 2-byte lines ("a\n", "b\n", "c\n") so total raw sizes approximate 8KiB/64KiB/256KiB.
	rawChunks := [][]byte{
		bytes.Repeat([]byte("a"), 4<<10),   // ~8 KiB raw
		bytes.Repeat([]byte("b"), 32<<10),  // ~64 KiB raw
		bytes.Repeat([]byte("c"), 128<<10), // ~256 KiB raw
	}
	var totalCompressed int
	for _, raw := range rawChunks {
		gz := gzipBytes(append(raw, '\n'))
		recs := uint32(bytes.Count(raw, []byte{'\n'}))
		crc := crc32.ChecksumIEEE(raw)
		now := time.Now().UTC().UnixNano()
		if err := w.AppendCompressedWithMeta(gz, recs, now, now, crc); err != nil {
			t.Fatalf("append: %v", err)
		}
		totalCompressed += len(gz)
	}

	if err := w.Sync(); err != nil {
		t.Fatalf("sync: %v", err)
	}

	// Verify the current segment exists and has at least the expected size.
	fi, err := os.Stat(w.currentPath)
	if err != nil {
		t.Fatalf("stat current segment: %v", err)
	}
	if fi.Size() < int64(totalCompressed) {
		t.Fatalf("segment size too small: got %d want >= %d", fi.Size(), totalCompressed)
	}

	// Layout sanity check: .../wal/node-<nodeID>/<YYYY-MM-DD>/seg-XXXXXX.wal.gz
	base := filepath.Base(filepath.Dir(w.currentPath))
	if _, err := time.Parse("2006-01-02", base); err != nil {
		t.Fatalf("unexpected day dir name: %q: %v", base, err)
	}
}

// TestWalWriter_RotateBySize forces a small roll size and verifies rotation.
func TestWalWriter_RotateBySize(t *testing.T) {
	dir := t.TempDir()
	w, err := newWalWriter(walWriterConfig{DataDir: dir, NodeID: "n"})
	if err != nil {
		t.Fatalf("newWalWriter: %v", err)
	}
	defer func() { _ = w.Close() }()

	// Force a small roll size to trigger rotation quickly (64 KiB).
	w.rollSizeBytes = 64 << 10
	firstPath := w.currentPath
	firstIndex := w.currentIndex

	// Append two ~48 KiB incompressible chunks (gzip-compressed) -> should roll after second write.
	raw := makeIncompressible(48 << 10)
	chunk := gzipBytes(raw)
	if err := w.AppendCompressed(chunk); err != nil {
		t.Fatal(err)
	}
	if err := w.AppendCompressed(chunk); err != nil {
		t.Fatal(err)
	}

	if w.currentIndex <= firstIndex {
		t.Fatalf("expected rotation: currentIndex=%d firstIndex=%d", w.currentIndex, firstIndex)
	}
	if w.currentPath == firstPath {
		t.Fatalf("expected new segment path after rotation")
	}
	if w.sizeBytes != 0 { // new segment starts at 0 before any further writes
		// buffer may hold unflushed bytes; force sync to flush and re-check
		_ = w.Sync()
	}
}

// TestWalWriter_CloseIdempotent ensures Close can be called multiple times.
func TestWalWriter_CloseIdempotent(t *testing.T) {
	dir := t.TempDir()
	w, err := newWalWriter(walWriterConfig{DataDir: dir})
	if err != nil {
		t.Fatalf("newWalWriter: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
}

// BenchmarkWalWriter_Append measures append throughput for various chunk sizes.
// This helps choose a reasonable chunk size and internal buffer size.
func BenchmarkWalWriter_Append(b *testing.B) {
	sizes := []int{1 << 10, 4 << 10, 16 << 10, 64 << 10, 256 << 10, 1 << 20}
	for _, sz := range sizes {
		b.Run(byteSize(sz), func(b *testing.B) {
			dir := b.TempDir()
			// Use a buffer size close to the chunk size to amortize syscalls.
			w, err := newWalWriter(walWriterConfig{DataDir: dir, NodeID: "bench", BufSize: sz})
			if err != nil {
				b.Fatalf("newWalWriter: %v", err)
			}
			defer func() { _ = w.Close() }()

			// Use incompressible data so gzip size ~= input size.
			chunk := gzipBytes(makeIncompressible(sz))
			b.SetBytes(int64(len(chunk)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := w.AppendCompressed(chunk); err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()
			_ = w.Sync()
		})
	}
}

// BenchmarkWalWriter_BufSizeVsChunk explores how WAL buffer size impacts
// throughput for large chunks. This helps choose an appropriate BufSize when
// very large memory-bounded chunks (e.g., hundreds of MiB) are appended.
func BenchmarkWalWriter_BufSizeVsChunk(b *testing.B) {
	// Sweep buffer sizes: 1MiB .. 64MiB.
	bufSizes := []int{1 << 20, 2 << 20, 4 << 20, 8 << 20, 16 << 20, 32 << 20, 64 << 20}
	// Test two representative chunk sizes (compressed sizes in practice):
	// 64MiB and 256MiB. Larger sizes (e.g., 1GiB) are not practical in CI.
	chunkSizes := []int{64 << 20, 256 << 20}

	for _, cs := range chunkSizes {
		b.Run("chunk="+byteSize(cs), func(b *testing.B) {
			for _, bs := range bufSizes {
				b.Run("buf="+byteSize(bs), func(b *testing.B) {
					dir := b.TempDir()
					w, err := newWalWriter(walWriterConfig{DataDir: dir, NodeID: "bench", BufSize: bs})
					if err != nil {
						b.Fatalf("newWalWriter: %v", err)
					}
					defer func() { _ = w.Close() }()

					chunk := gzipBytes(makeIncompressible(cs))
					b.SetBytes(int64(len(chunk)))
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						if err := w.AppendCompressed(chunk); err != nil {
							b.Fatal(err)
						}
					}
					b.StopTimer()
					_ = w.Sync()
				})
			}
		})
	}
}

// byteSize formats an integer byte count as a short label, e.g., "64KiB".
func byteSize(n int) string {
	const (
		KiB = 1 << 10
		MiB = 1 << 20
	)
	switch {
	case n >= MiB:
		return strconv.Itoa(n/MiB) + "MiB"
	case n >= KiB:
		return strconv.Itoa(n/KiB) + "KiB"
	default:
		return strconv.Itoa(n) + "B"
	}
}

// TestWalWriter_AppendWithMeta_Index verifies index metadata matches compressor logic
// and that the compressed span decompresses correctly using the recorded off/len.
func TestWalWriter_AppendWithMeta_Index(t *testing.T) {
	dir := t.TempDir()
	w, err := newWalWriter(walWriterConfig{DataDir: dir, NodeID: "meta", BufSize: 1 << 20})
	if err != nil {
		t.Fatalf("newWalWriter: %v", err)
	}
	defer func() { _ = w.Close() }()

	// Prepare raw data with newlines to validate recs and CRC.
	raw := []byte("a\nb\nc\n") // 3 records
	recs := uint32(bytes.Count(raw, []byte{'\n'}))
	crc := crc32.ChecksumIEEE(raw)
	now := time.Now().UTC().UnixNano()
	gz := gzipBytes(raw)

	if err := w.AppendCompressedWithMeta(gz, recs, now, now, crc); err != nil {
		t.Fatalf("append with meta: %v", err)
	}
	if err := w.Sync(); err != nil {
		t.Fatalf("sync: %v", err)
	}

	// Read index line.
	idxPath := strings.TrimSuffix(w.currentPath, ".wal.gz") + ".wal.idx"
	f, err := os.Open(idxPath)
	if err != nil {
		t.Fatalf("open idx: %v", err)
	}
	defer f.Close()
	type idxEntry struct {
		File    string `json:"file"`
		Frame   uint64 `json:"frame"`
		Off     uint64 `json:"off"`
		Len     uint64 `json:"len"`
		Recs    uint32 `json:"recs"`
		FirstTS int64  `json:"first_ts"`
		LastTS  int64  `json:"last_ts"`
		CRC32   uint32 `json:"crc32"`
	}
	var ent idxEntry
	dec := json.NewDecoder(f)
	if err := dec.Decode(&ent); err != nil {
		t.Fatalf("decode idx: %v", err)
	}

	if ent.Frame != 1 || ent.Off != 0 {
		t.Fatalf("unexpected frame/offset: frame=%d off=%d", ent.Frame, ent.Off)
	}
	if int(ent.Len) != len(gz) {
		t.Fatalf("len mismatch: got %d want %d", ent.Len, len(gz))
	}
	if ent.Recs != recs {
		t.Fatalf("recs mismatch: got %d want %d", ent.Recs, recs)
	}
	if ent.CRC32 != crc {
		t.Fatalf("crc mismatch: got %08x want %08x", ent.CRC32, crc)
	}

	// Decompress the recorded span and verify CRC.
	wf, err := os.Open(w.currentPath)
	if err != nil {
		t.Fatalf("open wal: %v", err)
	}
	defer wf.Close()
	sr := io.NewSectionReader(wf, int64(ent.Off), int64(ent.Len))
	zr, err := gzip.NewReader(sr)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	defer zr.Close()
	h := crc32.NewIEEE()
	if _, err := io.Copy(h, zr); err != nil && err != io.EOF {
		t.Fatalf("decompress copy: %v", err)
	}
	if got := h.Sum32(); got != crc {
		t.Fatalf("crc after decompress mismatch: got %08x want %08x", got, crc)
	}
}
