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

	"cosmossdk.io/depinject"
	"cosmossdk.io/math"

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
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authclitestutil "github.com/cosmos/cosmos-sdk/x/auth/client/testutil"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/testutil"
	bank "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	govtestutil "github.com/cosmos/cosmos-sdk/x/gov/client/testutil"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewE2ETestSuite(cfg network.Config) *E2ETestSuite {
	return &E2ETestSuite{cfg: cfg}
}

func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")
	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	kb := s.network.Validators[0].ClientCtx.Keyring
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
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	s.network.Cleanup()
}

func (s *E2ETestSuite) TestCLISignGenOnly() {
	val := s.network.Validators[0]
	val2 := s.network.Validators[1]

	k, err := val.ClientCtx.Keyring.KeyByAddress(val.Address)
	s.Require().NoError(err)
	keyName := k.Name

	addr, err := k.GetAddress()
	s.Require().NoError(err)

	account, err := val.ClientCtx.AccountRetriever.GetAccount(val.ClientCtx, addr)
	s.Require().NoError(err)

	sendTokens := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10)))
	args := []string{
		keyName, // from keyname
		val2.Address.String(),
		sendTokens.String(),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly), // shouldn't break if we use keyname with --generate-only flag
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
	}
	generatedStd, err := clitestutil.ExecTestCLICmd(val.ClientCtx, bank.NewSendTxCmd(addresscodec.NewBech32Codec("cosmos")), args)
	s.Require().NoError(err)
	opFile := testutil.WriteToNewTempFile(s.T(), generatedStd.String())
	defer opFile.Close()

	commonArgs := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		fmt.Sprintf("--%s=%s", flags.FlagHome, strings.Replace(val.ClientCtx.HomeDir, "simd", "simcli", 1)),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID),
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
		s.Run(tc.name, func() {
			cmd := authcli.GetSignCommand()
			cmd.PersistentFlags().String(flags.FlagHome, val.ClientCtx.HomeDir, "directory for config and data")
			out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, append(tc.args, commonArgs...))
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errMsg)
			} else {
				s.Require().NoError(err)
				func() {
					signedTx := testutil.WriteToNewTempFile(s.T(), out.String())
					defer signedTx.Close()
					_, err := authclitestutil.TxBroadcastExec(val.ClientCtx, signedTx.Name())
					s.Require().NoError(err, out.String())
				}()
			}
		})
	}
}

func (s *E2ETestSuite) TestCLISignBatch() {
	val := s.network.Validators[0]
	sendTokens := sdk.NewCoins(
		sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), math.NewInt(10)),
		sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10)),
	)

	generatedStd, err := s.createBankMsg(val, val.Address,
		sendTokens, fmt.Sprintf("--%s=true", flags.FlagGenerateOnly))
	s.Require().NoError(err)

	outputFile := testutil.WriteToNewTempFile(s.T(), strings.Repeat(generatedStd.String(), 3))
	defer outputFile.Close()
	val.ClientCtx.HomeDir = strings.Replace(val.ClientCtx.HomeDir, "simd", "simcli", 1)

	// sign-batch file - offline is set but account-number and sequence are not
	_, err = authclitestutil.TxSignBatchExec(val.ClientCtx, val.Address, outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--offline")
	s.Require().EqualError(err, "required flag(s) \"account-number\", \"sequence\" not set")

	// sign-batch file - offline and sequence is set but account-number is not set
	_, err = authclitestutil.TxSignBatchExec(val.ClientCtx, val.Address, outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), fmt.Sprintf("--%s=%s", flags.FlagSequence, "1"), "--offline")
	s.Require().EqualError(err, "required flag(s) \"account-number\" not set")

	// sign-batch file - offline and account-number is set but sequence is not set
	_, err = authclitestutil.TxSignBatchExec(val.ClientCtx, val.Address, outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), fmt.Sprintf("--%s=%s", flags.FlagAccountNumber, "1"), "--offline")
	s.Require().EqualError(err, "required flag(s) \"sequence\" not set")

	// sign-batch file - sequence and account-number are set when offline is false
	res, err := authclitestutil.TxSignBatchExec(val.ClientCtx, val.Address, outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), fmt.Sprintf("--%s=%s", flags.FlagSequence, "1"), fmt.Sprintf("--%s=%s", flags.FlagAccountNumber, "1"))
	s.Require().NoError(err)
	s.Require().Equal(3, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))

	// sign-batch file
	res, err = authclitestutil.TxSignBatchExec(val.ClientCtx, val.Address, outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID))
	s.Require().NoError(err)
	s.Require().Equal(3, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))

	// sign-batch file signature only
	res, err = authclitestutil.TxSignBatchExec(val.ClientCtx, val.Address, outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--signature-only")
	s.Require().NoError(err)
	s.Require().Equal(3, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))

	// Sign batch malformed tx file.
	malformedFile := testutil.WriteToNewTempFile(s.T(), fmt.Sprintf("%smalformed", generatedStd))
	defer malformedFile.Close()
	_, err = authclitestutil.TxSignBatchExec(val.ClientCtx, val.Address, malformedFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID))
	s.Require().Error(err)

	// Sign batch malformed tx file signature only.
	_, err = authclitestutil.TxSignBatchExec(val.ClientCtx, val.Address, malformedFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--signature-only")
	s.Require().Error(err)

	// make a txn to increase the sequence of sender
	_, seq, err := val.ClientCtx.AccountRetriever.GetAccountNumberSequence(val.ClientCtx, val.Address)
	s.Require().NoError(err)

	account1, err := val.ClientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	addr, err := account1.GetAddress()
	s.Require().NoError(err)

	// Send coins from validator to multisig.
	_, err = s.createBankMsg(
		val,
		addr,
		sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 1000)),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	// fetch the sequence after a tx, should be incremented.
	_, seq1, err := val.ClientCtx.AccountRetriever.GetAccountNumberSequence(val.ClientCtx, val.Address)
	s.Require().NoError(err)
	s.Require().Equal(seq+1, seq1)

	// signing sign-batch should start from the last sequence.
	signed, err := authclitestutil.TxSignBatchExec(val.ClientCtx, val.Address, outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--signature-only")
	s.Require().NoError(err)
	signedTxs := strings.Split(strings.Trim(signed.String(), "\n"), "\n")
	s.Require().GreaterOrEqual(len(signedTxs), 1)

	sigs, err := s.cfg.TxConfig.UnmarshalSignatureJSON([]byte(signedTxs[0]))
	s.Require().NoError(err)
	s.Require().Equal(sigs[0].Sequence, seq1)
}

func (s *E2ETestSuite) TestCLIQueryTxCmdByHash() {
	val := s.network.Validators[0]

	account2, err := val.ClientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)

	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)

	addr, err := account2.GetAddress()
	s.Require().NoError(err)

	// Send coins.
	out, err := s.createBankMsg(
		val, addr,
		sdk.NewCoins(sendTokens),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	var txRes sdk.TxResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txRes))

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
		s.Run(tc.name, func() {
			cmd := authcli.QueryTxCmd()
			clientCtx := val.ClientCtx
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
				s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
				s.Require().NotNil(result.Height)
				if ok := s.deepContains(result.Events, tc.rawLogContains); !ok {
					s.Require().Fail("raw log does not contain the expected value, expected value: %s", tc.rawLogContains)
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestCLIQueryTxCmdByEvents() {
	val := s.network.Validators[0]

	account2, err := val.ClientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)

	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)

	addr2, err := account2.GetAddress()
	s.Require().NoError(err)

	// Send coins.
	out, err := s.createBankMsg(
		val, addr2,
		sdk.NewCoins(sendTokens),
	)
	s.Require().NoError(err)
	var txRes sdk.TxResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txRes))
	s.Require().NoError(s.network.WaitForNextBlock())

	// Query the tx by hash to get the inner tx.
	err = s.network.RetryForBlocks(func() error {
		out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, authcli.QueryTxCmd(), []string{txRes.TxHash, fmt.Sprintf("--%s=json", flags.FlagOutput)})
		return err
	}, 3)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txRes))
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
				fmt.Sprintf("%s/%d", val.Address, protoTx.AuthInfo.SignerInfos[0].Sequence),
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
		s.Run(tc.name, func() {
			cmd := authcli.QueryTxCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectErrStr)
			} else {
				var result sdk.TxResponse
				s.Require().NoError(err)
				s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
				s.Require().NotNil(result.Height)
			}
		})
	}
}

func (s *E2ETestSuite) TestCLIQueryTxsCmdByEvents() {
	val := s.network.Validators[0]

	account2, err := val.ClientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)

	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)

	addr2, err := account2.GetAddress()
	s.Require().NoError(err)
	// Send coins.
	out, err := s.createBankMsg(
		val,
		addr2,
		sdk.NewCoins(sendTokens),
	)
	s.Require().NoError(err)
	var txRes sdk.TxResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txRes))
	s.Require().NoError(s.network.WaitForNextBlock())

	// Query the tx by hash to get the inner tx.
	err = s.network.RetryForBlocks(func() error {
		out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, authcli.QueryTxCmd(), []string{txRes.TxHash, fmt.Sprintf("--%s=json", flags.FlagOutput)})
		return err
	}, 3)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txRes))

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
		s.Run(tc.name, func() {
			cmd := authcli.QueryTxsByEventsCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result sdk.SearchTxsResult
			s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))

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
	val1 := s.network.Validators[0]

	account, err := val1.ClientCtx.Keyring.Key("newAccount")
	s.Require().NoError(err)

	sendTokens := sdk.NewCoin(s.cfg.BondDenom, sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction))

	addr, err := account.GetAddress()
	s.Require().NoError(err)
	normalGeneratedTx, err := s.createBankMsg(val1, addr,
		sdk.NewCoins(sendTokens), fmt.Sprintf("--%s=true", flags.FlagGenerateOnly))
	s.Require().NoError(err)

	txCfg := val1.ClientCtx.TxConfig

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
		sdk.NewCoins(sendTokens), fmt.Sprintf("--gas=%d", 100),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
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

	resp, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", val1.APIAddress, val1.Address))
	s.Require().NoError(err)

	var balRes banktypes.QueryAllBalancesResponse
	err = val1.ClientCtx.Codec.UnmarshalJSON(resp, &balRes)
	s.Require().NoError(err)
	startTokens := balRes.Balances.AmountOf(s.cfg.BondDenom)

	// Test generate sendTx, estimate gas
	finalGeneratedTx, err := s.createBankMsg(val1, addr,
		sdk.NewCoins(sendTokens), fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly))
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
	res, err := authclitestutil.TxValidateSignaturesExec(val1.ClientCtx, unsignedTxFile.Name())
	s.Require().EqualError(err, "signatures validation failed")
	s.Require().True(strings.Contains(res.String(), fmt.Sprintf("Signers:\n  0: %v\n\nSignatures:\n\n", val1.Address.String())))

	// Test sign

	// Does not work in offline mode
	_, err = authclitestutil.TxSignExec(val1.ClientCtx, val1.Address, unsignedTxFile.Name(), "--offline")
	s.Require().EqualError(err, "required flag(s) \"account-number\", \"sequence\" not set")

	// But works offline if we set account number and sequence
	val1.ClientCtx.HomeDir = strings.Replace(val1.ClientCtx.HomeDir, "simd", "simcli", 1)
	_, err = authclitestutil.TxSignExec(val1.ClientCtx, val1.Address, unsignedTxFile.Name(), "--offline", "--account-number", "1", "--sequence", "1")
	s.Require().NoError(err)

	// Sign transaction
	signedTx, err := authclitestutil.TxSignExec(val1.ClientCtx, val1.Address, unsignedTxFile.Name())
	s.Require().NoError(err)
	signedFinalTx, err := txCfg.TxJSONDecoder()(signedTx.Bytes())
	s.Require().NoError(err)
	txBuilder, err = val1.ClientCtx.TxConfig.WrapTxBuilder(signedFinalTx)
	s.Require().NoError(err)
	s.Require().Equal(len(txBuilder.GetTx().GetMsgs()), 1)
	sigs, err = txBuilder.GetTx().GetSignaturesV2()
	s.Require().NoError(err)
	s.Require().Equal(1, len(sigs))
	signers, err := txBuilder.GetTx().GetSigners()
	s.Require().NoError(err)
	s.Require().Equal([]byte(val1.Address), signers[0])

	// Write the output to disk
	signedTxFile := testutil.WriteToNewTempFile(s.T(), signedTx.String())
	defer signedTxFile.Close()

	// validate Signature
	res, err = authclitestutil.TxValidateSignaturesExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)
	s.Require().True(strings.Contains(res.String(), "[OK]"))
	s.Require().NoError(s.network.WaitForNextBlock())

	// Ensure foo has right amount of funds
	resp, err = testutil.GetRequest(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", val1.APIAddress, val1.Address))
	s.Require().NoError(err)
	err = val1.ClientCtx.Codec.UnmarshalJSON(resp, &balRes)
	s.Require().NoError(err)
	s.Require().Equal(startTokens, balRes.Balances.AmountOf(s.cfg.BondDenom))

	// Test broadcast

	// Does not work in offline mode
	_, err = authclitestutil.TxBroadcastExec(val1.ClientCtx, signedTxFile.Name(), "--offline")
	s.Require().EqualError(err, "cannot broadcast tx during offline mode")
	s.Require().NoError(s.network.WaitForNextBlock())

	// Broadcast correct transaction.
	val1.ClientCtx.BroadcastMode = flags.BroadcastSync
	_, err = authclitestutil.TxBroadcastExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	// Ensure destiny account state
	err = s.network.RetryForBlocks(func() error {
		resp, err = testutil.GetRequest(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", val1.APIAddress, addr))
		s.Require().NoError(err)
		return err
	}, 3)
	s.Require().NoError(err)

	err = val1.ClientCtx.Codec.UnmarshalJSON(resp, &balRes)
	s.Require().NoError(err)
	s.Require().Equal(sendTokens.Amount, balRes.Balances.AmountOf(s.cfg.BondDenom))

	// Ensure origin account state
	resp, err = testutil.GetRequest(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", val1.APIAddress, val1.Address))
	s.Require().NoError(err)
	err = val1.ClientCtx.Codec.UnmarshalJSON(resp, &balRes)
	s.Require().NoError(err)
}

func (s *E2ETestSuite) TestCLIMultisignInsufficientCosigners() {
	val1 := s.network.Validators[0]

	// Fetch account and a multisig info
	account1, err := val1.ClientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	multisigRecord, err := val1.ClientCtx.Keyring.Key("multi")
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
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	// Generate multisig transaction.
	multiGeneratedTx, err := clitestutil.MsgSendExec(
		val1.ClientCtx,
		addr,
		val1.Address,
		sdk.NewCoins(
			sdk.NewInt64Coin(s.cfg.BondDenom, 5),
		),
		addresscodec.NewBech32Codec("cosmos"),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	// Save tx to file
	multiGeneratedTxFile := testutil.WriteToNewTempFile(s.T(), multiGeneratedTx.String())
	defer multiGeneratedTxFile.Close()

	// Multisign, sign with one signature
	val1.ClientCtx.HomeDir = strings.Replace(val1.ClientCtx.HomeDir, "simd", "simcli", 1)
	addr1, err := account1.GetAddress()
	s.Require().NoError(err)
	account1Signature, err := authclitestutil.TxSignExec(val1.ClientCtx, addr1, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	s.Require().NoError(err)

	sign1File := testutil.WriteToNewTempFile(s.T(), account1Signature.String())
	defer sign1File.Close()

	multiSigWith1Signature, err := authclitestutil.TxMultiSignExec(val1.ClientCtx, multisigRecord.Name, multiGeneratedTxFile.Name(), sign1File.Name())
	s.Require().NoError(err)

	// Save tx to file
	multiSigWith1SignatureFile := testutil.WriteToNewTempFile(s.T(), multiSigWith1Signature.String())
	defer multiSigWith1SignatureFile.Close()

	_, err = authclitestutil.TxValidateSignaturesExec(val1.ClientCtx, multiSigWith1SignatureFile.Name())
	s.Require().Error(err)
}

func (s *E2ETestSuite) TestCLIEncode() {
	val1 := s.network.Validators[0]

	sendTokens := sdk.NewCoin(s.cfg.BondDenom, sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction))

	normalGeneratedTx, err := s.createBankMsg(
		val1, val1.Address,
		sdk.NewCoins(sendTokens),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
		fmt.Sprintf("--%s=deadbeef", flags.FlagNote),
	)
	s.Require().NoError(err)
	savedTxFile := testutil.WriteToNewTempFile(s.T(), normalGeneratedTx.String())
	defer savedTxFile.Close()

	// Encode
	encodeExec, err := authclitestutil.TxEncodeExec(val1.ClientCtx, savedTxFile.Name())
	s.Require().NoError(err)
	trimmedBase64 := strings.Trim(encodeExec.String(), "\"\n")

	// Check that the transaction decodes as expected
	decodedTx, err := authclitestutil.TxDecodeExec(val1.ClientCtx, trimmedBase64)
	s.Require().NoError(err)

	txCfg := val1.ClientCtx.TxConfig
	theTx, err := txCfg.TxJSONDecoder()(decodedTx.Bytes())
	s.Require().NoError(err)
	txBuilder, err := val1.ClientCtx.TxConfig.WrapTxBuilder(theTx)
	s.Require().NoError(err)
	s.Require().Equal("deadbeef", txBuilder.GetTx().GetMemo())
}

func (s *E2ETestSuite) TestCLIMultisignSortSignatures() {
	val1 := s.network.Validators[0]

	// Generate 2 accounts and a multisig.
	account1, err := val1.ClientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	account2, err := val1.ClientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)

	multisigRecord, err := val1.ClientCtx.Keyring.Key("multi")
	s.Require().NoError(err)

	// Generate dummy account which is not a part of multisig.
	dummyAcc, err := val1.ClientCtx.Keyring.Key("dummyAccount")
	s.Require().NoError(err)

	addr, err := multisigRecord.GetAddress()
	s.Require().NoError(err)
	resp, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", val1.APIAddress, addr))
	s.Require().NoError(err)

	var balRes banktypes.QueryAllBalancesResponse
	err = val1.ClientCtx.Codec.UnmarshalJSON(resp, &balRes)
	s.Require().NoError(err)
	intialCoins := balRes.Balances

	// Send coins from validator to multisig.
	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)
	_, err = s.createBankMsg(
		val1,
		addr,
		sdk.NewCoins(sendTokens),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	resp, err = testutil.GetRequest(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", val1.APIAddress, addr))
	s.Require().NoError(err)
	err = val1.ClientCtx.Codec.UnmarshalJSON(resp, &balRes)
	s.Require().NoError(err)
	diff, _ := balRes.Balances.SafeSub(intialCoins...)
	s.Require().Equal(sendTokens.Amount, diff.AmountOf(s.cfg.BondDenom))

	// Generate multisig transaction.
	multiGeneratedTx, err := clitestutil.MsgSendExec(
		val1.ClientCtx,
		addr,
		val1.Address,
		sdk.NewCoins(
			sdk.NewInt64Coin(s.cfg.BondDenom, 5),
		),
		addresscodec.NewBech32Codec("cosmos"),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	// Save tx to file
	multiGeneratedTxFile := testutil.WriteToNewTempFile(s.T(), multiGeneratedTx.String())
	defer multiGeneratedTxFile.Close()

	// Sign with account1
	addr1, err := account1.GetAddress()
	s.Require().NoError(err)
	val1.ClientCtx.HomeDir = strings.Replace(val1.ClientCtx.HomeDir, "simd", "simcli", 1)
	account1Signature, err := authclitestutil.TxSignExec(val1.ClientCtx, addr1, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	s.Require().NoError(err)

	sign1File := testutil.WriteToNewTempFile(s.T(), account1Signature.String())
	defer sign1File.Close()

	// Sign with account2
	addr2, err := account2.GetAddress()
	s.Require().NoError(err)
	account2Signature, err := authclitestutil.TxSignExec(val1.ClientCtx, addr2, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	s.Require().NoError(err)

	sign2File := testutil.WriteToNewTempFile(s.T(), account2Signature.String())
	defer sign2File.Close()

	// Sign with dummy account
	dummyAddr, err := dummyAcc.GetAddress()
	s.Require().NoError(err)
	_, err = authclitestutil.TxSignExec(val1.ClientCtx, dummyAddr, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "signing key is not a part of multisig key")

	multiSigWith2Signatures, err := authclitestutil.TxMultiSignExec(val1.ClientCtx, multisigRecord.Name, multiGeneratedTxFile.Name(), sign1File.Name(), sign2File.Name())
	s.Require().NoError(err)

	// Write the output to disk
	signedTxFile := testutil.WriteToNewTempFile(s.T(), multiSigWith2Signatures.String())
	defer signedTxFile.Close()

	_, err = authclitestutil.TxValidateSignaturesExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	val1.ClientCtx.BroadcastMode = flags.BroadcastSync
	_, err = authclitestutil.TxBroadcastExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *E2ETestSuite) TestSignWithMultisig() {
	val1 := s.network.Validators[0]

	// Generate a account for signing.
	account1, err := val1.ClientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	addr1, err := account1.GetAddress()
	s.Require().NoError(err)

	// Create an address that is not in the keyring, will be used to simulate `--multisig`
	multisig := "cosmos1hd6fsrvnz6qkp87s3u86ludegq97agxsdkwzyh"
	multisigAddr, err := sdk.AccAddressFromBech32(multisig)
	s.Require().NoError(err)

	// Generate a transaction for testing --multisig with an address not in the keyring.
	multisigTx, err := clitestutil.MsgSendExec(
		val1.ClientCtx,
		val1.Address,
		val1.Address,
		sdk.NewCoins(
			sdk.NewInt64Coin(s.cfg.BondDenom, 5),
		),
		addresscodec.NewBech32Codec("cosmos"),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	// Save multi tx to file
	multiGeneratedTx2File := testutil.WriteToNewTempFile(s.T(), multisigTx.String())
	defer multiGeneratedTx2File.Close()

	// Sign using multisig. We're signing a tx on behalf of the multisig address,
	// even though the tx signer is NOT the multisig address. This is fine though,
	// as the main point of this test is to test the `--multisig` flag with an address
	// that is not in the keyring.
	_, err = authclitestutil.TxSignExec(val1.ClientCtx, addr1, multiGeneratedTx2File.Name(), "--multisig", multisigAddr.String())
	s.Require().Contains(err.Error(), "error getting account from keybase")
}

func (s *E2ETestSuite) TestCLIMultisign() {
	val1 := s.network.Validators[0]

	// Generate 2 accounts and a multisig.
	account1, err := val1.ClientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	account2, err := val1.ClientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)

	multisigRecord, err := val1.ClientCtx.Keyring.Key("multi")
	s.Require().NoError(err)

	addr, err := multisigRecord.GetAddress()
	s.Require().NoError(err)

	// Send coins from validator to multisig.
	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)
	s.Require().NoError(s.network.WaitForNextBlock())
	_, err = s.createBankMsg(
		val1, addr,
		sdk.NewCoins(sendTokens),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	var balRes banktypes.QueryAllBalancesResponse
	err = s.network.RetryForBlocks(func() error {
		resp, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", val1.APIAddress, addr))
		if err != nil {
			return err
		}
		return val1.ClientCtx.Codec.UnmarshalJSON(resp, &balRes)
	}, 3)
	s.Require().NoError(err)
	s.Require().True(sendTokens.Amount.Equal(balRes.Balances.AmountOf(s.cfg.BondDenom)))

	// Generate multisig transaction.
	multiGeneratedTx, err := clitestutil.MsgSendExec(
		val1.ClientCtx,
		addr,
		val1.Address,
		sdk.NewCoins(
			sdk.NewInt64Coin(s.cfg.BondDenom, 5),
		),
		addresscodec.NewBech32Codec("cosmos"),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	// Save tx to file
	multiGeneratedTxFile := testutil.WriteToNewTempFile(s.T(), multiGeneratedTx.String())
	defer multiGeneratedTxFile.Close()

	addr1, err := account1.GetAddress()
	s.Require().NoError(err)
	// Sign with account1
	val1.ClientCtx.HomeDir = strings.Replace(val1.ClientCtx.HomeDir, "simd", "simcli", 1)
	account1Signature, err := authclitestutil.TxSignExec(val1.ClientCtx, addr1, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	s.Require().NoError(err)

	sign1File := testutil.WriteToNewTempFile(s.T(), account1Signature.String())
	defer sign1File.Close()

	addr2, err := account2.GetAddress()
	s.Require().NoError(err)
	// Sign with account2
	account2Signature, err := authclitestutil.TxSignExec(val1.ClientCtx, addr2, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	s.Require().NoError(err)

	sign2File := testutil.WriteToNewTempFile(s.T(), account2Signature.String())
	defer sign2File.Close()

	// Work in offline mode.
	multisigAccNum, multisigSeq, err := val1.ClientCtx.AccountRetriever.GetAccountNumberSequence(val1.ClientCtx, addr)
	s.Require().NoError(err)
	_, err = authclitestutil.TxMultiSignExec(
		val1.ClientCtx,
		multisigRecord.Name,
		multiGeneratedTxFile.Name(),
		fmt.Sprintf("--%s", flags.FlagOffline),
		fmt.Sprintf("--%s=%d", flags.FlagAccountNumber, multisigAccNum),
		fmt.Sprintf("--%s=%d", flags.FlagSequence, multisigSeq),
		sign1File.Name(),
		sign2File.Name(),
	)
	s.Require().NoError(err)

	val1.ClientCtx.Offline = false
	multiSigWith2Signatures, err := authclitestutil.TxMultiSignExec(val1.ClientCtx, multisigRecord.Name, multiGeneratedTxFile.Name(), sign1File.Name(), sign2File.Name())
	s.Require().NoError(err)

	// Write the output to disk
	signedTxFile := testutil.WriteToNewTempFile(s.T(), multiSigWith2Signatures.String())
	defer signedTxFile.Close()

	_, err = authclitestutil.TxValidateSignaturesExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	val1.ClientCtx.BroadcastMode = flags.BroadcastSync
	_, err = authclitestutil.TxBroadcastExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *E2ETestSuite) TestSignBatchMultisig() {
	val := s.network.Validators[0]

	// Fetch 2 accounts and a multisig.
	account1, err := val.ClientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)
	account2, err := val.ClientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)
	multisigRecord, err := val.ClientCtx.Keyring.Key("multi")
	s.Require().NoError(err)

	addr, err := multisigRecord.GetAddress()
	s.Require().NoError(err)

	// val may not have processed last tx yet
	s.Require().NoError(s.network.WaitForNextBlock())
	s.Require().NoError(s.network.WaitForNextBlock())

	// Send coins from validator to multisig.
	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)
	_, err = s.createBankMsg(
		val,
		addr,
		sdk.NewCoins(sendTokens),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	generatedStd, err := clitestutil.MsgSendExec(
		val.ClientCtx,
		addr,
		val.Address,
		sdk.NewCoins(
			sdk.NewCoin(s.cfg.BondDenom, math.NewInt(1)),
		),
		addresscodec.NewBech32Codec("cosmos"),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	// Write the output to disk
	filename := testutil.WriteToNewTempFile(s.T(), strings.Repeat(generatedStd.String(), 1))
	defer filename.Close()
	val.ClientCtx.HomeDir = strings.Replace(val.ClientCtx.HomeDir, "simd", "simcli", 1)

	addr1, err := account1.GetAddress()
	s.Require().NoError(err)
	// sign-batch file
	res, err := authclitestutil.TxSignBatchExec(val.ClientCtx, addr1, filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--multisig", addr.String(), "--signature-only")
	s.Require().NoError(err)
	s.Require().Equal(1, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))
	// write sigs to file
	file1 := testutil.WriteToNewTempFile(s.T(), res.String())
	defer file1.Close()

	addr2, err := account2.GetAddress()
	s.Require().NoError(err)
	// sign-batch file with account2
	res, err = authclitestutil.TxSignBatchExec(val.ClientCtx, addr2, filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--multisig", addr.String(), "--signature-only")
	s.Require().NoError(err)
	s.Require().Equal(1, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))
	// write sigs to file2
	file2 := testutil.WriteToNewTempFile(s.T(), res.String())
	defer file2.Close()
	_, err = authclitestutil.TxMultiSignExec(val.ClientCtx, multisigRecord.Name, filename.Name(), file1.Name(), file2.Name())
	s.Require().NoError(err)
}

func (s *E2ETestSuite) TestMultisignBatch() {
	val := s.network.Validators[0]

	// Fetch 2 accounts and a multisig.
	account1, err := val.ClientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)
	account2, err := val.ClientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)
	multisigRecord, err := val.ClientCtx.Keyring.Key("multi")
	s.Require().NoError(err)

	addr, err := multisigRecord.GetAddress()
	s.Require().NoError(err)

	// val may not have processed last tx yet
	s.Require().NoError(s.network.WaitForNextBlock())
	s.Require().NoError(s.network.WaitForNextBlock())

	// Send coins from validator to multisig.
	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 1000)
	_, err = s.createBankMsg(
		val,
		addr,
		sdk.NewCoins(sendTokens),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	generatedStd, err := clitestutil.MsgSendExec(
		val.ClientCtx,
		addr,
		val.Address,
		sdk.NewCoins(
			sdk.NewCoin(s.cfg.BondDenom, math.NewInt(1)),
		),
		addresscodec.NewBech32Codec("cosmos"),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	// Write the output to disk
	filename := testutil.WriteToNewTempFile(s.T(), strings.Repeat(generatedStd.String(), 3))
	defer filename.Close()
	val.ClientCtx.HomeDir = strings.Replace(val.ClientCtx.HomeDir, "simd", "simcli", 1)

	account, err := val.ClientCtx.AccountRetriever.GetAccount(val.ClientCtx, addr)
	s.Require().NoError(err)

	// sign-batch file
	addr1, err := account1.GetAddress()
	s.Require().NoError(err)
	res, err := authclitestutil.TxSignBatchExec(val.ClientCtx, addr1, filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--multisig", addr.String(), fmt.Sprintf("--%s", flags.FlagOffline), fmt.Sprintf("--%s=%s", flags.FlagAccountNumber, fmt.Sprint(account.GetAccountNumber())), fmt.Sprintf("--%s=%s", flags.FlagSequence, fmt.Sprint(account.GetSequence())), "--signature-only")
	s.Require().NoError(err)
	s.Require().Equal(3, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))
	// write sigs to file
	file1 := testutil.WriteToNewTempFile(s.T(), res.String())
	defer file1.Close()

	// sign-batch file with account2
	addr2, err := account2.GetAddress()
	s.Require().NoError(err)
	res, err = authclitestutil.TxSignBatchExec(val.ClientCtx, addr2, filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--multisig", addr.String(), fmt.Sprintf("--%s", flags.FlagOffline), fmt.Sprintf("--%s=%s", flags.FlagAccountNumber, fmt.Sprint(account.GetAccountNumber())), fmt.Sprintf("--%s=%s", flags.FlagSequence, fmt.Sprint(account.GetSequence())), "--signature-only")
	s.Require().NoError(err)
	s.Require().Equal(3, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))

	// multisign the file
	file2 := testutil.WriteToNewTempFile(s.T(), res.String())
	defer file2.Close()
	res, err = authclitestutil.TxMultiSignBatchExec(val.ClientCtx, filename.Name(), multisigRecord.Name, file1.Name(), file2.Name())
	s.Require().NoError(err)
	signedTxs := strings.Split(strings.Trim(res.String(), "\n"), "\n")

	// Broadcast transactions.
	for _, signedTx := range signedTxs {
		func() {
			signedTxFile := testutil.WriteToNewTempFile(s.T(), signedTx)
			defer signedTxFile.Close()
			val.ClientCtx.BroadcastMode = flags.BroadcastSync
			_, err = authclitestutil.TxBroadcastExec(val.ClientCtx, signedTxFile.Name())
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
	from, err := sdk.AccAddressFromBech32("cosmos1cxlt8kznps92fwu3j6npahx4mjfutydyene2qw")
	require.NoError(t, err)
	to, err := sdk.AccAddressFromBech32("cosmos1cxlt8kznps92fwu3j6npahx4mjfutydyene2qw")
	require.NoError(t, err)
	err = builder.SetMsgs(banktypes.NewMsgSend(from, to, sdk.Coins{sdk.NewInt64Coin("stake", 10000)}))
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
// public key causes an error - node will panic but recover.
// See https://github.com/cosmos/cosmos-sdk/issues/7585 for more details.
func (s *E2ETestSuite) TestTxWithoutPublicKey() {
	val1 := s.network.Validators[0]
	txCfg := val1.ClientCtx.TxConfig

	// Create a txBuilder with an unsigned tx.
	txBuilder := txCfg.NewTxBuilder()
	msg := banktypes.NewMsgSend(val1.Address, val1.Address, sdk.NewCoins(
		sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10)),
	))
	err := txBuilder.SetMsgs(msg)
	s.Require().NoError(err)
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(150))))
	txBuilder.SetGasLimit(testdata.NewTestGasLimit())
	// Set empty signature to set signer infos.
	sigV2 := signing.SignatureV2{
		PubKey: val1.PubKey,
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
	signedTx, err := authclitestutil.TxSignExec(val1.ClientCtx, val1.Address, unsignedTxFile.Name(), fmt.Sprintf("--%s=true", cli.FlagOverwrite))
	s.Require().NoError(err)

	// Remove the signerInfo's `public_key` field manually from the signedTx.
	// Note: this method is only used for test purposes! In general, one should
	// use txBuilder and TxEncoder/TxDecoder to manipulate txs.
	var tx tx.Tx
	err = val1.ClientCtx.Codec.UnmarshalJSON(signedTx.Bytes(), &tx)
	s.Require().NoError(err)
	tx.AuthInfo.SignerInfos[0].PublicKey = nil
	// Re-encode the tx again, to another file.
	txJSON, err = val1.ClientCtx.Codec.MarshalJSON(&tx)
	s.Require().NoError(err)
	signedTxFile := testutil.WriteToNewTempFile(s.T(), string(txJSON))
	defer signedTxFile.Close()
	s.Require().True(strings.Contains(string(txJSON), "\"public_key\":null"))

	// val may not have processed last tx yet
	s.Require().NoError(s.network.WaitForNextBlock())
	s.Require().NoError(s.network.WaitForNextBlock())

	// Broadcast tx, test that it should panic internally, recover and error.
	val1.ClientCtx.BroadcastMode = flags.BroadcastSync
	_, err = authclitestutil.TxBroadcastExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().Error(err)
}

// TestSignWithMultiSignersAminoJSON tests the case where a transaction with 2
// messages which has to be signed with 2 different keys. Sign and append the
// signatures using the CLI with Amino signing mode. Finally, send the
// transaction to the blockchain.
func (s *E2ETestSuite) TestSignWithMultiSignersAminoJSON() {
	require := s.Require()
	val0, val1 := s.network.Validators[0], s.network.Validators[1]
	val0Coin := sdk.NewCoin(fmt.Sprintf("%stoken", val0.Moniker), math.NewInt(10))
	val1Coin := sdk.NewCoin(fmt.Sprintf("%stoken", val1.Moniker), math.NewInt(10))
	_, _, addr1 := testdata.KeyTestPubAddr()

	// Creating a tx with 2 msgs from 2 signers: val0 and val1.
	// The validators need to sign with SIGN_MODE_LEGACY_AMINO_JSON,
	// because DIRECT doesn't support multi signers via the CLI.
	// Since we use amino, we don't need to pre-populate signer_infos.
	txBuilder := val0.ClientCtx.TxConfig.NewTxBuilder()
	require.NoError(txBuilder.SetMsgs(
		banktypes.NewMsgSend(val0.Address, addr1, sdk.NewCoins(val0Coin)),
		banktypes.NewMsgSend(val1.Address, addr1, sdk.NewCoins(val1Coin)),
	))
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))))
	txBuilder.SetGasLimit(testdata.NewTestGasLimit() * 2)
	signers, err := txBuilder.GetTx().GetSigners()
	require.NoError(err)
	require.Equal([][]byte{val0.Address, val1.Address}, signers)

	// Write the unsigned tx into a file.
	txJSON, err := val0.ClientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	require.NoError(err)
	unsignedTxFile := testutil.WriteToNewTempFile(s.T(), string(txJSON))
	defer unsignedTxFile.Close()

	// Let val0 sign first the file with the unsignedTx.
	signedByVal0, err := authclitestutil.TxSignExec(val0.ClientCtx, val0.Address, unsignedTxFile.Name(), "--overwrite", "--sign-mode=amino-json")
	require.NoError(err)
	signedByVal0File := testutil.WriteToNewTempFile(s.T(), signedByVal0.String())
	defer signedByVal0File.Close()

	// Then let val1 sign the file with signedByVal0.
	val1AccNum, val1Seq, err := val0.ClientCtx.AccountRetriever.GetAccountNumberSequence(val0.ClientCtx, val1.Address)
	require.NoError(err)

	signedTx, err := authclitestutil.TxSignExec(
		val1.ClientCtx,
		val1.Address,
		signedByVal0File.Name(),
		"--offline",
		fmt.Sprintf("--account-number=%d", val1AccNum),
		fmt.Sprintf("--sequence=%d", val1Seq),
		"--sign-mode=amino-json",
	)
	require.NoError(err)
	signedTxFile := testutil.WriteToNewTempFile(s.T(), signedTx.String())
	defer signedTxFile.Close()

	// val may not have processed last tx yet
	s.Require().NoError(s.network.WaitForNextBlock())
	s.Require().NoError(s.network.WaitForNextBlock())

	res, err := authclitestutil.TxBroadcastExec(
		val0.ClientCtx,
		signedTxFile.Name(),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
	)
	require.NoError(err)
	require.NoError(s.network.WaitForNextBlock())

	var txRes sdk.TxResponse
	require.NoError(val0.ClientCtx.Codec.UnmarshalJSON(res.Bytes(), &txRes))
	require.Equal(uint32(0), txRes.Code, txRes.RawLog)

	// Make sure the addr1's balance got funded.
	resp, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", val0.APIAddress, addr1))
	s.Require().NoError(err)
	var queryRes banktypes.QueryAllBalancesResponse
	err = val0.ClientCtx.Codec.UnmarshalJSON(resp, &queryRes)
	require.NoError(err)
	require.Equal(sdk.NewCoins(val0Coin, val1Coin), queryRes.Balances)
}

func (s *E2ETestSuite) TestAuxSigner() {
	require := s.Require()
	val := s.network.Validators[0]
	val0Coin := sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), math.NewInt(10))

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
		s.Run(tc.name, func() {
			_, err := govtestutil.MsgSubmitLegacyProposal(
				val.ClientCtx,
				val.Address.String(),
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
	val := s.network.Validators[0]

	kb := s.network.Validators[0].ClientCtx.Keyring
	acc, _, err := kb.NewMnemonic("tipperAccount", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(err)

	tipper, err := acc.GetAddress()
	require.NoError(err)
	tipperInitialBal := sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), math.NewInt(10000))

	feePayer := val.Address
	fee := sdk.NewCoin(s.cfg.BondDenom, math.NewInt(1000))
	tip := sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), math.NewInt(1000))

	require.NoError(s.network.WaitForNextBlock())
	_, err = s.createBankMsg(val, tipper, sdk.NewCoins(tipperInitialBal))
	require.NoError(err)
	require.NoError(s.network.WaitForNextBlock())

	bal := s.getBalances(val.ClientCtx, tipper, tip.Denom)
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
			tip:      sdk.Coin{Denom: fmt.Sprintf("%stoken", val.Moniker), Amount: math.NewInt(0)},
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
			tip:      sdk.Coin{Denom: fmt.Sprintf("%stoken", val.Moniker), Amount: math.NewInt(0)},
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
			tip:      sdk.Coin{Denom: fmt.Sprintf("%stoken", val.Moniker), Amount: math.NewInt(0)},
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
		s.Run(tc.name, func() {
			res, err := govtestutil.MsgSubmitLegacyProposal(
				val.ClientCtx,
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
					require.NoError(val.ClientCtx.Codec.UnmarshalJSON(res.Bytes(), &txRes))

					require.Contains(txRes.RawLog, tc.errMsg)

				default:
					require.NoError(err)

					var txRes sdk.TxResponse
					require.NoError(val.ClientCtx.Codec.UnmarshalJSON(res.Bytes(), &txRes))

					require.Equal(uint32(0), txRes.Code)
					require.NotNil(int64(0), txRes.Height)

					bal = s.getBalances(val.ClientCtx, tipper, tc.tip.Denom)
					tipperInitialBal = tipperInitialBal.Sub(tc.tip)
					require.True(bal.Equal(tipperInitialBal.Amount))
				}
			}
		})
	}
}

func (s *E2ETestSuite) createBankMsg(val *network.Validator, toAddr sdk.AccAddress, amount sdk.Coins, extraFlags ...string) (testutil.BufferWriter, error) {
	flags := []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees,
			sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
	}

	flags = append(flags, extraFlags...)
	return clitestutil.MsgSendExec(val.ClientCtx, val.Address, toAddr, amount, addresscodec.NewBech32Codec("cosmos"), flags...)
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
