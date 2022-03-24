package snapshots

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func GetTempDir(t *testing.T) string {
	// ioutil.TempDir() is used instead of testing.T.TempDir()
	// see https://github.com/cosmos/cosmos-sdk/pull/8475 for
	// this change's rationale.
	tempdir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(tempdir) })
	return tempdir
}
