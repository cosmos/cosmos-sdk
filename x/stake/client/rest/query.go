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
	r.HandleFunc(
		"/stake/{delegator}/bonding_status/{validator}",
		bondingStatusHandlerFn(ctx, "stake", cdc),
	).Methods("GET")
	r.HandleFunc(
		"/stake/validators",
		validatorsHandlerFn(ctx, "stake", cdc),
	).Methods("GET")
}

// http request handler to query delegator bonding status
func bondingStatusHandlerFn(ctx context.CoreContext, storeName string, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// read parameters
		vars := mux.Vars(r)
		delegator := vars["delegator"]
		validator := vars["validator"]

		bz, err := hex.DecodeString(delegator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		delegatorAddr := sdk.Address(bz)

		bz, err = hex.DecodeString(validator)
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

// http request handler to query list of validators
func validatorsHandlerFn(ctx context.CoreContext, storeName string, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		kvs, err := ctx.QuerySubspace(cdc, stake.ValidatorsKey, storeName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Couldn't query validators. Error: %s", err.Error())))
			return
		}

		// the query will return empty if there are no validators
		if len(kvs) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// parse out the validators
		var validators []stake.Validator
		for _, kv := range kvs {
			var validator stake.Validator
			err = cdc.UnmarshalBinary(kv.Value, &validator)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Couldn't decode validator. Error: %s", err.Error())))
				return
			}
			validators = append(validators, validator)
		}

		output, err := cdc.MarshalJSON(validators)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}
