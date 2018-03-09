package keys

import (
	"encoding/json"
	"fmt"
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

// REST

type NewKeyBody struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	// TODO make seed mandatory
	// Seed     string `json="seed"`
	Type string `json:"type"`
}

func AddNewKeyRequestHandler(w http.ResponseWriter, r *http.Request) {
	var kb keys.Keybase
	var m NewKeyBody

	kb, err := GetKeyBase()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&m)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	if m.Name == "" {
		w.WriteHeader(400)
		w.Write([]byte("You have to specify a name for the locally stored account."))
		return
	}

	// algo type defaults to ed25519
	if m.Type == "" {
		m.Type = "ed25519"
	}
	algo := keys.CryptoAlgo(m.Type)

	_, _, err = kb.Create(m.Name, m.Password, algo)
	// TODO handle different errors
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(200)
}

// function to just a new seed to display in the UI before actually persisting it in the keybase
func getSeed(algo keys.CryptoAlgo) string {
	kb := client.MockKeyBase()
	pass := "throwing-this-key-away"
	name := "inmemorykey"

	_, seed, _ := kb.Create(name, pass, algo)
	return seed
}

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
