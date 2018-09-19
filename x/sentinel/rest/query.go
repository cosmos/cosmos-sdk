package rest

import (
	"encoding/json"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sent "github.com/cosmos/cosmos-sdk/x/sentinel"
	senttype "github.com/cosmos/cosmos-sdk/x/sentinel/types"
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/wire"
)

const (
	storeName = "sentinel"
)

func querySessionHandlerFn(cdc *wire.Codec, ctx context.CoreContext, k sent.Keeper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		var clientSession senttype.Session
		sessionId := vars["sessionId"]
		res, err := ctx.QueryStore([]byte(sessionId), storeName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("couldn't query session."))
			return
		}
		err = cdc.UnmarshalBinary(res, &clientSession)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("couldn't marshall query result. Error: "))
			return
		}
		bz, err := json.Marshal(clientSession)
		w.Write(bz)
	}
}
