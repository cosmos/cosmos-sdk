package keys

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
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

func TestAddBIP85(t *testing.T) {
	const defaultHash = "35d464b10e832cce796da2b2759476363e502576c992a44f621f6b8a367b856a"
	const accountHash = "f2419a2b2a1f6cd4a35117e40a039ad549ef7f6086d1a1651c82ba6d547e0730"
	const coinHash = "e18897fc498f18b9ceb794f2a7a3b973a15c11875cb1f302defc8ad21b1fe2ca"
	mnemonic := strings.Repeat("abandon ", 23) + "art"
	for _, tc := range []struct {
		name, input, pubHash, err string
		args                      []string
	}{
		{name: "recover default", input: mnemonic + "\n", pubHash: defaultHash, args: []string{"--recover"}},
		{name: "normalized mnemonic", input: strings.ReplaceAll(mnemonic, "a", "\uff41") + "\n", pubHash: defaultHash, args: []string{"--recover"}},
		{name: "mnemonic whitespace", input: "\u2003" + strings.ReplaceAll(mnemonic, " ", "\t\u00a0 ") + " \n", pubHash: defaultHash, args: []string{"--recover"}},
		{name: "normalized interactive mnemonic", input: strings.ReplaceAll(mnemonic, "a", "\uff41") + "\n\n", pubHash: defaultHash, args: []string{"--interactive"}},
		{name: "passphrase whitespace", input: mnemonic + "\n \tcafé\u00a0 \n \tcafé\u00a0 \n", pubHash: "68f42cfb1c24865f52be3b3a49b964cd850c702b5a6d7ee0924363e55e55c3ac", args: []string{"--recover", "--interactive"}},
		{name: "whitespace only passphrase", input: mnemonic + "\n \n \n", pubHash: "1c96be005bca83dc53d635f1d5790b38d4afe5a77b866f9fc29f5de963db92f4", args: []string{"--recover", "--interactive"}},
		{name: "passphrase whitespace mismatch", input: mnemonic + "\n secret\nsecret\n", err: "passphrases don't match", args: []string{"--recover", "--interactive"}},
		{name: "empty recovery", input: "\n", err: "invalid mnemonic", args: []string{"--recover"}},
		{name: "invalid checksum", input: strings.Repeat("abandon ", 24) + "\n", err: "Invalid byte", args: []string{"--recover"}},
		{name: "account and index", input: mnemonic + "\nTREZOR\nTREZOR\n", pubHash: accountHash, args: []string{"--interactive", "--account=7", "--index=9"}},
		{name: "explicit path recovery", input: mnemonic + "\nTREZOR\nTREZOR\n", pubHash: accountHash, args: []string{"--recover", "--interactive", "--hd-path=m/83696968'/505153'/118'/201'/7'/9'"}},
		{name: "coin type scope", input: mnemonic + "\n", pubHash: coinHash, args: []string{"--recover", "--coin-type=60", "--account=1", "--index=7"}},
		{name: "path overrides scope", input: mnemonic + "\n", pubHash: coinHash, args: []string{"--recover", "--coin-type=118", "--account=7", "--index=9", "--hd-path=m/83696968'/505153'/60'/201'/1'/7'"}},
		{name: "legacy mnemonic too short", input: strings.Repeat("abandon ", 11) + "about\n", err: "24-word", args: []string{"--recover"}},
		{name: "wrong algorithm", err: "requires --key-type", args: []string{"--key-type=secp256k1"}},
		{name: "unhardened path", input: mnemonic + "\n", err: "must be hardened", args: []string{"--recover", "--hd-path=m/83696968'/505153'/118'/201'/0'/1"}},
		{name: "obsolete scheme", input: mnemonic + "\n", err: "path must be", args: []string{"--recover", "--hd-path=m/83696968'/505153'/118'/3'/0'/0'"}},
		{name: "reserved scheme", input: mnemonic + "\n", err: "path must be", args: []string{"--recover", "--hd-path=m/83696968'/505153'/118'/400'/0'/0'"}},
		{name: "wrong namespace", input: mnemonic + "\n", err: "path must be", args: []string{"--recover", "--hd-path=m/83696968'/503131'/0'"}},
		{name: "coin type overflow", err: "must be less than", args: []string{"--coin-type=2147483648"}},
		{name: "account overflow", err: "must be less than", args: []string{"--account=2147483648"}},
		{name: "index overflow", err: "must be less than", args: []string{"--index=2147483648"}},
		{name: "ledger", err: "none of the others", args: []string{"--ledger"}},
		{name: "public key", err: "none of the others", args: []string{"--pubkey={}"}},
		{name: "multisig", err: "none of the others", args: []string{"--multisig=other"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := AddKeyCommand()
			cmd.Flags().AddFlagSet(Commands().PersistentFlags())
			mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
			mockIn.Reset(tc.input)
			cdc := moduletestutil.MakeTestEncodingConfig().Codec
			kb := keyring.NewInMemory(cdc)
			clientCtx := client.Context{}.WithKeyring(kb).WithCodec(cdc).WithInput(mockIn)
			ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)
			cmd.SetArgs(append([]string{"pq", "--key-type=ml_dsa_65", "--bip85", "--no-backup"}, tc.args...))
			err := cmd.ExecuteContext(ctx)
			if tc.err != "" {
				require.ErrorContains(t, err, tc.err)
				_, err = kb.Key("pq")
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			record, err := kb.Key("pq")
			require.NoError(t, err)
			pub, err := record.GetPubKey()
			require.NoError(t, err)
			require.Equal(t, "ml_dsa_65", pub.Type())
			hash := sha256.Sum256(pub.Bytes())
			require.Equal(t, tc.pubHash, hex.EncodeToString(hash[:]))
		})
	}
}
