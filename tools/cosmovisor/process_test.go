//go:build linux
// +build linux

package cosmovisor_test

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/tools/cosmovisor"
	upgradetypes "cosmossdk.io/x/upgrade/types"
)

// TestLaunchProcess will try running the script a few times and watch upgrades work properly
// and args are passed through
func TestLaunchProcess(t *testing.T) {
	// binaries from testdata/validate directory
	home := copyTestData(t, "validate")
	cfg := &cosmovisor.Config{Home: home, Name: "dummyd", PollInterval: 15, UnsafeSkipBackup: true}
	logger := log.NewTestLogger(t).With(log.ModuleKey, "cosmosvisor")

	// should run the genesis binary and produce expected output
	stdin, _ := os.Open(os.DevNull)
	stdout, stderr := newBuffer(), newBuffer()
	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)
	require.Equal(t, cfg.GenesisBin(), currentBin)

	launcher, err := cosmovisor.NewLauncher(logger, cfg)
	require.NoError(t, err)

	upgradeFile := cfg.UpgradeInfoFilePath()

	args := []string{"foo", "bar", "1234", upgradeFile}
	doUpgrade, err := launcher.Run(args, stdin, stdout, stderr)
	require.NoError(t, err)
	require.True(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, fmt.Sprintf("Genesis foo bar 1234 %s\nUPGRADE \"chain2\" NEEDED at height: 49: {}\n", upgradeFile), stdout.String())

	// ensure this is upgraded now and produces new output
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)

	require.Equal(t, cfg.UpgradeBin("chain2"), currentBin)
	args = []string{"second", "run", "--verbose"}
	stdout.Reset()
	stderr.Reset()

	doUpgrade, err = launcher.Run(args, stdin, stdout, stderr)
	require.NoError(t, err)
	require.False(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, "Chain 2 is live!\nArgs: second run --verbose\nFinished successfully\n", stdout.String())

	// ended without other upgrade
	require.Equal(t, cfg.UpgradeBin("chain2"), currentBin)
}

// TestPlanDisableRecase will test upgrades without lower case plan names
func TestPlanDisableRecase(t *testing.T) {
	// binaries from testdata/validate directory
	home := copyTestData(t, "norecase")
	cfg := &cosmovisor.Config{Home: home, Name: "dummyd", PollInterval: 20, UnsafeSkipBackup: true, DisableRecase: true}
	logger := log.NewTestLogger(t).With(log.ModuleKey, "cosmosvisor")

	// should run the genesis binary and produce expected output
	stdin, _ := os.Open(os.DevNull)
	stdout, stderr := newBuffer(), newBuffer()
	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)
	require.Equal(t, cfg.GenesisBin(), currentBin)

	launcher, err := cosmovisor.NewLauncher(logger, cfg)
	require.NoError(t, err)

	upgradeFile := cfg.UpgradeInfoFilePath()

	args := []string{"foo", "bar", "1234", upgradeFile}
	doUpgrade, err := launcher.Run(args, stdin, stdout, stderr)
	require.NoError(t, err)
	require.True(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, fmt.Sprintf("Genesis foo bar 1234 %s\nUPGRADE \"Chain2\" NEEDED at height: 49: {}\n", upgradeFile), stdout.String())

	// ensure this is upgraded now and produces new output
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)

	require.Equal(t, cfg.UpgradeBin("Chain2"), currentBin)
	args = []string{"second", "run", "--verbose"}
	stdout.Reset()
	stderr.Reset()

	doUpgrade, err = launcher.Run(args, stdin, stdout, stderr)
	require.NoError(t, err)
	require.False(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, "Chain 2 is live!\nArgs: second run --verbose\nFinished successfully\n", stdout.String())

	// ended without other upgrade
	require.Equal(t, cfg.UpgradeBin("Chain2"), currentBin)
}

func TestLaunchProcessWithRestartDelay(t *testing.T) {
	// binaries from testdata/validate directory
	home := copyTestData(t, "validate")
	cfg := &cosmovisor.Config{Home: home, Name: "dummyd", RestartDelay: 5 * time.Second, PollInterval: 20, UnsafeSkipBackup: true}
	logger := log.NewTestLogger(t).With(log.ModuleKey, "cosmosvisor")

	// should run the genesis binary and produce expected output
	stdin, _ := os.Open(os.DevNull)
	stdout, stderr := newBuffer(), newBuffer()
	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)
	require.Equal(t, cfg.GenesisBin(), currentBin)

	launcher, err := cosmovisor.NewLauncher(logger, cfg)
	require.NoError(t, err)

	upgradeFile := cfg.UpgradeInfoFilePath()

	start := time.Now()
	doUpgrade, err := launcher.Run([]string{"foo", "bar", "1234", upgradeFile}, stdin, stdout, stderr)
	require.NoError(t, err)
	require.True(t, doUpgrade)

	// may not be the best way but the fastest way to check we meet the delay
	// in addition to comparing both the runtime of this test and TestLaunchProcess in addition
	if time.Since(start) < cfg.RestartDelay {
		require.FailNow(t, "restart delay not met")
	}
}

// TestPlanShutdownGrace will test upgrades without lower case plan names
func TestPlanShutdownGrace(t *testing.T) {
	// binaries from testdata/validate directory
	home := copyTestData(t, "dontdie")
	cfg := &cosmovisor.Config{Home: home, Name: "dummyd", PollInterval: 15, UnsafeSkipBackup: true, ShutdownGrace: 2 * time.Second}
	logger := log.NewTestLogger(t).With(log.ModuleKey, "cosmosvisor")

	// should run the genesis binary and produce expected output
	stdin, _ := os.Open(os.DevNull)
	stdout, stderr := newBuffer(), newBuffer()
	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)
	require.Equal(t, cfg.GenesisBin(), currentBin)

	launcher, err := cosmovisor.NewLauncher(logger, cfg)
	require.NoError(t, err)

	upgradeFile := cfg.UpgradeInfoFilePath()

	args := []string{"foo", "bar", "1234", upgradeFile}
	doUpgrade, err := launcher.Run(args, stdin, stdout, stderr)
	require.NoError(t, err)
	require.True(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, fmt.Sprintf("Genesis foo bar 1234 %s\nUPGRADE \"Chain2\" NEEDED at height: 49: {}\nWARN Need Flush\nFlushed\n", upgradeFile), stdout.String())

	// ensure this is upgraded now and produces new output
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)

	require.Equal(t, cfg.UpgradeBin("chain2"), currentBin)
	args = []string{"second", "run", "--verbose"}
	stdout.Reset()
	stderr.Reset()

	doUpgrade, err = launcher.Run(args, stdin, stdout, stderr)
	require.NoError(t, err)
	require.False(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, "Chain 2 is live!\nArgs: second run --verbose\nFinished successfully\n", stdout.String())

	// ended without other upgrade
	require.Equal(t, cfg.UpgradeBin("chain2"), currentBin)
}

// TestLaunchProcess will try running the script a few times and watch upgrades work properly
// and args are passed through
func TestLaunchProcessWithDownloads(t *testing.T) {
	// test case upgrade path (binaries from testdata/download directory):
	// genesis -> chain2-zip_bin
	// chain2-zip_bin -> ref_to_chain3-zip_dir.json = (json for the next download instructions) -> chain3-zip_dir
	// chain3-zip_dir - doesn't upgrade
	home := copyTestData(t, "download")
	cfg := &cosmovisor.Config{Home: home, Name: "autod", AllowDownloadBinaries: true, PollInterval: 100, UnsafeSkipBackup: true}
	logger := log.NewTestLogger(t).With(log.ModuleKey, "cosmovisor")
	upgradeFilename := cfg.UpgradeInfoFilePath()

	// should run the genesis binary and produce expected output
	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)
	require.Equal(t, cfg.GenesisBin(), currentBin)

	launcher, err := cosmovisor.NewLauncher(logger, cfg)
	require.NoError(t, err)

	stdin, _ := os.Open(os.DevNull)
	stdout, stderr := newBuffer(), newBuffer()
	args := []string{"some", "args", upgradeFilename}
	doUpgrade, err := launcher.Run(args, stdin, stdout, stderr)
	require.NoError(t, err)
	require.True(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, "Genesis autod. Args: some args "+upgradeFilename+"\n"+`ERROR: UPGRADE "chain2" NEEDED at height: 49: zip_binary`+"\n", stdout.String())
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)
	require.Equal(t, cfg.UpgradeBin("chain2"), currentBin)

	// start chain2
	stdout.Reset()
	stderr.Reset()
	args = []string{"run", "--fast", upgradeFilename}
	doUpgrade, err = launcher.Run(args, stdin, stdout, stderr)
	require.NoError(t, err)

	require.Empty(t, stderr.String())
	require.Equal(t, "Chain 2 from zipped binary\nArgs: run --fast "+upgradeFilename+"\n"+`ERROR: UPGRADE "chain3" NEEDED at height: 936: ref_to_chain3-zip_dir.json module=main`+"\n", stdout.String())
	// ended with one more upgrade
	require.True(t, doUpgrade)
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)
	require.Equal(t, cfg.UpgradeBin("chain3"), currentBin)

	// run the last chain
	args = []string{"end", "--halt", upgradeFilename}
	stdout.Reset()
	stderr.Reset()
	doUpgrade, err = launcher.Run(args, stdin, stdout, stderr)
	require.NoError(t, err)
	require.False(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, "Chain 3 from zipped directory\nArgs: end --halt "+upgradeFilename+"\n", stdout.String())

	// and this doesn't upgrade
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)
	require.Equal(t, cfg.UpgradeBin("chain3"), currentBin)
}

// TestLaunchProcessWithDownloadsAndMissingPreupgrade will try running the script a few times and watch upgrades work properly
// and args are passed through
func TestLaunchProcessWithDownloadsAndMissingPreupgrade(t *testing.T) {
	// test case upgrade path (binaries from testdata/download directory):
	// genesis -> chain2-zip_bin
	// chain2-zip_bin -> ref_to_chain3-zip_dir.json = (json for the next download instructions) -> chain3-zip_dir
	// chain3-zip_dir - doesn't upgrade
	home := copyTestData(t, "download")
	cfg := &cosmovisor.Config{
		Home:                  home,
		Name:                  "autod",
		AllowDownloadBinaries: true,
		PollInterval:          100,
		UnsafeSkipBackup:      true,
		CustomPreUpgrade:      "missing.sh",
	}
	logger := log.NewTestLogger(t).With(log.ModuleKey, "cosmovisor")
	upgradeFilename := cfg.UpgradeInfoFilePath()

	// should run the genesis binary and produce expected output
	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)
	require.Equal(t, cfg.GenesisBin(), currentBin)
	launcher, err := cosmovisor.NewLauncher(logger, cfg)
	require.NoError(t, err)

	// Missing Preupgrade Script
	stdin, _ := os.Open(os.DevNull)
	stdout, stderr := newBuffer(), newBuffer()
	args := []string{"some", "args", upgradeFilename}
	_, err = launcher.Run(args, stdin, stdout, stderr)

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
	home := copyTestData(t, "download")
	cfg := &cosmovisor.Config{
		Home:                  home,
		Name:                  "autod",
		AllowDownloadBinaries: true,
		PollInterval:          100,
		UnsafeSkipBackup:      true,
		CustomPreUpgrade:      "preupgrade.sh",
	}
	buf := newBuffer() // inspect output using buf.String()
	logger := log.NewLogger(buf).With(log.ModuleKey, "cosmovisor")
	upgradeFilename := cfg.UpgradeInfoFilePath()

	// should run the genesis binary and produce expected output
	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)
	require.Equal(t, cfg.GenesisBin(), currentBin)
	launcher, err := cosmovisor.NewLauncher(logger, cfg)
	require.NoError(t, err)

	stdin, _ := os.Open(os.DevNull)
	stdout, stderr := newBuffer(), newBuffer()
	args := []string{"some", "args", upgradeFilename}
	doUpgrade, err := launcher.Run(args, stdin, stdout, stderr)

	require.NoError(t, err)
	require.True(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, "Genesis autod. Args: some args "+upgradeFilename+"\n"+`ERROR: UPGRADE "chain2" NEEDED at height: 49: zip_binary`+"\n", stdout.String())
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)
	require.Equal(t, cfg.UpgradeBin("chain2"), currentBin)

	// should have preupgrade.sh results
	require.FileExists(t, filepath.Join(home, "upgrade_name_chain2_height_49"))

	// start chain2
	stdout.Reset()
	stderr.Reset()
	args = []string{"run", "--fast", upgradeFilename}
	doUpgrade, err = launcher.Run(args, stdin, stdout, stderr)
	require.NoError(t, err)

	require.Empty(t, stderr.String())
	require.Equal(t, "Chain 2 from zipped binary\nArgs: run --fast "+upgradeFilename+"\n"+`ERROR: UPGRADE "chain3" NEEDED at height: 936: ref_to_chain3-zip_dir.json module=main`+"\n", stdout.String())
	// ended with one more upgrade
	require.True(t, doUpgrade)
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)
	require.Equal(t, cfg.UpgradeBin("chain3"), currentBin)

	// should have preupgrade.sh results
	require.FileExists(t, filepath.Join(home, "upgrade_name_chain3_height_936"))

	// run the last chain
	args = []string{"end", "--halt", upgradeFilename}
	stdout.Reset()
	stderr.Reset()
	doUpgrade, err = launcher.Run(args, stdin, stdout, stderr)
	require.NoError(t, err)
	require.False(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, "Chain 3 from zipped directory\nArgs: end --halt "+upgradeFilename+"\n", stdout.String())

	// and this doesn't upgrade
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)
	require.Equal(t, cfg.UpgradeBin("chain3"), currentBin)
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
		h := cosmovisor.IsSkipUpgradeHeight(tc.args, tc.upgradeInfo)
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
		h := cosmovisor.UpgradeSkipHeights(tc.args)
		require.Equal(h, tc.expectRes)
	}
}

// buffer is a thread safe bytes buffer
type buffer struct {
	b bytes.Buffer
	m sync.Mutex
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
