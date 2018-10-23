package keys

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/cosmos/cosmos-sdk/crypto/keys/keyerror"
)

const (
	// FlagAddress is the flag for the user's address on the command line.
	FlagAddress = "address"
	// FlagPublicKey represents the user's public key on the command line.
	FlagPublicKey = "pubkey"
	// FlagBechPrefix defines a desired Bech32 prefix encoding for a key.
	FlagBechPrefix = "bech"
)

func showKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [name]",
		Short: "Show key info for the given name",
		Long:  `Return public details of one local key.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runShowCmd,
	}

	cmd.Flags().String(FlagBechPrefix, "acc", "The Bech32 prefix encoding for a key (acc|val|cons)")
	cmd.Flags().Bool(FlagAddress, false, "output the address only (overrides --output)")
	cmd.Flags().Bool(FlagPublicKey, false, "output the public key only (overrides --output)")

	return cmd
}

func runShowCmd(cmd *cobra.Command, args []string) error {
	name := args[0]

	info, err := GetKeyInfo(name)
	if err != nil {
		return err
	}

	isShowAddr := viper.GetBool(FlagAddress)
	isShowPubKey := viper.GetBool(FlagPublicKey)
	isOutputSet := cmd.Flag(cli.OutputFlag).Changed

	if isShowAddr && isShowPubKey {
		return errors.New("cannot use both --address and --pubkey at once")
	}

	if isOutputSet && (isShowAddr || isShowPubKey) {
		return errors.New("cannot use --output with --address or --pubkey")
	}

	bechKeyOut, err := getBechKeyOut(viper.GetString(FlagBechPrefix))
	if err != nil {
		return err
	}

	switch {
	case isShowAddr:
		printKeyAddress(info, bechKeyOut)
	case isShowPubKey:
		printPubKey(info, bechKeyOut)
	default:
		printKeyInfo(info, bechKeyOut)
	}

	return nil
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

///////////////////////////
// REST

// get key REST handler
func GetKeyRequestHandler(indent bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		info, err := GetKeyInfo(name)
		if keyerror.IsErrKeyNotFound(err) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		keyOutput, err := bechKeyOut(info)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		PostProcessResponse(w, cdc, keyOutput, indent)
	}
}
