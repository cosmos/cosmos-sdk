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
	nodeEnv = "NODE"
)

func initContext(t *testing.T, testNode string) context.Context {
	home := t.TempDir()

	clientCtx := client.Context{}.
		WithHomeDir(home).
		WithViper()

	clientCtx.Viper.BindEnv(nodeEnv)
	os.Setenv(nodeEnv, testNode)

	clientCtx, err := config.ReadFromClientConfig(clientCtx)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)
	return ctx
}

/*
First
env var > config
NODE=tcp://localhost:127 ./build/simd config node tcp://localhost:128
./build/simd config node //tcp://localhost:127
*/
func TestConfigCmdFirst(t *testing.T) {

	const (
		testNode1 = "tcp://localhost:127"
		testNode2 = "tcp://localhost:128"
	)

	ctx := initContext(t, testNode1)

	cmd := config.Cmd()

	// NODE=tcp://localhost:127 ./build/simd config node tcp://localhost:128
	cmd.SetArgs([]string{"node", testNode2})
	require.NoError(t, cmd.ExecuteContext(ctx))

	//./build/simd config node //tcp://localhost:127
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"node"})
	cmd.Execute()
	out, err := ioutil.ReadAll(b)
	require.NoError(t, err)
	require.Equal(t, string(out), testNode1+"\n")

}

/*
Second
env var > config WORKS
./build/simd config node // tcp://localhost:127 //done already
NODE=tcp://localhost:1 ./build/simd q staking validators
Error: post failed: Post "http://localhost:1": dial tcp 127.0.0.1:1: connect: connection refused

Third
flags > env var > config WORKS
./build/simd config node // tcp://localhost:127
NODE=tcp://localhost:1 ./build/simd q staking validators --node tcp://localhost:2
Error: post failed: Post "http://localhost:2": dial tcp 127.0.0.1:2: connect: connection refused
*/
func TestConfigCmdSecondThird(t *testing.T) {

	const (
		testNode1 = "http://localhost:1"
		testNode2 = "http://localhost:2"
	)

	ctx := initContext(t, testNode1)

	/*
		"no flag" Error: post failed: Post "http://localhost:1": dial tcp 127.0.0.1:1: connect: connection refused
		"flag" Error: post failed: Post "http://localhost:2": dial tcp 127.0.0.1:2: connect: connection refused
	*/

	tt := []struct {
		name    string
		args    []string
		expNode string
	}{
		{"no flag", []string{"validators"}, testNode1},
		{"flag", []string{"validators", fmt.Sprintf("--%s=%s", flags.FlagNode, testNode2)}, testNode2},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.GetQueryCmd()
			cmd.SetArgs(tc.args)
			err := cmd.ExecuteContext(ctx)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expNode)
		})
	}

}
