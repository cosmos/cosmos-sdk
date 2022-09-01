package cli_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	rpcclientmock "github.com/tendermint/tendermint/rpc/client/mock"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/evidence/client/cli"
)

var _ client.TendermintRPC = (*mockTendermintRPC)(nil)

type mockTendermintRPC struct {
	rpcclientmock.Client

	responseQuery abci.ResponseQuery
}

func newMockTendermintRPC(respQuery abci.ResponseQuery) mockTendermintRPC {
	return mockTendermintRPC{responseQuery: respQuery}
}

func (_ mockTendermintRPC) BroadcastTxCommit(_ context.Context, _ tmtypes.Tx) (*coretypes.ResultBroadcastTxCommit, error) {
	return &coretypes.ResultBroadcastTxCommit{}, nil
}

func (m mockTendermintRPC) ABCIQueryWithOptions(
	_ context.Context,
	_ string, _ tmbytes.HexBytes,
	_ rpcclient.ABCIQueryOptions,
) (*coretypes.ResultABCIQuery, error) {
	return &coretypes.ResultABCIQuery{Response: m.responseQuery}, nil
}

func TestGetQueryCmd(t *testing.T) {
	cmd := cli.GetQueryCmd()
	cmd.SetOut(io.Discard)
	assert.NotNil(t, cmd)

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
		ctxGen         func() client.Context
		expectedOutput string
		expectErr      bool
	}{
		"all evidence (default pagination)": {
			[]string{},
			func() client.Context {
				bz, _ := encCfg.Codec.Marshal(&sdk.TxResponse{})
				c := newMockTendermintRPC(abci.ResponseQuery{
					Value: bz,
				})
				return baseCtx.WithClient(c)
			},
			"evidence: []\npagination:\n  next_key: null\n  total: \"0\"",
			false,
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			var outBuf bytes.Buffer

			clientCtx := tc.ctxGen().WithOutput(&outBuf)
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			assert.NoError(t, client.SetCmdClientContextHandler(clientCtx, cmd))

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
