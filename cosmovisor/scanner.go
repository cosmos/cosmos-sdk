package cosmovisor

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// UpgradeInfo is the update details created by `x/upgrade/keeper.DumpUpgradeInfoToDisk`.
type UpgradeInfo struct {
	Name   string `json:"name"`
	Info   string `json:"info"`
	Height uint   `json:"height"`
}

type fileWatcher struct {
	// full path to a watched file
	filename string
	interval time.Duration

	currentInfo UpgradeInfo
	lastModTime time.Time
	cancel      chan bool
	ticker      *time.Ticker
	needsUpdate bool

	initialized bool
}

func newUpgradeFileWatcher(filename string, interval time.Duration) (*fileWatcher, error) {
	if filename == "" {
		return nil, errors.New("filename undefined")
	}
	filenameAbs, err := filepath.Abs(filename)
	if err != nil {
		return nil,
			fmt.Errorf("wrong path, %s must be a valid file path, [%w]", filename, err)
	}
	dirname := filepath.Dir(filename)
	info, err := os.Stat(dirname)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("wrong path, %s must be an existing directory, [%w]", dirname, err)
	}

	return &fileWatcher{filenameAbs, interval, UpgradeInfo{}, time.Time{}, make(chan bool), time.NewTicker(interval), false, false}, nil
}

func (fw *fileWatcher) Stop() {
	close(fw.cancel)
}

// pools the filesystem to check for new upgrade currentInfo. currentName is the name
// of currently running upgrade. The check is rejected if it finds an upgrade with the same
// name.
func (fw *fileWatcher) MonitorUpdate(currentUpgrade UpgradeInfo) <-chan struct{} {
	fw.ticker.Reset(fw.interval)
	done := make(chan struct{})
	fw.cancel = make(chan bool)
	fw.needsUpdate = false

	go func() {
		for {
			select {
			case <-fw.ticker.C:
				if fw.CheckUpdate(currentUpgrade) {
					done <- struct{}{}
					return
				}
			case <-fw.cancel:
				return
			}
		}
	}()
	return done
}

// CheckUpdate reads update plan from file and checks if there is a new update request
// currentName is the name of currently running upgrade. The check is rejected if it finds
// an upgrade with the same name.
func (fw *fileWatcher) CheckUpdate(currentUpgrade UpgradeInfo) bool {
	if fw.needsUpdate {
		return true
	}
	stat, err := os.Stat(fw.filename)
	if err != nil { // file doesn't exists
		return false
	}
	if !stat.ModTime().After(fw.lastModTime) {
		return false
	}
	info, err := parseUpgradeInfoFile(fw.filename)
	if err != nil {
		log.Fatal("Can't parse upgrade info file. Err: ", err)
		return false
	}
	if !fw.initialized { // daemon has restarted
		fw.initialized = true
		fw.currentInfo = info
		fw.lastModTime = stat.ModTime()
		// heuristic: deamon has restarted, so we don't know if we successfully downloaded the upgrade or not.
		// so we try to compare the running upgrade name (read from the cosmovisor file) with the upgrade info
		if currentUpgrade.Name != fw.currentInfo.Name {
			fw.needsUpdate = true
			return true
		}
	}

	if info.Height > fw.currentInfo.Height {
		fw.currentInfo = info
		fw.lastModTime = stat.ModTime()
		fw.needsUpdate = true
		return true
	}
	return false
}

func parseUpgradeInfoFile(filename string) (UpgradeInfo, error) {
	var ui UpgradeInfo
	f, err := os.Open(filename)
	if err != nil {
		return ui, err
	}
	defer f.Close()
	d := json.NewDecoder(f)
	err = d.Decode(&ui)
	if err != nil {
		return ui, err
	}
	// required values must be set
	if ui.Height == 0 || ui.Name == "" {
		return UpgradeInfo{}, fmt.Errorf("invalid upgrade-info.json content. Name and Hight must be not empty. Got: %v", ui)
	}
	return ui, err
}
