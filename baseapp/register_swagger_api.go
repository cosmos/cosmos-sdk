package baseapp

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
)

// RegisterSwaggerAPI - a common function which registers swagger route with API Server
func RegisterSwaggerAPI(_ client.Context, rtr *mux.Router, swaggerEnabled bool) {
	if swaggerEnabled {
		statikFS, err := fs.New()
		if err != nil {
			panic(err)
		}
		staticServer := http.FileServer(statikFS)
		rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
	}
}
