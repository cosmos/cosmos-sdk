package cosmovisor

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/fsnotify/fsnotify"
)

// Trim off whitespace around the info - match least greedy, grab as much space on both sides
// Defined here: https://github.com/cosmos/cosmos-sdk/blob/release/v0.38.2/x/upgrade/abci.go#L38
//  fmt.Sprintf("UPGRADE \"%s\" NEEDED at %s: %s", plan.Name, plan.DueAt(), plan.Info)
// DueAt defined here: https://github.com/cosmos/cosmos-sdk/blob/release/v0.38.2/x/upgrade/internal/types/plan.go#L73-L78
//
//    if !p.Time.IsZero() {
//      return fmt.Sprintf("time: %s", p.Time.UTC().Format(time.RFC3339))
//    }
//    return fmt.Sprintf("height: %d", p.Height)
var upgradeRegex = regexp.MustCompile(`UPGRADE "(.*)" NEEDED at ((height): (\d+)|(time): (\S+)):\s+(\S*)`)

// UpgradeInfo is the details from the regexp
type UpgradeInfo struct {
	Name string
	Info string
}

// WaitForUpdate will listen to the scanner until a line matches upgradeRegexp.
// It returns (info, nil) on a matching line
// It returns (nil, err) if the input stream errored
// It returns (nil, nil) if the input closed without ever matching the regexp
func WaitForUpdate(scanner *bufio.Scanner, res *WaitResult) {
	for scanner.Scan() {
		line := scanner.Text()
		if upgradeRegex.MatchString(line) {
			subs := upgradeRegex.FindStringSubmatch(line)
			info := UpgradeInfo{
				Name: subs[1],
				Info: subs[7],
			}
			res.SetUpgrade(&info)
			return
		}
	}
	res.SetError(scanner.Err())
}

type fileWatcher struct {
	filename string
	dirname  string
	ok       bool
}

func newUpgradeFileWatcher(filename string) (fileWatcher, error) {
	if filename == "" {
		return fileWatcher{}, nil
	}
	dirname := filepath.Dir(filename)
	fw := fileWatcher{filename, dirname, true}

	info, err := os.Stat(dirname)
	if err != nil || !info.IsDir() {
		return fw, fmt.Errorf("wrong path, %s must be an existing directory, [%w]", dirname, err)
	}
	return fw, nil
}

func (fw fileWatcher) WaitForUpdate(res *WaitResult) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("can't create file watcher: %w", err)
	}
	defer watcher.Close()

	err = watcher.Add(fw.dirname)
	if err != nil {
		return fmt.Errorf("can't add directory '%s', to the file watcher: %w", fw.dirname, err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				break
			}
			log.Println("event:", event)
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("modified file:", event.Name)
			}
		case err, ok := <-watcher.Errors:
			if ok {
				res.SetError(err)
			}
			break
		}
	}
	return nil
}
