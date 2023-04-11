package config_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

const (
	nodeEnv   = "NODE"
	testNode1 = "http://localhost:1"
	testNode2 = "http://localhost:2"
)

// initClientContext initiates client Context for tests
func initClientContext(t *testing.T, envVar string) (client.Context, func()) {
	home := t.TempDir()
	chainID := "test-chain"
	clientCtx := client.Context{}.
		WithHomeDir(home).
		WithViper("").
		WithCodec(codec.NewProtoCodec(codectypes.NewInterfaceRegistry())).
		WithChainID(chainID)

	require.NoError(t, clientCtx.Viper.BindEnv(nodeEnv))
	if envVar != "" {
		require.NoError(t, os.Setenv(nodeEnv, envVar))
	}

	clientCtx, err := config.ReadFromClientConfig(clientCtx)
	require.NoError(t, err)
	require.Equal(t, clientCtx.ChainID, chainID)

	return clientCtx, func() { _ = os.RemoveAll(home) }
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
		tc := tc
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
				_ = os.Unsetenv(nodeEnv)
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
