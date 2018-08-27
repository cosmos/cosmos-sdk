package keys

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/pkg/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
)

const (
	flagDefault = "default"
)

func newCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "new <name>",
		Args: cobra.ExactArgs(1),
		Short: "Creates a new key interactively, or with sensible defaults.",
		RunE: runNewCommand,
	}
	cmd.Flags().Bool(flagDefault, true, "Use system entropy to generate a new mnemonic and derive a key using default parameters")
	return cmd
}

// nolint: gocyclo
func runNewCommand(cmd *cobra.Command, args []string) error {
	name := args[0]

	if name == "" {
		return errors.New("you must provide a name for the key")
	}

	kb, err := GetKeyBase()
	if err != nil {
		return err
	}

	isDefault := viper.GetBool(flagDefault)
	if isDefault {
		_, seed, err := kb.CreateMnemonic(name, keys.English, "", keys.Secp256k1)
		if err != nil {
			return err
		}
		fmt.Printf("Seed: %s\n", seed)
		fmt.Printf("Successfully wrote encrypted priv key named %s\n", name)
		return nil
	}

	var mnemonic string
	var bip39Pw string
	var bip44Path string
	var encryptionPw string

	stdin := client.BufferStdin()
	printPrefixed("Enter your bip39 mnemonic.")
	printPrefixed("If you don't have one, just hit enter and one will be generated for you.")
	mnemonic, err = client.GetSeed("", stdin)
	if err != nil {
		return err
	}

	if mnemonic == "" {
		mnemonic, err = kb.GenerateMnemonic(keys.English, nil)
		if err != nil {
			return err
		}
	}

	printStep()
	printPrefixed("Enter your bip39 passphrase.")
	printPrefixed("If you don't have one, just hit enter and the default \"\" will be used.")
	bip39Pw, err = client.GetString("", stdin)
	if err != nil {
		return err
	}

	printStep()
	printPrefixed("Enter your bip44 path. If you press enter, the default of m/44'/0'/0'/0/0 will be used.")
	bip44Path, err = client.GetString("", stdin)
	if err != nil {
		return err
	}

	if bip44Path == "" {
		bip44Path = "m/44'/0'/0'/0/0"
	}

	printStep()
	printPrefixed("Enter a password to encrypt the derived private key with.")
	encryptionPw, err = client.GetString("", stdin)
	if err != nil {
		return err
	}

	if encryptionPw == "" {
		return errors.New("you must define an encryption password")
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

func printPrefixed(msg string) {
	fmt.Printf("> %s\n", msg)
}

func printStep() {
	printPrefixed("-------------------------------------")
}