package keys

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bip39 "github.com/cosmos/go-bip39"
)

func Test_runAddCmdBasic(t *testing.T) {
	cmd := AddKeyCommand()
	cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())

	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
	kbHome := t.TempDir()

	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockIn)
	require.NoError(t, err)

	clientCtx := client.Context{}.WithKeyringDir(kbHome).WithInput(mockIn)
	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	t.Cleanup(func() {
		_ = kb.Delete("keyname1")
		_ = kb.Delete("keyname2")
	})

	cmd.SetArgs([]string{
		"keyname1",
		fmt.Sprintf("--%s=%s", flags.FlagHome, kbHome),
		fmt.Sprintf("--%s=%s", cli.OutputFlag, OutputFormatText),
		fmt.Sprintf("--%s=%s", flags.FlagKeyAlgorithm, string(hd.Secp256k1Type)),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})
	mockIn.Reset("y\n")
	require.NoError(t, cmd.ExecuteContext(ctx))

	mockIn.Reset("N\n")
	require.Error(t, cmd.ExecuteContext(ctx))

	cmd.SetArgs([]string{
		"keyname2",
		fmt.Sprintf("--%s=%s", flags.FlagHome, kbHome),
		fmt.Sprintf("--%s=%s", cli.OutputFlag, OutputFormatText),
		fmt.Sprintf("--%s=%s", flags.FlagKeyAlgorithm, string(hd.Secp256k1Type)),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})

	require.NoError(t, cmd.ExecuteContext(ctx))
	require.Error(t, cmd.ExecuteContext(ctx))

	mockIn.Reset("y\n")
	require.NoError(t, cmd.ExecuteContext(ctx))

	cmd.SetArgs([]string{
		"keyname4",
		fmt.Sprintf("--%s=%s", flags.FlagHome, kbHome),
		fmt.Sprintf("--%s=%s", cli.OutputFlag, OutputFormatText),
		fmt.Sprintf("--%s=%s", flags.FlagKeyAlgorithm, string(hd.Secp256k1Type)),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})

	require.NoError(t, cmd.ExecuteContext(ctx))
	require.Error(t, cmd.ExecuteContext(ctx))

	cmd.SetArgs([]string{
		"keyname5",
		fmt.Sprintf("--%s=%s", flags.FlagHome, kbHome),
		fmt.Sprintf("--%s=true", flags.FlagDryRun),
		fmt.Sprintf("--%s=%s", cli.OutputFlag, OutputFormatText),
		fmt.Sprintf("--%s=%s", flags.FlagKeyAlgorithm, string(hd.Secp256k1Type)),
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

func TestAddRecoverFileBackend(t *testing.T) {
	cmd := AddKeyCommand()
	cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())

	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
	kbHome := t.TempDir()

	clientCtx := client.Context{}.WithKeyringDir(kbHome).WithInput(mockIn)
	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	cmd.SetArgs([]string{
		"keyname1",
		fmt.Sprintf("--%s=%s", flags.FlagHome, kbHome),
		fmt.Sprintf("--%s=%s", cli.OutputFlag, OutputFormatText),
		fmt.Sprintf("--%s=%s", flags.FlagKeyAlgorithm, string(hd.Secp256k1Type)),
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

	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendFile, kbHome, mockIn)
	require.NoError(t, err)

	t.Cleanup(func() {
		mockIn.Reset(fmt.Sprintf("%s\n%s\n", keyringPassword, keyringPassword))
		_ = kb.Delete("keyname1")
	})

	mockIn.Reset(fmt.Sprintf("%s\n%s\n", keyringPassword, keyringPassword))
	info, err := kb.Key("keyname1")
	require.NoError(t, err)
	require.Equal(t, "keyname1", info.GetName())
}
