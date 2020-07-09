package testutil_test

import (
	"io/ioutil"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
)

func TestNewTestCaseDir(t *testing.T) {
	dir1, cleanup1 := testutil.NewTestCaseDir(t)
	dir2, cleanup2 := testutil.NewTestCaseDir(t)

	require.NotEqual(t, dir1, dir2)
	require.DirExists(t, dir1)
	require.DirExists(t, dir2)

	cleanup1()

	require.NoDirExists(t, dir1)
	require.DirExists(t, dir2)

	cleanup2()
	require.NoDirExists(t, dir2)
}

func TestApplyMockIO(t *testing.T) {
	cmd := &cobra.Command{}

	oldStdin := cmd.InOrStdin()
	oldStdout := cmd.OutOrStdout()
	oldStderr := cmd.ErrOrStderr()

	testutil.ApplyMockIO(cmd)

	require.NotEqual(t, cmd.InOrStdin(), oldStdin)
	require.NotEqual(t, cmd.OutOrStdout(), oldStdout)
	require.NotEqual(t, cmd.ErrOrStderr(), oldStderr)
}

func TestWriteToNewTempFile(t *testing.T) {
	tempfile, cleanup := testutil.WriteToNewTempFile(t, "test string")
	tempfile.Close()

	bs, err := ioutil.ReadFile(tempfile.Name())
	require.NoError(t, err)
	require.Equal(t, "test string", string(bs))

	cleanup()

	require.NoFileExists(t, tempfile.Name())
}
