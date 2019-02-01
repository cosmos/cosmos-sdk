package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/common"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router,
	cdc *codec.Codec, queryRoute string) {

	// Get the total rewards balance from all delegations
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards",
		delegatorRewardsHandlerFn(cliCtx, cdc, queryRoute),
	).Methods("GET")

	// Query a delegation reward
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards/{validatorAddr}",
		delegationRewardsHandlerFn(cliCtx, cdc, queryRoute),
	).Methods("GET")

	// Get the rewards withdrawal address
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/withdraw_address",
		delegatorWithdrawalAddrHandlerFn(cliCtx, cdc, queryRoute),
	).Methods("GET")

	// Validator distribution information
	r.HandleFunc(
		"/distribution/validators/{validatorAddr}",
		validatorInfoHandlerFn(cliCtx, cdc, queryRoute),
	).Methods("GET")

	// Commission and self-delegation rewards of a single a validator
	r.HandleFunc(
		"/distribution/validators/{validatorAddr}/rewards",
		validatorRewardsHandlerFn(cliCtx, cdc, queryRoute),
	).Methods("GET")

	// Get the current distribution parameter values
	r.HandleFunc(
		"/distribution/parameters",
		paramsHandlerFn(cliCtx, cdc, queryRoute),
	).Methods("GET")

	// Get the current distribution pool
	r.HandleFunc(
		"/distribution/outstanding_rewards",
		outstandingRewardsHandlerFn(cliCtx, cdc, queryRoute),
	).Methods("GET")
}

// HTTP request handler to query delegators rewards
func delegatorRewardsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// query for rewards from a particular delegator
		res, abort := checkResponseQueryRewards(w, cliCtx, cdc, queryRoute,
			mux.Vars(r)["delegatorAddr"], "")
		if abort {
			return
		}

		utils.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

// HTTP request handler to query a delegation rewards
func delegationRewardsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// query for rewards from a particular delegation
		res, abort := checkResponseQueryRewards(w, cliCtx, cdc, queryRoute,
			mux.Vars(r)["delegatorAddr"], mux.Vars(r)["validatorAddr"])
		if abort {
			return
		}

		utils.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

// HTTP request handler to query a delegation rewards
func delegatorWithdrawalAddrHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		delegatorAddr, err := sdk.AccAddressFromBech32(mux.Vars(r)["delegatorAddr"])
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		params := distribution.NewQueryDelegatorWithdrawAddrParams(delegatorAddr)
		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/withdraw_addr", queryRoute), bz)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		utils.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

// ValidatorDistInfo defines the properties of
// validator distribution information response.
type ValidatorDistInfo struct {
	OperatorAddress     sdk.AccAddress                       `json:"operator_addr"`
	SelfBondRewards     sdk.DecCoins                         `json:"self_bond_rewards"`
	ValidatorCommission types.ValidatorAccumulatedCommission `json:"val_commission"`
}

// NewValidatorDistInfo creates a new instance of ValidatorDistInfo.
func NewValidatorDistInfo(operatorAddr sdk.AccAddress, rewards sdk.DecCoins,
	commission types.ValidatorAccumulatedCommission) ValidatorDistInfo {
	return ValidatorDistInfo{
		OperatorAddress:     operatorAddr,
		SelfBondRewards:     rewards,
		ValidatorCommission: commission,
	}
}

// HTTP request handler to query validator's distribution information
func validatorInfoHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		valAddr := mux.Vars(r)["validatorAddr"]
		validatorAddr, err := sdk.ValAddressFromBech32(valAddr)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// query commission
		commissionRes, err := common.QueryValidatorCommission(cliCtx, cdc, queryRoute, validatorAddr)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var valCom types.ValidatorAccumulatedCommission
		cdc.MustUnmarshalJSON(commissionRes, &valCom)

		// self bond rewards
		delAddr := sdk.AccAddress(validatorAddr)
		rewardsRes, abort := checkResponseQueryRewards(w, cliCtx, cdc, queryRoute,
			delAddr.String(), valAddr)
		if abort {
			return
		}

		var rewards sdk.DecCoins
		cdc.MustUnmarshalJSON(rewardsRes, &rewards)

		// Prepare response
		res := cdc.MustMarshalJSON(NewValidatorDistInfo(delAddr, rewards, valCom))
		utils.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

// HTTP request handler to query validator's commission and self-delegation rewards
func validatorRewardsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		valAddr := mux.Vars(r)["validatorAddr"]
		validatorAddr, err := sdk.ValAddressFromBech32(valAddr)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		delAddr := sdk.AccAddress(validatorAddr).String()
		res, abort := checkResponseQueryRewards(w, cliCtx, cdc, queryRoute, delAddr, valAddr)
		if abort {
			return
		}

		utils.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

// HTTP request handler to query the distribution params values
func paramsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		params, err := common.QueryParams(cliCtx, queryRoute)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		utils.PostProcessResponse(w, cdc, params, cliCtx.Indent)
	}
}

// HTTP request handler to query the outstanding rewards
func outstandingRewardsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/outstanding_rewards", queryRoute), []byte{})
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		utils.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func checkResponseQueryRewards(w http.ResponseWriter, cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute, delAddr, valAddr string) (res []byte, abort bool) {

	res, err := common.QueryRewards(cliCtx, cdc, queryRoute, delAddr, valAddr)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return nil, true
	}

	return res, false
}
