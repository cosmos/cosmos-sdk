package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/x/distribution/client/common"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, queryRoute string) {
	// Get the total rewards balance from all delegations
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards",
		delegatorRewardsHandlerFn(cliCtx, queryRoute),
	).Methods("GET")

	// Query a delegation reward
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards/{validatorAddr}",
		delegationRewardsHandlerFn(cliCtx, queryRoute),
	).Methods("GET")

	// Get the rewards withdrawal address
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/withdraw_address",
		delegatorWithdrawalAddrHandlerFn(cliCtx, queryRoute),
	).Methods("GET")

	// Validator distribution information
	r.HandleFunc(
		"/distribution/validators/{validatorAddr}",
		validatorInfoHandlerFn(cliCtx, queryRoute),
	).Methods("GET")

	// Commission and self-delegation rewards of a single a validator
	r.HandleFunc(
		"/distribution/validators/{validatorAddr}/rewards",
		validatorRewardsHandlerFn(cliCtx, queryRoute),
	).Methods("GET")

	// Outstanding rewards of a single validator
	r.HandleFunc(
		"/distribution/validators/{validatorAddr}/outstanding_rewards",
		outstandingRewardsHandlerFn(cliCtx, queryRoute),
	).Methods("GET")

	// Get the current distribution parameter values
	r.HandleFunc(
		"/distribution/parameters",
		paramsHandlerFn(cliCtx, queryRoute),
	).Methods("GET")

	// Get the amount held in the community pool
	r.HandleFunc(
		"/distribution/community_pool",
		communityPoolHandler(cliCtx, queryRoute),
	).Methods("GET")

}

// coinsReturn helps generate documentation for REST routes
type coinsReturn struct {
	Height int64  `json:"height"`
	Result []coin `json:"result"`
}

// coin helps generate documentation for REST routes
type coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// delegatorRewardsHandlerFn implements a delegator rewards query route
//
// @Summary Query all delegation rewards from a delegator
// @Description Query all delegation rewards from a single delegator address
// @Tags distribution
// @Produce json
// @Param delegatorAddr path string true "The address of the delegator"
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.coinsReturn
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have a valid height"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /distribution/delegators/{delegatorAddr}/rewards [get]
func delegatorRewardsHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// query for rewards from a particular delegator
		res, ok := checkResponseQueryDelegatorTotalRewards(w, cliCtx, queryRoute, mux.Vars(r)["delegatorAddr"])
		if !ok {
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// delegationRewardsHandlerFn implements a individual delegation reward query route
//
// @Summary Query delegation rewards
// @Description Query delegation rewards from a single delegator address
// @Tags distribution
// @Produce json
// @Param delegatorAddr path string true "The address of the delegator"
// @Param validatorAddr path string true "The address of the validator"
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.coinsReturn
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have a valid height"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /distribution/delegators/{delegatorAddr}/rewards/{validatorAddr} [get]
func delegationRewardsHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// query for rewards from a particular delegation
		res, ok := checkResponseQueryDelegationRewards(w, cliCtx, queryRoute, mux.Vars(r)["delegatorAddr"], mux.Vars(r)["validatorAddr"])
		if !ok {
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// delegatorWithdrawalAddr helps generate documentation for delegatorWithdrawalAddrHandlerFn
type delegatorWithdrawalAddr struct {
	Height int64       `json:"height"`
	Result sdk.Address `json:"result"`
}

// delegatorWithdrawalAddrHandlerFn implements a delegator withdraw address query route
//
// @Summary Query delegator withdraw address
// @Description Query withdraw address from a single delegator address
// @Tags distribution
// @Produce json
// @Param delegatorAddr path string true "The address of the delegator"
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.delegatorWithdrawalAddr
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have a valid height"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /distribution/delegators/{delegatorAddr}/withdraw_address [get]
func delegatorWithdrawalAddrHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		delegatorAddr, ok := checkDelegatorAddressVar(w, r)
		if !ok {
			return
		}

		cliCtx, ok = rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		bz := cliCtx.Codec.MustMarshalJSON(types.NewQueryDelegatorWithdrawAddrParams(delegatorAddr))
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/withdraw_addr", queryRoute), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// ValidatorDistInfo defines the properties of
// validator distribution information response.
type ValidatorDistInfo struct {
	OperatorAddress     sdk.AccAddress `json:"operator_address" yaml:"operator_address"`
	SelfBondRewards     sdk.DecCoins   `json:"self_bond_rewards" yaml:"self_bond_rewards"`
	ValidatorCommission sdk.DecCoins   `json:"val_commission" yaml:"val_commission"`
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

// validatorInfo helps generate documentation for validatorInfoHandlerFn
type validatorInfo struct {
	Height int64             `json:"height"`
	Result ValidatorDistInfo `json:"result"`
}

// validatorInfoHandlerFn implements a validator distribution information route
//
// @Summary Query validator distribution information
// @Description Query validator distribution information from a single validator
// @Tags distribution
// @Produce json
// @Param validatorAddr path string true "The address of the validator"
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.validatorInfo
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have a valid height"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /distribution/validators/{validatorAddr} [get]
func validatorInfoHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		valAddr := mux.Vars(r)["validatorAddr"]
		validatorAddr, ok := checkValidatorAddressVar(w, r)
		if !ok {
			return
		}

		cliCtx, ok = rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// query commission
		commissionRes, err := common.QueryValidatorCommission(cliCtx, queryRoute, validatorAddr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var valCom types.ValidatorAccumulatedCommission
		cliCtx.Codec.MustUnmarshalJSON(commissionRes, &valCom)

		// self bond rewards
		delAddr := sdk.AccAddress(validatorAddr)
		rewardsRes, ok := checkResponseQueryDelegationRewards(w, cliCtx, queryRoute, delAddr.String(), valAddr)
		if !ok {
			return
		}

		var rewards sdk.DecCoins
		cliCtx.Codec.MustUnmarshalJSON(rewardsRes, &rewards)

		// Prepare response
		res := cliCtx.Codec.MustMarshalJSON(NewValidatorDistInfo(delAddr, rewards, valCom))
		rest.PostProcessResponse(w, cliCtx, res)
	}

}

// validatorRewardsHandlerFn implements a validator rewards query route
//
// @Summary Query validator rewards information
// @Description Query validator rewards information from a single validator
// @Tags distribution
// @Produce json
// @Param validatorAddr path string true "The address of the validator"
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.coinsReturn
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have a valid height"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /distribution/validators/{validatorAddr}/rewards [get]
func validatorRewardsHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		valAddr := mux.Vars(r)["validatorAddr"]
		validatorAddr, ok := checkValidatorAddressVar(w, r)
		if !ok {
			return
		}

		cliCtx, ok = rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		delAddr := sdk.AccAddress(validatorAddr).String()
		res, ok := checkResponseQueryDelegationRewards(w, cliCtx, queryRoute, delAddr, valAddr)
		if !ok {
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// params helps generate documentation for paramsHandlerFn
type params struct {
	Height int64               `json:"height"`
	Result common.PrettyParams `json:"result"`
}

// paramsHandlerFn implements a params query route
//
// @Summary Query distribution params data
// @Description Query distribution params data
// @Tags distribution
// @Produce json
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.params
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have a valid height"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /distribution/parameters [get]
func paramsHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params, err := common.QueryParams(cliCtx, queryRoute)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, params)
	}
}

// communityPoolHandler implements a community pool query route
//
// @Summary Query community pool data
// @Description Query community pool data
// @Tags distribution
// @Produce json
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.coinsReturn
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have a valid height"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /distribution/community_pool [get]
func communityPoolHandler(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/community_pool", queryRoute), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var result sdk.DecCoins
		if err := cliCtx.Codec.UnmarshalJSON(res, &result); err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, result)
	}
}

// outstandingRewardsHandlerFn implements a validator outstanding rewards query route.
//
// @Summary Query validator outstanding rewards information
// @Description Query validator rewards information from a single validator
// @Tags distribution
// @Produce json
// @Param validatorAddr path string true "The address of the validator"
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.coinsReturn
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have a valid height"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /distribution/validators/{validatorAddr}/outstanding_rewards [get]
func outstandingRewardsHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		validatorAddr, ok := checkValidatorAddressVar(w, r)
		if !ok {
			return
		}

		cliCtx, ok = rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		bin := cliCtx.Codec.MustMarshalJSON(types.NewQueryValidatorOutstandingRewardsParams(validatorAddr))
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/validator_outstanding_rewards", queryRoute), bin)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func checkResponseQueryDelegatorTotalRewards(
	w http.ResponseWriter, cliCtx context.CLIContext, queryRoute, delAddr string,
) (res []byte, ok bool) {

	res, err := common.QueryDelegatorTotalRewards(cliCtx, queryRoute, delAddr)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return nil, false
	}

	return res, true
}

func checkResponseQueryDelegationRewards(
	w http.ResponseWriter, cliCtx context.CLIContext, queryRoute, delAddr, valAddr string,
) (res []byte, ok bool) {

	res, err := common.QueryDelegationRewards(cliCtx, queryRoute, delAddr, valAddr)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return nil, false
	}

	return res, true
}
