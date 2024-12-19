package cli_test

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/cli"
	v1 "cosmossdk.io/x/accounts/v1"
	"cosmossdk.io/x/bank"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

type CLITestSuite struct {
	suite.Suite

	kr        keyring.Keyring
	encCfg    testutilmod.TestEncodingConfig
	baseCtx   client.Context
	clientCtx client.Context
}

func TestCLITestSuite(t *testing.T) {
	suite.Run(t, new(CLITestSuite))
}

func (s *CLITestSuite) SetupSuite() {
	s.encCfg = testutilmod.MakeTestEncodingConfig(codectestutil.CodecOptions{}, bank.AppModule{})
	s.kr = keyring.NewInMemory(s.encCfg.Codec)

	s.baseCtx = client.Context{}.
		WithKeyring(s.kr).
		WithTxConfig(s.encCfg.TxConfig).
		WithCodec(s.encCfg.Codec).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithAddressCodec(addresscodec.NewBech32Codec("cosmos")).
		WithValidatorAddressCodec(addresscodec.NewBech32Codec("cosmosvaloper")).
		WithConsensusAddressCodec(addresscodec.NewBech32Codec("cosmosvalcons"))
}

func (s *CLITestSuite) TestTxInitCmd() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)
	accountStr := make([]string, len(accounts))
	for i, acc := range accounts {
		addrStr, err := s.baseCtx.AddressCodec.BytesToString(acc.Address)
		s.Require().NoError(err)
		accountStr[i] = addrStr
	}

	s.baseCtx = s.baseCtx.WithFromAddress(accounts[0].Address)

	extraArgs := []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("photon", math.NewInt(10))).String()),
		fmt.Sprintf("--%s=test-chain", flags.FlagChainID),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, accountStr[0]),
	}

	cmd := cli.GetTxInitCmd()
	cmd.SetOutput(io.Discard)

	ctxGen := func() client.Context {
		bz, _ := s.encCfg.Codec.Marshal(&v1.SchemaResponse{
			InitSchema: &v1.SchemaResponse_Handler{
				Request:  sdk.MsgTypeURL(&types.Empty{})[1:],
				Response: sdk.MsgTypeURL(&types.Empty{})[1:],
			},
		})
		c := clitestutil.NewMockCometRPCWithResponseQueryValue(bz)
		return s.baseCtx.WithClient(c)
	}
	s.clientCtx = ctxGen()

	testCases := []struct {
		name         string
		accountType  string
		jsonMsg      string
		extraArgs    []string
		expectErrMsg string
	}{
		{
			name:         "valid json msg",
			accountType:  "test",
			jsonMsg:      `{}`,
			extraArgs:    extraArgs,
			expectErrMsg: "",
		},
		{
			name:         "invalid json msg",
			accountType:  "test",
			jsonMsg:      `{"test": "jsonmsg"}`,
			extraArgs:    extraArgs,
			expectErrMsg: "provided message is not valid",
		},
		{
			name:         "invalid sender",
			accountType:  "test",
			jsonMsg:      `{}`,
			extraArgs:    append(extraArgs, fmt.Sprintf("--%s=%s", flags.FlagFrom, "bar")),
			expectErrMsg: "failed to convert address field to address",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			var args []string
			args = append(args, tc.accountType)
			args = append(args, tc.jsonMsg)
			args = append(args, tc.extraArgs...)

			cmd.SetContext(ctx)
			cmd.SetArgs(args)

			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				msg := &sdk.TxResponse{}
				s.Require().NoError(s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), msg), out.String())
			}
		})
	}
}
