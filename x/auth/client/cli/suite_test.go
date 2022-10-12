package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	rpcclientmock "github.com/tendermint/tendermint/rpc/client/mock"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/client/testutil"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
)

var _ client.TendermintRPC = (*mockTendermintRPC)(nil)

type mockTendermintRPC struct {
	rpcclientmock.Client

	responseQuery abci.ResponseQuery
}

func newMockTendermintRPC(respQuery abci.ResponseQuery) mockTendermintRPC {
	return mockTendermintRPC{responseQuery: respQuery}
}

func (mockTendermintRPC) BroadcastTxSync(context.Context, tmtypes.Tx) (*coretypes.ResultBroadcastTx, error) {
	return &coretypes.ResultBroadcastTx{Code: 0}, nil
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

	kr        keyring.Keyring
	encCfg    testutilmod.TestEncodingConfig
	baseCtx   client.Context
	clientCtx client.Context
}

func TestCLITestSuite(t *testing.T) {
	suite.Run(t, new(CLITestSuite))
}

func (s *CLITestSuite) SetupSuite() {
	s.encCfg = testutilmod.MakeTestEncodingConfig(auth.AppModuleBasic{})
	s.kr = keyring.NewInMemory(s.encCfg.Codec)
	s.baseCtx = client.Context{}.
		WithKeyring(s.kr).
		WithTxConfig(s.encCfg.TxConfig).
		WithCodec(s.encCfg.Codec).
		WithClient(mockTendermintRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain")

	var outBuf bytes.Buffer
	ctxGen := func() client.Context {
		bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
		c := newMockTendermintRPC(abci.ResponseQuery{
			Value: bz,
		})
		return s.baseCtx.WithClient(c)
	}
	s.clientCtx = ctxGen().WithOutput(&outBuf)

	kb := s.clientCtx.Keyring
	_, _, err := kb.NewMnemonic("newAccount", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	account1, _, err := kb.NewMnemonic("newAccount1", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	account2, _, err := kb.NewMnemonic("newAccount2", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)
	pub1, err := account1.GetPubKey()
	s.Require().NoError(err)
	pub2, err := account2.GetPubKey()
	s.Require().NoError(err)

	// Create a dummy account for testing purpose
	_, _, err = kb.NewMnemonic("dummyAccount", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	multi := kmultisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{pub1, pub2})
	_, err = kb.SaveMultisig("multi", multi)
	s.Require().NoError(err)
}

func (s *CLITestSuite) TestCLIValidateSignatures() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	sendTokens := sdk.NewCoins(
		sdk.NewCoin("monikertoken", sdk.NewInt(10)),
		sdk.NewCoin("stake", sdk.NewInt(10)))

	res, err := s.createBankMsg(s.clientCtx, accounts[0].Address, sendTokens,
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly))
	s.Require().NoError(err)

	// write  unsigned tx to file
	unsignedTx := testutil.WriteToNewTempFile(s.T(), res.String())
	defer unsignedTx.Close()
	res, err = authtestutil.TxSignExec(s.clientCtx, accounts[0].Address, unsignedTx.Name())
	s.Require().NoError(err)
	signedTx, err := s.clientCtx.TxConfig.TxJSONDecoder()(res.Bytes())
	s.Require().NoError(err)

	signedTxFile := testutil.WriteToNewTempFile(s.T(), res.String())
	defer signedTxFile.Close()
	txBuilder, err := s.clientCtx.TxConfig.WrapTxBuilder(signedTx)
	s.Require().NoError(err)
	_, err = authtestutil.TxValidateSignaturesExec(s.clientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	txBuilder.SetMemo("MODIFIED TX")
	bz, err := s.clientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	modifiedTxFile := testutil.WriteToNewTempFile(s.T(), string(bz))
	defer modifiedTxFile.Close()

	_, err = authtestutil.TxValidateSignaturesExec(s.clientCtx, modifiedTxFile.Name())
	s.Require().EqualError(err, "signatures validation failed")
}

func (s *CLITestSuite) createBankMsg(clientCtx client.Context, toAddr sdk.AccAddress, amount sdk.Coins, extraFlags ...string) (testutil.BufferWriter, error) {
	flags := []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees,
			sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10))).String()),
	}

	flags = append(flags, extraFlags...)
	return clitestutil.MsgSendExec(clientCtx, toAddr, toAddr, amount, flags...)
}

func (s *CLITestSuite) TestCLISignGenOnly() {
	// val := s.network.Validators[0]
	// val2 := s.network.Validators[1]
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 2)
	val := accounts[0]
	val2 := accounts[1]

	k, err := s.clientCtx.Keyring.KeyByAddress(val.Address)
	s.Require().NoError(err)
	keyName := k.Name

	addr, err := k.GetAddress()
	s.Require().NoError(err)

	account, err := s.clientCtx.AccountRetriever.GetAccount(s.clientCtx, addr)
	s.Require().NoError(err)

	sendTokens := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10)))
	args := []string{
		keyName, // from keyname
		val2.Address.String(),
		sendTokens.String(),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly), // shouldn't break if we use keyname with --generate-only flag
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
	}
	generatedStd, err := clitestutil.ExecTestCLICmd(s.clientCtx, bankcli.NewSendTxCmd(), args)
	s.Require().NoError(err)
	opFile := testutil.WriteToNewTempFile(s.T(), generatedStd.String())
	defer opFile.Close()

	commonArgs := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		fmt.Sprintf("--%s=%s", flags.FlagHome, strings.Replace(s.clientCtx.HomeDir, "simd", "simcli", 1)),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, s.clientCtx.ChainID),
	}

	cases := []struct {
		name   string
		args   []string
		expErr bool
		errMsg string
	}{
		{
			"offline mode with account-number, sequence and keyname (valid)",
			[]string{
				opFile.Name(),
				fmt.Sprintf("--%s=true", flags.FlagOffline),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, keyName),
				fmt.Sprintf("--%s=%d", flags.FlagAccountNumber, account.GetAccountNumber()),
				fmt.Sprintf("--%s=%d", flags.FlagSequence, account.GetSequence()),
			},
			false,
			"",
		},
		{
			"offline mode with account-number, sequence and address key (valid)",
			[]string{
				opFile.Name(),
				fmt.Sprintf("--%s=true", flags.FlagOffline),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%d", flags.FlagAccountNumber, account.GetAccountNumber()),
				fmt.Sprintf("--%s=%d", flags.FlagSequence, account.GetSequence()),
			},
			false,
			"",
		},
		{
			"offline mode without account-number and keyname (invalid)",
			[]string{
				opFile.Name(),
				fmt.Sprintf("--%s=true", flags.FlagOffline),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, keyName),
				fmt.Sprintf("--%s=%d", flags.FlagSequence, account.GetSequence()),
			},
			true,
			`required flag(s) "account-number" not set`,
		},
		{
			"offline mode without sequence and keyname (invalid)",
			[]string{
				opFile.Name(),
				fmt.Sprintf("--%s=true", flags.FlagOffline),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, keyName),
				fmt.Sprintf("--%s=%d", flags.FlagAccountNumber, account.GetAccountNumber()),
			},
			true,
			`required flag(s) "sequence" not set`,
		},
		{
			"offline mode without account-number, sequence and keyname (invalid)",
			[]string{
				opFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, keyName),
				fmt.Sprintf("--%s=true", flags.FlagOffline),
			},
			true,
			`required flag(s) "account-number", "sequence" not set`,
		},
	}

	for _, tc := range cases {
		cmd := authcli.GetSignCommand()
		tmcli.PrepareBaseCmd(cmd, "", "")
		out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, append(tc.args, commonArgs...))
		if tc.expErr {
			s.Require().Error(err)
			s.Require().Contains(err.Error(), tc.errMsg)
		} else {
			s.Require().NoError(err)
			func() {
				signedTx := testutil.WriteToNewTempFile(s.T(), out.String())
				defer signedTx.Close()
				_, err := authtestutil.TxBroadcastExec(s.clientCtx, signedTx.Name())
				s.Require().NoError(err)
			}()
		}
	}
}
