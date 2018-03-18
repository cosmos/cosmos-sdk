package rest

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/builder"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

type commander struct {
	storeName string
	cdc       *wire.Codec
	parser    sdk.ParseAccount
}

func QueryAccountRequestHandler(storeName string, cdc *wire.Codec, parser sdk.ParseAccount) func(http.ResponseWriter, *http.Request) {
	c := commander{storeName, cdc, parser}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		addr := vars["address"]

		bz, err := hex.DecodeString(addr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		key := sdk.Address(bz)

		res, err := builder.Query(key, c.storeName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Could't query account. Error: %s", err.Error())))
			return
		}

		// the query will return empty if there is no data for this account
		if len(res) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// parse out the value
		account, err := c.parser(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Could't parse query result. Result: %s. Error: %s", res, err.Error())))
			return
		}

		// print out whole account
		output, err := json.MarshalIndent(account, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Could't marshall query result. Error: %s", err.Error())))
			return
		}

		w.Write(output)
	}
}
