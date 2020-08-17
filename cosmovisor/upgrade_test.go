// +build linux

package cosmovisor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	copy2 "github.com/otiai10/copy"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrentBin(t *testing.T) {
	home, err := copyTestData("validate")
	require.NoError(t, err)
	defer os.RemoveAll(home)

	cfg := Config{Home: home, Name: "dummyd"}

	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)

	assert.Equal(t, cfg.GenesisBin(), currentBin)

	// ensure we cannot set this to an invalid value
	for _, name := range []string{"missing", "nobin", "noexec"} {
		err = cfg.SetCurrentUpgrade(name)
		require.Error(t, err, name)

		currentBin, err := cfg.CurrentBin()
		require.NoError(t, err)

		assert.Equal(t, cfg.GenesisBin(), currentBin, name)
	}

	// try a few times to make sure this can be reproduced
	for _, upgrade := range []string{"chain2", "chain3", "chain2"} {
		// now set it to a valid upgrade and make sure CurrentBin is now set properly
		err = cfg.SetCurrentUpgrade(upgrade)
		require.NoError(t, err)
		// we should see current point to the new upgrade dir
		currentBin, err := cfg.CurrentBin()
		require.NoError(t, err)

		assert.Equal(t, cfg.UpgradeBin(upgrade), currentBin)
	}
}

func TestCurrentAlwaysSymlinkToDirectory(t *testing.T) {
	home, err := copyTestData("validate")
	require.NoError(t, err)
	defer os.RemoveAll(home)

	cfg := Config{Home: home, Name: "dummyd"}

	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)
	assert.Equal(t, cfg.GenesisBin(), currentBin)
	assertCurrentLink(t, cfg, "genesis")

	err = cfg.SetCurrentUpgrade("chain2")
	require.NoError(t, err)
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)
	assert.Equal(t, cfg.UpgradeBin("chain2"), currentBin)
	assertCurrentLink(t, cfg, filepath.Join("upgrades", "chain2"))
}

func assertCurrentLink(t *testing.T, cfg Config, target string) {
	link := filepath.Join(cfg.Root(), currentLink)
	// ensure this is a symlink
	info, err := os.Lstat(link)
	require.NoError(t, err)
	require.Equal(t, os.ModeSymlink, info.Mode()&os.ModeSymlink)

	dest, err := os.Readlink(link)
	require.NoError(t, err)
	expected := filepath.Join(cfg.Root(), target)
	require.Equal(t, expected, dest)
}

// TODO: test with download (and test all download functions)
func TestDoUpgradeNoDownloadUrl(t *testing.T) {
	home, err := copyTestData("validate")
	require.NoError(t, err)
	defer os.RemoveAll(home)

	cfg := &Config{Home: home, Name: "dummyd", AllowDownloadBinaries: true}

	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)

	assert.Equal(t, cfg.GenesisBin(), currentBin)

	// do upgrade ignores bad files
	for _, name := range []string{"missing", "nobin", "noexec"} {
		info := &UpgradeInfo{Name: name}
		err = DoUpgrade(cfg, info)
		require.Error(t, err, name)
		currentBin, err := cfg.CurrentBin()
		require.NoError(t, err)
		assert.Equal(t, cfg.GenesisBin(), currentBin, name)
	}

	// make sure it updates a few times
	for _, upgrade := range []string{"chain2", "chain3"} {
		// now set it to a valid upgrade and make sure CurrentBin is now set properly
		info := &UpgradeInfo{Name: upgrade}
		err = DoUpgrade(cfg, info)
		require.NoError(t, err)
		// we should see current point to the new upgrade dir
		upgradeBin := cfg.UpgradeBin(upgrade)
		currentBin, err := cfg.CurrentBin()
		require.NoError(t, err)

		assert.Equal(t, upgradeBin, currentBin)
	}
}

func TestOsArch(t *testing.T) {
	// all download tests will fail if we are not on linux...
	assert.Equal(t, "linux/amd64", osArch())
}

func TestGetDownloadURL(t *testing.T) {
	// all download tests will fail if we are not on linux...
	ref, err := filepath.Abs(filepath.FromSlash("./testdata/repo/ref_zipped"))
	require.NoError(t, err)
	badref, err := filepath.Abs(filepath.FromSlash("./testdata/repo/zip_binary/autod.zip"))
	require.NoError(t, err)

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
		"missing binary": {
			info:  `{"binaries": {"linux/arm": "https://foo.bar/"}}`,
			isErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			url, err := GetDownloadURL(&UpgradeInfo{Info: tc.info})
			if tc.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.url, url)
			}
		})
	}
}

func TestDownloadBinary(t *testing.T) {
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

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// make temp dir
			home, err := copyTestData("download")
			require.NoError(t, err)
			defer os.RemoveAll(home)

			cfg := &Config{
				Home:                  home,
				Name:                  "autod",
				AllowDownloadBinaries: true,
			}

			// if we have a relative path, make it absolute, but don't change eg. https://... urls
			url := tc.url
			if strings.HasPrefix(url, "./") {
				url, err = filepath.Abs(url)
				require.NoError(t, err)
			}

			upgrade := "amazonas"
			info := &UpgradeInfo{
				Name: upgrade,
				Info: fmt.Sprintf(`{"binaries":{"%s": "%s"}}`, osArch(), url),
			}

			err = DownloadBinary(cfg, info)
			if !tc.canDownload {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			err = EnsureBinary(cfg.UpgradeBin(upgrade))
			if tc.validBinary {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// copyTestData will make a tempdir and then
// "cp -r" a subdirectory under testdata there
// returns the directory (which can now be used as Config.Home) and modified safely
func copyTestData(subdir string) (string, error) {
	tmpdir, err := ioutil.TempDir("", "upgrade-manager-test")
	if err != nil {
		return "", errors.Wrap(err, "create temp dir")
	}

	src := filepath.Join("testdata", subdir)

	err = copy2.Copy(src, tmpdir)
	if err != nil {
		os.RemoveAll(tmpdir)
		return "", errors.Wrap(err, "copying files")
	}
	return tmpdir, nil
}
