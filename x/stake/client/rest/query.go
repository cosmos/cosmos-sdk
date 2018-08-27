package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/cosmos/cosmos-sdk/x/stake/tags"
	"github.com/cosmos/cosmos-sdk/x/stake/types"

	"github.com/gorilla/mux"
)

const storeName = "stake"

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *wire.Codec) {

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

	// Query all unbonding_delegations between a delegator and a validator
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr}",
		unbondingDelegationsHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Get all validators
	r.HandleFunc(
		"/stake/validators",
		validatorsHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Get a single validator info
	r.HandleFunc(
		"/stake/validators/{addr}",
		validatorHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Get the current state of the staking pool
	r.HandleFunc(
		"/stake/pool",
		poolHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Get the current staking parameter values
	r.HandleFunc(
		"/stake/parameters",
		paramsHandlerFn(cliCtx, cdc),
	).Methods("GET")

}

// already resolve the rational shares to not handle this in the client

// defines a delegation without type Rat for shares
type DelegationWithoutRat struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
	ValidatorAddr sdk.AccAddress `json:"validator_addr"`
	Shares        string         `json:"shares"`
	Height        int64          `json:"height"`
}

// aggregation of all delegations, unbondings and redelegations
type DelegationSummary struct {
	Delegations          []DelegationWithoutRat      `json:"delegations"`
	UnbondingDelegations []stake.UnbondingDelegation `json:"unbonding_delegations"`
	Redelegations        []stake.Redelegation        `json:"redelegations"`
}

// HTTP request handler to query a delegator delegations
func delegatorHandlerFn(cliCtx context.CLIContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var validatorAddr sdk.AccAddress
		var delegationSummary = DelegationSummary{}

		// read parameters
		vars := mux.Vars(r)
		bech32delegator := vars["delegatorAddr"]

		delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		// Get all validators using key
		validators, statusCode, errMsg, err := getBech32Validators(storeName, cliCtx, cdc)
		if err != nil {
			w.WriteHeader(statusCode)
			w.Write([]byte(fmt.Sprintf("%s%s", errMsg, err.Error())))
			return
		}

		for _, validator := range validators {
			validatorAddr = validator.Operator

			// Delegations
			delegations, statusCode, errMsg, err := getDelegatorDelegations(cliCtx, cdc, delegatorAddr, validatorAddr)
			if err != nil {
				w.WriteHeader(statusCode)
				w.Write([]byte(fmt.Sprintf("%s%s", errMsg, err.Error())))
				return
			}
			if statusCode != http.StatusNoContent {
				delegationSummary.Delegations = append(delegationSummary.Delegations, delegations)
			}

			// Undelegations
			unbondingDelegation, statusCode, errMsg, err := getDelegatorUndelegations(cliCtx, cdc, delegatorAddr, validatorAddr)
			if err != nil {
				w.WriteHeader(statusCode)
				w.Write([]byte(fmt.Sprintf("%s%s", errMsg, err.Error())))
				return
			}
			if statusCode != http.StatusNoContent {
				delegationSummary.UnbondingDelegations = append(delegationSummary.UnbondingDelegations, unbondingDelegation)
			}

			// Redelegations
			// only querying redelegations to a validator as this should give us already all relegations
			// if we also would put in redelegations from, we would have every redelegation double
			redelegations, statusCode, errMsg, err := getDelegatorRedelegations(cliCtx, cdc, delegatorAddr, validatorAddr)
			if err != nil {
				w.WriteHeader(statusCode)
				w.Write([]byte(fmt.Sprintf("%s%s", errMsg, err.Error())))
				return
			}
			if statusCode != http.StatusNoContent {
				delegationSummary.Redelegations = append(delegationSummary.Redelegations, redelegations)
			}
		}

		output, err := cdc.MarshalJSON(delegationSummary)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}

// nolint gocyclo
// HTTP request handler to query all staking txs (msgs) from a delegator
func delegatorTxsHandlerFn(cliCtx context.CLIContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var output []byte
		var typesQuerySlice []string
		vars := mux.Vars(r)
		delegatorAddr := vars["delegatorAddr"]

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
			foundTxs, errQuery := queryTxs(node, cdc, action, delegatorAddr)
			if errQuery != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("error querying transactions. Error: %s", errQuery.Error())))
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
func unbondingDelegationsHandlerFn(cliCtx context.CLIContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32delegator := vars["delegatorAddr"]
		bech32validator := vars["validatorAddr"]

		delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		validatorAddr, err := sdk.AccAddressFromBech32(bech32validator)
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

		res, err := cliCtx.QueryWithData("custom/stake/unbonding-delegation", bz)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))

			return
		}

		w.Write(res)
	}
}

// HTTP request handler to query a bonded validator
func delegationHandlerFn(cliCtx context.CLIContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// read parameters
		vars := mux.Vars(r)
		bech32delegator := vars["delegatorAddr"]
		bech32validator := vars["validatorAddr"]

		delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		validatorAddr, err := sdk.AccAddressFromBech32(bech32validator)
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

		key := stake.GetDelegationKey(delegatorAddr, validatorAddr)

		delegation, err := types.UnmarshalDelegation(cdc, key, res)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		outputDelegation := DelegationWithoutRat{
			DelegatorAddr: delegation.DelegatorAddr,
			ValidatorAddr: delegation.ValidatorAddr,
			Height:        delegation.Height,
			Shares:        delegation.Shares.String(),
		}

		output, err := cdc.MarshalJSON(outputDelegation)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}

// HTTP request handler to query all delegator bonded validators
func delegatorValidatorsHandlerFn(cliCtx context.CLIContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// read parameters
		vars := mux.Vars(r)
		bech32delegator := vars["delegatorAddr"]

		delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		params := stake.QueryAddressParams{
			AccountAddr: delegatorAddr,
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

		var validators []stake.Validator
		cdc.UnmarshalJSON(res, &validators)

		if len(validators) == 0 {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.Write(res)
	}
}

// HTTP request handler to get information from a currently bonded validator
func delegatorValidatorHandlerFn(cliCtx context.CLIContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		bech32delegator := vars["delegatorAddr"]
		bech32validator := vars["validatorAddr"]

		delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
		validatorAccAddr, err := sdk.AccAddressFromBech32(bech32validator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Error: %s", err.Error())))
			return
		}

		params := stake.QueryBondsParams{
			DelegatorAddr: delegatorAddr,
			ValidatorAddr: validatorAccAddr,
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

// TODO bech32
// http request handler to query list of validators
func validatorsHandlerFn(cliCtx context.CLIContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		res, err := cliCtx.Query("custom/stake/validators")
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
func validatorHandlerFn(cliCtx context.CLIContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		bech32validatorAddr := vars["addr"]
		valAddress, err := sdk.AccAddressFromBech32(bech32validatorAddr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("error: %s", err.Error())))
			return
		}

		params := stake.QueryAddressParams{
			AccountAddr: valAddress,
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
func poolHandlerFn(cliCtx context.CLIContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		res, err := cliCtx.Query("custom/stake/pool")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))

			return
		}

		_, err = types.UnmarshalPool(cdc, res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(res)
	}
}

// HTTP request handler to query the staking params values
func paramsHandlerFn(cliCtx context.CLIContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		res, err := cliCtx.Query("custom/stake/parameters")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))

			return
		}

		_, err = types.UnmarshalParams(cdc, res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(res)
	}
}
