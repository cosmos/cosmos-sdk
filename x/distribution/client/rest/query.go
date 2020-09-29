package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/x/distribution/client/common"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

func registerQueryRoutes(clientCtx client.Context, r *mux.Router) {
	// Get the total rewards balance from all delegations
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards",
		delegatorRewardsHandlerFn(clientCtx),
	).Methods("GET")

	// Query a delegation reward
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards/{validatorAddr}",
		delegationRewardsHandlerFn(clientCtx),
	).Methods("GET")

	// Get the rewards withdrawal address
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/withdraw_address",
		delegatorWithdrawalAddrHandlerFn(clientCtx),
	).Methods("GET")

	// Validator distribution information
	r.HandleFunc(
		"/distribution/validators/{validatorAddr}",
		validatorInfoHandlerFn(clientCtx),
	).Methods("GET")

	// Commission and self-delegation rewards of a single a validator
	r.HandleFunc(
		"/distribution/validators/{validatorAddr}/rewards",
		validatorRewardsHandlerFn(clientCtx),
	).Methods("GET")

	// Outstanding rewards of a single validator
	r.HandleFunc(
		"/distribution/validators/{validatorAddr}/outstanding_rewards",
		outstandingRewardsHandlerFn(clientCtx),
	).Methods("GET")

	// Get the current distribution parameter values
	r.HandleFunc(
		"/distribution/parameters",
		paramsHandlerFn(clientCtx),
	).Methods("GET")

	// Get the amount held in the community pool
	r.HandleFunc(
		"/distribution/community_pool",
		communityPoolHandler(clientCtx),
	).Methods("GET")

}

// HTTP request handler to query the total rewards balance from all delegations
func delegatorRewardsHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		delegatorAddr, ok := checkDelegatorAddressVar(w, r)
		if !ok {
			return
		}

		params := types.NewQueryDelegatorParams(delegatorAddr)
		bz, err := clientCtx.LegacyAmino.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal params: %s", err))
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryDelegatorTotalRewards)
		res, height, err := clientCtx.QueryWithData(route, bz)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

// HTTP request handler to query a delegation rewards
func delegationRewardsHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		delAddr := mux.Vars(r)["delegatorAddr"]
		valAddr := mux.Vars(r)["validatorAddr"]

		// query for rewards from a particular delegation
		res, height, ok := checkResponseQueryDelegationRewards(w, clientCtx, delAddr, valAddr)
		if !ok {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

// HTTP request handler to query a delegation rewards
func delegatorWithdrawalAddrHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		delegatorAddr, ok := checkDelegatorAddressVar(w, r)
		if !ok {
			return
		}

		clientCtx, ok = rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		bz := clientCtx.LegacyAmino.MustMarshalJSON(types.NewQueryDelegatorWithdrawAddrParams(delegatorAddr))
		res, height, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/withdraw_addr", types.QuerierRoute), bz)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

// ValidatorDistInfo defines the properties of
// validator distribution information response.
type ValidatorDistInfo struct {
	OperatorAddress     sdk.AccAddress                       `json:"operator_address" yaml:"operator_address"`
	SelfBondRewards     sdk.DecCoins                         `json:"self_bond_rewards" yaml:"self_bond_rewards"`
	ValidatorCommission types.ValidatorAccumulatedCommission `json:"val_commission" yaml:"val_commission"`
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
func validatorInfoHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		valAddr, ok := checkValidatorAddressVar(w, r)
		if !ok {
			return
		}

		clientCtx, ok = rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		// query commission
		bz, err := common.QueryValidatorCommission(clientCtx, valAddr)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		var commission types.ValidatorAccumulatedCommission
		if rest.CheckInternalServerError(w, clientCtx.LegacyAmino.UnmarshalJSON(bz, &commission)) {
			return
		}

		// self bond rewards
		delAddr := sdk.AccAddress(valAddr)
		bz, height, ok := checkResponseQueryDelegationRewards(w, clientCtx, delAddr.String(), valAddr.String())
		if !ok {
			return
		}

		var rewards sdk.DecCoins
		if rest.CheckInternalServerError(w, clientCtx.LegacyAmino.UnmarshalJSON(bz, &rewards)) {
			return
		}

		bz, err = clientCtx.LegacyAmino.MarshalJSON(NewValidatorDistInfo(delAddr, rewards, commission))
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, bz)
	}
}

// HTTP request handler to query validator's commission and self-delegation rewards
func validatorRewardsHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		valAddr := mux.Vars(r)["validatorAddr"]
		validatorAddr, ok := checkValidatorAddressVar(w, r)
		if !ok {
			return
		}

		clientCtx, ok = rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		delAddr := sdk.AccAddress(validatorAddr).String()
		bz, height, ok := checkResponseQueryDelegationRewards(w, clientCtx, delAddr, valAddr)
		if !ok {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, bz)
	}
}

// HTTP request handler to query the distribution params values
func paramsHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryParams)
		res, height, err := clientCtx.QueryWithData(route, nil)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

func communityPoolHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		res, height, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/community_pool", types.QuerierRoute), nil)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		var result sdk.DecCoins
		if rest.CheckInternalServerError(w, clientCtx.LegacyAmino.UnmarshalJSON(res, &result)) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, result)
	}
}

// HTTP request handler to query the outstanding rewards
func outstandingRewardsHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		validatorAddr, ok := checkValidatorAddressVar(w, r)
		if !ok {
			return
		}

		clientCtx, ok = rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		bin := clientCtx.LegacyAmino.MustMarshalJSON(types.NewQueryValidatorOutstandingRewardsParams(validatorAddr))
		res, height, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/validator_outstanding_rewards", types.QuerierRoute), bin)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

func checkResponseQueryDelegationRewards(
	w http.ResponseWriter, clientCtx client.Context, delAddr, valAddr string,
) (res []byte, height int64, ok bool) {

	res, height, err := common.QueryDelegationRewards(clientCtx, delAddr, valAddr)
	if rest.CheckInternalServerError(w, err) {
		return nil, 0, false
	}

	return res, height, true
}
