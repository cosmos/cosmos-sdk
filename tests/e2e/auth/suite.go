package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/address"
	"cosmossdk.io/depinject"
	"cosmossdk.io/math"
	authcli "cosmossdk.io/x/auth/client/cli"
	authclitestutil "cosmossdk.io/x/auth/client/testutil"
	authtestutil "cosmossdk.io/x/auth/testutil"
	banktypes "cosmossdk.io/x/bank/types"
	govtestutil "cosmossdk.io/x/gov/client/testutil"
	govtypes "cosmossdk.io/x/gov/types/v1beta1"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
)

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	ac      address.Codec
	network network.NetworkI
}

func NewE2ETestSuite(cfg network.Config) *E2ETestSuite {
	return &E2ETestSuite{cfg: cfg}
}

func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")
	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	kb := s.network.GetValidators()[0].GetClientCtx().Keyring
	_, _, err = kb.NewMnemonic("newAccount", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
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
	s.Require().NoError(s.network.WaitForNextBlock())

	s.ac = addresscodec.NewBech32Codec("cosmos")
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	s.network.Cleanup()
}

func (s *E2ETestSuite) TestCLISignGenOnly() {
	val := s.network.GetValidators()[0]
	val2 := s.network.GetValidators()[1]

	k, err := val.GetClientCtx().Keyring.KeyByAddress(val.GetAddress())
	s.Require().NoError(err)
	keyName := k.Name

	addr, err := k.GetAddress()
	s.Require().NoError(err)

	account, err := val.GetClientCtx().AccountRetriever.GetAccount(val.GetClientCtx(), addr)
	s.Require().NoError(err)

	sendTokens := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10)))
	msgSend := &banktypes.MsgSend{
		FromAddress: val.GetAddress().String(),
		ToAddress:   val2.GetAddress().String(),
		Amount:      sendTokens,
	}

	generatedStd, err := clitestutil.SubmitTestTx(
		val.GetClientCtx(),
		msgSend,
		val.GetAddress(),
		clitestutil.TestTxConfig{
			GenOnly: true,
		},
	)
	s.Require().NoError(err)
	opFile := testutil.WriteToNewTempFile(s.T(), generatedStd.String())
	defer opFile.Close()

	commonArgs := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		fmt.Sprintf("--%s=%s", flags.FlagHome, strings.Replace(val.GetClientCtx().HomeDir, "simd", "simcli", 1)),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, val.GetClientCtx().ChainID),
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
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.GetAddress().String()),
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
		cmd.PersistentFlags().String(flags.FlagHome, val.GetClientCtx().HomeDir, "directory for config and data")
		out, err := clitestutil.ExecTestCLICmd(val.GetClientCtx(), cmd, append(tc.args, commonArgs...))
		if tc.expErr {
			s.Require().Error(err)
			s.Require().Contains(err.Error(), tc.errMsg)
		} else {
			s.Require().NoError(err)
			func() {
				signedTx := testutil.WriteToNewTempFile(s.T(), out.String())
				defer signedTx.Close()
				_, err := authclitestutil.TxBroadcastExec(val.GetClientCtx(), signedTx.Name())
				s.Require().NoError(err)
			}()
		}
	}
}

func (s *E2ETestSuite) TestCLISignBatch() {
	val := s.network.GetValidators()[0]
	clientCtx := val.GetClientCtx()
	sendTokens := sdk.NewCoins(
		sdk.NewCoin(fmt.Sprintf("%stoken", val.GetMoniker()), math.NewInt(10)),
		sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10)),
	)

	generatedStd, err := s.createBankMsg(
		val,
		val.GetAddress(),
		sendTokens, clitestutil.TestTxConfig{
			GenOnly: true,
		},
	)
	s.Require().NoError(err)

	outputFile := testutil.WriteToNewTempFile(s.T(), strings.Repeat(generatedStd.String()+"\n", 3))
	defer outputFile.Close()
	clientCtx.HomeDir = strings.Replace(clientCtx.HomeDir, "simd", "simcli", 1)

	// sign-batch file - offline is set but account-number and sequence are not
	_, err = authclitestutil.TxSignBatchExec(clientCtx, val.GetAddress(), outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID), "--offline")
	s.Require().EqualError(err, "required flag(s) \"account-number\", \"sequence\" not set")

	// sign-batch file - offline and sequence is set but account-number is not set
	_, err = authclitestutil.TxSignBatchExec(clientCtx, val.GetAddress(), outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID), fmt.Sprintf("--%s=%s", flags.FlagSequence, "1"), "--offline")
	s.Require().EqualError(err, "required flag(s) \"account-number\" not set")

	// sign-batch file - offline and account-number is set but sequence is not set
	_, err = authclitestutil.TxSignBatchExec(clientCtx, val.GetAddress(), outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID), fmt.Sprintf("--%s=%s", flags.FlagAccountNumber, "1"), "--offline")
	s.Require().EqualError(err, "required flag(s) \"sequence\" not set")

	// sign-batch file - sequence and account-number are set when offline is false
	res, err := authclitestutil.TxSignBatchExec(clientCtx, val.GetAddress(), outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID), fmt.Sprintf("--%s=%s", flags.FlagSequence, "1"), fmt.Sprintf("--%s=%s", flags.FlagAccountNumber, "1"))
	s.Require().NoError(err)
	s.Require().Equal(3, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))

	// sign-batch file
	res, err = authclitestutil.TxSignBatchExec(clientCtx, val.GetAddress(), outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID))
	s.Require().NoError(err)
	s.Require().Equal(3, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))

	// sign-batch file signature only
	res, err = authclitestutil.TxSignBatchExec(clientCtx, val.GetAddress(), outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID), "--signature-only")
	s.Require().NoError(err)
	s.Require().Equal(3, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))

	// Sign batch malformed tx file.
	malformedFile := testutil.WriteToNewTempFile(s.T(), fmt.Sprintf("malformed%s", generatedStd))
	defer malformedFile.Close()
	_, err = authclitestutil.TxSignBatchExec(clientCtx, val.GetAddress(), malformedFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID))
	s.Require().Error(err)

	// Sign batch malformed tx file signature only.
	_, err = authclitestutil.TxSignBatchExec(clientCtx, val.GetAddress(), malformedFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID), "--signature-only")
	s.Require().Error(err)

	// make a txn to increase the sequence of sender
	_, seq, err := clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, val.GetAddress())
	s.Require().NoError(err)

	account1, err := clientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	addr, err := account1.GetAddress()
	s.Require().NoError(err)

	// Send coins from validator to multisig.
	_, err = s.createBankMsg(
		val,
		addr,
		sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 1000)),
		clitestutil.TestTxConfig{},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	// fetch the sequence after a tx, should be incremented.
	_, seq1, err := clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, val.GetAddress())
	s.Require().NoError(err)
	s.Require().Equal(seq+1, seq1)

	// signing sign-batch should start from the last sequence.
	signed, err := authclitestutil.TxSignBatchExec(clientCtx, val.GetAddress(), outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID), "--signature-only")
	s.Require().NoError(err)
	signedTxs := strings.Split(strings.Trim(signed.String(), "\n"), "\n")
	s.Require().GreaterOrEqual(len(signedTxs), 1)

	sigs, err := s.cfg.TxConfig.UnmarshalSignatureJSON([]byte(signedTxs[0]))
	s.Require().NoError(err)
	s.Require().Equal(sigs[0].Sequence, seq1)
}

func (s *E2ETestSuite) TestCLIQueryTxCmdByHash() {
	val := s.network.GetValidators()[0]

	account2, err := val.GetClientCtx().Keyring.Key("newAccount2")
	s.Require().NoError(err)

	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)

	addr, err := account2.GetAddress()
	s.Require().NoError(err)

	// Send coins.
	res, err := s.createBankMsg(
		val,
		addr,
		sdk.NewCoins(sendTokens),
		clitestutil.TestTxConfig{},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	var txRes sdk.TxResponse
	s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(res.Bytes(), &txRes))

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		rawLogContains string
	}{
		{
			"not enough args",
			[]string{},
			true, "",
		},
		{
			"with invalid hash",
			[]string{"somethinginvalid", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			true, "",
		},
		{
			"with valid and not existing hash",
			[]string{"C7E7D3A86A17AB3A321172239F3B61357937AF0F25D9FA4D2F4DCCAD9B0D7747", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			true, "",
		},
		{
			"happy case",
			[]string{txRes.TxHash, fmt.Sprintf("--%s=json", flags.FlagOutput)},
			false,
			sdk.MsgTypeURL(&banktypes.MsgSend{}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := authcli.QueryTxCmd()
			clientCtx := val.GetClientCtx()
			var (
				out testutil.BufferWriter
				err error
			)

			err = s.network.RetryForBlocks(func() error {
				out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
				return err
			}, 2)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().NotEqual("internal", err.Error())
			} else {
				var result sdk.TxResponse
				s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(out.Bytes(), &result))
				s.Require().NotNil(result.Height)
				if ok := s.deepContains(result.Events, tc.rawLogContains); !ok {
					s.Require().Fail("raw log does not contain the expected value, expected value: %s", tc.rawLogContains)
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestCLIQueryTxCmdByEvents() {
	val := s.network.GetValidators()[0]

	account2, err := val.GetClientCtx().Keyring.Key("newAccount2")
	s.Require().NoError(err)

	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)

	addr2, err := account2.GetAddress()
	s.Require().NoError(err)

	// Send coins.
	res, err := s.createBankMsg(
		val,
		addr2,
		sdk.NewCoins(sendTokens),
		clitestutil.TestTxConfig{},
	)
	s.Require().NoError(err)
	var txRes sdk.TxResponse
	s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(res.Bytes(), &txRes))
	s.Require().NoError(s.network.WaitForNextBlock())

	var out testutil.BufferWriter
	// Query the tx by hash to get the inner tx.
	err = s.network.RetryForBlocks(func() error {
		out, err = clitestutil.ExecTestCLICmd(val.GetClientCtx(), authcli.QueryTxCmd(), []string{txRes.TxHash, fmt.Sprintf("--%s=json", flags.FlagOutput)})
		return err
	}, 3)
	s.Require().NoError(err)
	s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(out.Bytes(), &txRes))
	protoTx := txRes.GetTx().(*tx.Tx)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrStr string
	}{
		{
			"invalid --type",
			[]string{
				fmt.Sprintf("--type=%s", "foo"),
				"bar",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true, "unknown --type value foo",
		},
		{
			"--type=acc_seq with no addr+seq",
			[]string{
				"--type=acc_seq",
				"",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true, "`acc_seq` type takes an argument '<addr>/<seq>'",
		},
		{
			"non-existing addr+seq combo",
			[]string{
				"--type=acc_seq",
				"foobar",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true, "found no txs matching given address and sequence combination",
		},
		{
			"addr+seq happy case",
			[]string{
				"--type=acc_seq",
				fmt.Sprintf("%s/%d", val.GetAddress(), protoTx.AuthInfo.SignerInfos[0].Sequence),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false, "",
		},
		{
			"--type=signature with no signature",
			[]string{
				"--type=signature",
				"",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true, "argument should be comma-separated signatures",
		},
		{
			"non-existing signatures",
			[]string{
				"--type=signature",
				"foo",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true, "found no txs matching given signatures",
		},
		{
			"with --signatures happy case",
			[]string{
				"--type=signature",
				base64.StdEncoding.EncodeToString(protoTx.Signatures[0]),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false, "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := authcli.QueryTxCmd()
			clientCtx := val.GetClientCtx()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectErrStr)
			} else {
				var result sdk.TxResponse
				s.Require().NoError(err)
				s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(out.Bytes(), &result))
				s.Require().NotNil(result.Height)
			}
		})
	}
}

func (s *E2ETestSuite) TestCLIQueryTxsCmdByEvents() {
	val := s.network.GetValidators()[0]

	account2, err := val.GetClientCtx().Keyring.Key("newAccount2")
	s.Require().NoError(err)

	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)

	addr2, err := account2.GetAddress()
	s.Require().NoError(err)
	// Send coins.
	res, err := s.createBankMsg(
		val,
		addr2,
		sdk.NewCoins(sendTokens),
		clitestutil.TestTxConfig{},
	)
	s.Require().NoError(err)
	var txRes sdk.TxResponse
	s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(res.Bytes(), &txRes))
	s.Require().NoError(s.network.WaitForNextBlock())

	var out testutil.BufferWriter
	// Query the tx by hash to get the inner tx.
	err = s.network.RetryForBlocks(func() error {
		out, err = clitestutil.ExecTestCLICmd(val.GetClientCtx(), authcli.QueryTxCmd(), []string{txRes.TxHash, fmt.Sprintf("--%s=json", flags.FlagOutput)})
		return err
	}, 3)
	s.Require().NoError(err)
	s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(out.Bytes(), &txRes))

	testCases := []struct {
		name        string
		args        []string
		expectEmpty bool
	}{
		{
			"fee event happy case",
			[]string{
				fmt.Sprintf(
					"--query=tx.fee='%s'",
					sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String(),
				),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
		},
		{
			"no matching fee event",
			[]string{
				fmt.Sprintf(
					"--query=tx.fee='%s'",
					sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(0))).String(),
				),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := authcli.QueryTxsByEventsCmd()
			clientCtx := val.GetClientCtx()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result sdk.SearchTxsResult
			s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(out.Bytes(), &result))

			if tc.expectEmpty {
				s.Require().Equal(0, len(result.Txs))
			} else {
				s.Require().NotEqual(0, len(result.Txs))
				s.Require().NotNil(result.Txs[0])
			}
		})
	}
}

func (s *E2ETestSuite) TestCLISendGenerateSignAndBroadcast() {
	val1 := s.network.GetValidators()[0]
	clientCtx := val1.GetClientCtx()

	account, err := clientCtx.Keyring.Key("newAccount")
	s.Require().NoError(err)

	sendTokens := sdk.NewCoin(s.cfg.BondDenom, sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction))

	addr, err := account.GetAddress()
	s.Require().NoError(err)
	normalGeneratedTx, err := s.createBankMsg(
		val1,
		addr,
		sdk.NewCoins(sendTokens),
		clitestutil.TestTxConfig{
			GenOnly: true,
		},
	)
	s.Require().NoError(err)
	txCfg := clientCtx.TxConfig

	normalGeneratedStdTx, err := txCfg.TxJSONDecoder()(normalGeneratedTx.Bytes())
	s.Require().NoError(err)
	txBuilder, err := txCfg.WrapTxBuilder(normalGeneratedStdTx)
	s.Require().NoError(err)
	s.Require().Equal(txBuilder.GetTx().GetGas(), uint64(flags.DefaultGasLimit))
	s.Require().Equal(len(txBuilder.GetTx().GetMsgs()), 1)
	sigs, err := txBuilder.GetTx().GetSignaturesV2()
	s.Require().NoError(err)
	s.Require().Equal(0, len(sigs))

	// Test generate sendTx with --gas=$amount
	limitedGasGeneratedTx, err := s.createBankMsg(val1, addr,
		sdk.NewCoins(sendTokens), clitestutil.TestTxConfig{
			GenOnly: true,
			Gas:     100,
		},
	)
	s.Require().NoError(err)

	limitedGasStdTx, err := txCfg.TxJSONDecoder()(limitedGasGeneratedTx.Bytes())
	s.Require().NoError(err)
	txBuilder, err = txCfg.WrapTxBuilder(limitedGasStdTx)
	s.Require().NoError(err)
	s.Require().Equal(txBuilder.GetTx().GetGas(), uint64(100))
	s.Require().Equal(len(txBuilder.GetTx().GetMsgs()), 1)
	sigs, err = txBuilder.GetTx().GetSignaturesV2()
	s.Require().NoError(err)
	s.Require().Equal(0, len(sigs))

	resp, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", val1.GetAPIAddress(), val1.GetAddress()))
	s.Require().NoError(err)

	var balRes banktypes.QueryAllBalancesResponse
	err = clientCtx.Codec.UnmarshalJSON(resp, &balRes)
	s.Require().NoError(err)
	startTokens := balRes.Balances.AmountOf(s.cfg.BondDenom)

	// Test generate sendTx, estimate gas
	finalGeneratedTx, err := s.createBankMsg(
		val1,
		addr,
		sdk.NewCoins(sendTokens), clitestutil.TestTxConfig{
			GenOnly: true,
			Gas:     flags.DefaultGasLimit,
		})
	s.Require().NoError(err)

	finalStdTx, err := txCfg.TxJSONDecoder()(finalGeneratedTx.Bytes())
	s.Require().NoError(err)
	txBuilder, err = txCfg.WrapTxBuilder(finalStdTx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(flags.DefaultGasLimit), txBuilder.GetTx().GetGas())
	s.Require().Equal(len(finalStdTx.GetMsgs()), 1)

	// Write the output to disk
	unsignedTxFile := testutil.WriteToNewTempFile(s.T(), finalGeneratedTx.String())
	defer unsignedTxFile.Close()

	// Test validate-signatures
	res, err := authclitestutil.TxValidateSignaturesExec(clientCtx, unsignedTxFile.Name())
	s.Require().EqualError(err, "signatures validation failed")
	s.Require().True(strings.Contains(res.String(), fmt.Sprintf("Signers:\n  0: %v\n\nSignatures:\n\n", val1.GetAddress().String())))

	// Test sign

	// Does not work in offline mode
	_, err = authclitestutil.TxSignExec(clientCtx, val1.GetAddress(), unsignedTxFile.Name(), "--offline")
	s.Require().EqualError(err, "required flag(s) \"account-number\", \"sequence\" not set")

	// But works offline if we set account number and sequence
	clientCtx.HomeDir = strings.Replace(clientCtx.HomeDir, "simd", "simcli", 1)
	_, err = authclitestutil.TxSignExec(clientCtx, val1.GetAddress(), unsignedTxFile.Name(), "--offline", "--account-number", "1", "--sequence", "1")
	s.Require().NoError(err)

	// Sign transaction
	signedTx, err := authclitestutil.TxSignExec(clientCtx, val1.GetAddress(), unsignedTxFile.Name())
	s.Require().NoError(err)
	signedFinalTx, err := txCfg.TxJSONDecoder()(signedTx.Bytes())
	s.Require().NoError(err)
	txBuilder, err = clientCtx.TxConfig.WrapTxBuilder(signedFinalTx)
	s.Require().NoError(err)
	s.Require().Equal(len(txBuilder.GetTx().GetMsgs()), 1)
	sigs, err = txBuilder.GetTx().GetSignaturesV2()
	s.Require().NoError(err)
	s.Require().Equal(1, len(sigs))
	signers, err := txBuilder.GetTx().GetSigners()
	s.Require().NoError(err)
	s.Require().Equal([]byte(val1.GetAddress()), signers[0])

	// Write the output to disk
	signedTxFile := testutil.WriteToNewTempFile(s.T(), signedTx.String())
	defer signedTxFile.Close()

	// validate Signature
	res, err = authclitestutil.TxValidateSignaturesExec(clientCtx, signedTxFile.Name())
	s.Require().NoError(err)
	s.Require().True(strings.Contains(res.String(), "[OK]"))
	s.Require().NoError(s.network.WaitForNextBlock())

	// Ensure foo has right amount of funds
	resp, err = testutil.GetRequest(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", val1.GetAPIAddress(), val1.GetAddress()))
	s.Require().NoError(err)
	err = clientCtx.Codec.UnmarshalJSON(resp, &balRes)
	s.Require().NoError(err)
	s.Require().Equal(startTokens, balRes.Balances.AmountOf(s.cfg.BondDenom))

	// Test broadcast

	// Does not work in offline mode
	_, err = authclitestutil.TxBroadcastExec(clientCtx, signedTxFile.Name(), "--offline")
	s.Require().EqualError(err, "cannot broadcast tx during offline mode")
	s.Require().NoError(s.network.WaitForNextBlock())

	// Broadcast correct transaction.
	clientCtx.BroadcastMode = flags.BroadcastSync
	_, err = authclitestutil.TxBroadcastExec(clientCtx, signedTxFile.Name())
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	// Ensure destiny account state
	err = s.network.RetryForBlocks(func() error {
		resp, err = testutil.GetRequest(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", val1.GetAPIAddress(), addr))
		s.Require().NoError(err)
		return err
	}, 3)
	s.Require().NoError(err)

	err = clientCtx.Codec.UnmarshalJSON(resp, &balRes)
	s.Require().NoError(err)
	s.Require().Equal(sendTokens.Amount, balRes.Balances.AmountOf(s.cfg.BondDenom))

	// Ensure origin account state
	resp, err = testutil.GetRequest(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", val1.GetAPIAddress(), val1.GetAddress()))
	s.Require().NoError(err)
	err = clientCtx.Codec.UnmarshalJSON(resp, &balRes)
	s.Require().NoError(err)
}

func (s *E2ETestSuite) TestCLIMultisignInsufficientCosigners() {
	val1 := s.network.GetValidators()[0]
	clientCtx := val1.GetClientCtx()

	// Fetch account and a multisig info
	account1, err := clientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	multisigRecord, err := clientCtx.Keyring.Key("multi")
	s.Require().NoError(err)

	addr, err := multisigRecord.GetAddress()
	s.Require().NoError(err)
	// Send coins from validator to multisig.
	_, err = s.createBankMsg(
		val1,
		addr,
		sdk.NewCoins(
			sdk.NewInt64Coin(s.cfg.BondDenom, 10),
		),
		clitestutil.TestTxConfig{},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	coins := sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 5))
	msgSend := &banktypes.MsgSend{
		FromAddress: addr.String(),
		ToAddress:   val1.GetAddress().String(),
		Amount:      coins,
	}

	// Generate multisig transaction.
	multiGeneratedTx, err := clitestutil.SubmitTestTx(
		clientCtx,
		msgSend,
		addr,
		clitestutil.TestTxConfig{
			GenOnly: true,
		},
	)
	s.Require().NoError(err)

	// Save tx to file
	multiGeneratedTxFile := testutil.WriteToNewTempFile(s.T(), multiGeneratedTx.String())
	defer multiGeneratedTxFile.Close()

	// Multisign, sign with one signature
	clientCtx.HomeDir = strings.Replace(clientCtx.HomeDir, "simd", "simcli", 1)
	addr1, err := account1.GetAddress()
	s.Require().NoError(err)
	account1Signature, err := authclitestutil.TxSignExec(clientCtx, addr1, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	s.Require().NoError(err)

	sign1File := testutil.WriteToNewTempFile(s.T(), account1Signature.String())
	defer sign1File.Close()

	multiSigWith1Signature, err := authclitestutil.TxMultiSignExec(clientCtx, multisigRecord.Name, multiGeneratedTxFile.Name(), sign1File.Name())
	s.Require().NoError(err)

	// Save tx to file
	multiSigWith1SignatureFile := testutil.WriteToNewTempFile(s.T(), multiSigWith1Signature.String())
	defer multiSigWith1SignatureFile.Close()

	_, err = authclitestutil.TxValidateSignaturesExec(clientCtx, multiSigWith1SignatureFile.Name())
	s.Require().Error(err)
}

func (s *E2ETestSuite) TestCLIEncode() {
	val1 := s.network.GetValidators()[0]

	sendTokens := sdk.NewCoin(s.cfg.BondDenom, sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction))

	normalGeneratedTx, err := s.createBankMsg(
		val1, val1.GetAddress(),
		sdk.NewCoins(sendTokens),
		clitestutil.TestTxConfig{
			GenOnly: true,
			Memo:    "deadbeef",
		},
	)
	s.Require().NoError(err)
	savedTxFile := testutil.WriteToNewTempFile(s.T(), normalGeneratedTx.String())
	defer savedTxFile.Close()

	// Encode
	encodeExec, err := authclitestutil.TxEncodeExec(val1.GetClientCtx(), savedTxFile.Name())
	s.Require().NoError(err)
	trimmedBase64 := strings.Trim(encodeExec.String(), "\"\n")

	// Check that the transaction decodes as expected
	decodedTx, err := authclitestutil.TxDecodeExec(val1.GetClientCtx(), trimmedBase64)
	s.Require().NoError(err)

	txCfg := val1.GetClientCtx().TxConfig
	theTx, err := txCfg.TxJSONDecoder()(decodedTx.Bytes())
	s.Require().NoError(err)
	txBuilder, err := val1.GetClientCtx().TxConfig.WrapTxBuilder(theTx)
	s.Require().NoError(err)
	s.Require().Equal("deadbeef", txBuilder.GetTx().GetMemo())
}

func (s *E2ETestSuite) TestCLIMultisignSortSignatures() {
	val1 := s.network.GetValidators()[0]
	clientCtx := val1.GetClientCtx()

	// Generate 2 accounts and a multisig.
	account1, err := clientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	account2, err := clientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)

	multisigRecord, err := clientCtx.Keyring.Key("multi")
	s.Require().NoError(err)

	// Generate dummy account which is not a part of multisig.
	dummyAcc, err := clientCtx.Keyring.Key("dummyAccount")
	s.Require().NoError(err)

	addr, err := multisigRecord.GetAddress()
	s.Require().NoError(err)
	resp, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", val1.GetAPIAddress(), addr))
	s.Require().NoError(err)

	var balRes banktypes.QueryAllBalancesResponse
	err = clientCtx.Codec.UnmarshalJSON(resp, &balRes)
	s.Require().NoError(err)
	initialCoins := balRes.Balances

	// Send coins from validator to multisig.
	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)
	_, err = s.createBankMsg(
		val1,
		addr,
		sdk.NewCoins(sendTokens),
		clitestutil.TestTxConfig{},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	resp, err = testutil.GetRequest(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", val1.GetAPIAddress(), addr))
	s.Require().NoError(err)
	err = clientCtx.Codec.UnmarshalJSON(resp, &balRes)
	s.Require().NoError(err)
	diff, _ := balRes.Balances.SafeSub(initialCoins...)
	s.Require().Equal(sendTokens.Amount, diff.AmountOf(s.cfg.BondDenom))

	tokens := sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 5))
	msgSend := &banktypes.MsgSend{
		FromAddress: addr.String(),
		ToAddress:   val1.GetAddress().String(),
		Amount:      tokens,
	}

	// Generate multisig transaction.
	multiGeneratedTx, err := clitestutil.SubmitTestTx(
		clientCtx,
		msgSend,
		addr,
		clitestutil.TestTxConfig{
			GenOnly: true,
		},
	)
	s.Require().NoError(err)

	// Save tx to file
	multiGeneratedTxFile := testutil.WriteToNewTempFile(s.T(), multiGeneratedTx.String())
	defer multiGeneratedTxFile.Close()

	// Sign with account1
	addr1, err := account1.GetAddress()
	s.Require().NoError(err)
	clientCtx.HomeDir = strings.Replace(clientCtx.HomeDir, "simd", "simcli", 1)
	account1Signature, err := authclitestutil.TxSignExec(clientCtx, addr1, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	s.Require().NoError(err)

	sign1File := testutil.WriteToNewTempFile(s.T(), account1Signature.String())
	defer sign1File.Close()

	// Sign with account2
	addr2, err := account2.GetAddress()
	s.Require().NoError(err)
	account2Signature, err := authclitestutil.TxSignExec(clientCtx, addr2, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	s.Require().NoError(err)

	sign2File := testutil.WriteToNewTempFile(s.T(), account2Signature.String())
	defer sign2File.Close()

	// Sign with dummy account
	dummyAddr, err := dummyAcc.GetAddress()
	s.Require().NoError(err)
	_, err = authclitestutil.TxSignExec(clientCtx, dummyAddr, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "signing key is not a part of multisig key")

	multiSigWith2Signatures, err := authclitestutil.TxMultiSignExec(clientCtx, multisigRecord.Name, multiGeneratedTxFile.Name(), sign1File.Name(), sign2File.Name())
	s.Require().NoError(err)

	// Write the output to disk
	signedTxFile := testutil.WriteToNewTempFile(s.T(), multiSigWith2Signatures.String())
	defer signedTxFile.Close()

	_, err = authclitestutil.TxValidateSignaturesExec(clientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	clientCtx.BroadcastMode = flags.BroadcastSync
	_, err = authclitestutil.TxBroadcastExec(clientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *E2ETestSuite) TestSignWithMultisig() {
	val1 := s.network.GetValidators()[0]

	// Generate a account for signing.
	account1, err := val1.GetClientCtx().Keyring.Key("newAccount1")
	s.Require().NoError(err)

	addr1, err := account1.GetAddress()
	s.Require().NoError(err)

	// Create an address that is not in the keyring, will be used to simulate `--multisig`
	multisig := "cosmos1hd6fsrvnz6qkp87s3u86ludegq97agxsdkwzyh"
	multisigAddr, err := sdk.AccAddressFromBech32(multisig)
	s.Require().NoError(err)

	tokens := sdk.NewCoins(
		sdk.NewInt64Coin(s.cfg.BondDenom, 5),
	)
	msgSend := &banktypes.MsgSend{
		FromAddress: val1.GetAddress().String(),
		ToAddress:   val1.GetAddress().String(),
		Amount:      tokens,
	}

	// Generate a transaction for testing --multisig with an address not in the keyring.
	multisigTx, err := clitestutil.SubmitTestTx(
		val1.GetClientCtx(),
		msgSend,
		val1.GetAddress(),
		clitestutil.TestTxConfig{
			GenOnly: true,
		},
	)
	s.Require().NoError(err)

	// Save multi tx to file
	multiGeneratedTx2File := testutil.WriteToNewTempFile(s.T(), multisigTx.String())
	defer multiGeneratedTx2File.Close()

	// Sign using multisig. We're signing a tx on behalf of the multisig address,
	// even though the tx signer is NOT the multisig address. This is fine though,
	// as the main point of this test is to test the `--multisig` flag with an address
	// that is not in the keyring.
	_, err = authclitestutil.TxSignExec(val1.GetClientCtx(), addr1, multiGeneratedTx2File.Name(), "--multisig", multisigAddr.String())
	s.Require().Contains(err.Error(), "error getting account from keybase")
}

func (s *E2ETestSuite) TestCLIMultisign() {
	val1 := s.network.GetValidators()[0]
	clientCtx := val1.GetClientCtx()

	// Generate 2 accounts and a multisig.
	account1, err := clientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	account2, err := clientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)

	multisigRecord, err := clientCtx.Keyring.Key("multi")
	s.Require().NoError(err)

	addr, err := multisigRecord.GetAddress()
	s.Require().NoError(err)

	// Send coins from validator to multisig.
	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)
	s.Require().NoError(s.network.WaitForNextBlock())
	_, err = s.createBankMsg(
		val1, addr,
		sdk.NewCoins(sendTokens),
		clitestutil.TestTxConfig{},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	var balRes banktypes.QueryAllBalancesResponse
	err = s.network.RetryForBlocks(func() error {
		resp, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", val1.GetAPIAddress(), addr))
		if err != nil {
			return err
		}
		return clientCtx.Codec.UnmarshalJSON(resp, &balRes)
	}, 3)
	s.Require().NoError(err)
	s.Require().True(sendTokens.Amount.Equal(balRes.Balances.AmountOf(s.cfg.BondDenom)))

	tokens := sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 5))
	msgSend := &banktypes.MsgSend{
		FromAddress: addr.String(),
		ToAddress:   val1.GetAddress().String(),
		Amount:      tokens,
	}

	// Generate multisig transaction.
	multiGeneratedTx, err := clitestutil.SubmitTestTx(
		clientCtx,
		msgSend,
		addr,
		clitestutil.TestTxConfig{
			GenOnly: true,
		},
	)
	s.Require().NoError(err)

	// Save tx to file
	multiGeneratedTxFile := testutil.WriteToNewTempFile(s.T(), multiGeneratedTx.String())
	defer multiGeneratedTxFile.Close()

	addr1, err := account1.GetAddress()
	s.Require().NoError(err)
	// Sign with account1
	clientCtx.HomeDir = strings.Replace(clientCtx.HomeDir, "simd", "simcli", 1)
	account1Signature, err := authclitestutil.TxSignExec(clientCtx, addr1, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	s.Require().NoError(err)

	sign1File := testutil.WriteToNewTempFile(s.T(), account1Signature.String())
	defer sign1File.Close()

	addr2, err := account2.GetAddress()
	s.Require().NoError(err)
	// Sign with account2
	account2Signature, err := authclitestutil.TxSignExec(clientCtx, addr2, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	s.Require().NoError(err)

	sign2File := testutil.WriteToNewTempFile(s.T(), account2Signature.String())
	defer sign2File.Close()

	// Work in offline mode.
	multisigAccNum, multisigSeq, err := clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, addr)
	s.Require().NoError(err)
	_, err = authclitestutil.TxMultiSignExec(
		clientCtx,
		multisigRecord.Name,
		multiGeneratedTxFile.Name(),
		fmt.Sprintf("--%s", flags.FlagOffline),
		fmt.Sprintf("--%s=%d", flags.FlagAccountNumber, multisigAccNum),
		fmt.Sprintf("--%s=%d", flags.FlagSequence, multisigSeq),
		sign1File.Name(),
		sign2File.Name(),
	)
	s.Require().NoError(err)

	clientCtx.Offline = false
	multiSigWith2Signatures, err := authclitestutil.TxMultiSignExec(clientCtx, multisigRecord.Name, multiGeneratedTxFile.Name(), sign1File.Name(), sign2File.Name())
	s.Require().NoError(err)

	// Write the output to disk
	signedTxFile := testutil.WriteToNewTempFile(s.T(), multiSigWith2Signatures.String())
	defer signedTxFile.Close()

	_, err = authclitestutil.TxValidateSignaturesExec(clientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	clientCtx.BroadcastMode = flags.BroadcastSync
	_, err = authclitestutil.TxBroadcastExec(clientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *E2ETestSuite) TestSignBatchMultisig() {
	val := s.network.GetValidators()[0]
	clientCtx := val.GetClientCtx()

	// Fetch 2 accounts and a multisig.
	account1, err := clientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)
	account2, err := clientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)
	multisigRecord, err := clientCtx.Keyring.Key("multi")
	s.Require().NoError(err)

	addr, err := multisigRecord.GetAddress()
	s.Require().NoError(err)
	// Send coins from validator to multisig.
	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)
	_, err = s.createBankMsg(
		val,
		addr,
		sdk.NewCoins(sendTokens),
		clitestutil.TestTxConfig{},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	tokens := sdk.NewCoins(
		sdk.NewCoin(s.cfg.BondDenom, math.NewInt(1)),
	)
	msgSend := &banktypes.MsgSend{
		FromAddress: addr.String(),
		ToAddress:   val.GetAddress().String(),
		Amount:      tokens,
	}

	generatedStd, err := clitestutil.SubmitTestTx(
		clientCtx,
		msgSend,
		addr,
		clitestutil.TestTxConfig{
			GenOnly: true,
		},
	)
	s.Require().NoError(err)

	// Write the output to disk
	filename := testutil.WriteToNewTempFile(s.T(), strings.Repeat(generatedStd.String(), 1))
	defer filename.Close()
	clientCtx.HomeDir = strings.Replace(clientCtx.HomeDir, "simd", "simcli", 1)

	addr1, err := account1.GetAddress()
	s.Require().NoError(err)
	// sign-batch file
	res, err := authclitestutil.TxSignBatchExec(clientCtx, addr1, filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID), "--multisig", addr.String(), "--signature-only")
	s.Require().NoError(err)
	s.Require().Equal(1, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))
	// write sigs to file
	file1 := testutil.WriteToNewTempFile(s.T(), res.String())
	defer file1.Close()

	addr2, err := account2.GetAddress()
	s.Require().NoError(err)
	// sign-batch file with account2
	res, err = authclitestutil.TxSignBatchExec(clientCtx, addr2, filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID), "--multisig", addr.String(), "--signature-only")
	s.Require().NoError(err)
	s.Require().Equal(1, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))
	// write sigs to file2
	file2 := testutil.WriteToNewTempFile(s.T(), res.String())
	defer file2.Close()
	_, err = authclitestutil.TxMultiSignExec(clientCtx, multisigRecord.Name, filename.Name(), file1.Name(), file2.Name())
	s.Require().NoError(err)
}

func (s *E2ETestSuite) TestMultisignBatch() {
	val := s.network.GetValidators()[0]
	clientCtx := val.GetClientCtx()

	// Fetch 2 accounts and a multisig.
	account1, err := clientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)
	account2, err := clientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)
	multisigRecord, err := clientCtx.Keyring.Key("multi")
	s.Require().NoError(err)

	addr, err := multisigRecord.GetAddress()
	s.Require().NoError(err)
	// Send coins from validator to multisig.
	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 1000)
	_, err = s.createBankMsg(
		val,
		addr,
		sdk.NewCoins(sendTokens),
		clitestutil.TestTxConfig{},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	tokens := sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(1)))
	msgSend := &banktypes.MsgSend{
		FromAddress: addr.String(),
		ToAddress:   val.GetAddress().String(),
		Amount:      tokens,
	}

	generatedStd, err := clitestutil.SubmitTestTx(
		clientCtx,
		msgSend,
		addr,
		clitestutil.TestTxConfig{
			GenOnly: true,
		},
	)
	s.Require().NoError(err)

	// Write the output to disk
	filename := testutil.WriteToNewTempFile(s.T(), strings.Repeat(generatedStd.String()+"\n", 3))
	defer filename.Close()
	clientCtx.HomeDir = strings.Replace(clientCtx.HomeDir, "simd", "simcli", 1)

	account, err := clientCtx.AccountRetriever.GetAccount(clientCtx, addr)
	s.Require().NoError(err)

	// sign-batch file
	addr1, err := account1.GetAddress()
	s.Require().NoError(err)
	res, err := authclitestutil.TxSignBatchExec(clientCtx, addr1, filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID), "--multisig", addr.String(), fmt.Sprintf("--%s", flags.FlagOffline), fmt.Sprintf("--%s=%s", flags.FlagAccountNumber, fmt.Sprint(account.GetAccountNumber())), fmt.Sprintf("--%s=%s", flags.FlagSequence, fmt.Sprint(account.GetSequence())), "--signature-only")
	s.Require().NoError(err)
	s.Require().Equal(3, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))
	// write sigs to file
	file1 := testutil.WriteToNewTempFile(s.T(), res.String())
	defer file1.Close()

	// sign-batch file with account2
	addr2, err := account2.GetAddress()
	s.Require().NoError(err)
	res, err = authclitestutil.TxSignBatchExec(clientCtx, addr2, filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID), "--multisig", addr.String(), fmt.Sprintf("--%s", flags.FlagOffline), fmt.Sprintf("--%s=%s", flags.FlagAccountNumber, fmt.Sprint(account.GetAccountNumber())), fmt.Sprintf("--%s=%s", flags.FlagSequence, fmt.Sprint(account.GetSequence())), "--signature-only")
	s.Require().NoError(err)
	s.Require().Equal(3, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))

	// multisign the file
	file2 := testutil.WriteToNewTempFile(s.T(), res.String())
	defer file2.Close()
	res, err = authclitestutil.TxMultiSignBatchExec(clientCtx, filename.Name(), multisigRecord.Name, file1.Name(), file2.Name())
	s.Require().NoError(err)
	signedTxs := strings.Split(strings.Trim(res.String(), "\n"), "\n")

	// Broadcast transactions.
	for _, signedTx := range signedTxs {
		func() {
			signedTxFile := testutil.WriteToNewTempFile(s.T(), signedTx)
			defer signedTxFile.Close()
			clientCtx.BroadcastMode = flags.BroadcastSync
			_, err = authclitestutil.TxBroadcastExec(clientCtx, signedTxFile.Name())
			s.Require().NoError(err)
			s.Require().NoError(s.network.WaitForNextBlock())
		}()
	}
}

func TestGetBroadcastCommandOfflineFlag(t *testing.T) {
	cmd := authcli.GetBroadcastCommand()
	_ = testutil.ApplyMockIODiscardOutErr(cmd)
	cmd.SetArgs([]string{fmt.Sprintf("--%s=true", flags.FlagOffline), ""})

	require.EqualError(t, cmd.Execute(), "cannot broadcast tx during offline mode")
}

func TestGetBroadcastCommandWithoutOfflineFlag(t *testing.T) {
	var txCfg client.TxConfig
	err := depinject.Inject(authtestutil.AppConfig, &txCfg)
	require.NoError(t, err)
	clientCtx := client.Context{}
	clientCtx = clientCtx.WithTxConfig(txCfg)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	cmd := authcli.GetBroadcastCommand()
	_, out := testutil.ApplyMockIO(cmd)

	// Create new file with tx
	builder := txCfg.NewTxBuilder()
	builder.SetGasLimit(200000)
	err = builder.SetMsgs(banktypes.NewMsgSend("cosmos1cxlt8kznps92fwu3j6npahx4mjfutydyene2qw", "cosmos1cxlt8kznps92fwu3j6npahx4mjfutydyene2qw", sdk.Coins{sdk.NewInt64Coin("stake", 10000)}))
	require.NoError(t, err)
	txContents, err := txCfg.TxJSONEncoder()(builder.GetTx())
	require.NoError(t, err)
	txFile := testutil.WriteToNewTempFile(t, string(txContents))
	defer txFile.Close()

	cmd.SetArgs([]string{txFile.Name()})
	err = cmd.ExecuteContext(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "connect: connection refused")
	require.Contains(t, out.String(), "connect: connection refused")
}

// TestTxWithoutPublicKey makes sure sending a proto tx message without the
// public key doesn't cause any error in the RPC layer (broadcast).
// See https://github.com/cosmos/cosmos-sdk/issues/7585 for more details.
func (s *E2ETestSuite) TestTxWithoutPublicKey() {
	val1 := s.network.GetValidators()[0]
	clientCtx := val1.GetClientCtx()
	txCfg := clientCtx.TxConfig

	// Create a txBuilder with an unsigned tx.
	txBuilder := txCfg.NewTxBuilder()
	msg := banktypes.NewMsgSend(val1.GetAddress().String(), val1.GetAddress().String(), sdk.NewCoins(
		sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10)),
	))
	err := txBuilder.SetMsgs(msg)
	s.Require().NoError(err)
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(150))))
	txBuilder.SetGasLimit(testdata.NewTestGasLimit())
	// Set empty signature to set signer infos.
	sigV2 := signing.SignatureV2{
		PubKey: val1.GetPubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
	}
	err = txBuilder.SetSignatures(sigV2)
	s.Require().NoError(err)

	// Create a file with the unsigned tx.
	txJSON, err := txCfg.TxJSONEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)
	unsignedTxFile := testutil.WriteToNewTempFile(s.T(), string(txJSON))
	defer unsignedTxFile.Close()

	// Sign the file with the unsignedTx.
	signedTx, err := authclitestutil.TxSignExec(clientCtx, val1.GetAddress(), unsignedTxFile.Name(), fmt.Sprintf("--%s=true", cli.FlagOverwrite))
	s.Require().NoError(err)

	// Remove the signerInfo's `public_key` field manually from the signedTx.
	// Note: this method is only used for test purposes! In general, one should
	// use txBuilder and TxEncoder/TxDecoder to manipulate txs.
	var tx tx.Tx
	err = clientCtx.Codec.UnmarshalJSON(signedTx.Bytes(), &tx)
	s.Require().NoError(err)
	tx.AuthInfo.SignerInfos[0].PublicKey = nil
	// Re-encode the tx again, to another file.
	txJSON, err = clientCtx.Codec.MarshalJSON(&tx)
	s.Require().NoError(err)
	signedTxFile := testutil.WriteToNewTempFile(s.T(), string(txJSON))
	defer signedTxFile.Close()
	s.Require().True(strings.Contains(string(txJSON), "\"public_key\":null"))

	// Broadcast tx, test that it shouldn't panic.
	clientCtx.BroadcastMode = flags.BroadcastSync
	out, err := authclitestutil.TxBroadcastExec(clientCtx, signedTxFile.Name())
	s.Require().NoError(err)
	var res sdk.TxResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
	s.Require().NotEqual(0, res.Code)
}

// TestSignWithMultiSignersAminoJSON tests the case where a transaction with 2
// messages which has to be signed with 2 different keys. Sign and append the
// signatures using the CLI with Amino signing mode. Finally, send the
// transaction to the blockchain.
func (s *E2ETestSuite) TestSignWithMultiSignersAminoJSON() {
	require := s.Require()
	val0, val1 := s.network.GetValidators()[0], s.network.GetValidators()[1]
	val0Coin := sdk.NewCoin(fmt.Sprintf("%stoken", val0.GetMoniker()), math.NewInt(10))
	val1Coin := sdk.NewCoin(fmt.Sprintf("%stoken", val1.GetMoniker()), math.NewInt(10))
	_, _, addr1 := testdata.KeyTestPubAddr()

	// Creating a tx with 2 msgs from 2 signers: val0 and val1.
	// The validators need to sign with SIGN_MODE_LEGACY_AMINO_JSON,
	// because DIRECT doesn't support multi signers via the CLI.
	// Since we use amino, we don't need to pre-populate signer_infos.
	txBuilder := val0.GetClientCtx().TxConfig.NewTxBuilder()
	val0Str, err := s.ac.BytesToString(val0.GetAddress())
	s.Require().NoError(err)
	val1Str, err := s.ac.BytesToString(val1.GetAddress())
	s.Require().NoError(err)
	addrStr, err := s.ac.BytesToString(addr1)
	s.Require().NoError(err)
	err = txBuilder.SetMsgs(
		banktypes.NewMsgSend(val0Str, addrStr, sdk.NewCoins(val0Coin)),
		banktypes.NewMsgSend(val1Str, addrStr, sdk.NewCoins(val1Coin)),
	)
	require.NoError(err)
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))))
	txBuilder.SetGasLimit(testdata.NewTestGasLimit() * 2)
	signers, err := txBuilder.GetTx().GetSigners()
	require.NoError(err)
	require.Equal([][]byte{val0.GetAddress(), val1.GetAddress()}, signers)

	// Write the unsigned tx into a file.
	txJSON, err := val0.GetClientCtx().TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	require.NoError(err)
	unsignedTxFile := testutil.WriteToNewTempFile(s.T(), string(txJSON))
	defer unsignedTxFile.Close()

	// Let val0 sign first the file with the unsignedTx.
	signedByVal0, err := authclitestutil.TxSignExec(val0.GetClientCtx(), val0.GetAddress(), unsignedTxFile.Name(), "--overwrite", "--sign-mode=amino-json")
	require.NoError(err)
	signedByVal0File := testutil.WriteToNewTempFile(s.T(), signedByVal0.String())
	defer signedByVal0File.Close()

	// Then let val1 sign the file with signedByVal0.
	val1AccNum, val1Seq, err := val0.GetClientCtx().AccountRetriever.GetAccountNumberSequence(val0.GetClientCtx(), val1.GetAddress())
	require.NoError(err)

	signedTx, err := authclitestutil.TxSignExec(
		val1.GetClientCtx(),
		val1.GetAddress(),
		signedByVal0File.Name(),
		"--offline",
		fmt.Sprintf("--account-number=%d", val1AccNum),
		fmt.Sprintf("--sequence=%d", val1Seq),
		"--sign-mode=amino-json",
	)
	require.NoError(err)
	signedTxFile := testutil.WriteToNewTempFile(s.T(), signedTx.String())
	defer signedTxFile.Close()

	res, err := authclitestutil.TxBroadcastExec(
		val0.GetClientCtx(),
		signedTxFile.Name(),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
	)
	require.NoError(err)
	require.NoError(s.network.WaitForNextBlock())

	var txRes sdk.TxResponse
	require.NoError(val0.GetClientCtx().Codec.UnmarshalJSON(res.Bytes(), &txRes))
	require.Equal(uint32(0), txRes.Code, txRes.RawLog)

	// Make sure the addr1's balance got funded.
	resp, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", val0.GetAPIAddress(), addr1))
	s.Require().NoError(err)
	var queryRes banktypes.QueryAllBalancesResponse
	err = val0.GetClientCtx().Codec.UnmarshalJSON(resp, &queryRes)
	require.NoError(err)
	require.Equal(sdk.NewCoins(val0Coin, val1Coin), queryRes.Balances)
}

func (s *E2ETestSuite) TestAuxSigner() {
	s.T().Skip("re-enable this when we bring back sign mode aux client testing")
	require := s.Require()
	val := s.network.GetValidators()[0]
	val0Coin := sdk.NewCoin(fmt.Sprintf("%stoken", val.GetMoniker()), math.NewInt(10))

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"error with SIGN_MODE_DIRECT_AUX and --aux unset",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
			},
			true,
		},
		{
			"no error with SIGN_MDOE_DIRECT_AUX mode and generate-only set (ignores generate-only)",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
			},
			false,
		},
		{
			"no error with SIGN_MDOE_DIRECT_AUX mode and generate-only, tip flag set",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
				fmt.Sprintf("--%s=%s", flags.FlagTip, val0Coin.String()),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			_, err := govtestutil.MsgSubmitLegacyProposal(
				val.GetClientCtx(),
				val.GetAddress().String(),
				"test",
				"test desc",
				govtypes.ProposalTypeText,
				tc.args...,
			)
			if tc.expectErr {
				require.Error(err)
			} else {
				require.NoError(err)
			}
		})
	}
}

func (s *E2ETestSuite) TestAuxToFeeWithTips() {
	// Skipping this test as it needs a simapp with the TipDecorator in post handler.
	s.T().Skip()

	require := s.Require()
	val := s.network.GetValidators()[0]

	kb := s.network.GetValidators()[0].GetClientCtx().Keyring
	acc, _, err := kb.NewMnemonic("tipperAccount", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(err)

	tipper, err := acc.GetAddress()
	require.NoError(err)
	tipperInitialBal := sdk.NewCoin(fmt.Sprintf("%stoken", val.GetMoniker()), math.NewInt(10000))

	feePayer := val.GetAddress()
	fee := sdk.NewCoin(s.cfg.BondDenom, math.NewInt(1000))
	tip := sdk.NewCoin(fmt.Sprintf("%stoken", val.GetMoniker()), math.NewInt(1000))

	require.NoError(s.network.WaitForNextBlock())
	_, err = s.createBankMsg(
		val,
		tipper,
		sdk.NewCoins(tipperInitialBal),
		clitestutil.TestTxConfig{},
	)
	require.NoError(err)
	require.NoError(s.network.WaitForNextBlock())

	bal := s.getBalances(val.GetClientCtx(), tipper, tip.Denom)
	require.True(bal.Equal(tipperInitialBal.Amount))

	testCases := []struct {
		name               string
		tipper             sdk.AccAddress
		feePayer           sdk.AccAddress
		tip                sdk.Coin
		expectErrAux       bool
		expectErrBroadCast bool
		errMsg             string
		tipperArgs         []string
		feePayerArgs       []string
	}{
		{
			name:     "when --aux and --sign-mode = direct set: error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      tip,
			tipperArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirect),
				fmt.Sprintf("--%s=%s", flags.FlagTip, tip),
				fmt.Sprintf("--%s=true", flags.FlagAux),
			},
			expectErrAux: true,
			feePayerArgs: []string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, feePayer),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
			},
		},
		{
			name:     "both tipper, fee payer uses AMINO: no error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      tip,
			tipperArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				fmt.Sprintf("--%s=%s", flags.FlagTip, tip),
				fmt.Sprintf("--%s=true", flags.FlagAux),
			},
			feePayerArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, feePayer),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
			},
		},
		{
			name:     "tipper uses DIRECT_AUX, fee payer uses AMINO: no error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      tip,
			tipperArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
				fmt.Sprintf("--%s=%s", flags.FlagTip, tip),
				fmt.Sprintf("--%s=true", flags.FlagAux),
			},
			feePayerArgs: []string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, feePayer),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
			},
		},
		{
			name:     "--tip flag unset: no error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      sdk.Coin{Denom: fmt.Sprintf("%stoken", val.GetMoniker()), Amount: math.NewInt(0)},
			tipperArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
				fmt.Sprintf("--%s=true", flags.FlagAux),
			},
			feePayerArgs: []string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, feePayer),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
			},
		},
		{
			name:     "legacy amino json: no error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      tip,
			tipperArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				fmt.Sprintf("--%s=%s", flags.FlagTip, tip),
				fmt.Sprintf("--%s=true", flags.FlagAux),
			},
			feePayerArgs: []string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, feePayer),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
			},
		},
		{
			name:     "tipper uses direct aux, fee payer uses direct: happy case",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      tip,
			tipperArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
				fmt.Sprintf("--%s=%s", flags.FlagTip, tip),
				fmt.Sprintf("--%s=true", flags.FlagAux),
			},
			feePayerArgs: []string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, feePayer),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
			},
		},
		{
			name:     "chain-id mismatch: error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      tip,
			tipperArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
				fmt.Sprintf("--%s=%s", flags.FlagTip, tip),
				fmt.Sprintf("--%s=true", flags.FlagAux),
			},
			expectErrAux: false,
			feePayerArgs: []string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, feePayer),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagChainID, "foobar"),
			},
			expectErrBroadCast: true,
		},
		{
			name:     "wrong denom in tip: error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      sdk.Coin{Denom: fmt.Sprintf("%stoken", val.GetMoniker()), Amount: math.NewInt(0)},
			tipperArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagTip, "1000wrongDenom"),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
				fmt.Sprintf("--%s=true", flags.FlagAux),
			},
			feePayerArgs: []string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirect),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, feePayer),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
			},
			errMsg: "insufficient funds",
		},
		{
			name:     "insufficient fees: error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      sdk.Coin{Denom: fmt.Sprintf("%stoken", val.GetMoniker()), Amount: math.NewInt(0)},
			tipperArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagTip, tip),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
				fmt.Sprintf("--%s=true", flags.FlagAux),
			},
			feePayerArgs: []string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirect),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, feePayer),
			},
			errMsg: "insufficient fees",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			res, err := govtestutil.MsgSubmitLegacyProposal(
				val.GetClientCtx(),
				tipper.String(),
				"test",
				"test desc",
				govtypes.ProposalTypeText,
				tc.tipperArgs...,
			)

			if tc.expectErrAux {
				require.Error(err)
			} else {
				require.NoError(err)
				genTxFile := testutil.WriteToNewTempFile(s.T(), string(res.Bytes()))
				defer genTxFile.Close()

				s.Require().NoError(s.network.WaitForNextBlock())

				switch {
				case tc.expectErrBroadCast:
					require.Error(err)

				case tc.errMsg != "":
					require.NoError(err)

					var txRes sdk.TxResponse
					require.NoError(val.GetClientCtx().Codec.UnmarshalJSON(res.Bytes(), &txRes))

					require.Contains(txRes.RawLog, tc.errMsg)

				default:
					require.NoError(err)

					var txRes sdk.TxResponse
					require.NoError(val.GetClientCtx().Codec.UnmarshalJSON(res.Bytes(), &txRes))

					require.Equal(uint32(0), txRes.Code)
					require.NotNil(int64(0), txRes.Height)

					bal = s.getBalances(val.GetClientCtx(), tipper, tc.tip.Denom)
					tipperInitialBal = tipperInitialBal.Sub(tc.tip)
					require.True(bal.Equal(tipperInitialBal.Amount))
				}
			}
		})
	}
}

func (s *E2ETestSuite) createBankMsg(val network.ValidatorI, toAddr sdk.AccAddress, amount sdk.Coins, config clitestutil.TestTxConfig) (testutil.BufferWriter, error) {
	msgSend := &banktypes.MsgSend{
		FromAddress: val.GetAddress().String(),
		ToAddress:   toAddr.String(),
		Amount:      amount,
	}

	return clitestutil.SubmitTestTx(val.GetClientCtx(), msgSend, val.GetAddress(), config)
}

func (s *E2ETestSuite) getBalances(clientCtx client.Context, addr sdk.AccAddress, denom string) math.Int {
	resp, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s/by_denom?denom=%s", s.cfg.APIAddress, addr.String(), denom))
	s.Require().NoError(err)

	var balRes banktypes.QueryAllBalancesResponse
	err = clientCtx.Codec.UnmarshalJSON(resp, &balRes)
	s.Require().NoError(err)
	startTokens := balRes.Balances.AmountOf(denom)
	return startTokens
}

func (s *E2ETestSuite) deepContains(events []abci.Event, value string) bool {
	for _, e := range events {
		for _, attr := range e.Attributes {
			if strings.Contains(attr.Value, value) {
				return true
			}
		}
	}
	return false
}
