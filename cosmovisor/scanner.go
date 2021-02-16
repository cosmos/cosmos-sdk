package cosmovisor

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// UpgradeInfo is the update details created by `x/upgrade/keeper.DumpUpgradeInfoToDisk`.
type UpgradeInfo struct {
	Name string
	Info string
}

type fileWatcher struct {
	filename string
	dirname  string
}

func newUpgradeFileWatcher(filename string) (fileWatcher, error) {
	if filename == "" {
		return fileWatcher{}, nil
	}
	filenameAbs, err := filepath.Abs(filename)
	if err != nil {
		return fileWatcher{},
			fmt.Errorf("wrong path, %s must be a valid file path, [%w]", filename, err)
	}
	dirname := filepath.Dir(filename)
	fw := fileWatcher{filenameAbs, dirname}

	info, err := os.Stat(dirname)
	if err != nil || !info.IsDir() {
		return fw, fmt.Errorf("wrong path, %s must be an existing directory, [%w]", dirname, err)
	}
	return fw, nil
}

func (fw fileWatcher) MonitorUpdate(res *WaitResult) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		res.SetError(fmt.Errorf("can't create file watcher: %w", err))
		return
	}
	defer watcher.Close()

	if err = watcher.Add(fw.dirname); err != nil {
		res.SetError(fmt.Errorf("can't add directory '%s', to the file watcher: %w", fw.dirname, err))
		return
	}

	// we don't stop the process on error - the blockchain node shouldn't be killed if the
	// watcher stopped working correctly.
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				fn, err := filepath.Abs(event.Name)
				if err != nil {
					log.Printf("ERROR: file watcher can't expand a filename '%s', %v", event.Name, err)
				} else if fw.filename == fn {
					ui, err := parseUpgradeInfoFile(fn)
					if err != nil {
						log.Printf("ERROR: file watcher can't expand a filename '%s', %v", event.Name, err)
					} else {
						res.SetUpgrade(&ui)
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if ok {
				log.Println("ERROR: file watcher can't monitor correctly for the update", err)
			}
		}
	}
}

func parseUpgradeInfoFile(filename string) (UpgradeInfo, error) {
	var ui UpgradeInfo
	f, err := os.Open(filename)
	if err != nil {
		return ui, err
	}
	defer f.Close()
	// byteValue, _ := ioutil.ReadAll(jsonFile)
	d := json.NewDecoder(f)
	err = d.Decode(&ui)
	return ui, err
}
