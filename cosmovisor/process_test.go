// +build linux

package cosmovisor_test

import (
	"bytes"
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
	home := copyTestData(s.T(), "validate")
	cfg := &cosmovisor.Config{Home: home, Name: "dummyd"}

	// should run the genesis binary and produce expected output
	var stdout, stderr bytes.Buffer
	currentBin, err := cfg.CurrentBin()
	s.Require().NoError(err)

	s.Require().Equal(cfg.GenesisBin(), currentBin)

	args := []string{"foo", "bar", "1234"}
	doUpgrade, err := cosmovisor.LaunchProcess(cfg, args, &stdout, &stderr)
	s.Require().NoError(err)
	s.Require().True(doUpgrade)
	s.Require().Equal("", stderr.String())
	s.Require().Equal("Genesis foo bar 1234\nUPGRADE \"chain2\" NEEDED at height: 49: {}\n", stdout.String())

	// ensure this is upgraded now and produces new output

	currentBin, err = cfg.CurrentBin()
	s.Require().NoError(err)
	s.Require().Equal(cfg.UpgradeBin("chain2"), currentBin)
	args = []string{"second", "run", "--verbose"}
	stdout.Reset()
	stderr.Reset()
	doUpgrade, err = cosmovisor.LaunchProcess(cfg, args, &stdout, &stderr)
	s.Require().NoError(err)
	s.Require().False(doUpgrade)
	s.Require().Equal("", stderr.String())
	s.Require().Equal("Chain 2 is live!\nArgs: second run --verbose\nFinished successfully\n", stdout.String())

	// ended without other upgrade
	s.Require().Equal(cfg.UpgradeBin("chain2"), currentBin)
}

// TestLaunchProcess will try running the script a few times and watch upgrades work properly
// and args are passed through
func (s *processTestSuite) TestLaunchProcessWithDownloads() {
	// this is a fun path
	// genesis -> "chain2" = zip_binary
	// zip_binary -> "chain3" = ref_zipped -> zip_directory
	// zip_directory no upgrade
	home := copyTestData(s.T(), "download")
	cfg := &cosmovisor.Config{Home: home, Name: "autod", AllowDownloadBinaries: true}

	// should run the genesis binary and produce expected output
	var stdout, stderr bytes.Buffer
	currentBin, err := cfg.CurrentBin()
	s.Require().NoError(err)

	s.Require().Equal(cfg.GenesisBin(), currentBin)
	args := []string{"some", "args"}
	doUpgrade, err := cosmovisor.LaunchProcess(cfg, args, &stdout, &stderr)
	s.Require().NoError(err)
	s.Require().True(doUpgrade)
	s.Require().Equal("", stderr.String())
	s.Require().Equal("Preparing auto-download some args\n"+`ERROR: UPGRADE "chain2" NEEDED at height: 49: {"binaries":{"linux/amd64":"https://github.com/cosmos/cosmos-sdk/raw/51249cb93130810033408934454841c98423ed4b/cosmovisor/testdata/repo/zip_binary/autod.zip?checksum=sha256:dc48829b4126ae95bc0db316c66d4e9da5f3db95e212665b6080638cca77e998"}} module=main`+"\n", stdout.String())

	// ensure this is upgraded now and produces new output
	currentBin, err = cfg.CurrentBin()
	s.Require().NoError(err)
	s.Require().Equal(cfg.UpgradeBin("chain2"), currentBin)
	args = []string{"run", "--fast"}
	stdout.Reset()
	stderr.Reset()
	doUpgrade, err = cosmovisor.LaunchProcess(cfg, args, &stdout, &stderr)
	s.Require().NoError(err)
	s.Require().True(doUpgrade)
	s.Require().Equal("", stderr.String())
	s.Require().Equal("Chain 2 from zipped binary link to referral\nArgs: run --fast\n"+`ERROR: UPGRADE "chain3" NEEDED at height: 936: https://github.com/cosmos/cosmos-sdk/raw/0eae1a50612b8bf803336d35055896fbddaa1ddd/cosmovisor/testdata/repo/ref_zipped?checksum=sha256:0a428575de718ed3cf0771c9687eefaf6f19359977eca4d94a0abd0e11ef8e64 module=main`+"\n", stdout.String())

	// ended with one more upgrade
	currentBin, err = cfg.CurrentBin()
	s.Require().NoError(err)
	s.Require().Equal(cfg.UpgradeBin("chain3"), currentBin)
	// make sure this is the proper binary now....
	args = []string{"end", "--halt"}
	stdout.Reset()
	stderr.Reset()
	doUpgrade, err = cosmovisor.LaunchProcess(cfg, args, &stdout, &stderr)
	s.Require().NoError(err)
	s.Require().False(doUpgrade)
	s.Require().Equal("", stderr.String())
	s.Require().Equal("Chain 2 from zipped directory\nArgs: end --halt\n", stdout.String())

	// and this doesn't upgrade
	currentBin, err = cfg.CurrentBin()
	s.Require().NoError(err)
	s.Require().Equal(cfg.UpgradeBin("chain3"), currentBin)
}
