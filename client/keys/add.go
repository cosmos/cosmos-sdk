package keys

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/cosmos/go-bip39"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	ccrypto "github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
)

const (
	flagInteractive = "interactive"
	flagBIP44Path   = "bip44-path"
	flagRecover     = "recover"
	flagNoBackup    = "no-backup"
	flagDryRun      = "dry-run"
	flagAccount     = "account"
	flagIndex       = "index"
)

func addKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add an encrypted private key (either newly generated or recovered), encrypt it, and save to disk",
		Long: `Derive a new private key and encrypt to disk.
Optionally specify a BIP39 mnemonic, a BIP39 passphrase to further secure the mnemonic,
and a bip32 HD path to derive a specific account. The key will be stored under the given name 
and encrypted with the given password. The only input that is required is the encryption password.

If run with -i, it will prompt the user for BIP44 path, BIP39 mnemonic, and passphrase.
The flag --recover allows one to recover a key from a seed passphrase.
If run with --dry-run, a key would be generated (or recovered) but not stored to the local keystore.
`,
		Args: cobra.ExactArgs(1),
		RunE: runAddCmd,
	}
	cmd.Flags().BoolP(flagInteractive, "i", false, "Interactively prompt user for BIP39 passphrase and mnemonic")
	cmd.Flags().Bool(client.FlagUseLedger, false, "Store a local reference to a private key on a Ledger device")
	cmd.Flags().String(flagBIP44Path, "44'/118'/0'/0/0", "BIP44 path from which to derive a private key")
	cmd.Flags().Bool(flagRecover, false, "Provide seed phrase to recover existing key instead of creating")
	cmd.Flags().Bool(flagNoBackup, false, "Don't print out seed phrase (if others are watching the terminal)")
	cmd.Flags().Bool(flagDryRun, false, "Perform action, but don't add key to local keystore")
	cmd.Flags().Uint32(flagAccount, 0, "Account number for HD derivation")
	cmd.Flags().Uint32(flagIndex, 0, "Index number for HD derivation")
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
func runAddCmd(cmd *cobra.Command, args []string) error {
	var kb keys.Keybase
	var err error
	var encrypPassword string

	buf := client.BufferStdin()
	name := args[0]
	if viper.GetBool(flagDryRun) {
		// we throw this away, so don't enforce args,
		// we want to get a new random seed phrase quickly
		kb = client.MockKeyBase()
		encrypPassword = "throwing-this-key-away"
	} else {
		kb, err = GetKeyBaseWithWritePerm()
		if err != nil {
			return err
		}

		_, err := kb.Get(name)
		if err == nil {
			// account exists, ask for user confirmation
			if response, err := client.GetConfirmation(
				fmt.Sprintf("override the existing name %s", name), buf); err != nil || !response {
				return err
			}
		}

		// ask for a password when generating a local key
		if !viper.GetBool(client.FlagUseLedger) {
			encrypPassword, err = client.GetCheckPassword(
				"Enter a passphrase to encrypt your key to disk:",
				"Repeat the passphrase:", buf)
			if err != nil {
				return err
			}
		}
	}

	interactive := viper.GetBool(flagInteractive)
	flags := cmd.Flags()
	bipFlag := flags.Lookup(flagBIP44Path)

	bip44Params, err := getBIP44ParamsAndPath(bipFlag.Value.String(), bipFlag.Changed || !interactive)
	if err != nil {
		return err
	}

	// If we're using ledger, only thing we need is the path. So generate key and
	// we're done.
	if viper.GetBool(client.FlagUseLedger) {
		account := uint32(viper.GetInt(flagAccount))
		index := uint32(viper.GetInt(flagIndex))
		path := ccrypto.DerivationPath{44, 118, account, 0, index}
		info, err := kb.CreateLedger(name, path, keys.Secp256k1)
		if err != nil {
			return err
		}

		printCreate(info, "")
		return nil
	}

	// Recover key from seed passphrase
	if viper.GetBool(flagRecover) {
		seed, err := client.GetSeed(
			"Enter your recovery seed phrase:", buf)
		if err != nil {
			return err
		}
		info, err := kb.CreateKey(name, seed, encrypPassword)
		if err != nil {
			return err
		}
		// print out results without the seed phrase
		viper.Set(flagNoBackup, true)
		printCreate(info, "")
		return nil
	}

	var mnemonic string
	if interactive {
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
	if interactive {
		bip39Passphrase, err = client.GetString(
			"Enter your bip39 passphrase. This is combined with the mnemonic to derive the seed. "+
				"Most users should just hit enter to use the default, \"\"", buf)
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

	info, err := kb.Derive(name, mnemonic, bip39Passphrase, encrypPassword, *bip44Params)
	if err != nil {
		return err
	}
	printCreate(info, mnemonic)
	return nil
}

func getBIP44ParamsAndPath(path string, flagSet bool) (*hd.BIP44Params, error) {
	buf := client.BufferStdin()
	bip44Path := path

	// if it wasn't set in the flag, give it a chance to overide interactively
	if !flagSet {
		var err error

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

func printCreate(info keys.Info, seed string) {
	output := viper.Get(cli.OutputFlag)
	switch output {
	case "text":
		fmt.Fprintln(os.Stderr, "")
		printKeyInfo(info, Bech32KeyOutput)

		// print seed unless requested not to.
		if !viper.GetBool(client.FlagUseLedger) && !viper.GetBool(flagNoBackup) {
			fmt.Fprintln(os.Stderr, "\n**Important** write this seed phrase in a safe place.")
			fmt.Fprintln(os.Stderr, "It is the only way to recover your account if you ever forget your password.")
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, seed)
		}
	case "json":
		out, err := Bech32KeyOutput(info)
		if err != nil {
			panic(err)
		}
		if !viper.GetBool(flagNoBackup) {
			out.Seed = seed
		}
		var jsonString []byte
		if viper.GetBool(client.FlagIndentResponse) {
			jsonString, err = cdc.MarshalJSONIndent(out, "", "  ")
		} else {
			jsonString, err = cdc.MarshalJSON(out)
		}
		if err != nil {
			panic(err) // really shouldn't happen...
		}
		fmt.Fprintln(os.Stderr, string(jsonString))
	default:
		panic(fmt.Sprintf("I can't speak: %s", output))
	}
}

// function to just a new seed to display in the UI before actually persisting it in the keybase
func getSeed(algo keys.SigningAlgo) string {
	kb := client.MockKeyBase()
	pass := "throwing-this-key-away"
	name := "inmemorykey"
	_, seed, _ := kb.CreateMnemonic(name, keys.English, pass, algo)
	return seed
}

func printPrefixed(msg string) {
	fmt.Fprintln(os.Stderr, msg)
}

func printStep() {
	printPrefixed("-------------------------------------")
}

/////////////////////////////
// REST

// new key request REST body
type NewKeyBody struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Seed     string `json:"seed"`
}

// add new key REST handler
func AddNewKeyRequestHandler(indent bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var kb keys.Keybase
		var m NewKeyBody

		kb, err := GetKeyBaseWithWritePerm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		err = json.Unmarshal(body, &m)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		if m.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			err = errMissingName()
			w.Write([]byte(err.Error()))
			return
		}
		if m.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			err = errMissingPassword()
			w.Write([]byte(err.Error()))
			return
		}

		// check if already exists
		infos, err := kb.List()
		for _, info := range infos {
			if info.GetName() == m.Name {
				w.WriteHeader(http.StatusConflict)
				err = errKeyNameConflict(m.Name)
				w.Write([]byte(err.Error()))
				return
			}
		}

		// create account
		seed := m.Seed
		if seed == "" {
			seed = getSeed(keys.Secp256k1)
		}
		info, err := kb.CreateKey(m.Name, seed, m.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		keyOutput, err := Bech32KeyOutput(info)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		keyOutput.Seed = seed

		PostProcessResponse(w, cdc, keyOutput, indent)
	}
}

// Seed REST request handler
func SeedRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	algoType := vars["type"]
	// algo type defaults to secp256k1
	if algoType == "" {
		algoType = "secp256k1"
	}
	algo := keys.SigningAlgo(algoType)

	seed := getSeed(algo)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(seed))
}

// RecoverKeyBody is recover key request REST body
type RecoverKeyBody struct {
	Password string `json:"password"`
	Seed     string `json:"seed"`
}

// RecoverRequestHandler performs key recover request
func RecoverRequestHandler(indent bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		var m RecoverKeyBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		err = cdc.UnmarshalJSON(body, &m)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		if name == "" {
			w.WriteHeader(http.StatusBadRequest)
			err = errMissingName()
			w.Write([]byte(err.Error()))
			return
		}
		if m.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			err = errMissingPassword()
			w.Write([]byte(err.Error()))
			return
		}
		if m.Seed == "" {
			w.WriteHeader(http.StatusBadRequest)
			err = errMissingSeed()
			w.Write([]byte(err.Error()))
			return
		}

		kb, err := GetKeyBaseWithWritePerm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		// check if already exists
		infos, err := kb.List()
		for _, info := range infos {
			if info.GetName() == name {
				w.WriteHeader(http.StatusConflict)
				err = errKeyNameConflict(name)
				w.Write([]byte(err.Error()))
				return
			}
		}

		info, err := kb.CreateKey(name, m.Seed, m.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		keyOutput, err := Bech32KeyOutput(info)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		PostProcessResponse(w, cdc, keyOutput, indent)
	}
}
