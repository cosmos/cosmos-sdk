package cosmovisor

import (
	"context"
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

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/store"

	"cosmossdk.io/tools/cosmovisor/internal/watchers"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

var errUntestAble = errors.New("untestable")

func initWatcher[T any](ctx context.Context, cfg *Config, dirWatcher *watchers.FSNotifyWatcher, filename string) watchers.Watcher[T] {
	if dirWatcher != nil {
		hybridWatcher := watchers.NewHybridWatcher(ctx, dirWatcher, filename, cfg.PollInterval)
		return watchers.NewDataWatcher[T](ctx, hybridWatcher)
	} else {
		pollWatcher := watchers.NewPollWatcher(ctx, filename, cfg.PollInterval)
		return watchers.NewDataWatcher[T](ctx, pollWatcher)
	}
}

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

// checkHeight checks if the current block height
func (fw *fileWatcher) checkHeight() (int64, error) {
	if testing.Testing() { // we cannot test the command in the test environment
		return 0, errUntestAble
	}

	if fw.IsStop() {
		result, err := exec.Command(fw.currentBin, "config", "get", "config", "db_backend", "--home", fw.daemonHome).CombinedOutput() //nolint:gosec // we want to execute the config command
		if err != nil {
			result = []byte("goleveldb") // set default value, old version may not have config command
		}
		blockStoreDB, err := dbm.NewDB("blockstore", dbm.BackendType(result), filepath.Join(fw.daemonHome, "data"))
		if err != nil {
			return 0, err
		}
		defer blockStoreDB.Close()
		return store.NewBlockStore(blockStoreDB).Height(), nil
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
