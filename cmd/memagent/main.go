package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	agent "github.com/cosmos/cosmos-sdk/tools/memlogger"
)

// memagent tails a MemLogger WAL index (.wal.idx) in near real-time and reads
// the corresponding compressed frames from the .wal.gz by byte range. It prints
// basic metadata and (optionally) verifies the stream by decompressing.
// It also follows segment/day rotation by switching to the latest .wal.idx.
func main() {
	var (
		root   = flag.String("root", ".", "application root (contains data/)")
		nodeID = flag.String("node", "default", "node id (directory suffix)")
		once   = flag.Bool("once", false, "process available frames and exit")
		verify = flag.Bool("verify", false, "verify CRC/line count while streaming output")
		meta   = flag.Bool("meta", false, "print frame metadata to stderr")
		poll   = flag.Duration("poll", 200*time.Millisecond, "poll interval when idle")
	)
	flag.Parse()

	walRoot := filepath.Join(*root, "data", "log.wal")

	idxPath, err := latestIndex(walRoot, *nodeID)
	if err != nil {
		fmt.Fprintln(os.Stderr, "memagent:", err)
		os.Exit(1)
	}

	var rdr agent.Reader
	if err := rdr.OpenIndex(filepath.Dir(idxPath), filepath.Base(idxPath)); err != nil {
		fmt.Fprintln(os.Stderr, "open index:", err)
		os.Exit(1)
	}
	defer rdr.Close()

	curIdxDir := filepath.Dir(idxPath)
	curIdxBase := filepath.Base(idxPath)

	for {
		fm, rc, err := rdr.NextFrame()
		if err != nil {
			if err == io.EOF {
				if *once {
					return
				}
				// Check if a newer index file exists (segment/day rotation)
				next, nerr := latestIndex(walRoot, *nodeID)
				if nerr == nil && (filepath.Base(next) != curIdxBase || filepath.Dir(next) != curIdxDir) {
					_ = rdr.Close()
					if err := rdr.OpenIndex(filepath.Dir(next), filepath.Base(next)); err != nil {
						fmt.Fprintln(os.Stderr, "reopen index:", err)
						time.Sleep(*poll)
						continue
					}
					curIdxDir = filepath.Dir(next)
					curIdxBase = filepath.Base(next)
					continue
				}
				time.Sleep(*poll)
				continue
			}
			fmt.Fprintln(os.Stderr, "next frame:", err)
			time.Sleep(*poll)
			continue
		}

		// Optional metadata (stderr to keep stdout clean for log stream)
		if *meta {
			fmt.Fprintf(os.Stderr, "file=%s frame=%d off=%d len=%d recs=%d\n", fm.File, fm.Frame, fm.Off, fm.Len, fm.Recs)
		}

		// Stream-decompress the frame and emit original log bytes to stdout.
		if err := streamFrame(fm, rc, os.Stdout, *verify); err != nil {
			fmt.Fprintln(os.Stderr, "stream:", err)
			// Continue on error to avoid stalling the tailer.
		}
	}
}

// latestIndex returns the absolute path of the newest index file for nodeID.
func latestIndex(base string, nodeID string) (string, error) {
	nodeDir := filepath.Join(base, fmt.Sprintf("node-%s", nodeID))
	days, err := os.ReadDir(nodeDir)
	if err != nil {
		return "", fmt.Errorf("scan node dir: %w", err)
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
		return "", fmt.Errorf("no day directory under %s", nodeDir)
	}
	dayDir := filepath.Join(nodeDir, latestDay)
	ents, err := os.ReadDir(dayDir)
	if err != nil {
		return "", err
	}
	var latest string
	for _, e := range ents {
		if strings.HasSuffix(e.Name(), ".wal.idx") && e.Name() > latest {
			latest = e.Name()
		}
	}
	if latest == "" {
		return "", fmt.Errorf("no index files in %s", dayDir)
	}
	return filepath.Join(dayDir, latest), nil
}

func verifyFrame(fm agent.FrameMeta, rc io.ReadCloser) error {
	defer rc.Close()
	zr, err := gzip.NewReader(rc)
	if err != nil {
		return err
	}
	defer zr.Close()

	// Stream through the member, count newlines and bytes, compute CRC32.
	buf := make([]byte, 64<<10)
	lines := 0
	var total int64
	h := crc32.NewIEEE()
	for {
		n, err := zr.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			total += int64(n)
			h.Write(chunk)
			lines += bytes.Count(chunk, []byte{'\n'})
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	fmt.Printf("\tdecompressed lines=%d bytes=%d\n", lines, total)
	if fm.Recs != 0 && uint32(lines) != fm.Recs {
		fmt.Printf("\twarn: lines mismatch index=%d counted=%d\n", fm.Recs, lines)
	}
	if fm.CRC32 != 0 {
		if got := h.Sum32(); got != fm.CRC32 {
			return fmt.Errorf("crc mismatch: got=%08x want=%08x", got, fm.CRC32)
		}
	}
	return nil
}

// streamFrame decompresses a single frame and writes uncompressed bytes to w.
// If verify is true, it also checks CRC32 (if present) and counts lines.
func streamFrame(fm agent.FrameMeta, rc io.ReadCloser, w io.Writer, verify bool) error {
	defer rc.Close()
	zr, err := gzip.NewReader(rc)
	if err != nil {
		return err
	}
	defer zr.Close()

	var (
		buf   = make([]byte, 64<<10)
		lines int
		h     hash.Hash32
	)
	if verify {
		h = crc32.NewIEEE()
	}
	for {
		n, rerr := zr.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			if verify {
				if h != nil {
					_, _ = h.Write(chunk)
				}
				lines += bytes.Count(chunk, []byte{'\n'})
			}
			if _, werr := w.Write(chunk); werr != nil {
				return werr
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return rerr
		}
	}
	if verify {
		if fm.Recs != 0 && uint32(lines) != fm.Recs {
			fmt.Fprintf(os.Stderr, "warn: lines mismatch index=%d counted=%d\n", fm.Recs, lines)
		}
		if fm.CRC32 != 0 {
			if got := h.Sum32(); got != fm.CRC32 {
				return fmt.Errorf("crc mismatch: got=%08x want=%08x", got, fm.CRC32)
			}
		}
	}
	return nil
}
