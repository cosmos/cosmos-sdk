package rest

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

func registerQueryRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec) {
	r.HandleFunc("/stake/{delegator}/bonding_status/{candidate}", bondingStatusHandlerFn("stake", cdc, ctx)).Methods("GET")
	r.HandleFunc("/stake/candidates", candidatesHandlerFn("stake", cdc, ctx)).Methods("GET")
}

// bondingStatusHandlerFn - http request handler to query delegator bonding status
func bondingStatusHandlerFn(storeName string, cdc *wire.Codec, ctx context.CoreContext) http.HandlerFunc {
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

// candidatesHandlerFn - http request handler to query list of candidates
func candidatesHandlerFn(storeName string, cdc *wire.Codec, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := ctx.QuerySubspace(cdc, stake.CandidatesKey, storeName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Couldn't query candidates. Error: %s", err.Error())))
			return
		}

		if len(res) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		candidates := make(stake.Candidates, 0, len(res))
		for _, kv := range res {
			var candidate stake.Candidate
			err = cdc.UnmarshalBinary(kv.Value, &candidate)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Couldn't decode candidate. Error: %s", err.Error())))
				return
			}
			candidates = append(candidates, candidate)
		}

		output, err := cdc.MarshalJSON(candidates)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}
