package keys

import (
	"fmt"

	"github.com/bartekn/go-bip39"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
)

const (
	flagNewDefault = "default"
	flagBIP44Path  = "bip44-path"
)

func newKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Interactive command to derive a new private key, encrypt it, and save to disk",
		Long: `Derive a new private key using an interactive command that will prompt you for each input.
Optionally specify a bip39 mnemonic, a bip39 passphrase to further secure the mnemonic,
and a bip32 HD path to derive a specific account. The key will be stored under the given name 
and encrypted with the given password. The only input that is required is the encryption password.`,
		Args: cobra.ExactArgs(1),
		RunE: runNewCmd,
	}
	cmd.Flags().Bool(flagNewDefault, false, "Skip the prompts and just use the default values for everything")
	cmd.Flags().Bool(client.FlagUseLedger, false, "Store a local reference to a private key on a Ledger device")
	cmd.Flags().String(flagBIP44Path, "44'/118'/0'/0/0", "BIP44 path from which to derive a private key")
	return cmd
}

/*
input
	- bip39 mnemonic
	- bip39 passphrase
	- bip44 path
	- local encryption password
output
	- armor encrypted private key (saved to file)
*/
func runNewCmd(cmd *cobra.Command, args []string) error {
	name := args[0]
	kb, err := GetKeyBaseWithWritePerm()
	if err != nil {
		return err
	}

	buf := client.BufferStdin()

	_, err = kb.Get(name)
	if err == nil {
		// account exists, ask for user confirmation
		if response, err := client.GetConfirmation(
			fmt.Sprintf("> override the existing name %s", name), buf); err != nil || !response {
			return err
		}
	}

	flags := cmd.Flags()
	useDefaults, _ := flags.GetBool(flagNewDefault)
	bipFlag := flags.Lookup(flagBIP44Path)

	bip44Params, err := getBIP44ParamsAndPath(bipFlag.Value.String(), bipFlag.Changed || useDefaults)
	if err != nil {
		return err
	}

	// If we're using ledger, only thing we need is the path. So generate key and
	// we're done.
	if viper.GetBool(client.FlagUseLedger) {
		algo := keys.Secp256k1
		path := bip44Params.DerivationPath() // ccrypto.DerivationPath{44, 118, account, 0, index}

		info, err := kb.CreateLedger(name, path, algo)
		if err != nil {
			return err
		}

		printCreate(info, "")
		return nil
	}

	var mnemonic string

	if !useDefaults {
		mnemonic, err = client.GetString("Enter your bip39 mnemonic, or hit enter to generate one.", buf)
		if err != nil {
			return err
		}
	}

	if len(mnemonic) == 0 {
		// read entropy seed straight from crypto.Rand and convert to mnemonic
		entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
		if err != nil {
			return err
		}

		mnemonic, err = bip39.NewMnemonic(entropySeed[:])
		if err != nil {
			return err
		}
	}

	// get bip39 passphrase
	var bip39Passphrase string
	if !useDefaults {
		printStep()
		printPrefixed("Enter your bip39 passphrase. This is combined with the mnemonic to derive the seed")

		bip39Passphrase, err = client.GetString("Most users should just hit enter to use the default, \"\"", buf)
		if err != nil {
			return err
		}

		// if they use one, make them re-enter it
		if len(bip39Passphrase) != 0 {
			p2, err := client.GetString("Repeat the passphrase:", buf)
			if err != nil {
				return err
			}

			if bip39Passphrase != p2 {
				return errors.New("passphrases don't match")
			}
		}
	}

	printStep()

	// get the encryption password
	encryptPassword, err := client.GetCheckPassword(
		"> Enter a passphrase to encrypt your key to disk:",
		"> Repeat the passphrase:", buf)
	if err != nil {
		return err
	}

	info, err := kb.Derive(name, mnemonic, bip39Passphrase, encryptPassword, *bip44Params)
	if err != nil {
		return err
	}

	_ = info
	return nil
}

func getBIP44ParamsAndPath(path string, flagSet bool) (*hd.BIP44Params, error) {
	buf := client.BufferStdin()
	bip44Path := path

	// if it wasn't set in the flag, give it a chance to overide interactively
	if !flagSet {
		var err error

		printStep()

		bip44Path, err = client.GetString(fmt.Sprintf("Enter your bip44 path. Default is %s\n", path), buf)
		if err != nil {
			return nil, err
		}

		if len(bip44Path) == 0 {
			bip44Path = path
		}
	}

	bip44params, err := hd.NewParamsFromPath(bip44Path)
	if err != nil {
		return nil, err
	}

	return bip44params, nil
}

func printPrefixed(msg string) {
	fmt.Printf("> %s\n", msg)
}

func printStep() {
	printPrefixed("-------------------------------------")
}
