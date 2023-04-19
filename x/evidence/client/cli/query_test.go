package cli_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"cosmossdk.io/x/evidence"
	"cosmossdk.io/x/evidence/client/cli"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

func TestGetQueryCmd(t *testing.T) {
	encCfg := testutilmod.MakeTestEncodingConfig(evidence.AppModuleBasic{})
	kr := keyring.NewInMemory(encCfg.Codec)
	baseCtx := client.Context{}.
		WithKeyring(kr).
		WithTxConfig(encCfg.TxConfig).
		WithCodec(encCfg.Codec).
		WithClient(clitestutil.MockCometRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain")

	testCases := map[string]struct {
		args           []string
		expectedOutput string
		expectErrMsg   string
	}{
		"invalid args": {
			[]string{"foo", "bar"},
			"",
			"accepts at most 1 arg(s)",
		},
		"all evidence (default pagination)": {
			[]string{},
			"evidence: []\npagination: null",
			"",
		},
		"all evidence (json output)": {
			[]string{
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			`{"evidence":[],"pagination":null}`,
			"",
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd := cli.GetQueryCmd()
			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)
			require.NoError(t, client.SetCmdClientContextHandler(baseCtx, cmd))

			out, err := clitestutil.ExecTestCLICmd(baseCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectErrMsg)
			} else {
				require.NoError(t, err)
			}

			require.Contains(t, strings.TrimSpace(out.String()), tc.expectedOutput)
		})
	}
}
