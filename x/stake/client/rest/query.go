package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/cosmos/cosmos-sdk/x/stake/tags"

	"github.com/gorilla/mux"
)

const storeName = "stake"

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {

	// Get all delegations (delegation, undelegation and redelegation) from a delegator
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}",
		delegatorHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Get all staking txs (i.e msgs) from a delegator
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}/txs",
		delegatorTxsHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Query all validators that a delegator is bonded to
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}/validators",
		delegatorValidatorsHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Query a validator that a delegator is bonded to
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}/validators/{validatorAddr}",
		delegatorValidatorHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Query a delegation between a delegator and a validator
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}/delegations/{validatorAddr}",
		delegationHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Query all unbonding delegations between a delegator and a validator
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr}",
		unbondingDelegationHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Get all validators
	r.HandleFunc(
		"/stake/validators",
		validatorsHandlerFn(cliCtx),
	).Methods("GET")

	// Get a single validator info
	r.HandleFunc(
		"/stake/validators/{addr}",
		validatorHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Get the current state of the staking pool
	r.HandleFunc(
		"/stake/pool",
		poolHandlerFn(cliCtx),
	).Methods("GET")

	// Get the current staking parameter values
	r.HandleFunc(
		"/stake/parameters",
		paramsHandlerFn(cliCtx),
	).Methods("GET")

}

// HTTP request handler to query a delegator delegations
func delegatorHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		bech32delegator := vars["delegatorAddr"]

		w.Header().Set("Content-Type", "application/json")

		delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		params := stake.QueryDelegatorParams{
			DelegatorAddr: delegatorAddr,
		}

		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := cliCtx.QueryWithData("custom/stake/delegator", bz)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))

			return
		}

		w.Write(res)
	}
}

// HTTP request handler to query all staking txs (msgs) from a delegator
func delegatorTxsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var output []byte
		var typesQuerySlice []string
		vars := mux.Vars(r)
		delegatorAddr := vars["delegatorAddr"]

		w.Header().Set("Content-Type", "application/json")

		_, err := sdk.AccAddressFromBech32(delegatorAddr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		node, err := cliCtx.GetNode()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Couldn't get current Node information. Error: %s", err.Error())))
			return
		}

		// Get values from query

		typesQuery := r.URL.Query().Get("type")
		trimmedQuery := strings.TrimSpace(typesQuery)
		if len(trimmedQuery) != 0 {
			typesQuerySlice = strings.Split(trimmedQuery, " ")
		}

		noQuery := len(typesQuerySlice) == 0
		isBondTx := contains(typesQuerySlice, "bond")
		isUnbondTx := contains(typesQuerySlice, "unbond")
		isRedTx := contains(typesQuerySlice, "redelegate")
		var txs = []tx.Info{}
		var actions []string

		switch {
		case isBondTx:
			actions = append(actions, string(tags.ActionDelegate))
		case isUnbondTx:
			actions = append(actions, string(tags.ActionBeginUnbonding))
			actions = append(actions, string(tags.ActionCompleteUnbonding))
		case isRedTx:
			actions = append(actions, string(tags.ActionBeginRedelegation))
			actions = append(actions, string(tags.ActionCompleteRedelegation))
		case noQuery:
			actions = append(actions, string(tags.ActionDelegate))
			actions = append(actions, string(tags.ActionBeginUnbonding))
			actions = append(actions, string(tags.ActionCompleteUnbonding))
			actions = append(actions, string(tags.ActionBeginRedelegation))
			actions = append(actions, string(tags.ActionCompleteRedelegation))
		default:
			w.WriteHeader(http.StatusNoContent)
			return
		}

		for _, action := range actions {
			foundTxs, errQuery := queryTxs(node, cliCtx, cdc, action, delegatorAddr)
			if errQuery != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(errQuery.Error()))
			}
			txs = append(txs, foundTxs...)
		}

		output, err = cdc.MarshalJSON(txs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(output)
	}
}

// HTTP request handler to query an unbonding-delegation
func unbondingDelegationHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32delegator := vars["delegatorAddr"]
		bech32validator := vars["validatorAddr"]

		w.Header().Set("Content-Type", "application/json")

		delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		validatorAddr, err := sdk.ValAddressFromBech32(bech32validator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		params := stake.QueryBondsParams{
			DelegatorAddr: delegatorAddr,
			ValidatorAddr: validatorAddr,
		}

		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := cliCtx.QueryWithData("custom/stake/unbondingDelegation", bz)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))

			return
		}

		w.Write(res)
	}
}

// HTTP request handler to query a bonded validator
func delegationHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// read parameters
		vars := mux.Vars(r)
		bech32delegator := vars["delegatorAddr"]
		bech32validator := vars["validatorAddr"]

		w.Header().Set("Content-Type", "application/json")

		delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		validatorAddr, err := sdk.ValAddressFromBech32(bech32validator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		params := stake.QueryBondsParams{
			DelegatorAddr: delegatorAddr,
			ValidatorAddr: validatorAddr,
		}

		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := cliCtx.QueryWithData("custom/stake/delegation", bz)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))

			return
		}

		w.Write(res)
	}
}

// HTTP request handler to query all delegator bonded validators
func delegatorValidatorsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// read parameters
		vars := mux.Vars(r)
		bech32delegator := vars["delegatorAddr"]

		w.Header().Set("Content-Type", "application/json")

		delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		params := stake.QueryDelegatorParams{
			DelegatorAddr: delegatorAddr,
		}

		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := cliCtx.QueryWithData("custom/stake/delegatorValidators", bz)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))

			return
		}

		w.Write(res)
	}
}

// HTTP request handler to get information from a currently bonded validator
func delegatorValidatorHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		bech32delegator := vars["delegatorAddr"]
		bech32validator := vars["validatorAddr"]

		w.Header().Set("Content-Type", "application/json")

		delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
		validatorAddr, err := sdk.ValAddressFromBech32(bech32validator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		params := stake.QueryBondsParams{
			DelegatorAddr: delegatorAddr,
			ValidatorAddr: validatorAddr,
		}

		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := cliCtx.QueryWithData("custom/stake/delegatorValidator", bz)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))

			return
		}

		w.Write(res)
	}
}

// HTTP request handler to query list of validators
func validatorsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		res, err := cliCtx.QueryWithData("custom/stake/validators", nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(res)
	}
}

// HTTP request handler to query the validator information from a given validator address
func validatorHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		bech32validatorAddr := vars["addr"]

		w.Header().Set("Content-Type", "application/json")

		validatorAddr, err := sdk.ValAddressFromBech32(bech32validatorAddr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		params := stake.QueryValidatorParams{
			ValidatorAddr: validatorAddr,
		}

		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := cliCtx.QueryWithData("custom/stake/validator", bz)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))

			return
		}

		w.Write(res)
	}
}

// HTTP request handler to query the pool information
func poolHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		res, err := cliCtx.QueryWithData("custom/stake/pool", nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))

			return
		}

		w.Write(res)
	}
}

// HTTP request handler to query the staking params values
func paramsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		res, err := cliCtx.QueryWithData("custom/stake/parameters", nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))

			return
		}

		w.Write(res)
	}
}
