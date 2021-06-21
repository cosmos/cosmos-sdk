package cosmovisor

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// UpgradeInfo is the update details created by `x/upgrade/keeper.DumpUpgradeInfoToDisk`.
type UpgradeInfo struct {
	Name   string
	Info   string
	Height uint
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

	return &fileWatcher{filenameAbs, interval, UpgradeInfo{}, time.Time{}, make(chan bool), time.NewTicker(interval), false}, nil
}

func (fw *fileWatcher) Stop() {
	close(fw.cancel)
}

func (fw *fileWatcher) MonitorUpdate() <-chan struct{} {
	fw.ticker.Reset(fw.interval)
	done := make(chan struct{})
	fw.cancel = make(chan bool)
	fw.needsUpdate = false

	go func() {
		for {
			select {
			case <-fw.ticker.C:
				if fw.CheckUpdate() {
					done <- struct{}{}
				}
			case <-fw.cancel:
				return
			}
		}
	}()
	return done
}

// CheckUpdate reads update plan from file and checks if there is a new update request
func (fw *fileWatcher) CheckUpdate() bool {
	if fw.needsUpdate {
		return true
	}
	stat, err := os.Stat(fw.filename)
	if err != nil {
		return false
	}
	if !stat.ModTime().After(fw.lastModTime) {
		return false
	}
	ui, err := parseUpgradeInfoFile(fw.filename)
	if err != nil {
		// TODO: print error!
		return false
	}
	if ui.Height > fw.currentInfo.Height {
		fw.currentInfo = ui
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
	// byteValue, _ := ioutil.ReadAll(jsonFile)
	d := json.NewDecoder(f)
	err = d.Decode(&ui)
	return ui, err
}
