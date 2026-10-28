//go:build linux || darwin

package cosmovisor_test

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"
	"cosmossdk.io/tools/cosmovisor"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

var workDir string

func init() {
	workDir, _ = os.Getwd()
}

// TestLaunchProcess will try running the script a few times and watch upgrades work properly
// and args are passed through
func TestLaunchProcess(t *testing.T) {
	// binaries from testdata/validate directory
	cfg := prepareConfig(
		t,
		fmt.Sprintf("%s/%s", workDir, "testdata/validate"),
		cosmovisor.Config{
			Name:             "dummyd",
			PollInterval:     15,
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

	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("chain2"))
	require.NoError(t, err)

	require.Equal(t, rPath, currentBin)
	args = []string{"second", "run", "--verbose"}
	stdout.Reset()
	stderr.Reset()

	doUpgrade, err = launcher.Run(args, stdin, stdout, stderr)
	require.NoError(t, err)
	require.False(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, "Chain 2 is live!\nArgs: second run --verbose\nFinished successfully\n", stdout.String())

	// ended without other upgrade
	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("chain2"))
	require.NoError(t, err)

	require.Equal(t, rPath, currentBin)
}

// TestLaunchProcessSignalAfterUpgrade will check that a SIGTERM received after an upgrade
// is forwarded to the upgraded binary, and that the pre-upgrade run does not exit cosmovisor
func TestLaunchProcessSignalAfterUpgrade(t *testing.T) {
	// binaries from testdata/signal directory
	cfg := prepareConfig(
		t,
		fmt.Sprintf("%s/%s", workDir, "testdata/signal"),
		cosmovisor.Config{
			Name:             "dummyd",
			PollInterval:     15,
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

	launcher, err := cosmovisor.NewLauncher(logger, cfg)
	require.NoError(t, err)

	upgradeFile := cfg.UpgradeInfoFilePath()

	args := []string{"foo", "bar", "1234", upgradeFile}
	doUpgrade, err := launcher.Run(args, stdin, stdout, stderr)
	require.NoError(t, err)
	require.True(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, fmt.Sprintf("Genesis foo bar 1234 %s\nUPGRADE \"chain2\" NEEDED at height: 49: {}\n", upgradeFile), stdout.String())

	// ensure this is upgraded now and runs until it is signaled
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)
	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("chain2"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
	args = []string{"second", "run", "--verbose"}
	stdout.Reset()
	stderr.Reset()

	runErr := make(chan error, 1)
	go func() {
		doUpgrade, err = launcher.Run(args, stdin, stdout, stderr)
		runErr <- err
	}()
	require.Eventually(t, func() bool {
		return strings.Contains(stdout.String(), "Args: second run --verbose\n")
	}, 5*time.Second, 10*time.Millisecond)

	// the signal goes to the whole test process, so no test in this package may run in parallel
	require.NoError(t, syscall.Kill(os.Getpid(), syscall.SIGTERM))
	select {
	case err = <-runErr:
	case <-time.After(10 * time.Second):
		require.FailNow(t, "SIGTERM was not forwarded to the upgraded binary")
	}
	require.NoError(t, err)
	require.False(t, doUpgrade)
	require.Empty(t, stderr.String())
	require.Equal(t, "Chain 2 is live!\nArgs: second run --verbose\nReceived SIGTERM\n", stdout.String())

	// a signal handler left over from the first run would call os.Exit(1)
	// once its batch upgrade watcher stops, which takes up to a second
	time.Sleep(2 * time.Second)
}

func TestHelperRunLauncher(t *testing.T) {
	if os.Getenv("COSMOVISOR_HELPER") != "1" {
		return
	}
	cfg := &cosmovisor.Config{
		Home:             os.Getenv("DAEMON_HOME"),
		Name:             "dummyd",
		PollInterval:     10 * time.Millisecond,
		UnsafeSkipBackup: true,
		CustomPreUpgrade: os.Getenv("COSMOVISOR_CUSTOM_PREUPGRADE"),
		GRPCAddress:      "localhost:9090",
	}
	require.NoError(t, os.Chdir(cfg.Root()))
	launcher, err := cosmovisor.NewLauncher(cfg.Logger(os.Stdout), cfg)
	require.NoError(t, err)
	for {
		upgrade, err := launcher.Run([]string{"start", cfg.UpgradeInfoFilePath()}, nil, os.Stdout, os.Stderr)
		require.NoError(t, err)
		if !upgrade {
			return
		}
		fmt.Println("upgrade detected, relaunching")
	}
}

func TestLaunchProcessStop(t *testing.T) {
	for name, hook := range map[string]string{"group": "", "pre-upgrade": "preupgrade.sh"} {
		t.Run(name, func(t *testing.T) {
			home := copyTestData(t, filepath.Join(workDir, "testdata", "stop"))
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			executable, err := os.Executable()
			require.NoError(t, err)
			cmd := exec.CommandContext(ctx, executable, "-test.run=^TestHelperRunLauncher$")
			cmd.Env = append(os.Environ(), "COSMOVISOR_HELPER=1", "DAEMON_HOME="+home, "COSMOVISOR_CUSTOM_PREUPGRADE="+hook)
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			cmd.WaitDelay = time.Second
			output := newBuffer()
			cmd.Stdout, cmd.Stderr = output, output
			require.NoError(t, cmd.Start())
			defer func() {
				_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
				if cmd.ProcessState == nil {
					_ = cmd.Wait()
				}
			}()
			require.Eventually(t, func() bool {
				if hook != "" {
					_, err := os.Stat(filepath.Join(home, "hook-started"))
					return err == nil
				}
				return strings.Contains(output.String(), "ready\n") &&
					strings.Contains(output.String(), "starting the batch watcher loop")
			}, 5*time.Second, 10*time.Millisecond)
			pid := cmd.Process.Pid
			if hook == "" {
				pid = -pid
			}
			require.NoError(t, syscall.Kill(pid, syscall.SIGTERM))
			require.NoError(t, cmd.Wait(), "helper output:\n%s", output.String())
			require.NotContains(t, output.String(), "process already finished")
			require.NotContains(t, output.String(), "upgrade detected, relaunching")
			require.NotContains(t, output.String(), "upgraded app started")
		})
	}
}

// TestPlanDisableRecase will test upgrades without lower case plan names
func TestPlanDisableRecase(t *testing.T) {
	// binaries from testdata/validate directory
	cfg := prepareConfig(
		t,
		fmt.Sprintf("%s/%s", workDir, "testdata/norecase"),
		cosmovisor.Config{
			Name:             "dummyd",
			PollInterval:     20,
			UnsafeSkipBackup: true,
			DisableRecase:    true,
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

	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("Chain2"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
	args = []string{"second", "run", "--verbose"}
	stdout.Reset()
	stderr.Reset()

	doUpgrade, err = launcher.Run(args, stdin, stdout, stderr)
	require.NoError(t, err)
	require.False(t, doUpgrade)
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

// TestPlanShutdownGrace will test the shutdown grace period functionality
func TestPlanShutdownGrace(t *testing.T) {
	// binaries from testdata/validate directory
	cfg := prepareConfig(
		t,
		fmt.Sprintf("%s/%s", workDir, "testdata/dontdie"),
		cosmovisor.Config{
			Name:             "dummyd",
			PollInterval:     15,
			UnsafeSkipBackup: true,
			ShutdownGrace:    2 * time.Second,
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

	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("chain2"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
	args = []string{"second", "run", "--verbose"}
	stdout.Reset()
	stderr.Reset()

	doUpgrade, err = launcher.Run(args, stdin, stdout, stderr)
	require.NoError(t, err)
	require.False(t, doUpgrade)
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

	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("chain2"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)

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
	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("chain3"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)

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

	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("chain2"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)

	// should have preupgrade.sh results
	require.FileExists(t, filepath.Join(cfg.Home, "upgrade_name_chain2_height_49"))

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
	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("chain3"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)

	// should have preupgrade.sh results
	require.FileExists(t, filepath.Join(cfg.Home, "upgrade_name_chain3_height_936"))

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
	rPath, err = filepath.EvalSymlinks(cfg.UpgradeBin("chain3"))
	require.NoError(t, err)
	require.Equal(t, rPath, currentBin)
}

// TestSkipUpgrade tests heights that are identified to be skipped and returns whether the upgrade height matches the skip heights
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
