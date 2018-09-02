package keys

import (
	"fmt"
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
	// FlagBechPrefix defines a desired Bech32 prefix encoding for a key
	FlagBechPrefix = "bech"
)

func showKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [name]",
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

			bechKeyOut, err := getBechKeyOut(viper.GetString(FlagBechPrefix))
			if err != nil {
				return err
			}

			switch {
			case showAddress:
				printKeyAddress(info, bechKeyOut)
			case showPublicKey:
				printPubKey(info, bechKeyOut)
			default:
				printKeyInfo(info, bechKeyOut)
			}
			return nil
		},
	}

	cmd.Flags().String(FlagBechPrefix, "acc", "The Bech32 prefix encoding for a key (acc|val|cons)")
	cmd.Flags().Bool(FlagAddress, false, "output the address only (overrides --output)")
	cmd.Flags().Bool(FlagPublicKey, false, "output the public key only (overrides --output)")

	return cmd
}

func getBechKeyOut(bechPrefix string) (bechKeyOutFn, error) {
	switch bechPrefix {
	case "acc":
		return Bech32KeyOutput, nil
	case "val":
		return Bech32ValKeyOutput, nil
	case "cons":
		return Bech32ConsKeyOutput, nil
	}

	return nil, fmt.Errorf("invalid Bech32 prefix encoding provided: %s", bechPrefix)
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
	bechPrefix := r.URL.Query().Get(FlagBechPrefix)

	if bechPrefix == "" {
		bechPrefix = "acc"
	}

	bechKeyOut, err := getBechKeyOut(bechPrefix)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	info, err := getKey(name)
	// TODO: check for the error if key actually does not exist, instead of
	// assuming this as the reason
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	keyOutput, err := bechKeyOut(info)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	output, err := cdc.MarshalJSONIndent(keyOutput, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(output)
}
