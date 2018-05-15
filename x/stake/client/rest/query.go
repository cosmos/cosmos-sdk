package rest

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tendermint/go-crypto/keys"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

func registerQueryRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/stake/{delegator}/bonding_status/{candidate}", bondingStatusHandlerFn("stake", cdc, kb, ctx)).Methods("GET")
}

// BondingStatusHandlerFn - http request handler to query delegator bonding status
func bondingStatusHandlerFn(storeName string, cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// read parameters
		vars := mux.Vars(r)
		delegator := vars["delegator"]
		candidate := vars["candidate"]

		bz, err := hex.DecodeString(delegator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		delegatorAddr := sdk.Address(bz)

		bz, err = hex.DecodeString(candidate)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		validatorAddr := sdk.Address(bz)

		key := stake.GetDelegationKey(delegatorAddr, validatorAddr, cdc)

		res, err := ctx.Query(key, storeName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Couldn't query bond. Error: %s", err.Error())))
			return
		}

		// the query will return empty if there is no data for this bond
		if len(res) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var bond stake.Delegation
		err = cdc.UnmarshalBinary(res, &bond)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Couldn't decode bond. Error: %s", err.Error())))
			return
		}

		output, err := cdc.MarshalJSON(bond)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}
