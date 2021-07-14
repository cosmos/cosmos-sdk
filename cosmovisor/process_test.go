// +build linux

package cosmovisor_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
)

type processTestSuite struct {
	suite.Suite
}

func TestProcessTestSuite(t *testing.T) {
	suite.Run(t, new(processTestSuite))
}

// TestLaunchProcess will try running the script a few times and watch upgrades work properly
// and args are passed through
func (s *processTestSuite) TestLaunchProcess() {
	// binaries from testdata/validate directory
	require := s.Require()
	home := copyTestData(s.T(), "validate")
	cfg := &cosmovisor.Config{Home: home, Name: "dummyd", PoolInterval: 20}

	// should run the genesis binary and produce expected output
	var stdout, stderr = NewBuffer(), NewBuffer()
	currentBin, err := cfg.CurrentBin()
	require.NoError(err)
	require.Equal(cfg.GenesisBin(), currentBin)

	launcher, err := cosmovisor.NewLauncher(cfg)
	require.NoError(err)

	upgradeFile := cfg.UpgradeInfoFilePath()
	args := []string{"foo", "bar", "1234", upgradeFile}
	doUpgrade, err := launcher.Run(args, stdout, stderr)
	require.NoError(err)
	require.True(doUpgrade)
	require.Equal("", stderr.String())
	require.Equal(fmt.Sprintf("Genesis foo bar 1234 %s\nUPGRADE \"chain2\" NEEDED at height: 49: {}\n", upgradeFile),
		stdout.String())

	// ensure this is upgraded now and produces new output

	currentBin, err = cfg.CurrentBin()
	require.NoError(err)

	require.Equal(cfg.UpgradeBin("chain2"), currentBin)
	args = []string{"second", "run", "--verbose"}
	stdout.Reset()
	stderr.Reset()

	doUpgrade, err = launcher.Run(args, stdout, stderr)
	require.NoError(err)
	require.False(doUpgrade)
	require.Equal("", stderr.String())
	require.Equal("Chain 2 is live!\nArgs: second run --verbose\nFinished successfully\n", stdout.String())

	// ended without other upgrade
	require.Equal(cfg.UpgradeBin("chain2"), currentBin)
}

// TestLaunchProcess will try running the script a few times and watch upgrades work properly
// and args are passed through
func (s *processTestSuite) TestLaunchProcessWithDownloads() {
	// test case upgrade path (binaries from testdata/download directory):
	// genesis -> "chain2" = zip_binary
	// zip_binary -> "chain3" = ref_zipped (json for the next download instructions) -> zip_directory
	// zip_directory no upgrade
	require := s.Require()
	home := copyTestData(s.T(), "download")
	cfg := &cosmovisor.Config{Home: home, Name: "autod", AllowDownloadBinaries: true, PoolInterval: 100}
	upgradeFilename := cfg.UpgradeInfoFilePath()

	// should run the genesis binary and produce expected output
	currentBin, err := cfg.CurrentBin()
	require.NoError(err)
	require.Equal(cfg.GenesisBin(), currentBin)

	launcher, err := cosmovisor.NewLauncher(cfg)
	require.NoError(err)

	var stdout, stderr = NewBuffer(), NewBuffer()
	args := []string{"some", "args", upgradeFilename}
	doUpgrade, err := launcher.Run(args, stdout, stderr)
	fmt.Println("stderr: ", stderr.String())
	fmt.Println("stdout: ", stdout.String())

	require.NoError(err)
	require.True(doUpgrade)
	require.Equal("", stderr.String())
	require.Equal("Genesis autod. Args: some args "+upgradeFilename+"\n"+`ERROR: UPGRADE "chain2" NEEDED at height: 49: zip_binary`+"\n", stdout.String())

	// ensure this is upgraded now and produces new output
	currentBin, err = cfg.CurrentBin()
	require.NoError(err)
	require.Equal(cfg.UpgradeBin("chain2"), currentBin)

	stdout.Reset()
	stderr.Reset()
	args = []string{"run", "--fast", upgradeFilename}
	doUpgrade, err = launcher.Run(args, stdout, stderr)
	require.NoError(err)
	fmt.Println("stderr: ", stderr.String())
	fmt.Println("stdout: ", stdout.String())

	require.Equal("", stderr.String())
	require.Equal("Chain 2 from zipped binary\nArgs: run --fast "+upgradeFilename+"\n"+`ERROR: UPGRADE "chain3" NEEDED at height: 936: ref_zipped module=main`+"\n", stdout.String())
	// ended with one more upgrade
	require.True(doUpgrade)
	currentBin, err = cfg.CurrentBin()
	require.NoError(err)
	require.Equal(cfg.UpgradeBin("chain3"), currentBin)

	// run the last upgrade
	args = []string{"end", "--halt", upgradeFilename}
	stdout.Reset()
	stderr.Reset()
	doUpgrade, err = launcher.Run(args, stdout, stderr)
	require.NoError(err)
	require.False(doUpgrade)
	require.Equal("", stderr.String())
	require.Equal("Chain 2 from zipped directory\nArgs: end --halt "+upgradeFilename+"\n", stdout.String())

	// and this doesn't upgrade
	currentBin, err = cfg.CurrentBin()
	require.NoError(err)
	require.Equal(cfg.UpgradeBin("chain3"), currentBin)
}
