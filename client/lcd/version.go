package lcd

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/celler/httputil"
	"errors"
	"github.com/cosmos/cosmos-sdk/client/utils"
)

// cli version REST handler endpoint
func CLIVersionRequestHandler(w http.ResponseWriter, r *http.Request) {
	v := version.GetVersion()
	w.Write([]byte(v))
}

// connected node version REST handler endpoint
func NodeVersionRequestHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		version, err := cliCtx.Query("/app/version")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Could't query version. Error: %s", err.Error())))
			return
		}

		w.Write(version)
	}
}

func CLIVersionRequest(gtx *gin.Context) {
	v := version.GetVersion()
	utils.Response(gtx,v)
}

func NodeVersionRequest(cliCtx context.CLIContext) gin.HandlerFunc {
	return func(gtx *gin.Context) {
		appVersion, err := cliCtx.Query("/app/version")
		if err != nil {
			httputil.NewError(gtx, http.StatusInternalServerError, errors.New(fmt.Sprintf("Could't query version. Error: %s", err.Error())))
			return
		}
		utils.Response(gtx,string(appVersion))
	}
}