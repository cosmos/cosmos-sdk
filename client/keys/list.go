package keys

import (
	"net/http"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys/common"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/rest"
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
	kb, err := common.NewKeyBaseFromHomeFlag()
	if err != nil {
		return err
	}

	infos, err := kb.List()
	if err == nil {
		common.PrintInfos(infos)
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
			rest.PostProcessResponse(w, codec.Cdc, []string{}, cliCtx.Indent)
			return
		}
		keysOutput, err := common.Bech32KeysOutput(infos)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		rest.PostProcessResponse(w, codec.Cdc, keysOutput, cliCtx.Indent)
	}
}
