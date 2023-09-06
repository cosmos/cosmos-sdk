//go:build linux
// +build linux

package cosmovisor_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
	"cosmossdk.io/tools/cosmovisor"
	upgradetypes "cosmossdk.io/x/upgrade/types"
)

type upgradeTestSuite struct {
	suite.Suite
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(upgradeTestSuite))
}

func (s *upgradeTestSuite) TestCurrentBin() {
	home := copyTestData(s.T(), "validate")
	cfg := cosmovisor.Config{Home: home, Name: "dummyd"}

	currentBin, err := cfg.CurrentBin()
	s.Require().NoError(err)

	s.Require().Equal(cfg.GenesisBin(), currentBin)

	// ensure we cannot set this to an invalid value
	for _, name := range []string{"missing", "nobin"} {
		s.Require().Error(cfg.SetCurrentUpgrade(upgradetypes.Plan{Name: name}), name)

		currentBin, err := cfg.CurrentBin()
		s.Require().NoError(err)

		s.Require().Equal(cfg.GenesisBin(), currentBin, name)
	}

	// try a few times to make sure this can be reproduced
	for _, name := range []string{"chain2", "chain3", "chain2"} {
		// now set it to a valid upgrade and make sure CurrentBin is now set properly
		err = cfg.SetCurrentUpgrade(upgradetypes.Plan{Name: name})
		s.Require().NoError(err)
		// we should see current point to the new upgrade dir
		currentBin, err := cfg.CurrentBin()
		s.Require().NoError(err)

		s.Require().Equal(cfg.UpgradeBin(name), currentBin)
	}
}

func (s *upgradeTestSuite) TestCurrentAlwaysSymlinkToDirectory() {
	home := copyTestData(s.T(), "validate")
	cfg := cosmovisor.Config{Home: home, Name: "dummyd"}

	currentBin, err := cfg.CurrentBin()
	s.Require().NoError(err)
	s.Require().Equal(cfg.GenesisBin(), currentBin)
	s.assertCurrentLink(cfg, "genesis")

	err = cfg.SetCurrentUpgrade(upgradetypes.Plan{Name: "chain2"})
	s.Require().NoError(err)
	currentBin, err = cfg.CurrentBin()
	s.Require().NoError(err)
	s.Require().Equal(cfg.UpgradeBin("chain2"), currentBin)
	s.assertCurrentLink(cfg, filepath.Join("upgrades", "chain2"))
}

func (s *upgradeTestSuite) assertCurrentLink(cfg cosmovisor.Config, target string) {
	link := filepath.Join(cfg.Root(), "current")
	// ensure this is a symlink
	info, err := os.Lstat(link)
	s.Require().NoError(err)
	s.Require().Equal(os.ModeSymlink, info.Mode()&os.ModeSymlink)

	dest, err := os.Readlink(link)
	s.Require().NoError(err)
	expected := filepath.Join(cfg.Root(), target)
	s.Require().Equal(expected, dest)
}

// TODO: test with download (and test all download functions)
func (s *upgradeTestSuite) TestUpgradeBinaryNoDownloadUrl() {
	home := copyTestData(s.T(), "validate")
	cfg := &cosmovisor.Config{Home: home, Name: "dummyd", AllowDownloadBinaries: true}
	logger := log.NewLogger(os.Stdout).With(log.ModuleKey, "cosmovisor")

	currentBin, err := cfg.CurrentBin()
	s.Require().NoError(err)

	s.Require().Equal(cfg.GenesisBin(), currentBin)

	// do upgrade ignores bad files
	for _, name := range []string{"missing", "nobin"} {
		info := upgradetypes.Plan{Name: name}
		err = cosmovisor.UpgradeBinary(logger, cfg, info)
		s.Require().Error(err, name)
		currentBin, err := cfg.CurrentBin()
		s.Require().NoError(err)
		s.Require().Equal(cfg.GenesisBin(), currentBin, name)
	}

	// make sure it updates a few times
	for _, upgrade := range []string{"chain2", "chain3"} {
		// now set it to a valid upgrade and make sure CurrentBin is now set properly
		info := upgradetypes.Plan{Name: upgrade}
		err = cosmovisor.UpgradeBinary(logger, cfg, info)
		s.Require().NoError(err)
		// we should see current point to the new upgrade dir
		upgradeBin := cfg.UpgradeBin(upgrade)
		currentBin, err := cfg.CurrentBin()
		s.Require().NoError(err)

		s.Require().Equal(upgradeBin, currentBin)
	}
}

func (s *upgradeTestSuite) TestUpgradeBinary() {
	logger := log.NewLogger(os.Stdout).With(log.ModuleKey, "cosmovisor")

	cases := map[string]struct {
		url         string
		canDownload bool
		validBinary bool
	}{
		"get raw binary with checksum": {
			// sha256sum ./testdata/repo/raw_binary/autod
			url:         "./testdata/repo/raw_binary/autod?checksum=sha256:e6bc7851600a2a9917f7bf88eb7bdee1ec162c671101485690b4deb089077b0d",
			canDownload: true,
			validBinary: true,
		},
		"get raw binary with invalid checksum": {
			url:         "./testdata/repo/raw_binary/autod?checksum=sha256:73e2bd6cbb99261733caf137015d5cc58e3f96248d8b01da68be8564989dd906",
			canDownload: false,
		},
		"get zipped directory with valid checksum": {
			// sha256sum ./testdata/repo/chain3-zip_dir/autod.zip
			url:         "./testdata/repo/chain3-zip_dir/autod.zip?checksum=sha256:8951f52a0aea8617de0ae459a20daf704c29d259c425e60d520e363df0f166b4",
			canDownload: true,
			validBinary: true,
		},
		"get zipped directory with invalid checksum": {
			url:         "./testdata/repo/chain3-zip_dir/autod.zip?checksum=sha256:73e2bd6cbb99261733caf137015d5cc58e3f96248d8b01da68be8564989dd906",
			canDownload: false,
		},
		"invalid url": {
			url:         "./testdata/repo/bad_dir/autod?checksum=sha256:73e2bd6cbb99261733caf137015d5cc58e3f96248d8b01da68be8564989dd906",
			canDownload: false,
		},
		"valid remote": {
			url:         "https://github.com/cosmos/cosmos-sdk/raw/main/tools/cosmovisor/testdata/repo/chain3-zip_dir/autod.zip?checksum=sha256:8951f52a0aea8617de0ae459a20daf704c29d259c425e60d520e363df0f166b4",
			canDownload: true,
			validBinary: true,
		},
	}

	for label, tc := range cases {
		s.Run(label, func() {
			var err error
			// make temp dir
			home := copyTestData(s.T(), "download")

			cfg := &cosmovisor.Config{
				Home:                  home,
				Name:                  "autod",
				AllowDownloadBinaries: true,
			}

			url := tc.url
			if strings.HasPrefix(url, "./") {
				url, err = filepath.Abs(url)
				s.Require().NoError(err)
			}

			plan := upgradetypes.Plan{
				Name: "amazonas",
				Info: fmt.Sprintf(`{"binaries":{"%s": "%s"}}`, cosmovisor.OSArch(), url),
			}

			err = cosmovisor.UpgradeBinary(logger, cfg, plan)
			if !tc.canDownload {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *upgradeTestSuite) TestOsArch() {
	// all download tests will fail if we are not on linux...
	s.Require().Equal("linux/amd64", cosmovisor.OSArch())
}

// copyTestData will make a tempdir and then
// "cp -r" a subdirectory under testdata there
// returns the directory (which can now be used as Config.Home) and modified safely
func copyTestData(t *testing.T, subdir string) string {
	t.Helper()
	tmpdir := t.TempDir()
	require.NoError(t, copy.Copy(filepath.Join("testdata", subdir), tmpdir))

	return tmpdir
}
