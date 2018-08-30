package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sent "github.com/cosmos/cosmos-sdk/x/sentinel"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/wire"
)

const (
	storeName = "sentinel"
)


func queryvpnHandlerFn(cdc *wire.Codec, ctx context.CoreContext, k sent.Keeper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		address := vars["address"]

		addr, err := sdk.AccAddressFromBech32(address)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := ctx.QueryStore([]byte(addr), storeName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("couldn't query account. Error: %s", err.Error())))
			return
		}

		if len(res) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		account, err := k.NewMsgDecoder(res)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("the vpn account unmarshal failed. Error: %s", err.Error())))
			return
		}

		output, err := cdc.MarshalJSON(account)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("couldn't marshall query result. Error: %s", err.Error())))
			return
		}

		w.Write(output)
	}
}
