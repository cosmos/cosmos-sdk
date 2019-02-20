package rest

import (
	"net/http"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/tags"

	"github.com/gorilla/mux"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {

	// Get all delegations from a delegator
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/delegations",
		delegatorDelegationsHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Get all unbonding delegations from a delegator
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/unbonding_delegations",
		delegatorUnbondingDelegationsHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Get all staking txs (i.e msgs) from a delegator
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/txs",
		delegatorTxsHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Query all validators that a delegator is bonded to
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/validators",
		delegatorValidatorsHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Query a validator that a delegator is bonded to
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/validators/{validatorAddr}",
		delegatorValidatorHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Query a delegation between a delegator and a validator
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/delegations/{validatorAddr}",
		delegationHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Query all unbonding delegations between a delegator and a validator
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr}",
		unbondingDelegationHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Query redelegations (filters in query params)
	r.HandleFunc(
		"/staking/redelegations",
		redelegationsHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Get all validators
	r.HandleFunc(
		"/staking/validators",
		validatorsHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Get a single validator info
	r.HandleFunc(
		"/staking/validators/{validatorAddr}",
		validatorHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Get all delegations to a validator
	r.HandleFunc(
		"/staking/validators/{validatorAddr}/delegations",
		validatorDelegationsHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Get all unbonding delegations from a validator
	r.HandleFunc(
		"/staking/validators/{validatorAddr}/unbonding_delegations",
		validatorUnbondingDelegationsHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Get the current state of the staking pool
	r.HandleFunc(
		"/staking/pool",
		poolHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Get the current staking parameter values
	r.HandleFunc(
		"/staking/parameters",
		paramsHandlerFn(cliCtx, cdc),
	).Methods("GET")

}

// HTTP request handler to query a delegator delegations
func delegatorDelegationsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return queryDelegator(cliCtx, cdc, "custom/staking/delegatorDelegations")
}

// HTTP request handler to query a delegator unbonding delegations
func delegatorUnbondingDelegationsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return queryDelegator(cliCtx, cdc, "custom/staking/delegatorUnbondingDelegations")
}

// HTTP request handler to query all staking txs (msgs) from a delegator
func delegatorTxsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var typesQuerySlice []string
		vars := mux.Vars(r)
		delegatorAddr := vars["delegatorAddr"]

		_, err := sdk.AccAddressFromBech32(delegatorAddr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		node, err := cliCtx.GetNode()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
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
		var txs = []sdk.TxResponse{}
		var actions []string

		switch {
		case isBondTx:
			actions = append(actions, staking.MsgDelegate{}.Type())
		case isUnbondTx:
			actions = append(actions, staking.MsgUndelegate{}.Type())
			actions = append(actions, tags.ActionCompleteUnbonding)
		case isRedTx:
			actions = append(actions, staking.MsgBeginRedelegate{}.Type())
			actions = append(actions, tags.ActionCompleteRedelegation)
		case noQuery:
			actions = append(actions, staking.MsgDelegate{}.Type())
			actions = append(actions, staking.MsgUndelegate{}.Type())
			actions = append(actions, tags.ActionCompleteUnbonding)
			actions = append(actions, staking.MsgBeginRedelegate{}.Type())
			actions = append(actions, tags.ActionCompleteRedelegation)
		default:
			w.WriteHeader(http.StatusNoContent)
			return
		}

		for _, action := range actions {
			foundTxs, errQuery := queryTxs(node, cliCtx, cdc, action, delegatorAddr)
			if errQuery != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			}
			txs = append(txs, foundTxs...)
		}

		res, err := cdc.MarshalJSON(txs)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

// HTTP request handler to query an unbonding-delegation
func unbondingDelegationHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return queryBonds(cliCtx, cdc, "custom/staking/unbondingDelegation")
}

// HTTP request handler to query redelegations
func redelegationsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params staking.QueryRedelegationParams

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

		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := cliCtx.QueryWithData("custom/staking/redelegations", bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

// HTTP request handler to query a delegation
func delegationHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return queryBonds(cliCtx, cdc, "custom/staking/delegation")
}

// HTTP request handler to query all delegator bonded validators
func delegatorValidatorsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return queryDelegator(cliCtx, cdc, "custom/staking/delegatorValidators")
}

// HTTP request handler to get information from a currently bonded validator
func delegatorValidatorHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return queryBonds(cliCtx, cdc, "custom/staking/delegatorValidator")
}

// HTTP request handler to query list of validators
func validatorsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := cliCtx.QueryWithData("custom/staking/validators", nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

// HTTP request handler to query the validator information from a given validator address
func validatorHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return queryValidator(cliCtx, cdc, "custom/staking/validator")
}

// HTTP request handler to query all unbonding delegations from a validator
func validatorDelegationsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return queryValidator(cliCtx, cdc, "custom/staking/validatorDelegations")
}

// HTTP request handler to query all unbonding delegations from a validator
func validatorUnbondingDelegationsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return queryValidator(cliCtx, cdc, "custom/staking/validatorUnbondingDelegations")
}

// HTTP request handler to query the pool information
func poolHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := cliCtx.QueryWithData("custom/staking/pool", nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

// HTTP request handler to query the staking params values
func paramsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := cliCtx.QueryWithData("custom/staking/parameters", nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}
