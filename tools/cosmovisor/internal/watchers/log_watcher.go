package watchers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
)

func NewHaltHeightLogWatcher(ctx context.Context, log io.Reader, checker HeightChecker) Watcher[uint64] {
	check := func(line string) (uint64, error) {
		height, err := parseHaltHeightLogMessage(line)
		if err != nil {
			return 0, err
		}
		actualHeight, err := checker.GetLatestBlockHeight()
		if err != nil {
			return 0, err
		}
		if actualHeight < height {
			// false positive, ignore this log line
			return 0, os.ErrNotExist
		}
		return actualHeight, nil
	}
	return NewDataWatcher[uint64](ctx, NewLogWatcher(ctx, log, haltHeightRegex), check)
}

var haltHeightRegex = regexp.MustCompile(`halt per configuration height (\d+)`)

func parseHaltHeightLogMessage(line string) (uint64, error) {
	matches := haltHeightRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return 0, os.ErrNotExist // No match found
	}
	var height uint64
	_, err := fmt.Sscanf(matches[1], "%d", &height)
	if err != nil {
		return 0, err // Error parsing height
	}
	return height, nil
}

type LogWatcher struct {
	outChan chan string
	errChan chan error
}

func NewLogWatcher(ctx context.Context, log io.Reader, regex *regexp.Regexp) Watcher[string] {
	out := make(chan string, 1)
	err := make(chan error, 1)
	go func() {
		defer close(out)
		defer close(err)
		scanner := bufio.NewScanner(log)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				line := scanner.Text()
				if regex.MatchString(line) {
					out <- line
				}
			}
		}
	}()
	return &LogWatcher{
		outChan: out,
		errChan: err,
	}
}

func (l *LogWatcher) Updated() <-chan string {
	return l.outChan
}

func (l *LogWatcher) Errors() <-chan error {
	return l.errChan
}
