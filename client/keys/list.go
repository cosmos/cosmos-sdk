package keys

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/spf13/cobra"
)

func listKeysCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all keys",
		Long: `Return a list of all public keys stored by this key manager
along with their associated name and address.`,
		RunE: runListCmd,
	}
}

func runListCmd(cmd *cobra.Command, args []string) error {
	kb, err := NewKeyBaseFromHomeFlag()
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
func QueryKeysRequestHandler(indent bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		kb, err := NewKeyBaseFromHomeFlag()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		infos, err := kb.List()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		// an empty list will be JSONized as null, but we want to keep the empty list
		if len(infos) == 0 {
			rest.PostProcessResponse(w, cdc, []string{}, indent)
			return
		}
		keysOutput, err := Bech32KeysOutput(infos)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		rest.PostProcessResponse(w, cdc, keysOutput, indent)
	}
}
