package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
)

// newValidateCmd builds the command under a root that installs the SAME mutating
// PersistentPreRunE the shipped root command uses, plus a --home flag.
//
// The earlier version of this test wired only a server context and omitted the
// hook, which is why it passed while `simd validate-app` on a fresh home created
// <home>/config/app.toml and then reported OK for the file it had just written.
// The hook is the thing under test, so it has to be present.
func newValidateCmd(t *testing.T, homeDir string) *cobra.Command {
	t.Helper()

	root := &cobra.Command{Use: "root"}

	// Stand-in for server.InterceptConfigsPreRunHandler: the part that matters is
	// that it runs for every child and writes the default config when missing.
	var hookRan bool
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		hookRan = true
		home, err := cmd.Flags().GetString(flags.FlagHome)
		if err != nil {
			return err
		}
		cfgDir := filepath.Join(home, "config")
		if err := os.MkdirAll(cfgDir, 0o755); err != nil {
			return err
		}
		cfgPath := filepath.Join(cfgDir, "app.toml")
		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			// A generated default with non-empty gas prices, as simd writes.
			if err := os.WriteFile(cfgPath, []byte("minimum-gas-prices = \"0.01stake\"\n"), 0o644); err != nil {
				return err
			}
		}
		return nil
	}
	root.PersistentFlags().String(flags.FlagHome, homeDir, "The application home directory")
	root.AddCommand(ValidateAppConfigCmd())

	t.Cleanup(func() { _ = hookRan })
	return root
}

func TestValidateAppConfigRootHookDoesNotWrite(t *testing.T) {
	home := t.TempDir()
	cfgPath := filepath.Join(home, "config", "app.toml")

	cmd := newValidateCmd(t, home)
	cmd.SetArgs([]string{"validate-app"})

	// With the real hook installed, the command must still refuse a missing
	// candidate rather than validating a default the hook generated for it.
	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "config file not found")

	// The guarantee that matters: nothing was written anywhere under home.
	require.NoFileExists(t, cfgPath)
	require.NoDirExists(t, filepath.Dir(cfgPath))
}

func TestValidateAppConfigMissingFileIsNotCreated(t *testing.T) {
	home := t.TempDir()
	cfgPath := filepath.Join(home, "config", "app.toml")

	cmd := newValidateCmd(t, home)
	cmd.SetArgs([]string{"validate-app", "--config", cfgPath})

	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "config file not found")

	require.NoFileExists(t, cfgPath)
	require.NoDirExists(t, filepath.Dir(cfgPath))
}

func TestValidateAppConfigRejectsInvalidConfig(t *testing.T) {
	home := t.TempDir()
	cfgDir := filepath.Join(home, "config")
	require.NoError(t, os.MkdirAll(cfgDir, 0o755))
	cfgPath := filepath.Join(cfgDir, "app.toml")

	require.NoError(t, os.WriteFile(cfgPath, []byte(`
minimum-gas-prices = "0.01stake"
block-executor = "not-a-real-executor"
`), 0o644))

	cmd := newValidateCmd(t, home)
	cmd.SetArgs([]string{"validate-app", "--config", cfgPath})

	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid block executor")
}

func TestValidateAppConfigAcceptsValidConfig(t *testing.T) {
	home := t.TempDir()
	cfgDir := filepath.Join(home, "config")
	require.NoError(t, os.MkdirAll(cfgDir, 0o755))
	cfgPath := filepath.Join(cfgDir, "app.toml")

	require.NoError(t, os.WriteFile(cfgPath, []byte(`
minimum-gas-prices = "0.01stake"
`), 0o644))

	before, err := os.Stat(cfgPath)
	require.NoError(t, err)

	cmd := newValidateCmd(t, home)
	cmd.SetArgs([]string{"validate-app", "--config", cfgPath})

	require.NoError(t, cmd.Execute())

	after, err := os.Stat(cfgPath)
	require.NoError(t, err)
	require.Equal(t, before.Size(), after.Size())
	require.Equal(t, before.ModTime(), after.ModTime())
}

// A staged candidate keeps the loader's TOML rules. Extension inference would
// reject this name outright, and would let a JSON file pass here while startup
// rejects it after the rename to app.toml.
func TestValidateAppConfigStagedFilenameUsesTOMLRules(t *testing.T) {
	home := t.TempDir()
	cfgDir := filepath.Join(home, "config")
	require.NoError(t, os.MkdirAll(cfgDir, 0o755))

	staged := filepath.Join(cfgDir, "app.toml.new")
	require.NoError(t, os.WriteFile(staged, []byte(`
minimum-gas-prices = "0.01stake"
block-executor = "not-a-real-executor"
`), 0o644))

	cmd := newValidateCmd(t, home)
	cmd.SetArgs([]string{"validate-app", "--config", staged})

	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid block executor")

	// A JSON candidate must not slip through under a name that startup reads as TOML.
	jsonCandidate := filepath.Join(cfgDir, "candidate.new")
	require.NoError(t, os.WriteFile(jsonCandidate, []byte(`{"minimum-gas-prices":"0.01stake"}`), 0o644))

	cmd = newValidateCmd(t, home)
	cmd.SetArgs([]string{"validate-app", "--config", jsonCandidate})
	require.Error(t, cmd.Execute())
}
