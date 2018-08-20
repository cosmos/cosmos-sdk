package keys

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	bip39 "github.com/bartekn/go-bip39"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	keys "github.com/cosmos/cosmos-sdk/crypto/keys"
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
		Long: `Derive a new private key using an interactive command that will prompt you for each inputs.
Optionally specify a bip39 mnemonic, a bip39 passphrase to further secure the mnemonic, 
and a bip32 HD path to derive a specific account. The key will be stored under the given name 
and encrypted with the given password. The only input that is required is the encryption password.`,
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
// nolint: gocyclo
func runNewCmd(cmd *cobra.Command, args []string) error {

	if len(args) != 1 || len(args[0]) == 0 {
		return errors.New("you must provide a name for the key")
	}

	name := args[0]
	kb, err := GetKeyBase()
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

	// if we're using ledger, only thing we need is the path.
	// generate key and we're done.
	if viper.GetBool(client.FlagUseLedger) {

		algo := keys.Secp256k1               // SigningAlgo(viper.GetString(flagType))
		path := bip44Params.DerivationPath() // ccrypto.DerivationPath{44, 118, account, 0, index}
		info, err := kb.CreateLedger(name, path, algo)
		if err != nil {
			return err
		}
		printCreate(info, "")
		return nil
	}

	// get the mnemonic
	var mnemonic string
	if !useDefaults {
		fmt.Println("> Enter your bip39 mnemonic.")
		fmt.Println("> If you don't have one, just hit enter, and one will be generated for you.")
		mnemonic, err = readStdIn()
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
	var mnemonicPassphrase string
	if !useDefaults {
		fmt.Println("-------------------------------------")
		fmt.Println("> Enter your bip39 passphrase.")
		fmt.Println("> If you don't have one, just hit enter, and the default \"\" will be used")
		mnemonicPassphrase, err = readStdIn()
		if err != nil {
			return err
		}
	}

	// get the encryption password
	fmt.Println("-------------------------------------")
	encryptPassword, err := client.GetCheckPassword(
		"> Enter a passphrase to encrypt your key:",
		"> Repeat the passphrase:", buf)
	if err != nil {
		return err
	}

	info, err := kb.Derive(name, mnemonic, mnemonicPassphrase, encryptPassword, *bip44Params)
	if err != nil {
		return err
	}
	_ = info

	return nil
}

func readStdIn() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	mnemonic, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(mnemonic), nil
}

func getBIP44ParamsAndPath(path string, flagSet bool) (*hd.BIP44Params, error) {
	bip44Path := path

	// if it wasnt set in the flag, give it a chance to overide interactively
	if !flagSet {
		fmt.Println("-------------------------------------")
		fmt.Printf("> Enter your bip44 path. Default is %s\n", path)
		var err error
		bip44Path, err = readStdIn()
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
