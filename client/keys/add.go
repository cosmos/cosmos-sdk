package keys

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
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
