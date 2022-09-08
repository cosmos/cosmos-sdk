package config_test

import (
	"fmt"
	"io"
	"os"
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

// initClientContext initiates client.Context for tests
func initClientContext(t *testing.T, envVar string) (client.Context, func()) {
	t.Helper()

	clientCtx, cleanup, err := initClientContextWithTemplate(t, envVar, "", nil)
	require.NoError(t, err)
	require.Equal(t, chainID, clientCtx.ChainID)

	return clientCtx, cleanup
}

// initClientContextWithTemplate initiates client.Context with custom config and template for tests
func initClientContextWithTemplate(t *testing.T, envVar, customTemplate string, customConfig interface{}) (client.Context, func(), error) {
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

	clientCtx, err := config.CreateClientConfig(clientCtx, customTemplate, customConfig)
	return clientCtx, func() {
		_ = os.RemoveAll(home)
		_ = os.Unsetenv(nodeEnv)
	}, err
}

func TestCustomTemplateAndConfig(t *testing.T) {
	type GasConfig struct {
		GasAdjustment float64 `mapstructure:"gas-adjustment"`
	}

	//./build/simd config node //http://localhost:1
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"node"})
	cmd.Execute()
	out, err := io.ReadAll(b)
	require.NoError(t, err)
	require.Equal(t, string(out), testNode1+"\n")
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

func TestGRPCConfig(t *testing.T) {
	expectedGRPCConfig := config.GRPCConfig{
		Address:  "localhost:7070",
		Insecure: true,
	}

	clientCfg := config.DefaultConfig()
	clientCfg.GRPC = expectedGRPCConfig

	t.Run("custom template with gRPC config", func(t *testing.T) {
		clientCtx, cleanup, err := initClientContextWithTemplate(t, "", config.DefaultClientConfigTemplate, clientCfg)
		defer cleanup()

		require.NoError(t, err)

		require.Equal(t, expectedGRPCConfig.Address, clientCtx.Viper.GetString("grpc-address"))
		require.Equal(t, expectedGRPCConfig.Insecure, clientCtx.Viper.GetBool("grpc-insecure"))
	})
}
