package blocklog

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"cosmossdk.io/log/v2"
)

const fileExt = ".jsonl"

// Store writes captured entries to <dir>/<height>.jsonl and prunes heights that
// fell out of the retention window. All errors are logged (once per kind) and
// swallowed: the store must never fail block execution.
type Store struct {
	dir    string
	retain uint64
	logger log.Logger

	open atomic.Bool // fast path for Logger.record outside a block

	// ponytail: one mutex for writer and metadata; per-height writers if the
	// lock ever shows up in a profile.
	mu         sync.Mutex
	f          *os.File
	w          *bufio.Writer
	height     int64
	lastClosed int64
	swept      bool
	logged     map[string]bool
}

// Open prepares dir for writing. retain is the number of most recent heights
// kept on disk. logger receives capture errors; it must not be a logger that
// captures into this store.
func Open(dir string, retain uint64, logger log.Logger) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create block log dir %s: %w", dir, err)
	}
	return &Store{dir: dir, retain: retain, logger: logger, logged: map[string]bool{}}, nil
}

// LastClosed returns the last height whose file was closed by Commit.
func (s *Store) LastClosed() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastClosed
}

// Retain returns the number of heights kept on disk.
func (s *Store) Retain() uint64 { return s.retain }

// Begin opens the file for height, truncating whatever a previous run left
// there: a height executes at most once per process, so an existing file is a
// partial write from a crashed run that CometBFT is now replaying.
func (s *Store) Begin(height int64) {
	s.mu.Lock()
	errs := s.closeLocked()
	f, err := os.OpenFile(s.path(height), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		errs = append(errs, captureError{"block log: open file", height, err})
	} else {
		s.f, s.w, s.height = f, bufio.NewWriter(f), height
		s.open.Store(true)
	}
	s.mu.Unlock()
	s.report(errs)
}

// Write records e into the open height's file. It is a no-op outside a block.
func (s *Store) Write(e Entry) {
	if !s.open.Load() {
		return
	}
	s.mu.Lock()
	if s.w == nil {
		s.mu.Unlock()
		return
	}
	e.Height = s.height
	err := json.NewEncoder(s.w).Encode(e)
	height := s.height
	s.mu.Unlock()
	if err != nil {
		s.report([]captureError{{"block log: write entry", height, err}})
	}
}

// Commit closes the open height's file and prunes the height that fell out of
// the retention window. The first Commit after start sweeps every file at or
// below the cutoff, which handles restarts and retention changes.
func (s *Store) Commit() {
	s.mu.Lock()
	if s.w == nil {
		s.mu.Unlock()
		return
	}
	height := s.height
	errs := s.closeLocked()
	s.lastClosed = height

	cutoff := height - int64(s.retain)
	if !s.swept {
		s.swept = true
		errs = append(errs, s.sweep(cutoff)...)
	} else if cutoff >= 1 {
		errs = append(errs, s.remove(cutoff)...)
	}
	s.mu.Unlock()
	s.report(errs)
}

func (s *Store) closeLocked() (errs []captureError) {
	if s.w == nil {
		return nil
	}
	s.open.Store(false)
	if err := s.w.Flush(); err != nil {
		errs = append(errs, captureError{"block log: flush", s.height, err})
	}
	if err := s.f.Close(); err != nil {
		errs = append(errs, captureError{"block log: close", s.height, err})
	}
	s.f, s.w = nil, nil
	return errs
}

func (s *Store) sweep(cutoff int64) (errs []captureError) {
	dirEntries, err := os.ReadDir(s.dir)
	if err != nil {
		return []captureError{{"block log: sweep", cutoff, err}}
	}
	for _, de := range dirEntries {
		if h, ok := parseHeight(de.Name()); ok && h <= cutoff {
			errs = append(errs, s.remove(h)...)
		}
	}
	return errs
}

func (s *Store) remove(height int64) []captureError {
	if err := os.Remove(s.path(height)); err != nil && !os.IsNotExist(err) {
		return []captureError{{"block log: prune", height, err}}
	}
	return nil
}

func (s *Store) path(height int64) string {
	return filepath.Join(s.dir, strconv.FormatInt(height, 10)+fileExt)
}

type captureError struct {
	msg    string
	height int64
	err    error
}

// report logs each error kind once per process, outside the store lock so a
// logger that happens to capture into this store cannot deadlock it.
func (s *Store) report(errs []captureError) {
	for _, ce := range errs {
		s.mu.Lock()
		seen := s.logged[ce.msg]
		s.logged[ce.msg] = true
		s.mu.Unlock()
		if !seen {
			s.logger.Error(ce.msg, "height", ce.height, "err", ce.err)
		}
	}
}

func parseHeight(name string) (int64, bool) {
	base, ok := strings.CutSuffix(name, fileExt)
	if !ok {
		return 0, false
	}
	h, err := strconv.ParseInt(base, 10, 64)
	return h, err == nil
}
