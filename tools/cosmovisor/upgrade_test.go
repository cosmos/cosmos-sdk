//go:build linux
// +build linux

package cosmovisor_test

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/otiai10/copy"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/tools/cosmovisor"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
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
	logger := cosmovisor.NewLogger()

	currentBin, err := cfg.CurrentBin()
	s.Require().NoError(err)

	s.Require().Equal(cfg.GenesisBin(), currentBin)

	// do upgrade ignores bad files
	for _, name := range []string{"missing", "nobin", "noexec"} {
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

func (s *upgradeTestSuite) TestOsArch() {
	// all download tests will fail if we are not on linux...
	s.Require().Equal("linux/amd64", cosmovisor.OSArch())
}

func (s *upgradeTestSuite) TestGetDownloadURL() {
	// all download tests will fail if we are not on linux...
	ref, err := filepath.Abs(filepath.FromSlash("./testdata/repo/ref_to_chain3-zip_dir.json"))
	s.Require().NoError(err)
	badref, err := filepath.Abs(filepath.FromSlash("./testdata/repo/chain2-zip_bin/autod.zip")) // "./testdata/repo/zip_binary/autod.zip"))
	s.Require().NoError(err)

	cases := map[string]struct {
		info string
		url  string
		err  interface{}

		// If err == nil, the test must not report an error.
		// If err is a string, the test must report an error whose string has err
		// as a substring.
		// If err is a func(suite.Suite, error), it is called to check the error
		// value.
	}{
		"missing": {
			err: "downloading reference link : invalid source string:",
		},
		"follow reference": {
			info: ref,
			url:  "https://github.com/cosmos/cosmos-sdk/raw/main/tools/cosmovisor/testdata/repo/chain3-zip_dir/autod.zip?checksum=sha256:8951f52a0aea8617de0ae459a20daf704c29d259c425e60d520e363df0f166b4",
		},
		"malformated reference target": {
			info: badref,
			err:  "upgrade info doesn't contain binary map",
		},
		"missing link": {
			info: "https://no.such.domain/exists.txt",
			err: func(s suite.Suite, err error) {
				var dns *net.DNSError
				s.Require().True(errors.As(err, &dns), "result is not a DNSError")
				s.Require().Equal("no.such.domain", dns.Name)
				s.Require().Equal(true, dns.IsNotFound)
			},
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
			info: `{"binaries": {"linux/arm": "https://foo.bar/"}}`,
			err:  "cannot find binary for",
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			url, err := cosmovisor.GetDownloadURL(upgradetypes.Plan{Info: tc.info})
			switch e := tc.err.(type) {
			case nil:
				s.Require().NoError(err)
				s.Require().Equal(tc.url, url)

			case string:
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.err)

			case func(suite.Suite, error):
				e(s.Suite, err)
			}
		})
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
			url:         "./testdata/repo/chain3-zip_dir/autod.zip",
			canDownload: true,
			validBinary: true,
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
			url:         "./testdata/repo/bad_dir/autod",
			canDownload: false,
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

			const upgrade = "amazonas"
			info := upgradetypes.Plan{
				Name: upgrade,
				Info: fmt.Sprintf(`{"binaries":{"%s": "%s"}}`, cosmovisor.OSArch(), url),
			}

			err = cosmovisor.DownloadBinary(cfg, info)
			if !tc.canDownload {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}

			err = cosmovisor.EnsureBinary(cfg.UpgradeBin(upgrade))
			if tc.validBinary {
				s.Require().NoError(err)
			}
		})
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
