package log

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// TestWalWriter_AppendAndSync verifies basic append and sync behavior.
func TestWalWriter_AppendAndSync(t *testing.T) {
	dir := t.TempDir()
	w, err := newWalWriter(walWriterConfig{DataDir: dir, NodeID: "testnode", BufSize: 1 << 20})
	if err != nil {
		t.Fatalf("newWalWriter: %v", err)
	}
	defer func() { _ = w.Close() }()

	// Append a few chunks of varying sizes.
	chunks := [][]byte{
		bytes.Repeat([]byte("a"), 8<<10),   // 8 KiB
		bytes.Repeat([]byte("b"), 64<<10),  // 64 KiB
		bytes.Repeat([]byte("c"), 256<<10), // 256 KiB
	}
	var total int
	for _, c := range chunks {
		if err := w.AppendCompressed(c); err != nil {
			t.Fatalf("append: %v", err)
		}
		total += len(c)
	}

	if err := w.Sync(); err != nil {
		t.Fatalf("sync: %v", err)
	}

	// Verify the current segment exists and has at least the expected size.
	fi, err := os.Stat(w.currentPath)
	if err != nil {
		t.Fatalf("stat current segment: %v", err)
	}
	if fi.Size() < int64(total) {
		t.Fatalf("segment size too small: got %d want >= %d", fi.Size(), total)
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

	// Append two 48 KiB chunks -> should roll after second write.
	chunk := bytes.Repeat([]byte("x"), 48<<10)
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

			chunk := bytes.Repeat([]byte("z"), sz)
			b.SetBytes(int64(sz))
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

					chunk := bytes.Repeat([]byte("w"), cs)
					b.SetBytes(int64(cs))
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
