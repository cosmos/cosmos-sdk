package keys

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	keys "github.com/tendermint/go-crypto/keys"
)

const (
	flagExportPubKey = "export-pubkey"
)

var showKeysCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show key info for the given name",
	Long:  `Return public details of one local key.`,
	RunE:  runShowCmd,
	Args:  cobra.ExactArgs(1),
}

func init() {
	showKeysCmd.Flags().Bool(flagExportPubKey, false, "Export public key.")
}

func getKey(name string) (keys.Info, error) {
	kb, err := GetKeyBase()
	if err != nil {
		return keys.Info{}, err
	}

	return kb.Get(name)
}

// CMD

func runShowCmd(cmd *cobra.Command, args []string) error {
	info, err := getKey(args[0])
	if err != nil {
		return err
	}
	if viper.GetBool(flagExportPubKey) {
		out, err := info.PubKey.MarshalJSON()
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		return nil
	}
	printInfo(info)
	return nil
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

	keyOutput := KeyOutput{Name: info.Name, Address: info.PubKey.Address().String()}
	output, err := json.MarshalIndent(keyOutput, "", "  ")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(output)
}
