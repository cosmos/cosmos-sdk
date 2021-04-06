package config_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/staking/client/cli"
)

const (
	nodeEnv   = "NODE"
	testNode1 = "http://localhost:1"
	testNode2 = "http://localhost:2"
)

func initContext(t *testing.T) (context.Context, func()) {
	home := t.TempDir()

	clientCtx := client.Context{}.
		WithHomeDir(home).
		WithViper()

	clientCtx.Viper.BindEnv(nodeEnv)

	clientCtx, err := config.ReadFromClientConfig(clientCtx)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)
	return ctx, func() { _ = os.RemoveAll(home) }
}

func TestConfigCmd(t *testing.T) {
	os.Setenv(nodeEnv, testNode1)
	ctx, cleanup := initContext(t)
	defer func() {
		os.Unsetenv(nodeEnv)
		cleanup()
	}()

	cmd := config.Cmd()
	// NODE=http://localhost:1 ./build/simd config node http://localhost:2
	cmd.SetArgs([]string{"node", testNode2})
	require.NoError(t, cmd.ExecuteContext(ctx))

	//./build/simd config node //http://localhost:1
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"node"})
	cmd.Execute()
	out, err := ioutil.ReadAll(b)
	require.NoError(t, err)
	require.Equal(t, string(out), testNode1+"\n")
}

func TestConfigCmdEnvFlag(t *testing.T) {

	const (
		defaultNode = "http://localhost:26657"
	)

	tt := []struct {
		name    string
		envVar  string
		args    []string
		expNode string
	}{
		{"env var is set with no flag", testNode1, []string{"validators"}, testNode1},
		{"env var is set with a flag", testNode1, []string{"validators", fmt.Sprintf("--%s=%s", flags.FlagNode, testNode2)}, testNode2},
		{"env var is not set with no flag", "", []string{"validators"}, defaultNode},
		{"env var is not set with a flag", "", []string{"validators", fmt.Sprintf("--%s=%s", flags.FlagNode, testNode2)}, testNode2},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.envVar != "" {
				os.Setenv(nodeEnv, tc.envVar)
				defer func() {
					os.Unsetenv(nodeEnv)
				}()
			}

			ctx, cleanup := initContext(t)
			defer cleanup()

			cmd := cli.GetQueryCmd()
			cmd.SetArgs(tc.args)
			err := cmd.ExecuteContext(ctx)

			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expNode)
		})
	}

}
