package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/tools/confix/cmd"

	"github.com/cosmos/cosmos-sdk/client"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
)

func TestValidateCmdMissingFile(t *testing.T) {
	clientCtx, cleanup := initClientContext(t)
	defer cleanup()

	_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd.ValidateCommand(), []string{"app"})
	assert.ErrorContains(t, err, "app.toml not found")
}

func TestValidateCmdDoesNotCreateFile(t *testing.T) {
	clientCtx, cleanup := initClientContext(t)
	defer cleanup()

	appPath := filepath.Join(clientCtx.HomeDir, "config", "app.toml")
	_, err := os.Stat(appPath)
	assert.Assert(t, os.IsNotExist(err))

	_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd.ValidateCommand(), []string{"app"})
	assert.ErrorContains(t, err, "app.toml not found")

	_, err = os.Stat(appPath)
	assert.Assert(t, os.IsNotExist(err), "validate must not create app.toml")
}

func TestValidateCmdRejectsEmptyMinGasPrices(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "config")
	assert.NilError(t, os.MkdirAll(cfgDir, 0o755))
	appPath := filepath.Join(cfgDir, "app.toml")
	// minimal app.toml without minimum-gas-prices
	assert.NilError(t, os.WriteFile(appPath, []byte("minimum-gas-prices = \"\"\n"), 0o644))

	_, err := clitestutil.ExecTestCLICmd(client.Context{}, cmd.ValidateCommand(), []string{"--config", appPath})
	assert.ErrorContains(t, err, "min gas price")
}

func TestValidateCmdOK(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "config")
	assert.NilError(t, os.MkdirAll(cfgDir, 0o755))
	appPath := filepath.Join(cfgDir, "app.toml")
	assert.NilError(t, os.WriteFile(appPath, []byte("minimum-gas-prices = \"0.001stake\"\n"), 0o644))

	out, err := clitestutil.ExecTestCLICmd(client.Context{}, cmd.ValidateCommand(), []string{"--config", appPath})
	assert.NilError(t, err)
	assert.Assert(t, len(out.String()) > 0)
}
