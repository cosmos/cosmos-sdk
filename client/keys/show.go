package keys

import (
	"fmt"
	"net/http"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"errors"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/crypto/keys/keyerror"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// FlagAddress is the flag for the user's address on the command line.
	FlagAddress = "address"
	// FlagPublicKey represents the user's public key on the command line.
	FlagPublicKey = "pubkey"
	// FlagBechPrefix defines a desired Bech32 prefix encoding for a key.
	FlagBechPrefix = "bech"

	flagMultiSigThreshold  = "multisig-threshold"
	defaultMultiSigKeyName = "multi"
)

var _ keys.Info = (*multiSigKey)(nil)

type multiSigKey struct {
	name string
	key  crypto.PubKey
}

func (m multiSigKey) GetName() string            { return m.name }
func (m multiSigKey) GetType() keys.KeyType      { return keys.TypeLocal }
func (m multiSigKey) GetPubKey() crypto.PubKey   { return m.key }
func (m multiSigKey) GetAddress() sdk.AccAddress { return sdk.AccAddress(m.key.Address()) }

func showKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [name]",
		Short: "Show key info for the given name",
		Long:  `Return public details of one local key.`,
		Args:  cobra.MinimumNArgs(1),
		RunE:  runShowCmd,
	}

	cmd.Flags().String(FlagBechPrefix, "acc", "The Bech32 prefix encoding for a key (acc|val|cons)")
	cmd.Flags().BoolP(FlagAddress, "a", false, "output the address only (overrides --output)")
	cmd.Flags().BoolP(FlagPublicKey, "p", false, "output the public key only (overrides --output)")
	cmd.Flags().Uint(flagMultiSigThreshold, 1, "K out of N required signatures")

	return cmd
}

func runShowCmd(cmd *cobra.Command, args []string) (err error) {
	var info keys.Info

	if len(args) == 1 {
		info, err = GetKeyInfo(args[0])
		if err != nil {
			return err
		}
	} else {
		pks := make([]crypto.PubKey, len(args))
		for i, keyName := range args {
			info, err := GetKeyInfo(keyName)
			if err != nil {
				return err
			}
			pks[i] = info.GetPubKey()
		}

		multisigThreshold := viper.GetInt(flagMultiSigThreshold)
		err = validateMultisigThreshold(multisigThreshold, len(args))
		if err != nil {
			return err
		}
		multikey := multisig.NewPubKeyMultisigThreshold(multisigThreshold, pks)
		info = multiSigKey{
			name: defaultMultiSigKeyName,
			key:  multikey,
		}
	}

	isShowAddr := viper.GetBool(FlagAddress)
	isShowPubKey := viper.GetBool(FlagPublicKey)

	isOutputSet := false
	tmp := cmd.Flag(cli.OutputFlag)
	if tmp != nil {
		isOutputSet = tmp.Changed
	}

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

func validateMultisigThreshold(k, nKeys int) error {
	if k <= 0 {
		return fmt.Errorf("threshold must be a positive integer")
	}
	if nKeys < k {
		return fmt.Errorf(
			"threshold k of n multisignature: %d < %d", nKeys, k)
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
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		info, err := GetKeyInfo(name)
		if keyerror.IsErrKeyNotFound(err) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(err.Error()))
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		keyOutput, err := bechKeyOut(info)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		rest.PostProcessResponse(w, cdc, keyOutput, indent)
	}
}
