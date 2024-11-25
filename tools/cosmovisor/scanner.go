package cosmovisor

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	upgradetypes "cosmossdk.io/x/upgrade/types"
)

var errUntestAble = errors.New("untestable")

type fileWatcher struct {
	daemonHome string
	filename   string // full path to a watched file
	interval   time.Duration

	currentBin  string
	currentInfo upgradetypes.Plan
	lastModTime time.Time
	cancel      chan bool
	ticker      *time.Ticker

	needsUpdate   bool
	initialized   bool
	disableRecase bool
}

func newUpgradeFileWatcher(cfg *Config) (*fileWatcher, error) {
	filename := cfg.UpgradeInfoFilePath()
	if filename == "" {
		return nil, errors.New("filename undefined")
	}

	filenameAbs, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %s must be a valid file path: %w", filename, err)
	}

	dirname := filepath.Dir(filename)
	if info, err := os.Stat(dirname); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("invalid path: %s must be an existing directory: %w", dirname, err)
	}

	bin, err := cfg.CurrentBin()
	if err != nil {
		return nil, fmt.Errorf("error creating symlink to genesis: %w", err)
	}

	return &fileWatcher{
		daemonHome:    cfg.Home,
		currentBin:    bin,
		filename:      filenameAbs,
		interval:      cfg.PollInterval,
		currentInfo:   upgradetypes.Plan{},
		lastModTime:   time.Time{},
		cancel:        make(chan bool),
		ticker:        time.NewTicker(cfg.PollInterval),
		needsUpdate:   false,
		initialized:   false,
		disableRecase: cfg.DisableRecase,
	}, nil
}

func (fw *fileWatcher) Stop() {
	close(fw.cancel)
	fw.ticker.Stop()
}

// MonitorUpdate pools the filesystem to check for new upgrade currentInfo.
// currentName is the name of currently running upgrade.  The check is rejected if it finds
// an upgrade with the same name.
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
		if os.IsNotExist(err) {
			return false
		} else {
			panic(fmt.Errorf("failed to stat upgrade info file: %w", err))
		}
	}

	// check https://github.com/cosmos/cosmos-sdk/issues/21086
	// If new file is still empty, wait a small amount of time for write to complete
	if stat.Size() == 0 {
		for range 10 {
			time.Sleep(2 * time.Millisecond)
			stat, err = os.Stat(fw.filename)
			if err != nil {
				if os.IsNotExist(err) {
					return false
				} else {
					panic(fmt.Errorf("failed to stat upgrade info file: %w", err))
				}
			}
			if stat.Size() == 0 {
				break
			}
		}
	}
	if stat.Size() == 0 {
		return false
	}

	// no update if the file already exists and has not been modified
	if !stat.ModTime().After(fw.lastModTime) {
		return false
	}

	info, err := parseUpgradeInfoFile(fw.filename, fw.disableRecase)
	if err != nil {
		panic(fmt.Errorf("failed to parse upgrade info file: %w", err))
	}

	// file exist but too early in height
	currentHeight, err := fw.checkHeight()
	if (err != nil || currentHeight < info.Height) && !errors.Is(err, errUntestAble) { // ignore this check for tests
		return false
	}

	if !fw.initialized {
		// daemon has restarted
		fw.initialized = true
		fw.currentInfo = info
		fw.lastModTime = stat.ModTime()

		// Heuristic: Daemon has restarted, so we don't know if we successfully
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

// checkHeight checks if the current block height
func (fw *fileWatcher) checkHeight() (int64, error) {
	if testing.Testing() { // we cannot test the command in the test environment
		return 0, errUntestAble
	}

	result, err := exec.Command(fw.currentBin, "status", "--home", fw.daemonHome).CombinedOutput() //nolint:gosec // we want to execute the status command
	if err != nil {
		return 0, err
	}

	type response struct {
		SyncInfo struct {
			LatestBlockHeight string `json:"latest_block_height"`
		} `json:"sync_info"`
		AnotherCasingSyncInfo struct {
			LatestBlockHeight string `json:"latest_block_height"`
		} `json:"SyncInfo"`
	}

	var resp response
	if err := json.Unmarshal(result, &resp); err != nil {
		return 0, err
	}

	if resp.SyncInfo.LatestBlockHeight != "" {
		return strconv.ParseInt(resp.SyncInfo.LatestBlockHeight, 10, 64)
	} else if resp.AnotherCasingSyncInfo.LatestBlockHeight != "" {
		return strconv.ParseInt(resp.AnotherCasingSyncInfo.LatestBlockHeight, 10, 64)
	}

	return 0, errors.New("latest block height is empty")
}

func parseUpgradeInfoFile(filename string, disableRecase bool) (upgradetypes.Plan, error) {
	f, err := os.ReadFile(filename)
	if err != nil {
		return upgradetypes.Plan{}, err
	}

	if len(f) == 0 {
		return upgradetypes.Plan{}, fmt.Errorf("empty upgrade-info.json in %q", filename)
	}

	var upgradePlan upgradetypes.Plan
	if err := json.Unmarshal(f, &upgradePlan); err != nil {
		return upgradetypes.Plan{}, err
	}

	// required values must be set
	if err := upgradePlan.ValidateBasic(); err != nil {
		return upgradetypes.Plan{}, fmt.Errorf("invalid upgrade-info.json content: %w, got: %v", err, upgradePlan)
	}

	// normalize name to prevent operator error in upgrade name case sensitivity errors.
	if !disableRecase {
		upgradePlan.Name = strings.ToLower(upgradePlan.Name)
	}

	return upgradePlan, nil
}
