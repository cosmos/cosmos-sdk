//go:build linux || darwin

package cosmovisor_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"
	"cosmossdk.io/tools/cosmovisor"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func writePendingUpgrade(t *testing.T, cfg *cosmovisor.Config, p upgradetypes.Plan) {
	t.Helper()
	bz, err := json.Marshal(p)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(cfg.UpgradeInfoFilePath()), 0o755))
	require.NoError(t, os.WriteFile(cfg.UpgradeInfoFilePath(), bz, 0o600))
}

func newPendingUpgradeLauncher(t *testing.T) (*cosmovisor.Config, cosmovisor.Launcher) {
	t.Helper()
	cfg := prepareConfig(
		t,
		fmt.Sprintf("%s/%s", workDir, "testdata/validate"),
		cosmovisor.Config{Name: "dummyd", PollInterval: 15},
	)
	logger := log.NewTestLogger(t).With(log.ModuleKey, "cosmovisor")
	launcher, err := cosmovisor.NewLauncher(logger, cfg)
	require.NoError(t, err)
	return cfg, launcher
}

func TestCheckPendingUpgrade_NoFile(t *testing.T) {
	cfg, launcher := newPendingUpgradeLauncher(t)

	require.NoError(t, launcher.CheckPendingUpgrade())

	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err := filepath.EvalSymlinks(cfg.GenesisBin())
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
}

func TestCheckPendingUpgrade_StagedBinarySwitches(t *testing.T) {
	cfg, launcher := newPendingUpgradeLauncher(t)

	writePendingUpgrade(t, cfg, upgradetypes.Plan{Name: "chain2", Height: 49})
	require.NoError(t, launcher.CheckPendingUpgrade())

	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err := filepath.EvalSymlinks(cfg.UpgradeBin("chain2"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
}

func TestCheckPendingUpgrade_AlreadyCurrentIsNoop(t *testing.T) {
	cfg, launcher := newPendingUpgradeLauncher(t)

	require.NoError(t, cfg.SetCurrentUpgrade(upgradetypes.Plan{Name: "chain2", Height: 49}))
	writePendingUpgrade(t, cfg, upgradetypes.Plan{Name: "chain2", Height: 49})

	require.NoError(t, launcher.CheckPendingUpgrade())

	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err := filepath.EvalSymlinks(cfg.UpgradeBin("chain2"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
}

func TestCheckPendingUpgrade_MissingBinaryDefersToWatcher(t *testing.T) {
	cfg, launcher := newPendingUpgradeLauncher(t)

	writePendingUpgrade(t, cfg, upgradetypes.Plan{Name: "not-staged-yet", Height: 100})

	require.NoError(t, launcher.CheckPendingUpgrade())

	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err := filepath.EvalSymlinks(cfg.GenesisBin())
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
}

func TestCheckPendingUpgrade_MalformedFileErrors(t *testing.T) {
	cfg, launcher := newPendingUpgradeLauncher(t)

	require.NoError(t, os.MkdirAll(filepath.Dir(cfg.UpgradeInfoFilePath()), 0o755))
	require.NoError(t, os.WriteFile(cfg.UpgradeInfoFilePath(), []byte("not json"), 0o600))

	require.Error(t, launcher.CheckPendingUpgrade())
}

func TestCheckPendingUpgrade_EmptyNameIsNoop(t *testing.T) {
	cfg, launcher := newPendingUpgradeLauncher(t)

	writePendingUpgrade(t, cfg, upgradetypes.Plan{})

	require.NoError(t, launcher.CheckPendingUpgrade())

	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err := filepath.EvalSymlinks(cfg.GenesisBin())
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
}
