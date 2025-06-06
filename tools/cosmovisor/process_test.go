//go:build linux || darwin

package cosmovisor_test

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
	"cosmossdk.io/tools/cosmovisor/internal"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

var workDir string

func init() {
	workDir, _ = os.Getwd()
}

// TODO all these tests share the same setup so we can extract it to a common function

// TestLaunchProcess will try running the script a few times and watch upgrades work properly
// and args are passed through
func TestLaunchProcess(t *testing.T) {
	// binaries from testdata/validate directory
	cfg := prepareConfig(
		t,
		fmt.Sprintf("%s/%s", workDir, "testdata/validate"),
		cosmovisor.Config{
			Name:              "dummyd",
			PollInterval:      15,
			UnsafeSkipBackup:  true,
			MaxRestartRetries: 1,
		},
	)

	logger := log.NewTestLogger(t).With(log.ModuleKey, "cosmosvisor")

	// should run the genesis binary and produce expected output
	stdin, _ := os.Open(os.DevNull)
	stdout, stderr := newBuffer(), newBuffer()
	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)

	rPath, err := filepath.EvalSymlinks(cfg.GenesisBin())
	require.NoError(t, err)

	require.Equal(t, rPath, currentBin)

	runCfg := internal.RunConfig{
		StdIn:  stdin,
		StdOut: stdout,
		StdErr: stderr,
	}

	runner := internal.NewRunner(cfg, runCfg, logger)
	upgradeFile := cfg.UpgradeInfoFilePath()

	args := []string{"foo", "bar", "1234", upgradeFile}
	err = runner.Start(context.Background(), args)
	require.NoError(t, err)
	//doUpgrade, err := launcher.Run(args, stdin, stdout, stderr)
	//require.NoError(t, err)
	//require.True(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, fmt.Sprintf("Genesis foo bar 1234 %s\nUPGRADE \"chain2\" NEEDED at height: 49: {}\n", upgradeFile), stdout.String())

	// ensure this is upgraded now and produces new output
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("chain2"))

	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
	args = []string{"second", "run", "--verbose"}
	stdout.Reset()
	stderr.Reset()

	err = runner.Start(context.Background(), args)
	require.ErrorContains(t, err, "maximum number of restarts reached")
	//doUpgrade, err = launcher.Run(args, stdin, stdout, stderr)
	//require.False(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, "Chain 2 is live!\nArgs: second run --verbose\nFinished successfully\n", stdout.String())

	// ended without other upgrade
	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("chain2"))
	require.NoError(t, err)

	require.Equal(t, rPath, currentBin)
}

// TestPlanDisableRecase will test upgrades without lower case plan names
func TestPlanDisableRecase(t *testing.T) {
	// binaries from testdata/validate directory
	cfg := prepareConfig(
		t,
		fmt.Sprintf("%s/%s", workDir, "testdata/norecase"),
		cosmovisor.Config{
			Name:              "dummyd",
			PollInterval:      20,
			UnsafeSkipBackup:  true,
			DisableRecase:     true,
			MaxRestartRetries: 1,
		},
	)

	logger := log.NewTestLogger(t).With(log.ModuleKey, "cosmosvisor")

	// should run the genesis binary and produce expected output
	stdin, _ := os.Open(os.DevNull)
	stdout, stderr := newBuffer(), newBuffer()
	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)

	rPath, err := filepath.EvalSymlinks(cfg.GenesisBin())
	require.NoError(t, err)

	require.Equal(t, rPath, currentBin)

	runCfg := internal.RunConfig{
		StdIn:  stdin,
		StdOut: stdout,
		StdErr: stderr,
	}
	runner := internal.NewRunner(cfg, runCfg, logger)

	upgradeFile := cfg.UpgradeInfoFilePath()

	args := []string{"foo", "bar", "1234", upgradeFile}
	//doUpgrade, err := launcher.Run(args, stdin, stdout, stderr)
	err = runner.Start(context.Background(), args)
	require.NoError(t, err)
	//require.True(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, fmt.Sprintf("Genesis foo bar 1234 %s\nUPGRADE \"Chain2\" NEEDED at height: 49: {}\n", upgradeFile), stdout.String())

	// ensure this is upgraded now and produces new output
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)

	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("Chain2"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
	args = []string{"second", "run", "--verbose"}
	stdout.Reset()
	stderr.Reset()

	err = runner.Start(context.Background(), args)
	//runner.RunProcess(ctx, args)
	require.ErrorContains(t, err, "maximum number of restarts reached")
	//require.False(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, "Chain 2 is live!\nArgs: second run --verbose\nFinished successfully\n", stdout.String())

	// ended without other upgrade
	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("Chain2"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
}

func TestLaunchProcessWithRestartDelay(t *testing.T) {
	// binaries from testdata/validate directory
	cfg := prepareConfig(
		t,
		fmt.Sprintf("%s/%s", workDir, "testdata/validate"),
		cosmovisor.Config{
			Name:             "dummyd",
			RestartDelay:     5 * time.Second,
			PollInterval:     20,
			UnsafeSkipBackup: true,
		},
	)

	logger := log.NewTestLogger(t).With(log.ModuleKey, "cosmosvisor")

	// should run the genesis binary and produce expected output
	stdin, _ := os.Open(os.DevNull)
	stdout, stderr := newBuffer(), newBuffer()
	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)

	rPath, err := filepath.EvalSymlinks(cfg.GenesisBin())
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)

	runner := internal.NewRunner(cfg, internal.RunConfig{
		StdIn:  stdin,
		StdOut: stdout,
		StdErr: stderr,
	}, logger)
	require.NoError(t, err)

	upgradeFile := cfg.UpgradeInfoFilePath()

	start := time.Now()
	err = runner.Start(context.Background(), []string{"foo", "bar", "1234", upgradeFile})
	require.NoError(t, err)
	//require.True(t, doUpgrade)

	// may not be the best way but the fastest way to check we meet the delay
	// in addition to comparing both the runtime of this test and TestLaunchProcess in addition
	if time.Since(start) < cfg.RestartDelay {
		require.FailNow(t, "restart delay not met")
	}
}

// TestPlanShutdownGrace will test upgrades without lower case plan names
func TestPlanShutdownGrace(t *testing.T) {
	// binaries from testdata/validate directory
	cfg := prepareConfig(
		t,
		fmt.Sprintf("%s/%s", workDir, "testdata/dontdie"),
		cosmovisor.Config{
			Name:              "dummyd",
			PollInterval:      15,
			UnsafeSkipBackup:  true,
			ShutdownGrace:     2 * time.Second,
			MaxRestartRetries: 1,
		},
	)

	logger := log.NewTestLogger(t).With(log.ModuleKey, "cosmosvisor")

	// should run the genesis binary and produce expected output
	stdin, _ := os.Open(os.DevNull)
	stdout, stderr := newBuffer(), newBuffer()
	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err := filepath.EvalSymlinks(cfg.GenesisBin())
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)

	runner := internal.NewRunner(cfg, internal.RunConfig{
		StdIn:  stdin,
		StdOut: stdout,
		StdErr: stderr,
	}, logger)

	upgradeFile := cfg.UpgradeInfoFilePath()

	args := []string{"foo", "bar", "1234", upgradeFile}
	err = runner.Start(context.Background(), args)
	require.NoError(t, err)
	//require.True(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, fmt.Sprintf("Genesis foo bar 1234 %s\nUPGRADE \"Chain2\" NEEDED at height: 49: {}\nWARN Need Flush\nFlushed\n", upgradeFile), stdout.String())

	// ensure this is upgraded now and produces new output
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)

	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("chain2"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
	args = []string{"second", "run", "--verbose"}
	stdout.Reset()
	stderr.Reset()

	err = runner.Start(context.Background(), args)
	require.ErrorContains(t, err, "maximum number of restarts reached")
	//require.False(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, "Chain 2 is live!\nArgs: second run --verbose\nFinished successfully\n", stdout.String())

	// ended without other upgrade
	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("chain2"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
}

// TestLaunchProcess will try running the script a few times and watch upgrades work properly
// and args are passed through
func TestLaunchProcessWithDownloads(t *testing.T) {
	// test case upgrade path (binaries from testdata/download directory):
	// genesis -> chain2-zip_bin
	// chain2-zip_bin -> ref_to_chain3-zip_dir.json = (json for the next download instructions) -> chain3-zip_dir
	// chain3-zip_dir - doesn't upgrade
	cfg := prepareConfig(
		t,
		fmt.Sprintf("%s/%s", workDir, "testdata/download"),
		cosmovisor.Config{
			Name:                  "autod",
			AllowDownloadBinaries: true,
			PollInterval:          100,
			UnsafeSkipBackup:      true,
			MaxRestartRetries:     1,
		},
	)

	logger := log.NewTestLogger(t).With(log.ModuleKey, "cosmovisor")
	upgradeFilename := cfg.UpgradeInfoFilePath()

	// should run the genesis binary and produce expected output
	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err := filepath.EvalSymlinks(cfg.GenesisBin())
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)

	stdin, _ := os.Open(os.DevNull)
	stdout, stderr := newBuffer(), newBuffer()
	launcher := internal.NewRunner(cfg, internal.RunConfig{
		StdIn:  stdin,
		StdOut: stdout,
		StdErr: stderr,
	}, logger)

	args := []string{"some", "args", upgradeFilename}
	err = launcher.Start(context.Background(), args)
	require.NoError(t, err)
	//require.True(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, "Genesis autod. Args: some args "+upgradeFilename+"\n"+`ERROR: UPGRADE "chain2" NEEDED at height: 49: zip_binary`+"\n", stdout.String())
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)

	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("chain2"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)

	// start chain2
	stdout.Reset()
	stderr.Reset()
	args = []string{"run", "--fast", upgradeFilename}
	err = launcher.Start(context.Background(), args)
	require.NoError(t, err)

	require.Empty(t, stderr.String())
	require.Equal(t, "Chain 2 from zipped binary\nArgs: run --fast "+upgradeFilename+"\n"+`ERROR: UPGRADE "chain3" NEEDED at height: 936: ref_to_chain3-zip_dir.json module=main`+"\n", stdout.String())
	// ended with one more upgrade
	//require.True(t, doUpgrade)
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("chain3"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)

	// run the last chain
	args = []string{"end", "--halt", upgradeFilename}
	stdout.Reset()
	stderr.Reset()
	err = launcher.Start(context.Background(), args)
	require.ErrorContains(t, err, "maximum number of restarts reached")
	//require.False(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, "Chain 3 from zipped directory\nArgs: end --halt "+upgradeFilename+"\n", stdout.String())

	// and this doesn't upgrade
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("chain3"))
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
	cfg := prepareConfig(
		t,
		fmt.Sprintf("%s/%s", workDir, "testdata/download"),
		cosmovisor.Config{
			Name:                  "autod",
			AllowDownloadBinaries: true,
			PollInterval:          100,
			UnsafeSkipBackup:      true,
			CustomPreUpgrade:      "missing.sh",
		},
	)

	logger := log.NewTestLogger(t).With(log.ModuleKey, "cosmovisor")
	upgradeFilename := cfg.UpgradeInfoFilePath()

	// should run the genesis binary and produce expected output
	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)

	rPath, err := filepath.EvalSymlinks(cfg.GenesisBin())
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
	stdin, _ := os.Open(os.DevNull)
	stdout, stderr := newBuffer(), newBuffer()
	runner := internal.NewRunner(cfg, internal.RunConfig{
		StdIn:  stdin,
		StdOut: stdout,
		StdErr: stderr,
	}, logger)
	require.NoError(t, err)

	// Missing Preupgrade Script
	args := []string{"some", "args", upgradeFilename}
	err = runner.Start(context.Background(), args)

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
	cfg := prepareConfig(
		t,
		fmt.Sprintf("%s/%s", workDir, "testdata/download"),
		cosmovisor.Config{
			Name:                  "autod",
			AllowDownloadBinaries: true,
			PollInterval:          100,
			UnsafeSkipBackup:      true,
			CustomPreUpgrade:      "preupgrade.sh",
		},
	)

	buf := newBuffer() // inspect output using buf.String()
	logger := log.NewLogger(buf).With(log.ModuleKey, "cosmovisor")
	upgradeFilename := cfg.UpgradeInfoFilePath()

	// should run the genesis binary and produce expected output
	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err := filepath.EvalSymlinks(cfg.GenesisBin())
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
	stdin, _ := os.Open(os.DevNull)
	stdout, stderr := newBuffer(), newBuffer()
	runner := internal.NewRunner(cfg, internal.RunConfig{
		StdIn:  stdin,
		StdOut: stdout,
		StdErr: stderr,
	}, logger)

	args := []string{"some", "args", upgradeFilename}
	err = runner.Start(context.Background(), args)

	require.NoError(t, err)
	//require.True(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, "Genesis autod. Args: some args "+upgradeFilename+"\n"+`ERROR: UPGRADE "chain2" NEEDED at height: 49: zip_binary`+"\n", stdout.String())
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)

	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("chain2"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)

	// should have preupgrade.sh results
	require.FileExists(t, filepath.Join(cfg.Home, "upgrade_name_chain2_height_49"))

	// start chain2
	stdout.Reset()
	stderr.Reset()
	args = []string{"run", "--fast", upgradeFilename}
	//doUpgrade, err = runner.Run(args, stdin, stdout, stderr)
	err = runner.Start(context.Background(), args)
	require.NoError(t, err)

	require.Empty(t, stderr.String())
	require.Equal(t, "Chain 2 from zipped binary\nArgs: run --fast "+upgradeFilename+"\n"+`ERROR: UPGRADE "chain3" NEEDED at height: 936: ref_to_chain3-zip_dir.json module=main`+"\n", stdout.String())
	// ended with one more upgrade
	//require.True(t, doUpgrade)
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("chain3"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)

	// should have preupgrade.sh results
	require.FileExists(t, filepath.Join(cfg.Home, "upgrade_name_chain3_height_936"))

	// run the last chain
	args = []string{"end", "--halt", upgradeFilename}
	stdout.Reset()
	stderr.Reset()
	err = runner.Start(context.Background(), args)
	require.NoError(t, err)
	//require.False(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, "Chain 3 from zipped directory\nArgs: end --halt "+upgradeFilename+"\n", stdout.String())

	// and this doesn't upgrade
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("chain3"))
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
