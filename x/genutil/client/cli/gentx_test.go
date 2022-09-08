package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	rpcclientmock "github.com/tendermint/tendermint/rpc/client/mock"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/nft/testutil"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
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

type CLITestSuite struct {
	suite.Suite

	network *network.Network
	kr      keyring.Keyring
	encCfg  testutilmod.TestEncodingConfig
	baseCtx client.Context
}

func TestCLITestSuite(t *testing.T) {
	suite.Run(t, new(CLITestSuite))
}

func (s *CLITestSuite) SetupSuite() {
	s.encCfg = testutilmod.MakeTestEncodingConfig(genutil.AppModuleBasic{})
	s.kr = keyring.NewInMemory(s.encCfg.Codec)
	s.baseCtx = client.Context{}.
		WithKeyring(s.kr).
		WithTxConfig(s.encCfg.TxConfig).
		WithCodec(s.encCfg.Codec).
		WithClient(mockTendermintRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain")

	cfg, err := network.DefaultConfigWithAppConfig(testutil.AppConfig)
	s.Require().NoError(err)
	cfg.NumValidators = 1

	s.network, err = network.New(s.T(), s.T().TempDir(), cfg)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *CLITestSuite) TestGenTxCmd() {
	amount := sdk.NewCoin("stake", sdk.NewInt(12))

	tests := []struct {
		name     string
		ctxGen   func() client.Context
		args     []string
		expError bool
	}{
		{
			name: "invalid commission rate returns error",
			ctxGen: func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
				c := newMockTendermintRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			args: []string{
				fmt.Sprintf("--%s=%s", flags.FlagChainID, s.baseCtx.ChainID),
				fmt.Sprintf("--%s=1", stakingcli.FlagCommissionRate),
				"node0",
				amount.String(),
			},
			expError: true,
		},
		{
			name: "valid gentx",
			ctxGen: func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
				c := newMockTendermintRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			args: []string{
				fmt.Sprintf("--%s=%s", flags.FlagChainID, s.baseCtx.ChainID),
				"node0",
				amount.String(),
			},
			expError: false,
		},
		{
			name: "invalid pubkey",
			ctxGen: func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
				c := newMockTendermintRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			args: []string{
				fmt.Sprintf("--%s=%s", flags.FlagChainID, "test-chain-1"),
				fmt.Sprintf("--%s={\"key\":\"BOIkjkFruMpfOFC9oNPhiJGfmY2pHF/gwHdLDLnrnS0=\"}", stakingcli.FlagPubKey),
				"node0",
				amount.String(),
			},
			expError: true,
		},
		{
			name: "valid pubkey flag",
			ctxGen: func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
				c := newMockTendermintRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			args: []string{
				fmt.Sprintf("--%s=%s", flags.FlagChainID, "test-chain-1"),
				fmt.Sprintf("--%s={\"@type\":\"/cosmos.crypto.ed25519.PubKey\",\"key\":\"BOIkjkFruMpfOFC9oNPhiJGfmY2pHF/gwHdLDLnrnS0=\"}", stakingcli.FlagPubKey),
				"node0",
				amount.String(),
			},
			expError: false,
		},
	}

	for _, tc := range tests {
		tc := tc

		dir := s.T().TempDir()
		genTxFile := filepath.Join(dir, "myTx")
		tc.args = append(tc.args, fmt.Sprintf("--%s=%s", flags.FlagOutputDocument, genTxFile))

		s.Run(tc.name, func() {

			var outBuf bytes.Buffer
			clientCtx := tc.ctxGen().WithOutput(&outBuf)
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd := cli.GenTxCmd(
				module.NewBasicManager(),
				clientCtx.TxConfig,
				banktypes.GenesisBalancesIterator{},
				clientCtx.HomeDir,
			)
			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			s.Require().NoError(client.SetCmdClientContextHandler(clientCtx, cmd))

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expError {
				s.Require().Error(err)

				_, err = os.Open(genTxFile)
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, "test: %s\noutput: %s", tc.name, out.String())

				// validate generated transaction.
				open, err := os.Open(genTxFile)
				s.Require().NoError(err)

				all, err := io.ReadAll(open)
				s.Require().NoError(err)

				tx, err := s.encCfg.TxConfig.TxJSONDecoder()(all)
				s.Require().NoError(err)

				msgs := tx.GetMsgs()
				s.Require().Len(msgs, 1)

				// s.Require().Equal(sdk.MsgTypeURL(&types.MsgCreateValidator{}), sdk.MsgTypeURL(msgs[0]))
				// s.Require().True(val.Address.Equals(msgs[0].GetSigners()[0]))
				// s.Require().Equal(amount, msgs[0].(*types.MsgCreateValidator).Value)
				// s.Require().NoError(tx.ValidateBasic())
			}
		})
	}
}
