package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
)

const (
	chainID   = "test-chain"
	nodeEnv   = "CONFIG_TEST_NODE"
	testNode1 = "http://localhost:1"
	testNode2 = "http://localhost:2"
)

// initClientContext initiates client Context for tests
func initClientContext(t *testing.T, envVar string) (client.Context, func()) {
	t.Helper()
	home := t.TempDir()
	clientCtx := client.Context{}.
		WithHomeDir(home).
		WithViper("").
		WithCodec(codec.NewProtoCodec(codectypes.NewInterfaceRegistry())).
		WithChainID(chainID)

	if envVar != "" {
		require.NoError(t, os.Setenv(nodeEnv, envVar))
	}

	clientCtx, err := config.ReadFromClientConfig(clientCtx)
	require.NoError(t, err)
	require.Equal(t, clientCtx.ChainID, chainID)

	return clientCtx, func() {
		_ = os.RemoveAll(home)
		_ = os.Unsetenv(nodeEnv)
	}
}

func TestConfigCmdEnvFlag(t *testing.T) {
	tt := []struct {
		name    string
		envVar  string
		args    []string
		expNode string
	}{
		{"env var is set with no flag", testNode1, []string{}, testNode1},
		{"env var is set with a flag", testNode1, []string{fmt.Sprintf("--%s=%s", flags.FlagNode, testNode2)}, testNode2},
		{"env var is not set with no flag", "", []string{}, "tcp://localhost:26657"},
		{"env var is not set with a flag", "", []string{fmt.Sprintf("--%s=%s", flags.FlagNode, testNode2)}, testNode2},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			testCmd := &cobra.Command{
				Use: "test",
				RunE: func(cmd *cobra.Command, args []string) error {
					clientCtx, err := client.GetClientQueryContext(cmd)
					if err != nil {
						return err
					}

					return fmt.Errorf("%s", clientCtx.NodeURI)
				},
			}
			flags.AddQueryFlagsToCmd(testCmd)

			clientCtx, cleanup := initClientContext(t, tc.envVar)
			defer func() {
				cleanup()
			}()
			/*
				env var is set with a flag

				NODE=http://localhost:1 test-cmd --node http://localhost:2
				Prints "http://localhost:2"

				It prints http://localhost:2 cause a flag has the higher priority than env variable.
			*/

			_, err := clitestutil.ExecTestCLICmd(clientCtx, testCmd, tc.args)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expNode)
		})
	}
}

func TestReadFromClientConfig(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, homeDir string) error
		chainID     string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, ctx client.Context)
	}{
		{
			name:    "creates config file when it does not exist",
			setup:   func(t *testing.T, homeDir string) error { return nil },
			chainID: "test-chain",
			wantErr: false,
			validate: func(t *testing.T, ctx client.Context) {
				require.Equal(t, "test-chain", ctx.ChainID)
				require.NotNil(t, ctx.Keyring)
				require.NotNil(t, ctx.Client)
			},
		},
		{
			name: "reads existing config file",
			setup: func(t *testing.T, homeDir string) error {
				configPath := filepath.Join(homeDir, "config")
				if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
					return err
				}
				configFile := filepath.Join(configPath, "client.toml")
				content := `chain-id = "existing-chain"
keyring-backend = "test"
output = "json"
node = "tcp://localhost:26658"
broadcast-mode = "async"
`
				return os.WriteFile(configFile, []byte(content), 0o600)
			},
			chainID: "",
			wantErr: false,
			validate: func(t *testing.T, ctx client.Context) {
				require.Equal(t, "existing-chain", ctx.ChainID)
				require.Equal(t, "json", ctx.OutputFormat)
			},
		},
		{
			name: "handles invalid TOML",
			setup: func(t *testing.T, homeDir string) error {
				configPath := filepath.Join(homeDir, "config")
				if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
					return err
				}
				configFile := filepath.Join(configPath, "client.toml")
				content := `invalid toml content [unclosed bracket`
				return os.WriteFile(configFile, []byte(content), 0o600)
			},
			chainID:     "",
			wantErr:     true,
			errContains: "couldn't get client config",
		},
		{
			name:    "preserves chain ID from context when creating new config",
			setup:   func(t *testing.T, homeDir string) error { return nil },
			chainID: "preserved-chain",
			wantErr: false,
			validate: func(t *testing.T, ctx client.Context) {
				require.Equal(t, "preserved-chain", ctx.ChainID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			if tt.setup != nil {
				require.NoError(t, tt.setup(t, home))
			}

			clientCtx := client.Context{}.
				WithHomeDir(home).
				WithViper("").
				WithCodec(codec.NewProtoCodec(codectypes.NewInterfaceRegistry()))

			if tt.chainID != "" {
				clientCtx = clientCtx.WithChainID(tt.chainID)
			}

			resultCtx, err := config.ReadFromClientConfig(clientCtx)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, resultCtx)
				}
			}
		})
	}
}

func TestReadDefaultValuesFromDefaultClientConfig(t *testing.T) {
	tests := []struct {
		name        string
		chainID     string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, ctx client.Context, originalHomeDir string)
	}{
		{
			name:    "successfully reads default values",
			chainID: "test-chain",
			wantErr: false,
			validate: func(t *testing.T, ctx client.Context, originalHomeDir string) {
				require.Equal(t, originalHomeDir, ctx.HomeDir, "HomeDir should be restored")
				require.NotNil(t, ctx.Keyring)
				require.NotNil(t, ctx.Client)
			},
		},
		{
			name:    "restores HomeDir after error",
			chainID: "",
			wantErr: false,
			validate: func(t *testing.T, ctx client.Context, originalHomeDir string) {
				require.Equal(t, originalHomeDir, ctx.HomeDir, "HomeDir should be restored even on success")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			originalHomeDir := home

			clientCtx := client.Context{}.
				WithHomeDir(home).
				WithViper("").
				WithCodec(codec.NewProtoCodec(codectypes.NewInterfaceRegistry()))

			if tt.chainID != "" {
				clientCtx = clientCtx.WithChainID(tt.chainID)
			}

			resultCtx, err := config.ReadDefaultValuesFromDefaultClientConfig(clientCtx)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, resultCtx, originalHomeDir)
				}
			}
		})
	}
}
