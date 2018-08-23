package keys

import (
	"encoding/json"
	"net/http"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/gorilla/mux"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tmlibs/cli"
)

const (
	// FlagAddress is the flag for the user's address on the command line.
	FlagAddress = "address"
	// FlagPublicKey represents the user's public key on the command line.
	FlagPublicKey = "pubkey"
)

var showKeysCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show key info for the given name",
	Long:  `Return public details of one local key.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		info, err := getKey(name)
		if err != nil {
			return err
		}

		showAddress := viper.GetBool(FlagAddress)
		showPublicKey := viper.GetBool(FlagPublicKey)
		outputSet := cmd.Flag(cli.OutputFlag).Changed
		if showAddress && showPublicKey {
			return errors.New("cannot use both --address and --pubkey at once")
		}
		if outputSet && (showAddress || showPublicKey) {
			return errors.New("cannot use --output with --address or --pubkey")
		}
		if showAddress {
			printKeyAddress(info)
			return nil
		}
		if showPublicKey {
			printPubKey(info)
			return nil
		}

		printInfo(info)
		return nil
	},
}

func init() {
	showKeysCmd.Flags().Bool(FlagAddress, false, "output the address only (overrides --output)")
	showKeysCmd.Flags().Bool(FlagPublicKey, false, "output the public key only (overrides --output)")
}

func getKey(name string) (keys.Info, error) {
	kb, err := GetKeyBase()
	if err != nil {
		return nil, err
	}

	return kb.Get(name)
}

///////////////////////////
// REST

// get key REST handler
func GetKeyRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	info, err := getKey(name)
	// TODO check for the error if key actually does not exist, instead of assuming this as the reason
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}

	keyOutput, err := Bech32KeyOutput(info)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	output, err := json.MarshalIndent(keyOutput, "", "  ")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(output)
}
