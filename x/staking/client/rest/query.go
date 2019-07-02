package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	// Get all delegations from a delegator
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/delegations",
		delegatorDelegationsHandlerFn(cliCtx),
	).Methods("GET")

	// Get all unbonding delegations from a delegator
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/unbonding_delegations",
		delegatorUnbondingDelegationsHandlerFn(cliCtx),
	).Methods("GET")

	// Get all staking txs (i.e msgs) from a delegator
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/txs",
		delegatorTxsHandlerFn(cliCtx),
	).Methods("GET")

	// Query all validators that a delegator is bonded to
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/validators",
		delegatorValidatorsHandlerFn(cliCtx),
	).Methods("GET")

	// Query a validator that a delegator is bonded to
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/validators/{validatorAddr}",
		delegatorValidatorHandlerFn(cliCtx),
	).Methods("GET")

	// Query a delegation between a delegator and a validator
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/delegations/{validatorAddr}",
		delegationHandlerFn(cliCtx),
	).Methods("GET")

	// Query all unbonding delegations between a delegator and a validator
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr}",
		unbondingDelegationHandlerFn(cliCtx),
	).Methods("GET")

	// Query redelegations (filters in query params)
	r.HandleFunc(
		"/staking/redelegations",
		redelegationsHandlerFn(cliCtx),
	).Methods("GET")

	// Get all validators
	r.HandleFunc(
		"/staking/validators",
		validatorsHandlerFn(cliCtx),
	).Methods("GET")

	// Get a single validator info
	r.HandleFunc(
		"/staking/validators/{validatorAddr}",
		validatorHandlerFn(cliCtx),
	).Methods("GET")

	// Get all delegations to a validator
	r.HandleFunc(
		"/staking/validators/{validatorAddr}/delegations",
		validatorDelegationsHandlerFn(cliCtx),
	).Methods("GET")

	// Get all unbonding delegations from a validator
	r.HandleFunc(
		"/staking/validators/{validatorAddr}/unbonding_delegations",
		validatorUnbondingDelegationsHandlerFn(cliCtx),
	).Methods("GET")

	// Get the current state of the staking pool
	r.HandleFunc(
		"/staking/pool",
		poolHandlerFn(cliCtx),
	).Methods("GET")

	// Get the current staking parameter values
	r.HandleFunc(
		"/staking/parameters",
		paramsHandlerFn(cliCtx),
	).Methods("GET")

}

// HTTP request handler to query a delegator delegations
func delegatorDelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryDelegator(cliCtx, fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryDelegatorDelegations))
}

// HTTP request handler to query a delegator unbonding delegations
func delegatorUnbondingDelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryDelegator(cliCtx, "custom/staking/delegatorUnbondingDelegations")
}

// HTTP request handler to query all staking txs (msgs) from a delegator
func delegatorTxsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var typesQuerySlice []string
		vars := mux.Vars(r)
		delegatorAddr := vars["delegatorAddr"]

		_, err := sdk.AccAddressFromBech32(delegatorAddr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		typesQuery := r.URL.Query().Get("type")
		trimmedQuery := strings.TrimSpace(typesQuery)
		if len(trimmedQuery) != 0 {
			typesQuerySlice = strings.Split(trimmedQuery, " ")
		}

		noQuery := len(typesQuerySlice) == 0
		isBondTx := contains(typesQuerySlice, "bond")
		isUnbondTx := contains(typesQuerySlice, "unbond")
		isRedTx := contains(typesQuerySlice, "redelegate")

		var (
			txs     []*sdk.SearchTxsResult
			actions []string
		)

		switch {
		case isBondTx:
			actions = append(actions, types.MsgDelegate{}.Type())

		case isUnbondTx:
			actions = append(actions, types.MsgUndelegate{}.Type())

		case isRedTx:
			actions = append(actions, types.MsgBeginRedelegate{}.Type())

		case noQuery:
			actions = append(actions, types.MsgDelegate{}.Type())
			actions = append(actions, types.MsgUndelegate{}.Type())
			actions = append(actions, types.MsgBeginRedelegate{}.Type())

		default:
			w.WriteHeader(http.StatusNoContent)
			return
		}

		for _, action := range actions {
			foundTxs, errQuery := queryTxs(cliCtx, action, delegatorAddr)
			if errQuery != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, errQuery.Error())
			}
			txs = append(txs, foundTxs)
		}

		res, err := cliCtx.Codec.MarshalJSON(txs)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// HTTP request handler to query an unbonding-delegation
func unbondingDelegationHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryBonds(cliCtx, "custom/staking/unbondingDelegation")
}

// HTTP request handler to query redelegations
func redelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params types.QueryRedelegationParams

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		bechDelegatorAddr := r.URL.Query().Get("delegator")
		bechSrcValidatorAddr := r.URL.Query().Get("validator_from")
		bechDstValidatorAddr := r.URL.Query().Get("validator_to")

		if len(bechDelegatorAddr) != 0 {
			delegatorAddr, err := sdk.AccAddressFromBech32(bechDelegatorAddr)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			params.DelegatorAddr = delegatorAddr
		}

		if len(bechSrcValidatorAddr) != 0 {
			srcValidatorAddr, err := sdk.ValAddressFromBech32(bechSrcValidatorAddr)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			params.SrcValidatorAddr = srcValidatorAddr
		}

		if len(bechDstValidatorAddr) != 0 {
			dstValidatorAddr, err := sdk.ValAddressFromBech32(bechDstValidatorAddr)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			params.DstValidatorAddr = dstValidatorAddr
		}

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData("custom/staking/redelegations", bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// HTTP request handler to query a delegation
func delegationHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryBonds(cliCtx, fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryDelegation))
}

// HTTP request handler to query all delegator bonded validators
func delegatorValidatorsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryDelegator(cliCtx, "custom/staking/delegatorValidators")
}

// HTTP request handler to get information from a currently bonded validator
func delegatorValidatorHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryBonds(cliCtx, "custom/staking/delegatorValidator")
}

// HTTP request handler to query list of validators
func validatorsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		status := r.FormValue("status")
		if status == "" {
			status = sdk.BondStatusBonded
		}

		params := types.NewQueryValidatorsParams(page, limit, status)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryValidators)
		res, height, err := cliCtx.QueryWithData(route, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// HTTP request handler to query the validator information from a given validator address
func validatorHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryValidator(cliCtx, "custom/staking/validator")
}

// HTTP request handler to query all unbonding delegations from a validator
func validatorDelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryValidator(cliCtx, fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryValidatorDelegations))
}

// HTTP request handler to query all unbonding delegations from a validator
func validatorUnbondingDelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryValidator(cliCtx, "custom/staking/validatorUnbondingDelegations")
}

// HTTP request handler to query the pool information
func poolHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.QueryWithData("custom/staking/pool", nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// HTTP request handler to query the staking params values
func paramsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.QueryWithData("custom/staking/parameters", nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
