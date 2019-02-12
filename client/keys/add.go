package keys

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"errors"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/go-bip39"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/libs/cli"
)

const (
	flagInteractive = "interactive"
	flagRecover     = "recover"
	flagNoBackup    = "no-backup"
	flagDryRun      = "dry-run"
	flagAccount     = "account"
	flagIndex       = "index"
	flagMultisig    = "multisig"
	flagNoSort      = "nosort"
)

const (
	maxValidAccountValue = int(0x80000000 - 1)
	maxValidIndexalue    = int(0x80000000 - 1)
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
If run with --dry-run, a key would be generated (or recovered) but not stored to the
local keystore.
Use the --pubkey flag to add arbitrary public keys to the keystore for constructing
multisig transactions.

You can add a multisig key by passing the list of key names you want the public
key to be composed of to the --multisig flag and the minimum number of signatures
required through --multisig-threshold. The keys are sorted by address, unless
the flag --nosort is set.
`,
		Args: cobra.ExactArgs(1),
		RunE: runAddCmd,
	}
	cmd.Flags().StringSlice(flagMultisig, nil, "Construct and store a multisig public key (implies --pubkey)")
	cmd.Flags().Uint(flagMultiSigThreshold, 1, "K out of N required signatures. For use in conjunction with --multisig")
	cmd.Flags().Bool(flagNoSort, false, "Keys passed to --multisig are taken in the order they're supplied")
	cmd.Flags().String(FlagPublicKey, "", "Parse a public key in bech32 format and save it to disk")
	cmd.Flags().BoolP(flagInteractive, "i", false, "Interactively prompt user for BIP39 passphrase and mnemonic")
	cmd.Flags().Bool(client.FlagUseLedger, false, "Store a local reference to a private key on a Ledger device")
	cmd.Flags().Bool(flagRecover, false, "Provide seed phrase to recover existing key instead of creating")
	cmd.Flags().Bool(flagNoBackup, false, "Don't print out seed phrase (if others are watching the terminal)")
	cmd.Flags().Bool(flagDryRun, false, "Perform action, but don't add key to local keystore")
	cmd.Flags().Uint32(flagAccount, 0, "Account number for HD derivation")
	cmd.Flags().Uint32(flagIndex, 0, "Address index number for HD derivation")
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
func runAddCmd(_ *cobra.Command, args []string) error {
	var kb keys.Keybase
	var err error
	var encryptPassword string

	buf := client.BufferStdin()
	name := args[0]

	interactive := viper.GetBool(flagInteractive)
	showMnemonic := !viper.GetBool(flagNoBackup)

	if viper.GetBool(flagDryRun) {
		// we throw this away, so don't enforce args,
		// we want to get a new random seed phrase quickly
		kb = keys.NewInMemory()
		encryptPassword = app.DefaultKeyPass
	} else {
		kb, err = NewKeyBaseFromHomeFlag()
		if err != nil {
			return err
		}

		_, err = kb.Get(name)
		if err == nil {
			// account exists, ask for user confirmation
			if response, err2 := client.GetConfirmation(
				fmt.Sprintf("override the existing name %s", name), buf); err2 != nil || !response {
				return err2
			}
		}

		multisigKeys := viper.GetStringSlice(flagMultisig)
		if len(multisigKeys) != 0 {
			var pks []crypto.PubKey

			multisigThreshold := viper.GetInt(flagMultiSigThreshold)
			if err := validateMultisigThreshold(multisigThreshold, len(multisigKeys)); err != nil {
				return err
			}

			for _, keyname := range multisigKeys {
				k, err := kb.Get(keyname)
				if err != nil {
					return err
				}
				pks = append(pks, k.GetPubKey())
			}

			// Handle --nosort
			if !viper.GetBool(flagNoSort) {
				sort.Slice(pks, func(i, j int) bool {
					return bytes.Compare(pks[i].Address(), pks[j].Address()) < 0
				})
			}

			pk := multisig.NewPubKeyMultisigThreshold(multisigThreshold, pks)
			if _, err := kb.CreateOffline(name, pk); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Key %q saved to disk.", name)
			return nil
		}

		// ask for a password when generating a local key
		if viper.GetString(FlagPublicKey) == "" && !viper.GetBool(client.FlagUseLedger) {
			encryptPassword, err = client.GetCheckPassword(
				"Enter a passphrase to encrypt your key to disk:",
				"Repeat the passphrase:", buf)
			if err != nil {
				return err
			}
		}
	}

	if viper.GetString(FlagPublicKey) != "" {
		pk, err := sdk.GetAccPubKeyBech32(viper.GetString(FlagPublicKey))
		if err != nil {
			return err
		}
		_, err = kb.CreateOffline(name, pk)
		if err != nil {
			return err
		}
		return nil
	}

	account := uint32(viper.GetInt(flagAccount))
	index := uint32(viper.GetInt(flagIndex))

	// If we're using ledger, only thing we need is the path. So generate key and we're done.
	if viper.GetBool(client.FlagUseLedger) {
		info, err := kb.CreateLedger(name, keys.Secp256k1, account, index)
		if err != nil {
			return err
		}

		return printCreate(info, false, "")
	}

	// Get bip39 mnemonic
	var mnemonic string
	var bip39Passphrase string

	if interactive || viper.GetBool(flagRecover) {
		bip39Message := "Enter your bip39 mnemonic"
		if !viper.GetBool(flagRecover) {
			bip39Message = "Enter your bip39 mnemonic, or hit enter to generate one."
		}

		mnemonic, err = client.GetString(bip39Message, buf)
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

	if !bip39.IsMnemonicValid(mnemonic) {
		fmt.Fprintf(os.Stderr, "Error: Mnemonic is not valid")
		return nil
	}

	// override bip39 passphrase
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

	info, err := kb.CreateAccount(name, mnemonic, keys.DefaultBIP39Passphrase, encryptPassword, account, index)
	if err != nil {
		return err
	}

	// Recover key from seed passphrase
	if viper.GetBool(flagRecover) {
		// Hide mnemonic from output
		showMnemonic = false
		mnemonic = ""
	}

	return printCreate(info, showMnemonic, mnemonic)
}

func printCreate(info keys.Info, showMnemonic bool, mnemonic string) error {
	output := viper.Get(cli.OutputFlag)

	switch output {
	case OutputFormatText:
		fmt.Fprintln(os.Stderr)
		printKeyInfo(info, Bech32KeyOutput)

		// print mnemonic unless requested not to.
		if showMnemonic {
			fmt.Fprintln(os.Stderr, "\n**Important** write this mnemonic phrase in a safe place.")
			fmt.Fprintln(os.Stderr, "It is the only way to recover your account if you ever forget your password.")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, mnemonic)
		}
	case OutputFormatJSON:
		out, err := Bech32KeyOutput(info)
		if err != nil {
			return err
		}

		if showMnemonic {
			out.Mnemonic = mnemonic
		}

		var jsonString []byte
		if viper.GetBool(client.FlagIndentResponse) {
			jsonString, err = cdc.MarshalJSONIndent(out, "", "  ")
		} else {
			jsonString, err = cdc.MarshalJSON(out)
		}

		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, string(jsonString))
	default:
		return fmt.Errorf("I can't speak: %s", output)
	}

	return nil
}

/////////////////////////////
// REST

// function to just create a new seed to display in the UI before actually persisting it in the keybase
func generateMnemonic(algo keys.SigningAlgo) string {
	kb := keys.NewInMemory()
	pass := app.DefaultKeyPass
	name := "inmemorykey"
	_, seed, _ := kb.CreateMnemonic(name, keys.English, pass, algo)
	return seed
}

// CheckAndWriteErrorResponse will check for errors and return
// a given error message when corresponding
//TODO: Move to utils/rest or similar
func CheckAndWriteErrorResponse(w http.ResponseWriter, httpErr int, err error) bool {
	if err != nil {
		w.WriteHeader(httpErr)
		_, _ = w.Write([]byte(err.Error()))
		return true
	}
	return false
}

// add new key REST handler
func AddNewKeyRequestHandler(indent bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var kb keys.Keybase
		var m AddNewKey

		kb, err := NewKeyBaseFromHomeFlag()
		if CheckAndWriteErrorResponse(w, http.StatusInternalServerError, err) {
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if CheckAndWriteErrorResponse(w, http.StatusBadRequest, err) {
			return
		}

		err = json.Unmarshal(body, &m)
		if CheckAndWriteErrorResponse(w, http.StatusBadRequest, err) {
			return
		}

		// Check parameters
		if m.Name == "" {
			CheckAndWriteErrorResponse(w, http.StatusBadRequest, errMissingName())
			return
		}
		if m.Password == "" {
			CheckAndWriteErrorResponse(w, http.StatusBadRequest, errMissingPassword())
			return
		}

		mnemonic := m.Mnemonic
		// if mnemonic is empty, generate one
		if mnemonic == "" {
			mnemonic = generateMnemonic(keys.Secp256k1)
		}
		if !bip39.IsMnemonicValid(mnemonic) {
			CheckAndWriteErrorResponse(w, http.StatusBadRequest, errInvalidMnemonic())
		}

		if m.Account < 0 || m.Account > maxValidAccountValue {
			CheckAndWriteErrorResponse(w, http.StatusBadRequest, errInvalidAccountNumber())
			return
		}

		if m.Index < 0 || m.Index > maxValidIndexalue {
			CheckAndWriteErrorResponse(w, http.StatusBadRequest, errInvalidIndexNumber())
			return
		}

		_, err = kb.Get(m.Name)
		if err == nil {
			CheckAndWriteErrorResponse(w, http.StatusConflict, errKeyNameConflict(m.Name))
			return
		}

		// create account
		account := uint32(m.Account)
		index := uint32(m.Index)
		info, err := kb.CreateAccount(m.Name, mnemonic, keys.DefaultBIP39Passphrase, m.Password, account, index)
		if CheckAndWriteErrorResponse(w, http.StatusInternalServerError, err) {
			return
		}

		keyOutput, err := Bech32KeyOutput(info)
		if CheckAndWriteErrorResponse(w, http.StatusInternalServerError, err) {
			return
		}

		keyOutput.Mnemonic = mnemonic

		rest.PostProcessResponse(w, cdc, keyOutput, indent)
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
	seed := generateMnemonic(algo)

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(seed))
}

// RecoverRequestHandler performs key recover request
func RecoverRequestHandler(indent bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		var m RecoverKey

		body, err := ioutil.ReadAll(r.Body)
		if CheckAndWriteErrorResponse(w, http.StatusBadRequest, err) {
			return
		}

		err = cdc.UnmarshalJSON(body, &m)
		if CheckAndWriteErrorResponse(w, http.StatusBadRequest, err) {
			return
		}

		kb, err := NewKeyBaseFromHomeFlag()
		CheckAndWriteErrorResponse(w, http.StatusInternalServerError, err)

		if name == "" {
			CheckAndWriteErrorResponse(w, http.StatusBadRequest, errMissingName())
			return
		}
		if m.Password == "" {
			CheckAndWriteErrorResponse(w, http.StatusBadRequest, errMissingPassword())
			return
		}

		mnemonic := m.Mnemonic
		if !bip39.IsMnemonicValid(mnemonic) {
			CheckAndWriteErrorResponse(w, http.StatusBadRequest, errInvalidMnemonic())
		}

		if m.Mnemonic == "" {
			CheckAndWriteErrorResponse(w, http.StatusBadRequest, errMissingMnemonic())
			return
		}

		if m.Account < 0 || m.Account > maxValidAccountValue {
			CheckAndWriteErrorResponse(w, http.StatusBadRequest, errInvalidAccountNumber())
			return
		}

		if m.Index < 0 || m.Index > maxValidIndexalue {
			CheckAndWriteErrorResponse(w, http.StatusBadRequest, errInvalidIndexNumber())
			return
		}

		_, err = kb.Get(name)
		if err == nil {
			CheckAndWriteErrorResponse(w, http.StatusConflict, errKeyNameConflict(name))
			return
		}

		account := uint32(m.Account)
		index := uint32(m.Index)

		info, err := kb.CreateAccount(name, mnemonic, keys.DefaultBIP39Passphrase, m.Password, account, index)
		if CheckAndWriteErrorResponse(w, http.StatusInternalServerError, err) {
			return
		}

		keyOutput, err := Bech32KeyOutput(info)
		if CheckAndWriteErrorResponse(w, http.StatusInternalServerError, err) {
			return
		}

		rest.PostProcessResponse(w, cdc, keyOutput, indent)
	}
}
