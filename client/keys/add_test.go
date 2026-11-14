package keys

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/cosmos/go-bip39"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

func Test_runAddCmdBasic(t *testing.T) {
	cmd := AddKeyCommand()
	cmd.Flags().AddFlagSet(Commands().PersistentFlags())

	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
	kbHome := t.TempDir()

	cdc := moduletestutil.MakeTestEncodingConfig().Codec
	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockIn, cdc)
	require.NoError(t, err)

	clientCtx := client.Context{}.WithKeyringDir(kbHome).WithInput(mockIn).WithCodec(cdc)
	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	t.Cleanup(func() {
		_ = kb.Delete("keyname1")
		_ = kb.Delete("keyname2")
	})

	// test empty name
	cmd.SetArgs([]string{
		"",
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=%s", flags.FlagOutput, flags.OutputFormatText),
		fmt.Sprintf("--%s=%s", flags.FlagKeyType, hd.Secp256k1Type),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})
	mockIn.Reset("y\n")
	require.ErrorContains(t, cmd.ExecuteContext(ctx), "the provided name is invalid or empty after trimming whitespace")

	cmd.SetArgs([]string{
		"keyname1",
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=%s", flags.FlagOutput, flags.OutputFormatText),
		fmt.Sprintf("--%s=%s", flags.FlagKeyType, hd.Secp256k1Type),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})
	mockIn.Reset("y\n")
	require.NoError(t, cmd.ExecuteContext(ctx))

	mockIn.Reset("N\n")
	require.Error(t, cmd.ExecuteContext(ctx))

	cmd.SetArgs([]string{
		"keyname2",
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=%s", flags.FlagOutput, flags.OutputFormatText),
		fmt.Sprintf("--%s=%s", flags.FlagKeyType, hd.Secp256k1Type),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})

	require.NoError(t, cmd.ExecuteContext(ctx))
	require.Error(t, cmd.ExecuteContext(ctx))

	mockIn.Reset("y\n")
	require.NoError(t, cmd.ExecuteContext(ctx))

	cmd.SetArgs([]string{
		"keyname4",
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=%s", flags.FlagOutput, flags.OutputFormatText),
		fmt.Sprintf("--%s=%s", flags.FlagKeyType, hd.Secp256k1Type),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})

	require.NoError(t, cmd.ExecuteContext(ctx))
	require.Error(t, cmd.ExecuteContext(ctx))

	cmd.SetArgs([]string{
		"keyname5",
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=true", flags.FlagDryRun),
		fmt.Sprintf("--%s=%s", flags.FlagOutput, flags.OutputFormatText),
		fmt.Sprintf("--%s=%s", flags.FlagKeyType, hd.Secp256k1Type),
	})

	require.NoError(t, cmd.ExecuteContext(ctx))

	// In recovery mode
	cmd.SetArgs([]string{
		"keyname6",
		fmt.Sprintf("--%s=true", flagRecover),
	})

	// use valid mnemonic and complete recovery key generation successfully
	mockIn.Reset("decide praise business actor peasant farm drastic weather extend front hurt later song give verb rhythm worry fun pond reform school tumble august one\n")
	require.NoError(t, cmd.ExecuteContext(ctx))

	// use invalid mnemonic and fail recovery key generation
	mockIn.Reset("invalid mnemonic\n")
	require.Error(t, cmd.ExecuteContext(ctx))

	// In interactive mode
	cmd.SetArgs([]string{
		"keyname7",
		"-i",
		fmt.Sprintf("--%s=false", flagRecover),
	})

	const password = "password1!"

	// set password and complete interactive key generation successfully
	mockIn.Reset("\n" + password + "\n" + password + "\n")
	require.NoError(t, cmd.ExecuteContext(ctx))

	// passwords don't match and fail interactive key generation
	mockIn.Reset("\n" + password + "\n" + "fail" + "\n")
	require.Error(t, cmd.ExecuteContext(ctx))
}

func Test_runAddCmdMultisigDupKeys(t *testing.T) {
	cmd := AddKeyCommand()
	cmd.Flags().AddFlagSet(Commands().PersistentFlags())

	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
	kbHome := t.TempDir()

	cdc := moduletestutil.MakeTestEncodingConfig().Codec
	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockIn, cdc)
	require.NoError(t, err)

	clientCtx := client.Context{}.
		WithKeyringDir(kbHome).
		WithInput(mockIn).
		WithCodec(cdc)

	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	t.Cleanup(func() {
		_ = kb.Delete("keyname1")
		_ = kb.Delete("keyname2")
		_ = kb.Delete("multisigname")
	})

	cmd.SetArgs([]string{
		"keyname1",
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=%s", flags.FlagOutput, flags.OutputFormatText),
		fmt.Sprintf("--%s=%s", flags.FlagKeyType, hd.Secp256k1Type),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})
	require.NoError(t, cmd.ExecuteContext(ctx))

	cmd.SetArgs([]string{
		"keyname2",
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=%s", flags.FlagOutput, flags.OutputFormatText),
		fmt.Sprintf("--%s=%s", flags.FlagKeyType, hd.Secp256k1Type),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})
	require.NoError(t, cmd.ExecuteContext(ctx))

	cmd.SetArgs([]string{
		"multisigname",
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=%s", flags.FlagOutput, flags.OutputFormatText),
		fmt.Sprintf("--%s=%s", flags.FlagKeyType, hd.Secp256k1Type),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		fmt.Sprintf("--%s=%s", flagMultisig, "keyname1,keyname2"),
		fmt.Sprintf("--%s=%s", flagMultiSigThreshold, "2"),
	})
	require.NoError(t, cmd.ExecuteContext(ctx))

	cmd.SetArgs([]string{
		"multisigname",
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=%s", flags.FlagOutput, flags.OutputFormatText),
		fmt.Sprintf("--%s=%s", flags.FlagKeyType, hd.Secp256k1Type),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		fmt.Sprintf("--%s=%s", flagMultisig, "keyname1,keyname1"),
		fmt.Sprintf("--%s=%s", flagMultiSigThreshold, "2"),
	})
	mockIn.Reset("y\n")
	require.Error(t, cmd.ExecuteContext(ctx))
	mockIn.Reset("y\n")
	require.EqualError(t, cmd.ExecuteContext(ctx), "duplicate multisig keys: keyname1")
}

func Test_runAddCmdDryRun(t *testing.T) {
	pubkey1 := `{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AtObiFVE4s+9+RX5SP8TN9r2mxpoaT4eGj9CJfK7VRzN"}`
	pubkey2 := `{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A/se1vkqgdQ7VJQCM4mxN+L+ciGhnnJ4XYsQCRBMrdRi"}`
	b64Pubkey := "QWhnOHhpdXBJcGZ2UlR2ak5la1ExclROUThTOW96YjdHK2RYQmFLVjl4aUo="
	cdc := moduletestutil.MakeTestEncodingConfig().Codec

	testData := []struct {
		name  string
		args  []string
		added bool
	}{
		{
			name: "account is added",
			args: []string{
				"testkey",
				fmt.Sprintf("--%s=%s", flags.FlagDryRun, "false"),
			},
			added: true,
		},
		{
			name: "account is not added with dry run",
			args: []string{
				"testkey",
				fmt.Sprintf("--%s=%s", flags.FlagDryRun, "true"),
			},
			added: false,
		},
		{
			name: "multisig account is added",
			args: []string{
				"testkey",
				fmt.Sprintf("--%s=%s", flags.FlagDryRun, "false"),
				fmt.Sprintf("--%s=%s", flagMultisig, "subkey"),
			},
			added: true,
		},
		{
			name: "multisig account is not added with dry run",
			args: []string{
				"testkey",
				fmt.Sprintf("--%s=%s", flags.FlagDryRun, "true"),
				fmt.Sprintf("--%s=%s", flagMultisig, "subkey"),
			},
			added: false,
		},
		{
			name: "pubkey account is added",
			args: []string{
				"testkey",
				fmt.Sprintf("--%s=%s", flags.FlagDryRun, "false"),
				fmt.Sprintf("--%s=%s", FlagPublicKey, pubkey1),
			},
			added: true,
		},
		{
			name: "pubkey account is not added with dry run",
			args: []string{
				"testkey",
				fmt.Sprintf("--%s=%s", flags.FlagDryRun, "true"),
				fmt.Sprintf("--%s=%s", FlagPublicKey, pubkey2),
			},
			added: false,
		},
		{
			name: "base64 pubkey account is added",
			args: []string{
				"testkey",
				fmt.Sprintf("--%s=%s", flags.FlagDryRun, "false"),
				fmt.Sprintf("--%s=%s", flagPubKeyBase64, b64Pubkey),
			},
			added: true,
		},
		{
			name: "base64 pubkey account is not added with dry run",
			args: []string{
				"testkey",
				fmt.Sprintf("--%s=%s", flags.FlagDryRun, "true"),
				fmt.Sprintf("--%s=%s", flagPubKeyBase64, b64Pubkey),
			},
			added: false,
		},
	}
	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			cmd := AddKeyCommand()
			cmd.Flags().AddFlagSet(Commands().PersistentFlags())

			kbHome := t.TempDir()
			mockIn := testutil.ApplyMockIODiscardOutErr(cmd)

			kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockIn, cdc)
			require.NoError(t, err)

			clientCtx := client.Context{}.
				WithCodec(cdc).
				WithKeyringDir(kbHome).
				WithKeyring(kb)
			ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

			path := sdk.GetConfig().GetFullBIP44Path()
			_, err = kb.NewAccount("subkey", testdata.TestMnemonic, "", path, hd.Secp256k1)
			require.NoError(t, err)

			t.Cleanup(func() {
				_ = kb.Delete("subkey")
			})

			b := bytes.NewBufferString("")
			cmd.SetOut(b)

			cmd.SetArgs(tt.args)
			require.NoError(t, cmd.ExecuteContext(ctx))

			if tt.added {
				_, err := kb.Key("testkey")
				require.NoError(t, err)

				out, err := io.ReadAll(b)
				require.NoError(t, err)
				require.Contains(t, string(out), "name: testkey")
			} else {
				_, err = kb.Key("testkey")
				require.Error(t, err)
				require.Equal(t, "testkey.info: key not found", err.Error())
			}
		})
	}
}

func TestAddRecoverFileBackend(t *testing.T) {
	cmd := AddKeyCommand()
	cmd.Flags().AddFlagSet(Commands().PersistentFlags())
	cdc := moduletestutil.MakeTestEncodingConfig().Codec

	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
	kbHome := t.TempDir()

	clientCtx := client.Context{}.WithKeyringDir(kbHome).WithInput(mockIn).WithCodec(cdc)
	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	cmd.SetArgs([]string{
		"keyname1",
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=%s", flags.FlagOutput, flags.OutputFormatText),
		fmt.Sprintf("--%s=%s", flags.FlagKeyType, hd.Secp256k1Type),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendFile),
		fmt.Sprintf("--%s", flagRecover),
	})

	keyringPassword := "12345678"

	entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
	require.NoError(t, err)

	mnemonic, err := bip39.NewMnemonic(entropySeed)
	require.NoError(t, err)

	mockIn.Reset(fmt.Sprintf("%s\n%s\n%s\n", mnemonic, keyringPassword, keyringPassword))
	require.NoError(t, cmd.ExecuteContext(ctx))

	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendFile, kbHome, mockIn, cdc)
	require.NoError(t, err)

	t.Cleanup(func() {
		mockIn.Reset(fmt.Sprintf("%s\n%s\n", keyringPassword, keyringPassword))
		_ = kb.Delete("keyname1")
	})

	mockIn.Reset(fmt.Sprintf("%s\n%s\n", keyringPassword, keyringPassword))
	k, err := kb.Key("keyname1")
	require.NoError(t, err)
	require.Equal(t, "keyname1", k.Name)
}

func Test_readMnemonicFromFile(t *testing.T) {
	const mnemonic = "equip will roof matter pink blind book anxiety banner elbow sun young"

	tests := []struct {
		name     string
		noFile   bool
		content  string
		expected string
		expErr   string
	}{
		{
			name:   "file does not exist",
			noFile: true,
			expErr: "no such file or directory",
		},
		{
			name:    "empty file",
			content: "",
			expErr:  "mnemonic source file is empty",
		},
		{
			name:    "only spaces",
			content: "   ",
			expErr:  "mnemonic source file is empty",
		},
		{
			name:    "only newlines",
			content: "\n\n",
			expErr:  "mnemonic source file is empty",
		},
		{
			name:    "mixed whitespace only",
			content: " \t\r\n \n\t ",
			expErr:  "mnemonic source file is empty",
		},
		{
			name:     "no surrounding whitespace",
			content:  mnemonic,
			expected: mnemonic,
		},
		{
			name:     "trailing newline",
			content:  mnemonic + "\n",
			expected: mnemonic,
		},
		{
			name:     "trailing CRLF",
			content:  mnemonic + "\r\n",
			expected: mnemonic,
		},
		{
			name:     "leading whitespace",
			content:  " \t\n" + mnemonic,
			expected: mnemonic,
		},
		{
			name:     "leading and trailing whitespace",
			content:  "\n  \t" + mnemonic + " \t\n\n",
			expected: mnemonic,
		},
		{
			name:     "inner whitespace is preserved",
			content:  "\n" + "equip  will\troof" + "\n",
			expected: "equip  will\troof",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filePath := filepath.Join(t.TempDir(), "mnemonic.txt")
			if !tc.noFile {
				require.NoError(t, os.WriteFile(filePath, []byte(tc.content), 0o600), "os.WriteFile")
			}

			var actual string
			var err error
			testFunc := func() {
				actual, err = readMnemonicFromFile(filePath)
			}
			require.NotPanics(t, testFunc, "readMnemonicFromFile")
			if len(tc.expErr) > 0 {
				require.ErrorContains(t, err, tc.expErr, "readMnemonicFromFile error")
			} else {
				require.NoError(t, err, "readMnemonicFromFile error")
			}
			require.Equal(t, tc.expected, actual, "readMnemonicFromFile result")
		})
	}
}

func Test_runAddCmdMnemonicSource(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig().Codec
	paddedMnemonic := "\n  " + testdata.TestMnemonic + " \t\n\n"

	// Derive the address we expect from the test mnemonic so we know the file content was used.
	expKb := keyring.NewInMemory(cdc)
	expRec, err := expKb.NewAccount("expected", testdata.TestMnemonic, "", sdk.GetConfig().GetFullBIP44Path(), hd.Secp256k1)
	require.NoError(t, err, "NewAccount for expected key")
	expAddr, err := expRec.GetAddress()
	require.NoError(t, err, "GetAddress for expected key")

	tests := []struct {
		name    string
		mode    string // either flagRecover or flagInteractive.
		noFile  bool
		content string
		stdin   string
		expErr  string
	}{
		{
			name:    "recover: padded mnemonic",
			mode:    flagRecover,
			content: paddedMnemonic,
		},
		{
			name:    "recover: empty file",
			mode:    flagRecover,
			content: "",
			expErr:  "mnemonic source file is empty",
		},
		{
			name:    "recover: whitespace only",
			mode:    flagRecover,
			content: " \t\n\r\n ",
			expErr:  "mnemonic source file is empty",
		},
		{
			name:   "recover: file does not exist",
			mode:   flagRecover,
			noFile: true,
			expErr: "no such file or directory",
		},
		{
			name:    "recover: invalid mnemonic",
			mode:    flagRecover,
			content: "invalid mnemonic\n",
			expErr:  "invalid mnemonic",
		},
		{
			name:    "interactive: padded mnemonic",
			mode:    flagInteractive,
			content: paddedMnemonic,
			stdin:   "\n", // empty bip39 passphrase
		},
		{
			name:    "interactive: empty file",
			mode:    flagInteractive,
			content: "",
			stdin:   "\n",
			expErr:  "mnemonic source file is empty",
		},
		{
			name:    "interactive: whitespace only",
			mode:    flagInteractive,
			content: " \t\n\r\n ",
			stdin:   "\n",
			expErr:  "mnemonic source file is empty",
		},
		{
			name:   "interactive: file does not exist",
			mode:   flagInteractive,
			noFile: true,
			stdin:  "\n",
			expErr: "no such file or directory",
		},
		{
			name:    "interactive: invalid mnemonic",
			mode:    flagInteractive,
			content: "invalid mnemonic\n",
			stdin:   "\n",
			expErr:  "invalid mnemonic",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := AddKeyCommand()
			cmd.Flags().AddFlagSet(Commands().PersistentFlags())

			kbHome := t.TempDir()
			mockIn := testutil.ApplyMockIODiscardOutErr(cmd)

			kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockIn, cdc)
			require.NoError(t, err, "keyring.New")

			clientCtx := client.Context{}.
				WithCodec(cdc).
				WithKeyringDir(kbHome).
				WithKeyring(kb).
				WithInput(mockIn)
			ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

			srcFile := filepath.Join(t.TempDir(), "mnemonic.txt")
			if !tc.noFile {
				require.NoError(t, os.WriteFile(srcFile, []byte(tc.content), 0o600), "os.WriteFile")
			}

			cmd.SetArgs([]string{
				"testkey",
				fmt.Sprintf("--%s=%s", flags.FlagOutput, flags.OutputFormatText),
				fmt.Sprintf("--%s", tc.mode),
				fmt.Sprintf("--%s=%s", flagMnemonicSrc, srcFile),
			})
			mockIn.Reset(tc.stdin)

			err = cmd.ExecuteContext(ctx)
			if len(tc.expErr) > 0 {
				require.ErrorContains(t, err, tc.expErr, "ExecuteContext error")
				_, err = kb.Key("testkey")
				require.Error(t, err, "key should not have been added")
				return
			}

			require.NoError(t, err, "ExecuteContext error")
			rec, err := kb.Key("testkey")
			require.NoError(t, err, "kb.Key")
			addr, err := rec.GetAddress()
			require.NoError(t, err, "GetAddress")
			require.Equal(t, expAddr.String(), addr.String(), "key address")
		})
	}
}
