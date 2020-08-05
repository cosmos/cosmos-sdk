// +build linux

package cosmovisor

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLaunchProcess will try running the script a few times and watch upgrades work properly
// and args are passed through
func TestLaunchProcess(t *testing.T) {
	home, err := copyTestData("validate")
	cfg := &Config{Home: home, Name: "dummyd"}
	require.NoError(t, err)
	defer os.RemoveAll(home)

	// should run the genesis binary and produce expected output
	var stdout, stderr bytes.Buffer
	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)

	require.Equal(t, cfg.GenesisBin(), currentBin)

	args := []string{"foo", "bar", "1234"}
	doUpgrade, err := LaunchProcess(cfg, args, &stdout, &stderr)
	require.NoError(t, err)
	assert.True(t, doUpgrade)
	assert.Equal(t, "", stderr.String())
	assert.Equal(t, "Genesis foo bar 1234\nUPGRADE \"chain2\" NEEDED at height: 49: {}\n", stdout.String())

	// ensure this is upgraded now and produces new output

	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)
	require.Equal(t, cfg.UpgradeBin("chain2"), currentBin)
	args = []string{"second", "run", "--verbose"}
	stdout.Reset()
	stderr.Reset()
	doUpgrade, err = LaunchProcess(cfg, args, &stdout, &stderr)
	require.NoError(t, err)
	assert.False(t, doUpgrade)
	assert.Equal(t, "", stderr.String())
	assert.Equal(t, "Chain 2 is live!\nArgs: second run --verbose\nFinished successfully\n", stdout.String())

	// ended without other upgrade
	require.Equal(t, cfg.UpgradeBin("chain2"), currentBin)
}

// TestLaunchProcess will try running the script a few times and watch upgrades work properly
// and args are passed through
func TestLaunchProcessWithDownloads(t *testing.T) {
	// this is a fun path
	// genesis -> "chain2" = zip_binary
	// zip_binary -> "chain3" = ref_zipped -> zip_directory
	// zip_directory no upgrade
	home, err := copyTestData("download")
	cfg := &Config{Home: home, Name: "autod", AllowDownloadBinaries: true}
	require.NoError(t, err)
	defer os.RemoveAll(home)

	// should run the genesis binary and produce expected output
	var stdout, stderr bytes.Buffer
	currentBin, err := cfg.CurrentBin()
	require.NoError(t, err)

	require.Equal(t, cfg.GenesisBin(), currentBin)
	args := []string{"some", "args"}
	doUpgrade, err := LaunchProcess(cfg, args, &stdout, &stderr)
	require.NoError(t, err)
	assert.True(t, doUpgrade)
	assert.Equal(t, "", stderr.String())
	assert.Equal(t, "Preparing auto-download some args\n"+`ERROR: UPGRADE "chain2" NEEDED at height: 49: {"binaries":{"linux/amd64":"https://github.com/cosmos/cosmos-sdk/raw/51249cb93130810033408934454841c98423ed4b/cosmovisor/testdata/repo/zip_binary/autod.zip?checksum=sha256:dc48829b4126ae95bc0db316c66d4e9da5f3db95e212665b6080638cca77e998"}} module=main`+"\n", stdout.String())

	// ensure this is upgraded now and produces new output
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)
	require.Equal(t, cfg.UpgradeBin("chain2"), currentBin)
	args = []string{"run", "--fast"}
	stdout.Reset()
	stderr.Reset()
	doUpgrade, err = LaunchProcess(cfg, args, &stdout, &stderr)
	require.NoError(t, err)
	assert.True(t, doUpgrade)
	assert.Equal(t, "", stderr.String())
	assert.Equal(t, "Chain 2 from zipped binary link to referral\nArgs: run --fast\n"+`ERROR: UPGRADE "chain3" NEEDED at height: 936: https://github.com/cosmos/cosmos-sdk/raw/0eae1a50612b8bf803336d35055896fbddaa1ddd/cosmovisor/testdata/repo/ref_zipped?checksum=sha256:0a428575de718ed3cf0771c9687eefaf6f19359977eca4d94a0abd0e11ef8e64 module=main`+"\n", stdout.String())

	// ended with one more upgrade
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)
	require.Equal(t, cfg.UpgradeBin("chain3"), currentBin)
	// make sure this is the proper binary now....
	args = []string{"end", "--halt"}
	stdout.Reset()
	stderr.Reset()
	doUpgrade, err = LaunchProcess(cfg, args, &stdout, &stderr)
	require.NoError(t, err)
	assert.False(t, doUpgrade)
	assert.Equal(t, "", stderr.String())
	assert.Equal(t, "Chain 2 from zipped directory\nArgs: end --halt\n", stdout.String())

	// and this doesn't upgrade
	currentBin, err = cfg.CurrentBin()
	require.NoError(t, err)
	require.Equal(t, cfg.UpgradeBin("chain3"), currentBin)
}
