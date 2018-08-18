package keys

import (
	"encoding/json"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/gin-gonic/gin"
	"github.com/cosmos/cosmos-sdk/client/httputils"
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
	keysOutput, err := Bech32KeysOutput(infos)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	output, err := json.MarshalIndent(keysOutput, "", "  ")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(output)
}

func QueryKeysRequest(gtx *gin.Context) {
	kb, err := GetKeyBase()
	if err != nil {
		httputils.NewError(gtx, http.StatusInternalServerError, err)
		return
	}
	infos, err := kb.List()
	if err != nil {
		httputils.NewError(gtx, http.StatusInternalServerError, err)
		return
	}
	// an empty list will be JSONized as null, but we want to keep the empty list
	if len(infos) == 0 {
		httputils.Response(gtx, nil)
		return
	}
	keysOutput, err := Bech32KeysOutput(infos)
	if err != nil {
		httputils.NewError(gtx, http.StatusInternalServerError, err)
		return
	}
	httputils.Response(gtx, keysOutput)
}