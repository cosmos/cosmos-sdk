package cmd_test

import (
	"path/filepath"
	"strings"
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/tools/confix/cmd"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
)

func TestMigrateCmd(t *testing.T) {
	clientCtx, cleanup := initClientContext(t)
	defer cleanup()

	_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd.MigrateCommand(), []string{"v0.0"})
	assert.ErrorContains(t, err, "unknown version")

	// clientCtx does not create app.toml, so this should fail
	_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd.MigrateCommand(), []string{"v0.45"})
	assert.ErrorContains(t, err, "no such file or directory")

	// try to migrate from unsupported.toml it should fail without --skip-validate
	_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd.MigrateCommand(), []string{"v0.46", filepath.Join(clientCtx.HomeDir, "config", "unsupported.toml")})
	assert.ErrorContains(t, err, "failed to migrate config")

	// try to migrate from unsupported.toml - it should work and give us a big diff
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd.MigrateCommand(), []string{"v0.46", filepath.Join(clientCtx.HomeDir, "config", "unsupported.toml"), "--skip-validate", "--verbose"})
	assert.NilError(t, err)
	assert.Assert(t, strings.Contains(out.String(), "add app-db-backend key"))

	// this should work
	out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd.MigrateCommand(), []string{"v0.51", filepath.Join(clientCtx.HomeDir, "config", "client.toml"), "--client", "--verbose"})
	assert.NilError(t, err)
	assert.Assert(t, strings.Contains(out.String(), "add keyring-default-keyname key"))
}
