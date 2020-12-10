// +build linux

package cosmovisor_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
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
	for _, name := range []string{"missing", "nobin", "noexec"} {
		s.Require().Error(cfg.SetCurrentUpgrade(name), name)

		currentBin, err := cfg.CurrentBin()
		s.Require().NoError(err)

		s.Require().Equal(cfg.GenesisBin(), currentBin, name)
	}

	// try a few times to make sure this can be reproduced
	for _, upgrade := range []string{"chain2", "chain3", "chain2"} {
		// now set it to a valid upgrade and make sure CurrentBin is now set properly
		err = cfg.SetCurrentUpgrade(upgrade)
		s.Require().NoError(err)
		// we should see current point to the new upgrade dir
		currentBin, err := cfg.CurrentBin()
		s.Require().NoError(err)

		s.Require().Equal(cfg.UpgradeBin(upgrade), currentBin)
	}
}

func (s *upgradeTestSuite) TestCurrentAlwaysSymlinkToDirectory() {
	home := copyTestData(s.T(), "validate")
	cfg := cosmovisor.Config{Home: home, Name: "dummyd"}

	currentBin, err := cfg.CurrentBin()
	s.Require().NoError(err)
	s.Require().Equal(cfg.GenesisBin(), currentBin)
	s.assertCurrentLink(cfg, "genesis")

	err = cfg.SetCurrentUpgrade("chain2")
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
func (s *upgradeTestSuite) TestDoUpgradeNoDownloadUrl() {
	home := copyTestData(s.T(), "validate")
	cfg := &cosmovisor.Config{Home: home, Name: "dummyd", AllowDownloadBinaries: true}

	currentBin, err := cfg.CurrentBin()
	s.Require().NoError(err)

	s.Require().Equal(cfg.GenesisBin(), currentBin)

	// do upgrade ignores bad files
	for _, name := range []string{"missing", "nobin", "noexec"} {
		info := &cosmovisor.UpgradeInfo{Name: name}
		err = cosmovisor.DoUpgrade(cfg, info)
		s.Require().Error(err, name)
		currentBin, err := cfg.CurrentBin()
		s.Require().NoError(err)
		s.Require().Equal(cfg.GenesisBin(), currentBin, name)
	}

	// make sure it updates a few times
	for _, upgrade := range []string{"chain2", "chain3"} {
		// now set it to a valid upgrade and make sure CurrentBin is now set properly
		info := &cosmovisor.UpgradeInfo{Name: upgrade}
		err = cosmovisor.DoUpgrade(cfg, info)
		s.Require().NoError(err)
		// we should see current point to the new upgrade dir
		upgradeBin := cfg.UpgradeBin(upgrade)
		currentBin, err := cfg.CurrentBin()
		s.Require().NoError(err)

		s.Require().Equal(upgradeBin, currentBin)
	}
}

func (s *upgradeTestSuite) TestOsArch() {
	// all download tests will fail if we are not on linux...
	s.Require().Equal("linux/amd64", cosmovisor.OSArch())
}

func (s *upgradeTestSuite) TestGetDownloadURL() {
	// all download tests will fail if we are not on linux...
	ref, err := filepath.Abs(filepath.FromSlash("./testdata/repo/ref_zipped"))
	s.Require().NoError(err)
	badref, err := filepath.Abs(filepath.FromSlash("./testdata/repo/zip_binary/autod.zip"))
	s.Require().NoError(err)

	cases := map[string]struct {
		info  string
		url   string
		isErr bool
	}{
		"missing": {
			isErr: true,
		},
		"follow reference": {
			info: ref,
			url:  "https://github.com/cosmos/cosmos-sdk/raw/aa5d6140ad4011bb33d472dca8246a0dcbe223ee/cosmovisor/testdata/repo/zip_directory/autod.zip?checksum=sha256:3784e4574cad69b67e34d4ea4425eff140063a3870270a301d6bb24a098a27ae",
		},
		"malformated reference target": {
			info:  badref,
			isErr: true,
		},
		"missing link": {
			info:  "https://no.such.domain/exists.txt",
			isErr: true,
		},
		"proper binary": {
			info: `{"binaries": {"linux/amd64": "https://foo.bar/", "windows/amd64": "https://something.else"}}`,
			url:  "https://foo.bar/",
		},
		"any architecture not used": {
			info: `{"binaries": {"linux/amd64": "https://foo.bar/", "*": "https://something.else"}}`,
			url:  "https://foo.bar/",
		},
		"any architecture used": {
			info: `{"binaries": {"linux/arm": "https://foo.bar/arm-only", "any": "https://foo.bar/portable"}}`,
			url:  "https://foo.bar/portable",
		},
		"missing binary": {
			info:  `{"binaries": {"linux/arm": "https://foo.bar/"}}`,
			isErr: true,
		},
	}

	for _, tc := range cases {
		url, err := cosmovisor.GetDownloadURL(&cosmovisor.UpgradeInfo{Info: tc.info})
		if tc.isErr {
			s.Require().Error(err)
		} else {
			s.Require().NoError(err)
			s.Require().Equal(tc.url, url)
		}
	}
}

func (s *upgradeTestSuite) TestDownloadBinary() {
	cases := map[string]struct {
		url         string
		canDownload bool
		validBinary bool
	}{
		"get raw binary": {
			url:         "./testdata/repo/raw_binary/autod",
			canDownload: true,
			validBinary: true,
		},
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
		"get zipped directory": {
			url:         "./testdata/repo/zip_directory/autod.zip",
			canDownload: true,
			validBinary: true,
		},
		"get zipped directory with valid checksum": {
			// sha256sum ./testdata/repo/zip_directory/autod.zip
			url:         "./testdata/repo/zip_directory/autod.zip?checksum=sha256:3784e4574cad69b67e34d4ea4425eff140063a3870270a301d6bb24a098a27ae",
			canDownload: true,
			validBinary: true,
		},
		"get zipped directory with invalid checksum": {
			url:         "./testdata/repo/zip_directory/autod.zip?checksum=sha256:73e2bd6cbb99261733caf137015d5cc58e3f96248d8b01da68be8564989dd906",
			canDownload: false,
		},
		"invalid url": {
			url:         "./testdata/repo/bad_dir/autod",
			canDownload: false,
		},
	}

	for _, tc := range cases {
		var err error
		// make temp dir
		home := copyTestData(s.T(), "download")

		cfg := &cosmovisor.Config{
			Home:                  home,
			Name:                  "autod",
			AllowDownloadBinaries: true,
		}

		// if we have a relative path, make it absolute, but don't change eg. https://... urls
		url := tc.url
		if strings.HasPrefix(url, "./") {
			url, err = filepath.Abs(url)
			s.Require().NoError(err)
		}

		upgrade := "amazonas"
		info := &cosmovisor.UpgradeInfo{
			Name: upgrade,
			Info: fmt.Sprintf(`{"binaries":{"%s": "%s"}}`, cosmovisor.OSArch(), url),
		}

		err = cosmovisor.DownloadBinary(cfg, info)
		if !tc.canDownload {
			s.Require().Error(err)
			return
		}
		s.Require().NoError(err)

		err = cosmovisor.EnsureBinary(cfg.UpgradeBin(upgrade))
		if tc.validBinary {
			s.Require().NoError(err)
		} else {
			s.Require().Error(err)
		}
	}
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
