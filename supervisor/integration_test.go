package supervisor

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/otiai10/copy"

	"github.com/stretchr/testify/require"
)

func TestUpgradeBin(t *testing.T) {
	// Create test folder.
	daemonFolder, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.Remove(daemonFolder)

	cosmosdFolder, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.Remove(cosmosdFolder)

	err = copy.Copy(filepath.Join("testdata", "simd"), cosmosdFolder)
	require.NoError(t, err)

	t.Log("the folder is in ", cosmosdFolder)
}
