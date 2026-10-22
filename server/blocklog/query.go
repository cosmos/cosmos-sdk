package blocklog

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	// DefaultLimit is the page size when the request does not set one.
	DefaultLimit = 100
	// MaxLimit is the largest page size; bigger requests are clamped.
	MaxLimit = 1000
)

// ErrInvalidArgument wraps every request validation failure.
var ErrInvalidArgument = errors.New("invalid argument")

var levelRank = map[string]int{"debug": 0, "info": 1, "warn": 2, "error": 3}

// Request selects entries from closed heights.
type Request struct {
	Start, End int64  // 0 means the last closed height
	Level      string // minimum level, inclusive; empty means info
	Module     string // exact match on Entry.Module when set
	Key        []byte // opaque cursor from a previous Result.NextKey
	Limit      uint64 // 0 means DefaultLimit
}

// Result is one page of entries. NextKey is set when more matches exist.
type Result struct {
	Entries []Entry
	NextKey []byte
}

// Query streams the files for the requested range and returns one page. It
// takes the store lock only to read the last closed height, so a slow query
// never stalls block execution.
func (s *Store) Query(req Request) (Result, error) {
	last := s.LastClosed()

	minLevel, ok := levelRank[req.Level]
	if req.Level == "" {
		minLevel = levelRank["info"]
	} else if !ok {
		return Result{}, fmt.Errorf("%w: unknown level %q (want debug, info, warn or error)", ErrInvalidArgument, req.Level)
	}
	if last == 0 {
		return Result{}, fmt.Errorf("%w: no block has been captured yet", ErrInvalidArgument)
	}
	end := req.End
	if end == 0 {
		end = last
	}
	start := req.Start
	if start == 0 {
		start = end
	}
	switch {
	case start < 1:
		return Result{}, fmt.Errorf("%w: start must be >= 1, got %d", ErrInvalidArgument, start)
	case start > end:
		return Result{}, fmt.Errorf("%w: start %d > end %d", ErrInvalidArgument, start, end)
	case end > last:
		return Result{}, fmt.Errorf("%w: end %d beyond last captured height %d", ErrInvalidArgument, end, last)
	}
	if lo := end - int64(s.retain) + 1; lo > start {
		start = lo // clip to the retention window
	}

	limit := req.Limit
	if limit == 0 {
		limit = DefaultLimit
	}
	limit = min(limit, MaxLimit)

	cursorHeight, offset := start, int64(0)
	if req.Key != nil {
		h, off, err := decodeKey(req.Key)
		if err != nil {
			return Result{}, err
		}
		if h < start || h > end {
			return Result{}, fmt.Errorf("%w: pagination key height %d outside range [%d, %d]", ErrInvalidArgument, h, start, end)
		}
		cursorHeight, offset = h, off
	}

	var res Result
	for h := cursorHeight; h <= end; h++ {
		if h != cursorHeight {
			offset = 0
		}
		next, err := s.scanFile(h, offset, func(e Entry) bool {
			return levelRank[e.Level] >= minLevel && (req.Module == "" || e.Module == req.Module)
		}, &res, limit)
		if err != nil {
			return Result{}, err
		}
		if next {
			return res, nil
		}
	}
	return res, nil
}

// scanFile appends matching entries from height's file starting at byte
// offset. It reports true when the page is full and another match exists, in
// which case res.NextKey points at that match.
func (s *Store) scanFile(height, offset int64, match func(Entry) bool, res *Result, limit uint64) (bool, error) {
	f, err := os.Open(s.path(height))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("open block log %d: %w", height, err)
	}
	defer f.Close()

	if offset > 0 {
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			return false, fmt.Errorf("seek block log %d: %w", height, err)
		}
	}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	pos := offset
	for sc.Scan() {
		line := sc.Bytes()
		lineStart := pos
		pos += int64(len(line)) + 1

		var e Entry
		if json.Unmarshal(line, &e) != nil {
			continue // partial line from a crashed run
		}
		if !match(e) {
			continue
		}
		if uint64(len(res.Entries)) >= limit {
			res.NextKey = encodeKey(height, lineStart)
			return true, nil
		}
		res.Entries = append(res.Entries, e)
	}
	// ponytail: a scanner error (line > 1 MiB) ends this file like EOF
	return false, nil
}

func encodeKey(height, offset int64) []byte {
	key := make([]byte, 16)
	binary.BigEndian.PutUint64(key, uint64(height))
	binary.BigEndian.PutUint64(key[8:], uint64(offset))
	return key
}

func decodeKey(key []byte) (height, offset int64, err error) {
	if len(key) != 16 {
		return 0, 0, fmt.Errorf("%w: malformed pagination key", ErrInvalidArgument)
	}
	return int64(binary.BigEndian.Uint64(key)), int64(binary.BigEndian.Uint64(key[8:])), nil
}
