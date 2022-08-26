package cli_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	rpcclientmock "github.com/tendermint/tendermint/rpc/client/mock"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintcli "github.com/cosmos/cosmos-sdk/x/mint/client/cli"
	tmcli "github.com/tendermint/tendermint/libs/cli"
)

var _ client.TendermintRPC = (*mockTendermintRPC)(nil)

type mockTendermintRPC struct {
	rpcclientmock.Client

	responseQuery abci.ResponseQuery
}

func (_ mockTendermintRPC) BroadcastTxCommit(_ context.Context, _ tmtypes.Tx) (*coretypes.ResultBroadcastTxCommit, error) {
	return &coretypes.ResultBroadcastTxCommit{}, nil
}

func TestGetCmdQueryParams(t *testing.T) {
	encCfg := testutilmod.MakeTestEncodingConfig(mint.AppModuleBasic{})
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
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			`{"mint_denom":"stake","inflation_rate_change":"0.130000000000000000","inflation_max":"1.000000000000000000","inflation_min":"1.000000000000000000","goal_bonded":"0.670000000000000000","blocks_per_year":"6311520"}`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			`blocks_per_year: "6311520"
	goal_bonded: "0.670000000000000000"
	inflation_max: "1.000000000000000000"
	inflation_min: "1.000000000000000000"
	inflation_rate_change: "0.130000000000000000"
	mint_denom: stake`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd := mintcli.GetCmdQueryParams()
			cmd.SetOut(io.Discard)
			assert.NotNil(t, cmd)

			cmd.SetContext(ctx)
			assert.NoError(t, client.SetCmdClientContextHandler(baseCtx, cmd))

			out, err := clitestutil.ExecTestCLICmd(baseCtx, cmd, tc.args)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedOutput, strings.TrimSpace(out.String()))

		})
	}
}

func TestGetCmdQueryInflation(t *testing.T) {
	encCfg := testutilmod.MakeTestEncodingConfig(mint.AppModuleBasic{})
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
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			`1.000000000000000000`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			`1.000000000000000000`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd := mintcli.GetCmdQueryInflation()
			cmd.SetOut(io.Discard)
			assert.NotNil(t, cmd)

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			assert.NoError(t, client.SetCmdClientContextHandler(baseCtx, cmd))

			out, err := clitestutil.ExecTestCLICmd(baseCtx, cmd, tc.args)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func TestGetCmdQueryAnnualProvisions(t *testing.T) {
	encCfg := testutilmod.MakeTestEncodingConfig(mint.AppModuleBasic{})
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
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			`500000000.000000000000000000`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			`500000000.000000000000000000`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd := mintcli.GetCmdQueryAnnualProvisions()
			cmd.SetOut(io.Discard)
			assert.NotNil(t, cmd)

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			assert.NoError(t, client.SetCmdClientContextHandler(baseCtx, cmd))

			out, err := clitestutil.ExecTestCLICmd(baseCtx, cmd, tc.args)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}
