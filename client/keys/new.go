package keys

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/pkg/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"bufio"
)

const (
	flagDefaultEntropy = "default"
)

func newCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "new <name>",
		Args: cobra.ExactArgs(1),
		Short: "Creates a new key interactively",
		RunE: runNewCommand,
	}
	cmd.Flags().Bool(flagDefaultEntropy, true, "Use system entropy to generate a new mnemonic and derive a key using default parameters")
	return cmd
}

func runNewCommand(cmd *cobra.Command, args []string) error {
	name := args[0]

	if name == "" {
		return errors.New("you must provide a name for the key")
	}

	kb, err := GetKeyBase()
	if err != nil {
		return err
	}

	stdin := client.BufferStdin()

	isDefault := viper.GetBool(flagDefaultEntropy)
	if isDefault {
		return handleDefault(name, kb, stdin)
	}

	var mnemonic string
	var bip39Pw string
	var bip44Path string
	var encryptionPw string

	mnemonic, err = handleMnemonic(kb, stdin)
	if err != nil {
		return err
	}
	printStep()
	bip39Pw, err = handleBip39Pw(stdin)
	if err != nil {
		return err
	}
	printStep()
	bip44Path, err = handleBip44Path(stdin)
	if err != nil {
		return err
	}
	encryptionPw, err = handleEncryptionPw(stdin)
	if err != nil {
		return err
	}

	printStep()

	params, err := hd.ParamsFromString(bip44Path)
	if err != nil {
		return err
	}
	_, err = kb.Derive(name, mnemonic, encryptionPw, bip39Pw, *params)
	if err != nil {
		return err
	}
	fmt.Printf("Mnemonic: %s\n", mnemonic)
	fmt.Printf("Successfully wrote encrypted priv key named %s\n", name)

	return nil
}

func handleDefault(name string, kb keys.Keybase, stdin *bufio.Reader) error {
	pw, err := handleEncryptionPw(stdin)
	if err != nil {
		return err
	}

	_, seed, err := kb.CreateMnemonic(name, keys.English, pw, keys.Secp256k1)
	if err != nil {
		return err
	}
	fmt.Printf("Seed: %s\n", seed)
	fmt.Printf("Successfully wrote encrypted priv key named %s\n", name)
	return nil
}

func handleEncryptionPw(stdin *bufio.Reader) (string, error) {
	printPrefixed("Enter a password to encrypt the derived private key with.")
	pw, err := client.GetString("", stdin)
	if err != nil {
		return "", err
	}

	if pw == "" {
		return "", errors.New("you must define an encryption password")
	}

	return pw, nil
}

func handleMnemonic(kb keys.Keybase, stdin *bufio.Reader) (string, error) {
	printPrefixed("Enter your bip39 mnemonic.")
	printPrefixed("If you don't have one, just hit enter and one will be generated for you.")
	mnemonic, err := client.GetSeed("", stdin)
	if err != nil {
		return "", err
	}

	if mnemonic == "" {
		mnemonic, err = kb.GenerateMnemonic(keys.English, nil)
		if err != nil {
			return "", err
		}
	}

	return mnemonic, nil
}

func handleBip39Pw(stdin *bufio.Reader) (string, error) {
	printPrefixed("Enter your bip39 passphrase.")
	printPrefixed("If you don't have one, just hit enter and the default \"\" will be used.")
	bip39Pw, err := client.GetString("", stdin)
	if err != nil {
		return "", err
	}

	return bip39Pw, nil
}

func handleBip44Path(stdin *bufio.Reader) (string, error) {
	printPrefixed("Enter your bip44 path. If you press enter, the default of m/44'/0'/0'/0/0 will be used.")
	bip44Path, err := client.GetString("", stdin)
	if err != nil {
		return "", err
	}

	if bip44Path == "" {
		bip44Path = "m/44'/0'/0'/0/0"
	}

	return bip44Path, nil
}

func printPrefixed(msg string) {
	fmt.Printf("> %s\n", msg)
}

func printStep() {
	printPrefixed("-------------------------------------")
}
