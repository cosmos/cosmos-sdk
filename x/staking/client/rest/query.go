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

// delegatorDelegationsHandlerFn implements a delegator delegations query route
//
// @Summary Query all delegations from a delegator
// @Description Query all delegations from a single delegator address
// @Tags staking
// @Produce json
// @Param delegatorAddr path string true "The delegator address"
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.delegatorDelegations
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/delegators/{delegatorAddr}/delegations [get]
func delegatorDelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryDelegator(cliCtx, fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryDelegatorDelegations))
}

// delegatorUnbondingDelegationsHandlerFn implements a delegator unbonding delegations query route
//
// @Summary Query all unbonding delegations from a delegator
// @Description Query all unbonding delegations from a single delegator address
// @Tags staking
// @Produce json
// @Param delegatorAddr path string true "The delegator address"
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.delegatorUnbondingDelegations
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/delegators/{delegatorAddr}/unbonding_delegations [get]
func delegatorUnbondingDelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryDelegator(cliCtx, "custom/staking/delegatorUnbondingDelegations")
}

// delegatorTxsHandlerFn implements a delegator transactions query route
//
// @Summary Query all staking transactions from a delegator
// @Description Query all staking transactions from a single delegator address
// @Description NOTE: In order to query staking transactions, the transaction
// @Description record must be available otherwise the query will fail. This requires a
// @Description node that is not pruning transaction history
// @Tags staking
// @Produce json
// @Param delegatorAddr path string true "The delegator address"
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Param type query string false "Type of staking transaction, either (bond | unbond | redelegate)"
// @Success 200 {object} rest.delegatorTxs
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/delegators/{delegatorAddr}/txs [get]
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

		rest.PostProcessResponseBare(w, cliCtx, res)
	}
}

// unbondingDelegationHandlerFn implements a delegator/validator unbonding delegations query route
//
// @Summary Query all unbonding delegations from a delegator/validator pair
// @Description Query all unbonding delegations from a single delegator/validator pair
// @Tags staking
// @Produce json
// @Param delegatorAddr path string true "The delegator address"
// @Param validatorAddr path string true "The validator address"
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.unbondingDelegation
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr} [get]
func unbondingDelegationHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryBonds(cliCtx, "custom/staking/unbondingDelegation")
}

// redelegationsHandlerFn implements a redelegations query route
//
// @Summary Query all redelegations with filters
// @Description Query all redelegations filtered by delegator, validator_from, and/or validator_to
// @Tags staking
// @Produce json
// @Param delegator query string false "The delegator address"
// @Param validator_from query string false "The source validator address"
// @Param validator_to query string false "The destination validator address"
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.redelegations
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/redelegations [get]
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

// delegationHandlerFn implements an individual delegation query route
//
// @Summary Query an individual delegation
// @Description Query an individual delegation
// @Tags staking
// @Produce json
// @Param delegatorAddr path string true "The delegator address"
// @Param validatorAddr path string true "The validator address"
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.delegation
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/delegators/{delegatorAddr}/delegations/{validatorAddr} [get]
func delegationHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryBonds(cliCtx, fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryDelegation))
}

// delegatorValidatorsHandlerFn implements a delegator's bonded validators query route
//
// @Summary Query a delegator's bonded validators
// @Description Query a delegator's bonded validators
// @Tags staking
// @Produce json
// @Param delegatorAddr path string true "The delegator address"
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.delegatorValidators
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/delegators/{delegatorAddr}/validators [get]
func delegatorValidatorsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryDelegator(cliCtx, "custom/staking/delegatorValidators")
}

// delegatorValidatorHandlerFn implements a delegator's bonded validator query route
//
// @Summary Query a delegator's bonded validator
// @Description Query a delegator's bonded validator
// @Tags staking
// @Produce json
// @Param delegatorAddr path string true "The delegator address"
// @Param validatorAddr path string true "The validator address"
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.delegatorValidator
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/delegators/{delegatorAddr}/validators/{validatorAddr} [get]
func delegatorValidatorHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryBonds(cliCtx, "custom/staking/delegatorValidator")
}

// validatorsHandlerFn implements a validators query route
//
// @Summary Query validators
// @Description Query paginated validators filtering by status
// @Tags staking
// @Produce json
// @Param status query string false "The validator status (bonded | unbonded | unbonding)" default(bonded)
// @Param page query int false "The page number to query" default(1)
// @Param limit query int false "The number of results per page" default(100)
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.validators
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/validators [get]
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

// validatorHandlerFn implements a validator query route.
//
// @Summary Query a validator
// @Description Query a validator by operator address
// @Tags staking
// @Produce json
// @Param validatorAddr path string true "The validator address"
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.validator
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/validators/{validatorAddr} [get]
func validatorHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryValidator(cliCtx, "custom/staking/validator")
}

// validatorDelegationsHandlerFn implements a validator delegations query route.
//
// @Summary Query a validator's delegations
// @Description Query a validator's delegations by operator address
// @Tags staking
// @Produce json
// @Param validatorAddr path string true "The validator address"
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} validatorDelegations
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/validators/{validatorAddr}/delegations [get]
func validatorDelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryValidator(cliCtx, fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryValidatorDelegations))
}

// validatorUnbondingDelegationsHandlerFn implements a validator unbonding delegations query route
//
// @Summary Query a validator's unbonding delegations
// @Description Query a validator's unbonding delegations by operator address
// @Tags staking
// @Produce json
// @Param validatorAddr path string true "The validator address"
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.validatorUnbondingDelegations
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/validators/{validatorAddr}/unbonding_delegations [get]
func validatorUnbondingDelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryValidator(cliCtx, "custom/staking/validatorUnbondingDelegations")
}

// poolHandlerFn implements a staking pool query route
//
// @Summary Query the staking pool
// @Description Query the staking pools
// @Tags staking
// @Produce json
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.pool
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/pool [get]
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

// paramsHandlerFn implements a staking parameters query route.
//
// @Summary Query the staking parameters
// @Description Query the staking parameters
// @Tags staking
// @Produce json
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.params
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/parameters [get]
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
