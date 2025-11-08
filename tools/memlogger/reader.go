package memlogger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Reader is a minimal helper to follow an .idx file and read the corresponding frames
// from the .gz file by byte range. This is a skeleton for building a custom agent.
type Reader struct {
	dir   string
	idx   *os.File
	gz    *os.File
	curGz string
	r     *bufio.Reader
}

// OpenIndex opens an index file (by path or basename). It also opens the matching .gz file.
func (rdr *Reader) OpenIndex(dir, idxBase string) error {
	rdr.dir = dir
	idxPath := idxBase
	if !strings.Contains(idxBase, string(os.PathSeparator)) {
		idxPath = filepath.Join(dir, idxBase)
	}
	f, err := os.Open(idxPath)
	if err != nil {
		return err
	}
	rdr.idx = f
	rdr.r = bufio.NewReaderSize(f, 64*1024)
	// infer gz name from first line or from idx name
	gzBase := strings.TrimSuffix(filepath.Base(idxBase), ".idx") + ".gz"
	gzPath := filepath.Join(dir, gzBase)
	gz, err := os.Open(gzPath)
	if err != nil {
		_ = f.Close()
		return err
	}
	rdr.gz = gz
	rdr.curGz = gzBase
	return nil
}

// NextFrame blocks until it can read the next complete index line.
// It returns the FrameMeta and a ReadCloser for the gzip member bytes.
//
// Note: fm.Off and fm.Len refer to COMPRESSED byte ranges in the .wal.gz file.
// The returned ReadCloser is already scoped to [off, off+len); callers should
// wrap it with gzip.NewReader to obtain the uncompressed payload.
func (rdr *Reader) NextFrame() (FrameMeta, io.ReadCloser, error) {
	if rdr.idx == nil || rdr.gz == nil {
		return FrameMeta{}, nil, fmt.Errorf("reader not opened")
	}
	for {
		line, err := rdr.r.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				// No new frame yet; caller may sleep and retry
				return FrameMeta{}, nil, io.EOF
			}
			return FrameMeta{}, nil, err
		}
		var fm FrameMeta
		if err := json.Unmarshal(line, &fm); err != nil {
			// skip malformed line
			continue
		}
		// If file switched (rotation), reopen gz accordingly
		if fm.File != rdr.curGz {
			_ = rdr.gz.Close()
			gz, err := os.Open(filepath.Join(rdr.dir, fm.File))
			if err != nil {
				return FrameMeta{}, nil, err
			}
			rdr.gz = gz
			rdr.curGz = fm.File
		}
		// Create a section reader for the frame bytes
		rc := io.NopCloser(io.NewSectionReader(rdr.gz, int64(fm.Off), int64(fm.Len)))
		return fm, rc, nil
	}
}

// Close closes files.
func (rdr *Reader) Close() error {
	var err error
	if rdr.idx != nil {
		if e := rdr.idx.Close(); err == nil {
			err = e
		}
	}
	if rdr.gz != nil {
		if e := rdr.gz.Close(); err == nil {
			err = e
		}
	}
	return err
}
