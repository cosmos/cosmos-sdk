//go:build linux || darwin

package internal

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"cosmossdk.io/log"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/tools/cosmovisor"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

var workDir string

func init() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	workDir = filepath.Join(dir, "..")
}

type launchProcessFixture struct {
	cfg    *cosmovisor.Config
	stdin  *os.File
	stdout *buffer
	stderr *buffer
	logger log.Logger
	runner *Runner
}

// TestLaunchProcess will try running the script a few times and watch upgrades work properly
// and args are passed through
func TestLaunchProcess(t *testing.T) {
	f := setupTestLaunchProcessFixture(t, "validate", cosmovisor.Config{
		Name:              "dummyd",
		PollInterval:      15,
		UnsafeSkipBackup:  true,
		MaxRestartRetries: 1,
	})

	currentBin, err := f.cfg.CurrentBin()
	require.NoError(t, err)

	rPath, err := filepath.EvalSymlinks(f.cfg.GenesisBin())
	require.NoError(t, err)

	require.Equal(t, rPath, currentBin)

	upgradeFile := f.cfg.UpgradeInfoFilePath()

	args := []string{"foo", "bar", "1234", upgradeFile}
	err = f.runner.Start(context.Background(), args)
	require.ErrorIs(t, err, ErrUpgradeNoDaemonRestart)
	require.Empty(t, f.stderr.String())
	require.Equal(t, fmt.Sprintf("Genesis foo bar 1234 %s\nUPGRADE \"chain2\" NEEDED at height: 49: {}\n", upgradeFile), f.stdout.String())

	// ensure this is upgraded now and produces new output
	currentBin, err = f.cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err = filepath.EvalSymlinks(f.cfg.UpgradeBin("chain2"))

	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
	args = []string{"second", "run", "--verbose"}
	f.stdout.Reset()
	f.stderr.Reset()

	err = f.runner.Start(context.Background(), args)
	require.ErrorContains(t, err, "maximum number of restarts reached")
	require.Empty(t, f.stderr.String())
	require.Equal(t, "Chain 2 is live!\nArgs: second run --verbose\nFinished successfully\n", f.stdout.String())

	// ended without other upgrade
	rPath, err = filepath.EvalSymlinks(f.cfg.UpgradeBin("chain2"))
	require.NoError(t, err)

	require.Equal(t, rPath, currentBin)
}

// TestPlanDisableRecase will test upgrades without lower case plan names
func TestPlanDisableRecase(t *testing.T) {
	f := setupTestLaunchProcessFixture(t, "norecase", cosmovisor.Config{
		Name:              "dummyd",
		PollInterval:      20,
		UnsafeSkipBackup:  true,
		DisableRecase:     true,
		MaxRestartRetries: 1,
	})

	currentBin, err := f.cfg.CurrentBin()
	require.NoError(t, err)

	rPath, err := filepath.EvalSymlinks(f.cfg.GenesisBin())
	require.NoError(t, err)

	require.Equal(t, rPath, currentBin)

	upgradeFile := f.cfg.UpgradeInfoFilePath()

	args := []string{"foo", "bar", "1234", upgradeFile}
	err = f.runner.Start(context.Background(), args)
	require.ErrorIs(t, err, ErrUpgradeNoDaemonRestart)
	require.Empty(t, f.stderr.String())
	require.Equal(t, fmt.Sprintf("Genesis foo bar 1234 %s\nUPGRADE \"Chain2\" NEEDED at height: 49: {}\n", upgradeFile), f.stdout.String())

	// ensure this is upgraded now and produces new output
	currentBin, err = f.cfg.CurrentBin()
	require.NoError(t, err)

	rPath, err = filepath.EvalSymlinks(f.cfg.UpgradeBin("Chain2"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
	args = []string{"second", "run", "--verbose"}
	f.stdout.Reset()
	f.stderr.Reset()

	err = f.runner.Start(context.Background(), args)
	require.ErrorContains(t, err, "maximum number of restarts reached")
	require.Empty(t, f.stderr.String())
	require.Equal(t, "Chain 2 is live!\nArgs: second run --verbose\nFinished successfully\n", f.stdout.String())

	// ended without other upgrade
	rPath, err = filepath.EvalSymlinks(f.cfg.UpgradeBin("Chain2"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
}

func TestLaunchProcessWithRestartDelay(t *testing.T) {
	f := setupTestLaunchProcessFixture(t, "validate", cosmovisor.Config{
		Name:             "dummyd",
		RestartDelay:     5 * time.Second,
		PollInterval:     20,
		UnsafeSkipBackup: true,
	})

	// should run the genesis binary and produce expected output
	currentBin, err := f.cfg.CurrentBin()
	require.NoError(t, err)

	rPath, err := filepath.EvalSymlinks(f.cfg.GenesisBin())
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)

	upgradeFile := f.cfg.UpgradeInfoFilePath()

	start := time.Now()
	err = f.runner.Start(context.Background(), []string{"foo", "bar", "1234", upgradeFile})
	require.ErrorIs(t, err, ErrUpgradeNoDaemonRestart)

	// may not be the best way but the fastest way to check we meet the delay
	// in addition to comparing both the runtime of this test and TestLaunchProcess in addition
	if time.Since(start) < f.cfg.RestartDelay {
		require.FailNow(t, "restart delay not met")
	}
}

// TestPlanShutdownGrace will test upgrades without lower case plan names
func TestPlanShutdownGrace(t *testing.T) {
	f := setupTestLaunchProcessFixture(t, "dontdie", cosmovisor.Config{
		Name:              "dummyd",
		PollInterval:      15,
		UnsafeSkipBackup:  true,
		ShutdownGrace:     2 * time.Second,
		MaxRestartRetries: 1,
	})

	// should run the genesis binary and produce expected output
	currentBin, err := f.cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err := filepath.EvalSymlinks(f.cfg.GenesisBin())
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)

	upgradeFile := f.cfg.UpgradeInfoFilePath()

	args := []string{"foo", "bar", "1234", upgradeFile}
	err = f.runner.Start(context.Background(), args)
	require.ErrorIs(t, err, ErrUpgradeNoDaemonRestart)
	require.Empty(t, f.stderr.String())
	require.Equal(t, fmt.Sprintf("Genesis foo bar 1234 %s\nUPGRADE \"Chain2\" NEEDED at height: 49: {}\nWARN Need Flush\nFlushed\n", upgradeFile), f.stdout.String())

	// ensure this is upgraded now and produces new output
	currentBin, err = f.cfg.CurrentBin()
	require.NoError(t, err)

	rPath, err = filepath.EvalSymlinks(f.cfg.UpgradeBin("chain2"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
	args = []string{"second", "run", "--verbose"}
	f.stdout.Reset()
	f.stderr.Reset()

	err = f.runner.Start(context.Background(), args)
	require.ErrorContains(t, err, "maximum number of restarts reached")
	require.Empty(t, f.stderr.String())
	require.Equal(t, "Chain 2 is live!\nArgs: second run --verbose\nFinished successfully\n", f.stdout.String())

	// ended without other upgrade
	rPath, err = filepath.EvalSymlinks(f.cfg.UpgradeBin("chain2"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
}

// TestLaunchProcess will try running the script a few times and watch upgrades work properly
// and args are passed through
func TestLaunchProcessWithDownloads(t *testing.T) {
	f := setupTestLaunchProcessFixture(t, "download", cosmovisor.Config{
		Name:                  "autod",
		AllowDownloadBinaries: true,
		PollInterval:          100,
		UnsafeSkipBackup:      true,
		MaxRestartRetries:     1,
	})

	// test case upgrade path (binaries from testdata/download directory):
	// genesis -> chain2-zip_bin
	// chain2-zip_bin -> ref_to_chain3-zip_dir.json = (json for the next download instructions) -> chain3-zip_dir
	// chain3-zip_dir - doesn't upgrade
	upgradeFilename := f.cfg.UpgradeInfoFilePath()

	// should run the genesis binary and produce expected output
	currentBin, err := f.cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err := filepath.EvalSymlinks(f.cfg.GenesisBin())

	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)

	args := []string{"some", "args", upgradeFilename}
	err = f.runner.Start(context.Background(), args)
	require.ErrorIs(t, err, ErrUpgradeNoDaemonRestart)
	require.Empty(t, f.stderr.String())
	require.Equal(t, "Genesis autod. Args: some args "+upgradeFilename+"\n"+`ERROR: UPGRADE "chain2" NEEDED at height: 49: zip_binary`+"\n", f.stdout.String())
	currentBin, err = f.cfg.CurrentBin()
	require.NoError(t, err)

	rPath, err = filepath.EvalSymlinks(f.cfg.UpgradeBin("chain2"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)

	// start chain2
	f.stdout.Reset()
	f.stderr.Reset()
	args = []string{"run", "--fast", upgradeFilename}
	err = f.runner.Start(context.Background(), args)
	// ended with one more upgrade
	require.ErrorIs(t, err, ErrUpgradeNoDaemonRestart)
	require.Empty(t, f.stderr.String())
	require.Equal(t, "Chain 2 from zipped binary\nArgs: run --fast "+upgradeFilename+"\n"+`ERROR: UPGRADE "chain3" NEEDED at height: 936: ref_to_chain3-zip_dir.json module=main`+"\n", f.stdout.String())
	currentBin, err = f.cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err = filepath.EvalSymlinks(f.cfg.UpgradeBin("chain3"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)

	// run the last chain
	args = []string{"end", "--halt", upgradeFilename}
	f.stdout.Reset()
	f.stderr.Reset()
	err = f.runner.Start(context.Background(), args)
	require.ErrorContains(t, err, "maximum number of restarts reached")
	require.Empty(t, f.stderr.String())
	require.Equal(t, "Chain 3 from zipped directory\nArgs: end --halt "+upgradeFilename+"\n", f.stdout.String())

	// and this doesn't upgrade
	currentBin, err = f.cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err = filepath.EvalSymlinks(f.cfg.UpgradeBin("chain3"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
}

// TestLaunchProcessWithDownloadsAndMissingPreupgrade will try running the script a few times and watch upgrades work properly
// and args are passed through
func TestLaunchProcessWithDownloadsAndMissingPreupgrade(t *testing.T) {
	// test case upgrade path (binaries from testdata/download directory):
	// genesis -> chain2-zip_bin
	// chain2-zip_bin -> ref_to_chain3-zip_dir.json = (json for the next download instructions) -> chain3-zip_dir
	// chain3-zip_dir - doesn't upgrade
	f := setupTestLaunchProcessFixture(t, "download", cosmovisor.Config{
		Name:                  "autod",
		AllowDownloadBinaries: true,
		PollInterval:          100,
		UnsafeSkipBackup:      true,
		CustomPreUpgrade:      "missing.sh",
	})

	upgradeFilename := f.cfg.UpgradeInfoFilePath()

	// should run the genesis binary and produce expected output
	currentBin, err := f.cfg.CurrentBin()
	require.NoError(t, err)

	rPath, err := filepath.EvalSymlinks(f.cfg.GenesisBin())
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
	require.NoError(t, err)

	// Missing Preupgrade Script
	args := []string{"some", "args", upgradeFilename}
	err = f.runner.Start(context.Background(), args)

	require.ErrorContains(t, err, "missing.sh")
	require.ErrorIs(t, err, fs.ErrNotExist)
}

// TestLaunchProcessWithDownloadsAndPreupgrade will try running the script a few times and watch upgrades work properly
// and args are passed through
func TestLaunchProcessWithDownloadsAndPreupgrade(t *testing.T) {
	// test case upgrade path (binaries from testdata/download directory):
	// genesis -> chain2-zip_bin
	// chain2-zip_bin -> ref_to_chain3-zip_dir.json = (json for the next download instructions) -> chain3-zip_dir
	// chain3-zip_dir - doesn't upgrade
	f := setupTestLaunchProcessFixture(t, "download", cosmovisor.Config{
		Name:                  "autod",
		AllowDownloadBinaries: true,
		PollInterval:          100,
		UnsafeSkipBackup:      true,
		CustomPreUpgrade:      "preupgrade.sh",
		MaxRestartRetries:     1,
	})

	upgradeFilename := f.cfg.UpgradeInfoFilePath()

	// should run the genesis binary and produce expected output
	currentBin, err := f.cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err := filepath.EvalSymlinks(f.cfg.GenesisBin())
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)

	args := []string{"some", "args", upgradeFilename}
	err = f.runner.Start(context.Background(), args)

	require.ErrorIs(t, err, ErrUpgradeNoDaemonRestart)
	require.Empty(t, f.stderr.String())
	require.Equal(t, "Genesis autod. Args: some args "+upgradeFilename+"\n"+`ERROR: UPGRADE "chain2" NEEDED at height: 49: zip_binary`+"\n", f.stdout.String())
	currentBin, err = f.cfg.CurrentBin()
	require.NoError(t, err)

	rPath, err = filepath.EvalSymlinks(f.cfg.UpgradeBin("chain2"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)

	// should have preupgrade.sh results
	require.FileExists(t, filepath.Join(f.cfg.Home, "upgrade_name_chain2_height_49"))

	// start chain2
	f.stdout.Reset()
	f.stderr.Reset()
	args = []string{"run", "--fast", upgradeFilename}
	err = f.runner.Start(context.Background(), args)
	// ended with one more upgrade
	require.ErrorIs(t, err, ErrUpgradeNoDaemonRestart)
	require.Empty(t, f.stderr.String())
	require.Equal(t, "Chain 2 from zipped binary\nArgs: run --fast "+upgradeFilename+"\n"+`ERROR: UPGRADE "chain3" NEEDED at height: 936: ref_to_chain3-zip_dir.json module=main`+"\n", f.stdout.String())
	currentBin, err = f.cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err = filepath.EvalSymlinks(f.cfg.UpgradeBin("chain3"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)

	// should have preupgrade.sh results
	require.FileExists(t, filepath.Join(f.cfg.Home, "upgrade_name_chain3_height_936"))

	// run the last chain
	args = []string{"end", "--halt", upgradeFilename}
	f.stdout.Reset()
	f.stderr.Reset()
	err = f.runner.Start(context.Background(), args)
	require.ErrorContains(t, err, "maximum number of restarts reached")
	require.Empty(t, f.stderr.String())
	require.Equal(t, "Chain 3 from zipped directory\nArgs: end --halt "+upgradeFilename+"\n", f.stdout.String())

	// and this doesn't upgrade
	currentBin, err = f.cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err = filepath.EvalSymlinks(f.cfg.UpgradeBin("chain3"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
}

// TestSkipUpgrade tests heights that are identified to be skipped and return if upgrade height matches the skip heights
func TestSkipUpgrade(t *testing.T) {
	cases := []struct {
		args        []string
		upgradeInfo upgradetypes.Plan
		expectRes   bool
	}{{
		args:        []string{"appb", "start", "--unsafe-skip-upgrades"},
		upgradeInfo: upgradetypes.Plan{Name: "upgrade1", Info: "some info", Height: 123},
		expectRes:   false,
	}, {
		args:        []string{"appb", "start", "--unsafe-skip-upgrades", "--abcd"},
		upgradeInfo: upgradetypes.Plan{Name: "upgrade1", Info: "some info", Height: 123},
		expectRes:   false,
	}, {
		args:        []string{"appb", "start", "--unsafe-skip-upgrades", "10", "--abcd"},
		upgradeInfo: upgradetypes.Plan{Name: "upgrade1", Info: "some info", Height: 11},
		expectRes:   false,
	}, {
		args:        []string{"appb", "start", "--unsafe-skip-upgrades", "10", "20", "--abcd"},
		upgradeInfo: upgradetypes.Plan{Name: "upgrade1", Info: "some info", Height: 20},
		expectRes:   true,
	}, {
		args:        []string{"appb", "start", "--unsafe-skip-upgrades", "10", "20", "--abcd", "34"},
		upgradeInfo: upgradetypes.Plan{Name: "upgrade1", Info: "some info", Height: 34},
		expectRes:   false,
	}}

	for i := range cases {
		tc := cases[i]
		require := require.New(t)
		h := IsSkipUpgradeHeight(tc.args, tc.upgradeInfo)
		require.Equal(h, tc.expectRes)
	}
}

// TestUpgradeSkipHeights tests if correct skip upgrade heights are identified from the cli args
func TestUpgradeSkipHeights(t *testing.T) {
	cases := []struct {
		args      []string
		expectRes []int
	}{{
		args:      []string{},
		expectRes: nil,
	}, {
		args:      []string{"appb", "start"},
		expectRes: nil,
	}, {
		args:      []string{"appb", "start", "--unsafe-skip-upgrades"},
		expectRes: nil,
	}, {
		args:      []string{"appb", "start", "--unsafe-skip-upgrades", "--abcd"},
		expectRes: nil,
	}, {
		args:      []string{"appb", "start", "--unsafe-skip-upgrades", "10", "--abcd"},
		expectRes: []int{10},
	}, {
		args:      []string{"appb", "start", "--unsafe-skip-upgrades", "10", "20", "--abcd"},
		expectRes: []int{10, 20},
	}, {
		args:      []string{"appb", "start", "--unsafe-skip-upgrades", "10", "20", "--abcd", "34"},
		expectRes: []int{10, 20},
	}, {
		args:      []string{"appb", "start", "--unsafe-skip-upgrades", "10", "as", "20", "--abcd"},
		expectRes: []int{10, 20},
	}}

	for i := range cases {
		tc := cases[i]
		require := require.New(t)
		h := UpgradeSkipHeights(tc.args)
		require.Equal(h, tc.expectRes)
	}
}

// buffer is a thread safe bytes buffer
type buffer struct {
	b bytes.Buffer
	m sync.Mutex
}

func setupTestLaunchProcessFixture(t *testing.T, testdataDir string, cfg cosmovisor.Config) *launchProcessFixture {
	// binaries from testdata/validate directory
	preppedCfg := prepareConfig(
		t,
		fmt.Sprintf("%s/testdata/%s", workDir, testdataDir),
		cfg,
	)

	logger := log.NewTestLogger(t).With(log.ModuleKey, "cosmosvisor")

	// should run the genesis binary and produce expected output
	stdin, _ := os.Open(os.DevNull)
	stdout, stderr := newBuffer(), newBuffer()

	return &launchProcessFixture{
		cfg:    preppedCfg,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		logger: logger,
		runner: NewRunner(preppedCfg, RunConfig{
			StdIn:  stdin,
			StdOut: stdout,
			StdErr: stderr,
		}, logger),
	}
}

func newBuffer() *buffer {
	return &buffer{}
}

func (b *buffer) Write(bz []byte) (int, error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(bz)
}

func (b *buffer) String() string {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.String()
}

func (b *buffer) Reset() {
	b.m.Lock()
	defer b.m.Unlock()
	b.b.Reset()
}
