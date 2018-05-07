package keys

import (
	"encoding/json"
	"net/http"

	"github.com/spf13/cobra"
)

// CMD

// listKeysCmd represents the list command
var listKeysCmd = &cobra.Command{
	Use:   "list",
	Short: "List all keys",
	Long: `Return a list of all public keys stored by this key manager
along with their associated name and address.`,
	RunE: runListCmd,
}

func runListCmd(cmd *cobra.Command, args []string) error {
	kb, err := GetKeyBase()
	if err != nil {
		return err
	}

	infos, err := kb.List()
	if err == nil {
		printInfos(infos)
	}
	return err
}

/////////////////////////
// REST

// query key list REST handler
func QueryKeysRequestHandler(w http.ResponseWriter, r *http.Request) {
	kb, err := GetKeyBase()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	infos, err := kb.List()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	// an empty list will be JSONized as null, but we want to keep the empty list
	if len(infos) == 0 {
		w.Write([]byte("[]"))
		return
	}
	keysOutput := make([]KeyOutput, len(infos))
	for i, info := range infos {
		keysOutput[i] = KeyOutput{Name: info.Name, Address: info.PubKey.Address().String()}
	}
	output, err := json.MarshalIndent(keysOutput, "", "  ")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(output)
}
