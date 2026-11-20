package cli_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// rekeyNetwork is a one-validator network with PubKeyChangeEnabled set.
type rekeyNetwork struct {
	net       *network.Network
	cfg       network.Config
	val       *network.Validator
	clientCtx client.Context
	fees      sdk.Coins
	// txFlags broadcast a tx in sync mode, paying fees for 2M gas.
	txFlags []string
}

func startRekeyNetwork(t *testing.T) rekeyNetwork {
	t.Helper()
	cfg, err := network.DefaultConfigWithAppConfig(authtestutil.AppConfig)
	require.NoError(t, err)
	cfg.NumValidators = 1

	var authGen authtypes.GenesisState
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[authtypes.ModuleName], &authGen))
	authGen.Params.PubKeyChangeEnabled = true
	cfg.GenesisState[authtypes.ModuleName], err = cfg.Codec.MarshalJSON(&authGen)
	require.NoError(t, err)

	net, err := network.New(t, t.TempDir(), cfg)
	require.NoError(t, err)
	t.Cleanup(net.Cleanup)
	_, err = net.WaitForHeight(1)
	require.NoError(t, err)

	val := net.Validators[0]
	fees := sdk.NewCoins(sdk.NewCoin(cfg.BondDenom, math.NewInt(100)))
	return rekeyNetwork{
		net:       net,
		cfg:       cfg,
		val:       val,
		clientCtx: val.ClientCtx,
		fees:      fees,
		txFlags: []string{
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%s", flags.FlagFees, fees.String()),
			fmt.Sprintf("--%s=%d", flags.FlagGas, 2_000_000),
		},
	}
}

// requireTxOK checks that a broadcast tx passed CheckTx and was included in a
// block with code 0.
func (rn rekeyNetwork) requireTxOK(t *testing.T, out testutil.BufferWriter) {
	t.Helper()
	var res sdk.TxResponse
	require.NoError(t, rn.clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res), out.String())
	require.Equal(t, uint32(0), res.Code, res.RawLog)
	require.NoError(t, clitestutil.CheckTxCode(rn.net, rn.clientCtx, res.TxHash, 0))
}

// TestChangePubKeyCmd_Multisig rotates a secp256k1 account to a 2-of-3
// ML-DSA-65 multisig with the change-pubkey CLI, using two sign-rekey-proof
// outputs as the proof, then spends from the account with the multisig via
// --signer-address.
func TestChangePubKeyCmd_Multisig(t *testing.T) {
	rn := startRekeyNetwork(t)
	cfg, val, clientCtx, txFlags := rn.cfg, rn.val, rn.clientCtx, rn.txFlags
	kr := clientCtx.Keyring
	ac := clientCtx.TxConfig.SigningContext().AddressCodec()

	// The account to rekey starts with a secp256k1 key.
	oldRec, _, err := kr.NewMnemonic("rekey-old", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)
	account, err := oldRec.GetAddress()
	require.NoError(t, err)

	// The new key is a 2-of-3 multisig of ML-DSA-65 keys.
	memberNames := []string{"rekey-m1", "rekey-m2", "rekey-m3"}
	members := make([]cryptotypes.PubKey, len(memberNames))
	for i, name := range memberNames {
		rec, _, err := kr.NewMnemonic(name, keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.MlDsa65)
		require.NoError(t, err)
		members[i], err = rec.GetPubKey()
		require.NoError(t, err)
	}
	msigPk := kmultisig.NewLegacyAminoPubKey(2, members)
	_, err = kr.SaveMultisig("rekey-msig", msigPk)
	require.NoError(t, err)
	msigAddr := sdk.AccAddress(msigPk.Address())
	require.NotEqual(t, account, msigAddr)

	// Fund the account, which also stores its secp256k1 pubkey on first use.
	out, err := clitestutil.MsgSendExec(clientCtx, val.Address, account,
		sdk.NewCoins(sdk.NewInt64Coin(cfg.BondDenom, 1_000_000)), ac, txFlags...)
	require.NoError(t, err)
	rn.requireTxOK(t, out)

	pkJSON, err := clientCtx.Codec.MarshalInterfaceJSON(msigPk)
	require.NoError(t, err)

	// Two members each sign a proof of possession.
	proofFiles := make([]string, 0, 2)
	for _, name := range memberNames[:2] {
		out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.NewSignRekeyProofCmd(), []string{
			account.String(), string(pkJSON),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, name),
		})
		require.NoError(t, err)
		proofFiles = append(proofFiles, testutil.WriteToNewTempFile(t, out.String()).Name())
	}

	// A proof signed by a key outside the multisig is refused.
	_, err = clitestutil.ExecTestCLICmd(clientCtx, cli.NewSignRekeyProofCmd(), []string{
		account.String(), string(pkJSON),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, "rekey-old"),
	})
	require.ErrorContains(t, err, "is not the new pubkey or a member of it")

	// The current secp256k1 key sends MsgChangePubKey.
	out, err = clitestutil.ExecTestCLICmd(clientCtx, cli.NewChangePubKeyCmd(), append([]string{
		account.String(), string(pkJSON),
		fmt.Sprintf("--proof=%s", strings.Join(proofFiles, ",")),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, "rekey-old"),
	}, txFlags...))
	require.NoError(t, err)
	rn.requireTxOK(t, out)

	acc, err := clientCtx.AccountRetriever.GetAccount(clientCtx, account)
	require.NoError(t, err)
	require.True(t, msigPk.Equals(acc.GetPubKey()), "account pubkey was not changed")

	rekeyed, err := authtypes.NewQueryClient(clientCtx).RekeyedAccounts(context.Background(),
		&authtypes.QueryRekeyedAccountsRequest{Address: msigAddr.String()})
	require.NoError(t, err)
	require.Equal(t, []string{account.String()}, rekeyed.Addresses)

	// Spend from the rekeyed account with the multisig: generate, have two
	// members sign, combine, broadcast, with both multisign and
	// multisign-batch. Only amino-json is covered: as for any multisig, the
	// direct sign doc commits to the AuthInfo, which differs between what a
	// member signs and the combined tx.
	recipient := sdk.AccAddress("rekey_test_recipient")
	sendAmt := sdk.NewInt64Coin(cfg.BondDenom, 1234)
	signerFlag := fmt.Sprintf("--%s=%s", flags.FlagSignerAddress, account.String())
	chainFlag := fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID)

	wantBal := sdk.NewInt64Coin(cfg.BondDenom, 0)
	for _, signMode := range []string{flags.SignModeLegacyAminoJSON} {
		for _, batch := range []bool{false, true} {
			name := fmt.Sprintf("%s/batch=%t", signMode, batch)
			signModeFlag := fmt.Sprintf("--%s=%s", flags.FlagSignMode, signMode)

			// A batch holds two txs, so multisign-batch has to give them
			// consecutive sequences of the rekeyed account.
			numTxs := 1
			if batch {
				numTxs = 2
			}
			var unsigned strings.Builder
			for range numTxs {
				out, err := clitestutil.ExecTestCLICmd(clientCtx, bankcli.NewSendTxCmd(ac), append([]string{
					"rekey-msig", recipient.String(), sendAmt.String(),
					signerFlag,
					fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
				}, txFlags...))
				require.NoError(t, err, name)
				unsigned.WriteString(out.String())
			}
			unsignedFile := testutil.WriteToNewTempFile(t, unsigned.String()).Name()

			signCmd, multisignCmd := cli.GetSignCommand, cli.GetMultiSignCommand
			if batch {
				signCmd, multisignCmd = cli.GetSignBatchCommand, cli.GetMultiSignBatchCmd
			}

			signArgs := []string{
				unsignedFile,
				"--multisig=rekey-msig",
				"--signature-only",
				signerFlag,
				chainFlag,
				signModeFlag,
			}
			if batch {
				// Online, sign-batch --multisig reads the sequence from
				// state for every tx, as for any multisig, so sign the
				// batch offline from the account's current sequence.
				accNum, seq, err := clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, account)
				require.NoError(t, err, name)
				signArgs = append(signArgs,
					fmt.Sprintf("--%s=true", flags.FlagOffline),
					fmt.Sprintf("--%s=%d", flags.FlagAccountNumber, accNum),
					fmt.Sprintf("--%s=%d", flags.FlagSequence, seq),
				)
			}

			sigFiles := make([]string, 0, 2)
			for _, member := range memberNames[1:] {
				out, err := clitestutil.ExecTestCLICmd(clientCtx, signCmd(), append([]string{
					fmt.Sprintf("--%s=%s", flags.FlagFrom, member),
				}, signArgs...))
				require.NoError(t, err, name)
				sigFiles = append(sigFiles, testutil.WriteToNewTempFile(t, out.String()).Name())
			}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, multisignCmd(), append([]string{
				unsignedFile, "rekey-msig",
			}, append(sigFiles, signerFlag, chainFlag, signModeFlag)...))
			require.NoError(t, err, name)
			signedTxs := strings.Split(strings.TrimSpace(out.String()), "\n")
			require.Len(t, signedTxs, numTxs, name)

			for _, signedTx := range signedTxs {
				signedFile := testutil.WriteToNewTempFile(t, signedTx).Name()

				// validate-signatures accepts the multisig signature for the
				// rekeyed account, checking the pubkey against the account's
				// stored one.
				out, err = clitestutil.ExecTestCLICmd(clientCtx, cli.GetValidateSignaturesCommand(), []string{
					signedFile, chainFlag, signModeFlag,
				})
				require.NoError(t, err, "%s: %s", name, out.String())
				require.NotContains(t, out.String(), "ERROR", name)
				require.Contains(t, out.String(), account.String(), name)

				out, err = clitestutil.ExecTestCLICmd(clientCtx, cli.GetBroadcastCommand(), []string{
					signedFile,
					fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				})
				require.NoError(t, err, name)
				rn.requireTxOK(t, out)

				wantBal = wantBal.Add(sendAmt)
			}
			bal, err := banktypes.NewQueryClient(clientCtx).Balance(context.Background(),
				&banktypes.QueryBalanceRequest{Address: recipient.String(), Denom: cfg.BondDenom})
			require.NoError(t, err, name)
			require.Equal(t, wantBal, *bal.Balance, name)
		}
	}
}

// TestChangePubKeyCmd_SignerMismatch checks that change-pubkey refuses to
// build a msg for an account other than the one the --from key signs for.
func TestChangePubKeyCmd_SignerMismatch(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{})
	kr := keyring.NewInMemory(encCfg.Codec)
	rec, _, err := kr.NewMnemonic("k", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)
	newPk, err := rec.GetPubKey()
	require.NoError(t, err)
	pkJSON, err := encCfg.Codec.MarshalInterfaceJSON(newPk)
	require.NoError(t, err)

	clientCtx := client.Context{}.
		WithKeyring(kr).
		WithCodec(encCfg.Codec).
		WithInterfaceRegistry(encCfg.InterfaceRegistry).
		WithTxConfig(encCfg.TxConfig).
		WithChainID("test-chain")

	other := sdk.AccAddress("some_other_account__")
	proof := testutil.WriteToNewTempFile(t, "{}").Name()
	_, err = clitestutil.ExecTestCLICmd(clientCtx, cli.NewChangePubKeyCmd(), []string{
		other.String(), string(pkJSON),
		"--proof=" + proof,
		"--from=k",
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	})
	require.ErrorContains(t, err, "--signer-address")
}

// TestChangePubKeyCmd_ProofBelowThreshold checks that change-pubkey refuses a
// multisig proof with fewer signatures than the new key's threshold, instead
// of sending a tx that pays gas and then fails on chain.
func TestChangePubKeyCmd_ProofBelowThreshold(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{})
	kr := keyring.NewInMemory(encCfg.Codec)
	oldRec, _, err := kr.NewMnemonic("old", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)
	account, err := oldRec.GetAddress()
	require.NoError(t, err)

	memberNames := []string{"m1", "m2", "m3"}
	members := make([]cryptotypes.PubKey, len(memberNames))
	for i, name := range memberNames {
		rec, _, err := kr.NewMnemonic(name, keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		require.NoError(t, err)
		members[i], err = rec.GetPubKey()
		require.NoError(t, err)
	}
	msigPk := kmultisig.NewLegacyAminoPubKey(2, members)
	pkJSON, err := encCfg.Codec.MarshalInterfaceJSON(msigPk)
	require.NoError(t, err)

	clientCtx := client.Context{}.
		WithKeyring(kr).
		WithCodec(encCfg.Codec).
		WithInterfaceRegistry(encCfg.InterfaceRegistry).
		WithTxConfig(encCfg.TxConfig).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithChainID("test-chain")

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.NewSignRekeyProofCmd(), []string{
		account.String(), string(pkJSON),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, memberNames[0]),
		fmt.Sprintf("--%s=true", flags.FlagOffline),
		fmt.Sprintf("--%s=0", flags.FlagAccountNumber),
	})
	require.NoError(t, err)
	proof := testutil.WriteToNewTempFile(t, out.String()).Name()

	_, err = clitestutil.ExecTestCLICmd(clientCtx, cli.NewChangePubKeyCmd(), []string{
		account.String(), string(pkJSON),
		"--proof=" + proof,
		"--from=old",
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	})
	require.ErrorContains(t, err, "threshold")
}

// TestMultisign_SignerAddress checks that multisign and multisign-batch
// verify member signatures for the --signer-address account, not for the
// multisig key's or a member's own address. SIGN_MODE_DIRECT_AUX is the sign
// mode that reads SignerData.Address: it refuses to sign for the fee payer,
// which here is the rekeyed account.
func TestMultisign_SignerAddress(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{})
	kr := keyring.NewInMemory(encCfg.Codec)
	txCfg := encCfg.TxConfig

	memberNames := []string{"m1", "m2", "m3"}
	members := make([]cryptotypes.PubKey, len(memberNames))
	for i, name := range memberNames {
		rec, _, err := kr.NewMnemonic(name, keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		require.NoError(t, err)
		members[i], err = rec.GetPubKey()
		require.NoError(t, err)
	}
	msigPk := kmultisig.NewLegacyAminoPubKey(2, members)
	_, err := kr.SaveMultisig("msig", msigPk)
	require.NoError(t, err)

	// The rekeyed account sends, so it is the fee payer.
	account := sdk.AccAddress("rekeyed_account_____")
	recipient := sdk.AccAddress("recipient___________")
	require.NotEqual(t, account, sdk.AccAddress(msigPk.Address()))

	const (
		chainID        = "test-chain"
		accNum  uint64 = 7
		seq     uint64 = 3
	)
	txBuilder := txCfg.NewTxBuilder()
	require.NoError(t, txBuilder.SetMsgs(banktypes.NewMsgSend(account, recipient, sdk.NewCoins(sdk.NewInt64Coin("stake", 1)))))
	txBuilder.SetGasLimit(200_000)
	txJSON, err := txCfg.TxJSONEncoder()(txBuilder.GetTx())
	require.NoError(t, err)
	unsignedFile := testutil.WriteToNewTempFile(t, string(txJSON)).Name()

	// signFiles writes two members' DIRECT_AUX signatures. SignDocDirectAux
	// commits to the pubkey in the signer data but not to the address;
	// multisign puts the member pubkey there, multisign-batch the multisig.
	signFiles := func(batch bool) []string {
		files := make([]string, 0, 2)
		for i, name := range memberNames[:2] {
			pk := members[i]
			if batch {
				pk = msigPk
			}
			bz, err := authsigning.GetSignBytesAdapter(context.Background(), txCfg.SignModeHandler(),
				signingtypes.SignMode_SIGN_MODE_DIRECT_AUX, authsigning.SignerData{
					Address:       sdk.AccAddress(members[i].Address()).String(),
					ChainID:       chainID,
					AccountNumber: accNum,
					Sequence:      seq,
					PubKey:        pk,
				}, txBuilder.GetTx())
			require.NoError(t, err)
			sig, _, err := kr.Sign(name, bz, signingtypes.SignMode_SIGN_MODE_DIRECT_AUX)
			require.NoError(t, err)
			sigJSON, err := txCfg.MarshalSignatureJSON([]signingtypes.SignatureV2{{
				PubKey:   members[i],
				Data:     &signingtypes.SingleSignatureData{SignMode: signingtypes.SignMode_SIGN_MODE_DIRECT_AUX, Signature: sig},
				Sequence: seq,
			}})
			require.NoError(t, err)
			files = append(files, testutil.WriteToNewTempFile(t, string(sigJSON)).Name())
		}
		return files
	}

	clientCtx := client.Context{}.
		WithKeyring(kr).
		WithCodec(encCfg.Codec).
		WithInterfaceRegistry(encCfg.InterfaceRegistry).
		WithTxConfig(txCfg).
		WithChainID(chainID)

	for _, batch := range []bool{false, true} {
		multisignCmd := cli.GetMultiSignCommand
		if batch {
			multisignCmd = cli.GetMultiSignBatchCmd
		}
		run := func(extra ...string) error {
			args := append([]string{unsignedFile, "msig"}, signFiles(batch)...)
			args = append(args,
				fmt.Sprintf("--%s=%s", flags.FlagChainID, chainID),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
				fmt.Sprintf("--%s=true", flags.FlagOffline),
				fmt.Sprintf("--%s=%d", flags.FlagAccountNumber, accNum),
				fmt.Sprintf("--%s=%d", flags.FlagSequence, seq),
			)
			_, err := clitestutil.ExecTestCLICmd(clientCtx, multisignCmd(), append(args, extra...))
			return err
		}

		// Without --signer-address the signer is the multisig account, which
		// is not the fee payer, so the signatures verify.
		require.NoError(t, run(), "batch=%t", batch)

		// With --signer-address the signer is the rekeyed account, which is
		// the fee payer.
		err := run(fmt.Sprintf("--%s=%s", flags.FlagSignerAddress, account.String()))
		require.ErrorContains(t, err, "fee payer "+account.String()+" cannot sign with SIGN_MODE_DIRECT_AUX", "batch=%t", batch)
	}
}

// TestAux_SignerAddress rotates a secp256k1 account to another secp256k1 key,
// then has the new key make SIGN_MODE_DIRECT_AUX aux signer data for the
// account with --aux --signer-address. A separate fee payer adds the data to
// a tx and signs it, and the tx passes the ante handler.
//
// The CLI sets the signer address in both the client context and the
// factory, so this test does not cover a factory-only signer address, which
// TestMakeAuxSignerData_WithSignerAddress in client/tx does.
func TestAux_SignerAddress(t *testing.T) {
	rn := startRekeyNetwork(t)
	cfg, val, clientCtx, txFlags := rn.cfg, rn.val, rn.clientCtx, rn.txFlags
	kr := clientCtx.Keyring
	ac := clientCtx.TxConfig.SigningContext().AddressCodec()

	oldRec, _, err := kr.NewMnemonic("aux-old", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)
	account, err := oldRec.GetAddress()
	require.NoError(t, err)
	newRec, _, err := kr.NewMnemonic("aux-new", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)
	newPk, err := newRec.GetPubKey()
	require.NoError(t, err)
	require.NotEqual(t, account, sdk.AccAddress(newPk.Address()))

	out, err := clitestutil.MsgSendExec(clientCtx, val.Address, account,
		sdk.NewCoins(sdk.NewInt64Coin(cfg.BondDenom, 1_000_000)), ac, txFlags...)
	require.NoError(t, err)
	rn.requireTxOK(t, out)

	// Rekey the account to the new key.
	pkJSON, err := clientCtx.Codec.MarshalInterfaceJSON(newPk)
	require.NoError(t, err)
	out, err = clitestutil.ExecTestCLICmd(clientCtx, cli.NewSignRekeyProofCmd(), []string{
		account.String(), string(pkJSON),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, "aux-new"),
	})
	require.NoError(t, err)
	proofFile := testutil.WriteToNewTempFile(t, out.String()).Name()
	out, err = clitestutil.ExecTestCLICmd(clientCtx, cli.NewChangePubKeyCmd(), append([]string{
		account.String(), string(pkJSON),
		"--proof=" + proofFile,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, "aux-old"),
	}, txFlags...))
	require.NoError(t, err)
	rn.requireTxOK(t, out)

	recipient := sdk.AccAddress("aux_test_recipient__")
	sendAmt := sdk.NewInt64Coin(cfg.BondDenom, 1234)
	auxArgs := []string{
		"aux-new", recipient.String(), sendAmt.String(),
		fmt.Sprintf("--%s=true", flags.FlagAux),
		fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
	}

	// Without --signer-address the aux signer is the new key's own address,
	// which has no account.
	_, err = clitestutil.ExecTestCLICmd(clientCtx, bankcli.NewSendTxCmd(ac), auxArgs)
	require.ErrorContains(t, err, "not found")

	out, err = clitestutil.ExecTestCLICmd(clientCtx, bankcli.NewSendTxCmd(ac), append(auxArgs,
		fmt.Sprintf("--%s=%s", flags.FlagSignerAddress, account.String())))
	require.NoError(t, err)
	var auxData txtypes.AuxSignerData
	require.NoError(t, clientCtx.Codec.UnmarshalJSON(out.Bytes(), &auxData), out.String())
	require.Equal(t, account.String(), auxData.Address)

	// The validator pays the fee and signs after the aux signer, as in
	// x/auth/tx/aux_test.go: there is no CLI that consumes aux signer data.
	txBuilder := clientCtx.TxConfig.NewTxBuilder()
	txBuilder.SetFeePayer(val.Address)
	txBuilder.SetFeeAmount(rn.fees)
	txBuilder.SetGasLimit(2_000_000)
	require.NoError(t, txBuilder.AddAuxSignerData(auxData))
	var auxPk cryptotypes.PubKey
	require.NoError(t, clientCtx.InterfaceRegistry.UnpackAny(auxData.SignDoc.PublicKey, &auxPk))
	require.True(t, newPk.Equals(auxPk))
	// Drop the fee payer's empty signer info slot, which Sign appends.
	require.NoError(t, txBuilder.SetSignatures(signingtypes.SignatureV2{
		PubKey:   auxPk,
		Data:     &signingtypes.SingleSignatureData{SignMode: auxData.Mode, Signature: auxData.Sig},
		Sequence: auxData.SignDoc.Sequence,
	}))

	valRec, err := kr.KeyByAddress(val.Address)
	require.NoError(t, err)
	txf, err := clienttx.Factory{}.
		WithTxConfig(clientCtx.TxConfig).
		WithKeybase(kr).
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithChainID(clientCtx.ChainID).
		WithSignMode(signingtypes.SignMode_SIGN_MODE_DIRECT).
		Prepare(clientCtx.WithFromAddress(val.Address))
	require.NoError(t, err)
	require.NoError(t, clienttx.Sign(context.Background(), txf, valRec.Name, txBuilder, false))

	txJSON, err := clientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	require.NoError(t, err)
	signedFile := testutil.WriteToNewTempFile(t, string(txJSON)).Name()

	out, err = clitestutil.ExecTestCLICmd(clientCtx, cli.GetBroadcastCommand(), []string{
		signedFile,
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
	})
	require.NoError(t, err)
	rn.requireTxOK(t, out)

	bal, err := banktypes.NewQueryClient(clientCtx).Balance(context.Background(),
		&banktypes.QueryBalanceRequest{Address: recipient.String(), Denom: cfg.BondDenom})
	require.NoError(t, err)
	require.Equal(t, sendAmt, *bal.Balance)
}
