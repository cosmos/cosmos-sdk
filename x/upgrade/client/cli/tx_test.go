package cli_test

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/bytes"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	rpcclientmock "github.com/tendermint/tendermint/rpc/client/mock"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradecli "github.com/cosmos/cosmos-sdk/x/upgrade/client/cli"
)

var _ client.TendermintRPC = (*mockTendermintRPC)(nil)

type mockTendermintRPC struct {
	rpcclientmock.Client

	responseQuery abci.ResponseQuery
}

func (mockTendermintRPC) BroadcastTxSync(_ context.Context, _ tmtypes.Tx) (*coretypes.ResultBroadcastTx, error) {
	return &coretypes.ResultBroadcastTx{}, nil
}

func (m mockTendermintRPC) ABCIQueryWithOptions(
	_ context.Context,
	_ string, _ bytes.HexBytes,
	_ rpcclient.ABCIQueryOptions,
) (*coretypes.ResultABCIQuery, error) {
	return &coretypes.ResultABCIQuery{Response: m.responseQuery}, nil
}

func TestModuleVersionsCLI(t *testing.T) {
	cmd := upgradecli.GetModuleVersionsCmd()
	cmd.SetOut(io.Discard)
	require.NotNil(t, cmd)

	encCfg := testutilmod.MakeTestEncodingConfig(upgrade.AppModuleBasic{})
	kr := keyring.NewInMemory(encCfg.Codec)
	baseCtx := client.Context{}.
		WithKeyring(kr).
		WithTxConfig(encCfg.TxConfig).
		WithCodec(encCfg.Codec).
		WithClient(mockTendermintRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain")

	testCases := []struct {
		msg          string
		args         []string
		expCmdOuptut string
	}{
		{
			msg:          "test full query with json output",
			args:         []string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			expCmdOuptut: `--height=1 --output=json`,
		},
		{
			msg:          "test full query with text output",
			args:         []string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			expCmdOuptut: `--height=1 --output=text`,
		},
		{
			msg:          "test single module",
			args:         []string{"bank", fmt.Sprintf("--%s=1", flags.FlagHeight)},
			expCmdOuptut: `bank --height=1`,
		},
		{
			msg:          "test non-existent module",
			args:         []string{"abcdefg", fmt.Sprintf("--%s=1", flags.FlagHeight)},
			expCmdOuptut: `abcdefg --height=1`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.msg, func(t *testing.T) {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetOut(io.Discard)
			require.NotNil(t, cmd)

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			require.NoError(t, client.SetCmdClientContextHandler(baseCtx, cmd))

			if len(tc.args) != 0 {
				require.Contains(t, fmt.Sprint(cmd), tc.expCmdOuptut)
			}
		})
	}
}
