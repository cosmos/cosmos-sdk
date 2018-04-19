package keys

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/tmlibs/cli"
)

const (
	flagType     = "type"
	flagRecover  = "recover"
	flagNoBackup = "no-backup"
	flagDryRun   = "dry-run"
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
	cmd.Flags().StringP(flagType, "t", "ed25519", "Type of private key (ed25519|secp256k1|ledger)")
	cmd.Flags().Bool(flagRecover, false, "Provide seed phrase to recover existing key instead of creating")
	cmd.Flags().Bool(flagNoBackup, false, "Don't print out seed phrase (if others are watching the terminal)")
	cmd.Flags().Bool(flagDryRun, false, "Perform action, but don't add key to local keystore")
	return cmd
}

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
			return errors.New("You must provide a name for the key")
		}
		name = args[0]
		kb, err = GetKeyBase()
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

		pass, err = client.GetCheckPassword(
			"Enter a passphrase for your key:",
			"Repeat the passphrase:", buf)
		if err != nil {
			return err
		}
	}

	if viper.GetBool(flagRecover) {
		seed, err := client.GetSeed(
			"Enter your recovery seed phrase:", buf)
		if err != nil {
			return err
		}
		info, err := kb.Recover(name, pass, seed)
		if err != nil {
			return err
		}
		// print out results without the seed phrase
		viper.Set(flagNoBackup, true)
		printCreate(info, "")
	} else {
		algo := keys.CryptoAlgo(viper.GetString(flagType))
		info, seed, err := kb.Create(name, pass, algo)
		if err != nil {
			return err
		}
		printCreate(info, seed)
	}
	return nil
}

// addOutput lets us json format the data
type addOutput struct {
	Key  keys.Info `json:"key"`
	Seed string    `json:"seed"`
}

func printCreate(info keys.Info, seed string) {
	output := viper.Get(cli.OutputFlag)
	switch output {
	case "text":
		printInfo(info)
		// print seed unless requested not to.
		if !viper.GetBool(flagNoBackup) {
			fmt.Println("**Important** write this seed phrase in a safe place.")
			fmt.Println("It is the only way to recover your account if you ever forget your password.")
			fmt.Println()
			fmt.Println(seed)
		}
	case "json":
		out := addOutput{Key: info}
		if !viper.GetBool(flagNoBackup) {
			out.Seed = seed
		}
		json, err := MarshalJSON(out)
		if err != nil {
			panic(err) // really shouldn't happen...
		}
		fmt.Println(string(json))
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
func AddNewKeyRequestHandler(w http.ResponseWriter, r *http.Request) {
	var kb keys.Keybase
	var m NewKeyBody

	kb, err := GetKeyBase()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &m)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if m.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("You have to specify a name for the locally stored account."))
		return
	}
	if m.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("You have to specify a password for the locally stored account."))
		return
	}
	if m.Seed == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("You have to specify a seed for the locally stored account."))
		return
	}

	// check if already exists
	infos, err := kb.List()
	for _, i := range infos {
		if i.Name == m.Name {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(fmt.Sprintf("Account with name %s already exists.", m.Name)))
			return
		}
	}

	// create account
	info, err := kb.Recover(m.Name, m.Password, m.Seed)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte(info.PubKey.Address().String()))
}

// function to just a new seed to display in the UI before actually persisting it in the keybase
func getSeed(algo keys.CryptoAlgo) string {
	kb := client.MockKeyBase()
	pass := "throwing-this-key-away"
	name := "inmemorykey"

	_, seed, _ := kb.Create(name, pass, algo)
	return seed
}

// Seed REST request handler
func SeedRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	algoType := vars["type"]
	// algo type defaults to ed25519
	if algoType == "" {
		algoType = "ed25519"
	}
	algo := keys.CryptoAlgo(algoType)

	seed := getSeed(algo)
	w.Write([]byte(seed))
}
