package cli_test

import (
	"context"
	"io"
	"strings"
	"testing"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	rpcclientmock "github.com/tendermint/tendermint/rpc/client/mock"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/evidence/client/cli"
)

var _ client.TendermintRPC = (*mockTendermintRPC)(nil)

type mockTendermintRPC struct {
	rpcclientmock.Client

	responseQuery abci.ResponseQuery
}

func (_ mockTendermintRPC) BroadcastTxCommit(_ context.Context, _ tmtypes.Tx) (*coretypes.ResultBroadcastTxCommit, error) {
	return &coretypes.ResultBroadcastTxCommit{}, nil
}

func TestGetQueryCmd(t *testing.T) {
	encCfg := testutilmod.MakeTestEncodingConfig(evidence.AppModuleBasic{})
	kr := keyring.NewInMemory(encCfg.Codec)
	baseCtx := client.Context{}.
		WithKeyring(kr).
		WithTxConfig(encCfg.TxConfig).
		WithCodec(encCfg.Codec).
		WithClient(mockTendermintRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain")

	testCases := map[string]struct {
		args           []string
		expectedOutput string
		expectErr      bool
	}{
		"non-existent evidence": {
			[]string{"DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660"},
			"evidence DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660 not found",
			true,
		},
		"all evidence (default pagination)": {
			[]string{},
			"evidence: []\npagination:\n  next_key: null\n  total: \"0\"",
			false,
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd := cli.GetQueryCmd()
			cmd.SetOut(io.Discard)
			assert.NotNil(t, cmd)

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			assert.NoError(t, client.SetCmdClientContextHandler(baseCtx, cmd))

			// err := cmd.Execute()
			out, err := clitestutil.ExecTestCLICmd(baseCtx, cmd, tc.args)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Contains(t, strings.TrimSpace(out.String()), tc.expectedOutput)
		})
	}
}
