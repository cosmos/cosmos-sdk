package log

import (
	"bytes"
	"compress/gzip"
	"hash/crc32"
	"io"
	"sync"
	"testing"
	"time"
)

// newTestAggregator returns a minimal memAggregator configured just for
// gzipWithPool usage in benchmarks (no goroutines started).
func newTestAggregator() *memAggregator {
	m := &memAggregator{}
	m.gzPool = sync.Pool{New: func() any { w, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed); return w }}
	m.bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}
	return m
}

// BenchmarkMemLogger_MetadataOverhead compares gzip cost alone vs.
// gzip plus computing (recs, crc32, now) as done in compressor().
func BenchmarkMemLogger_MetadataOverhead(b *testing.B) {
	sizes := []int{4 << 10, 64 << 10, 1 << 20}
	for _, sz := range sizes {
		data := makeIncompressible(sz)
		// Ensure there are a few newlines to reflect JSONL-like data; place ~1% newlines.
		for i := 100; i < len(data); i += 100 {
			data[i] = '\n'
		}

		b.Run("gzip_only/"+byteSize(sz), func(b *testing.B) {
			m := newTestAggregator()
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := m.gzipWithPool(data); err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("gzip_plus_meta/"+byteSize(sz), func(b *testing.B) {
			m := newTestAggregator()
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Metadata calculations mirrored from compressor()
				_ = uint32(bytes.Count(data, []byte{'\n'}))
				_ = crc32.ChecksumIEEE(data)
				_ = time.Now().UTC().UnixNano()
				if _, err := m.gzipWithPool(data); err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("gzip_plus_meta_opt/"+byteSize(sz), func(b *testing.B) {
			m := newTestAggregator()
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				gz, err := m.gzipWithPool(data)
				if err != nil {
					b.Fatal(err)
				}
				if _, ok := gzipCRC32FromMember(gz); !ok {
					b.Fatal("failed to parse gzip trailer CRC32")
				}
			}
		})
	}
}

// BenchmarkWalWriter_AppendMetaOverhead compares appending a gzip member to WAL
// with and without writing index metadata. Metadata is precomputed so the
// benchmark isolates the append-path difference (JSON encode + index fs writes).
func BenchmarkWalWriter_AppendMetaOverhead(b *testing.B) {
	sizes := []int{64 << 10, 256 << 10}
	for _, sz := range sizes {
		b.Run(byteSize(sz), func(b *testing.B) {
			dir := b.TempDir()
			w, err := newWalWriter(walWriterConfig{DataDir: dir, NodeID: "bench", BufSize: 4 << 20})
			if err != nil {
				b.Fatalf("newWalWriter: %v", err)
			}
			defer func() { _ = w.Close() }()

			// Prepare an incompressible chunk to avoid overly tiny gzip members.
			raw := makeIncompressible(sz)
			chunk := gzipBytes(raw)
			recs := uint32(bytes.Count(raw, []byte{'\n'}))
			crc := crc32.ChecksumIEEE(raw)
			now := time.Now().UTC().UnixNano()

			b.Run("append_only", func(b *testing.B) {
				b.ReportAllocs()
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

			b.Run("append_with_meta", func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(chunk)))
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if err := w.AppendCompressedWithMeta(chunk, recs, now, now, crc); err != nil {
						b.Fatal(err)
					}
				}
				b.StopTimer()
				_ = w.Sync()
			})
		})
	}
}
