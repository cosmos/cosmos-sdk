package keys

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	ccrypto "github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keys"

	"github.com/tendermint/tendermint/libs/cli"
)

const (
	flagType     = "type"
	flagRecover  = "recover"
	flagNoBackup = "no-backup"
	flagDryRun   = "dry-run"
	flagAccount  = "account"
	flagIndex    = "index"
)

func addKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Create a new key, or import from seed",
		Long: `Add a public/private key pair to the key store.
If you select --seed/-s you can recover a key from the seed
phrase, otherwise, a new key will be generated.`,
		RunE: runAddCmd,
	}
	cmd.Flags().StringP(flagType, "t", "secp256k1", "Type of private key (secp256k1|ed25519)")
	cmd.Flags().Bool(client.FlagUseLedger, false, "Store a local reference to a private key on a Ledger device")
	cmd.Flags().Bool(flagRecover, false, "Provide seed phrase to recover existing key instead of creating")
	cmd.Flags().Bool(flagNoBackup, false, "Don't print out seed phrase (if others are watching the terminal)")
	cmd.Flags().Bool(flagDryRun, false, "Perform action, but don't add key to local keystore")
	cmd.Flags().Uint32(flagAccount, 0, "Account number for HD derivation")
	cmd.Flags().Uint32(flagIndex, 0, "Index number for HD derivation")
	return cmd
}

// TODO remove the above when addressing #1446
func runAddCmd(cmd *cobra.Command, args []string) error {
	var kb keys.Keybase
	var err error
	var name, pass string

	buf := client.BufferStdin()
	if viper.GetBool(flagDryRun) {
		// we throw this away, so don't enforce args,
		// we want to get a new random seed phrase quickly
		kb = client.MockKeyBase()
		pass = "throwing-this-key-away"
		name = "inmemorykey"
	} else {
		if len(args) != 1 || len(args[0]) == 0 {
			return errMissingName()
		}
		name = args[0]
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
			pass, err = client.GetCheckPassword(
				"Enter a passphrase for your key:",
				"Repeat the passphrase:", buf)
			if err != nil {
				return err
			}
		}
	}

	if viper.GetBool(client.FlagUseLedger) {
		account := uint32(viper.GetInt(flagAccount))
		index := uint32(viper.GetInt(flagIndex))
		path := ccrypto.DerivationPath{44, 118, account, 0, index}
		algo := keys.SigningAlgo(viper.GetString(flagType))
		info, err := kb.CreateLedger(name, path, algo)
		if err != nil {
			return err
		}
		printCreate(info, "")
	} else if viper.GetBool(flagRecover) {
		seed, err := client.GetSeed(
			"Enter your recovery seed phrase:", buf)
		if err != nil {
			return err
		}
		info, err := kb.CreateKey(name, seed, pass)
		if err != nil {
			return err
		}
		// print out results without the seed phrase
		viper.Set(flagNoBackup, true)
		printCreate(info, "")
	} else {
		algo := keys.SigningAlgo(viper.GetString(flagType))
		info, seed, err := kb.CreateMnemonic(name, keys.English, pass, algo)
		if err != nil {
			return err
		}
		printCreate(info, seed)
	}
	return nil
}

func printCreate(info keys.Info, seed string) {
	output := viper.Get(cli.OutputFlag)
	switch output {
	case "text":
		printKeyInfo(info, Bech32KeyOutput)

		// print seed unless requested not to.
		if !viper.GetBool(client.FlagUseLedger) && !viper.GetBool(flagNoBackup) {
			fmt.Println("**Important** write this seed phrase in a safe place.")
			fmt.Println("It is the only way to recover your account if you ever forget your password.")
			fmt.Println()
			fmt.Println(seed)
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
		fmt.Println(string(jsonString))
	default:
		panic(fmt.Sprintf("I can't speak: %s", output))
	}
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

// function to just a new seed to display in the UI before actually persisting it in the keybase
func getSeed(algo keys.SigningAlgo) string {
	kb := client.MockKeyBase()
	pass := "throwing-this-key-away"
	name := "inmemorykey"
	_, seed, _ := kb.CreateMnemonic(name, keys.English, pass, algo)
	return seed
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
