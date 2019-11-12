package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/ibc/channel/channel/open-init", channelOpenInitHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc("/ibc/channel/channel/open-try", channelOpenTryHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/ibc/channel/channels/{%s}/open-ack", RestChannelID), channelOpenAckHandlerFn(cliCtx)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("/ibc/channel/channels/{%s}/open-confirm", RestChannelID), channelOpenConfirmHandlerFn(cliCtx)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("/ibc/channel/channels/{%s}/close-init", RestChannelID), channelCloseInitHandlerFn(cliCtx)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("/ibc/channel/channels/{%s}/close-confirm", RestChannelID), channelCloseConfirmHandlerFn(cliCtx)).Methods("PUT")
}

func channelOpenInitHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func channelOpenTryHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func channelOpenAckHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func channelOpenConfirmHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func channelCloseInitHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func channelCloseConfirmHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
