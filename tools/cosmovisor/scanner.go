package cosmovisor

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/rs/zerolog"
)

type fileWatcher struct {
	logger *zerolog.Logger

	// full path to a watched file
	filename string
	interval time.Duration

	currentInfo upgradetypes.Plan
	lastModTime time.Time
	cancel      chan bool
	ticker      *time.Ticker
	needsUpdate bool

	initialized bool
}

func newUpgradeFileWatcher(logger *zerolog.Logger, filename string, interval time.Duration) (*fileWatcher, error) {
	if filename == "" {
		return nil, errors.New("filename undefined")
	}

	filenameAbs, err := filepath.Abs(filename)
	if err != nil {
		return nil,
			fmt.Errorf("invalid path; %s must be a valid file path: %w", filename, err)
	}

	dirname := filepath.Dir(filename)
	info, err := os.Stat(dirname)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("invalid path; %s must be an existing directory: %w", dirname, err)
	}

	return &fileWatcher{
		logger:      logger,
		filename:    filenameAbs,
		interval:    interval,
		currentInfo: upgradetypes.Plan{},
		lastModTime: time.Time{},
		cancel:      make(chan bool),
		ticker:      time.NewTicker(interval),
		needsUpdate: false,
		initialized: false,
	}, nil
}

func (fw *fileWatcher) Stop() {
	close(fw.cancel)
}

// pools the filesystem to check for new upgrade currentInfo. currentName is the name
// of currently running upgrade. The check is rejected if it finds an upgrade with the same
// name.
func (fw *fileWatcher) MonitorUpdate(currentUpgrade upgradetypes.Plan) <-chan struct{} {
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
func (fw *fileWatcher) CheckUpdate(currentUpgrade upgradetypes.Plan) bool {
	if fw.needsUpdate {
		return true
	}

	stat, err := os.Stat(fw.filename)
	if err != nil {
		// file doesn't exists
		return false
	}

	if !stat.ModTime().After(fw.lastModTime) {
		return false
	}

	info, err := parseUpgradeInfoFile(fw.filename)
	if err != nil {
		fw.logger.Fatal().Err(err).Msg("failed to parse upgrade info file")
		return false
	}

	if !fw.initialized {
		// daemon has restarted
		fw.initialized = true
		fw.currentInfo = info
		fw.lastModTime = stat.ModTime()

		// Heuristic: Deamon has restarted, so we don't know if we successfully
		// downloaded the upgrade or not. So we try to compare the running upgrade
		// name (read from the cosmovisor file) with the upgrade info.
		if !strings.EqualFold(currentUpgrade.Name, fw.currentInfo.Name) {
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

func parseUpgradeInfoFile(filename string) (upgradetypes.Plan, error) {
	var ui upgradetypes.Plan

	f, err := os.Open(filename)
	if err != nil {
		return upgradetypes.Plan{}, err
	}
	defer f.Close()

	d := json.NewDecoder(f)
	if err := d.Decode(&ui); err != nil {
		return upgradetypes.Plan{}, err
	}

	// required values must be set
	if ui.Height <= 0 || ui.Name == "" {
		return upgradetypes.Plan{}, fmt.Errorf("invalid upgrade-info.json content; name and height must be not empty; got: %v", ui)
	}

	// normalize name to prevent operator error in upgrade name case sensitivity errors.
	ui.Name = strings.ToLower(ui.Name)

	return ui, err
}
