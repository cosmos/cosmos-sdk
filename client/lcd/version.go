package lcd

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/celler/httputil"
	"github.com/cosmos/cosmos-sdk/client/httputils"
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

// CLIVersionRequest is the handler of getting rest server version
func CLIVersionRequest(gtx *gin.Context) {
	v := version.GetVersion()
	httputils.NormalResponse(gtx, []byte(v))
}

// NodeVersionRequest is the handler of getting connected node version
func NodeVersionRequest(cliCtx context.CLIContext) gin.HandlerFunc {
	return func(gtx *gin.Context) {
		appVersion, err := cliCtx.Query("/app/version")
		if err != nil {
			httputil.NewError(gtx, http.StatusInternalServerError, fmt.Errorf("could't query version. error: %s", err.Error()))
			return
		}
		httputils.NormalResponse(gtx, appVersion)
	}
}