package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/common"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

type (
	withdrawRewardsReq struct {
		BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	}

	setWithdrawalAddrReq struct {
		BaseReq         rest.BaseReq   `json:"base_req" yaml:"base_req"`
		WithdrawAddress sdk.AccAddress `json:"withdraw_address" yaml:"withdraw_address"`
	}

	fundCommunityPoolReq struct {
		BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
		Amount  sdk.Coins    `json:"amount" yaml:"amount"`
	}
)

func registerTxHandlers(cliCtx context.CLIContext, m codec.Marshaler, txg tx.Generator, r *mux.Router) {
	// Withdraw all delegator rewards
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards",
		newWithdrawDelegatorRewardsHandlerFn(cliCtx, m, txg),
	).Methods("POST")

	// Withdraw delegation rewards
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards/{validatorAddr}",
		newWithdrawDelegationRewardsHandlerFn(cliCtx, m, txg),
	).Methods("POST")

	// Replace the rewards withdrawal address
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/withdraw_address",
		newSetDelegatorWithdrawalAddrHandlerFn(cliCtx, m, txg),
	).Methods("POST")

	// Withdraw validator rewards and commission
	r.HandleFunc(
		"/distribution/validators/{validatorAddr}/rewards",
		newWithdrawValidatorRewardsHandlerFn(cliCtx, m, txg),
	).Methods("POST")

	// Fund the community pool
	r.HandleFunc(
		"/distribution/community_pool",
		newFundCommunityPoolHandlerFn(cliCtx, m, txg),
	).Methods("POST")
}

func newWithdrawDelegatorRewardsHandlerFn(cliCtx context.CLIContext, m codec.Marshaler, txg tx.Generator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx = cliCtx.WithMarshaler(m)
		var req withdrawRewardsReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// read and validate URL's variables
		delAddr, ok := checkDelegatorAddressVar(w, r)
		if !ok {
			return
		}

		msgs, err := common.WithdrawAllDelegatorRewards(cliCtx, types.QuerierRoute, delAddr)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, txg, req.BaseReq, msgs...)
	}
}

func newWithdrawDelegationRewardsHandlerFn(cliCtx context.CLIContext, m codec.Marshaler, txg tx.Generator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx = cliCtx.WithMarshaler(m)
		var req withdrawRewardsReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// read and validate URL's variables
		delAddr, ok := checkDelegatorAddressVar(w, r)
		if !ok {
			return
		}

		valAddr, ok := checkValidatorAddressVar(w, r)
		if !ok {
			return
		}

		msg := types.NewMsgWithdrawDelegatorReward(delAddr, valAddr)
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, txg, req.BaseReq, msg)
	}
}

func newSetDelegatorWithdrawalAddrHandlerFn(cliCtx context.CLIContext, m codec.Marshaler, txg tx.Generator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx = cliCtx.WithMarshaler(m)
		var req setWithdrawalAddrReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// read and validate URL's variables
		delAddr, ok := checkDelegatorAddressVar(w, r)
		if !ok {
			return
		}

		msg := types.NewMsgSetWithdrawAddress(delAddr, req.WithdrawAddress)
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, txg, req.BaseReq, msg)
	}
}

func newWithdrawValidatorRewardsHandlerFn(cliCtx context.CLIContext, m codec.Marshaler, txg tx.Generator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx = cliCtx.WithMarshaler(m)
		var req withdrawRewardsReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// read and validate URL's variable
		valAddr, ok := checkValidatorAddressVar(w, r)
		if !ok {
			return
		}

		// prepare multi-message transaction
		msgs, err := common.WithdrawValidatorRewardsAndCommission(valAddr)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, txg, req.BaseReq, msgs...)
	}
}

func newFundCommunityPoolHandlerFn(cliCtx context.CLIContext, m codec.Marshaler, txg tx.Generator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx = cliCtx.WithMarshaler(m)
		var req fundCommunityPoolReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		msg := types.NewMsgFundCommunityPool(req.Amount, fromAddr)
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, txg, req.BaseReq, msg)
	}
}

// ---------------------------------------------------------------------------
// Deprecated
//
// TODO: Remove once client-side Protobuf migration has been completed.
// ---------------------------------------------------------------------------
func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router, queryRoute string) {
	// Withdraw all delegator rewards
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards",
		withdrawDelegatorRewardsHandlerFn(cliCtx, queryRoute),
	).Methods("POST")

	// Withdraw delegation rewards
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards/{validatorAddr}",
		withdrawDelegationRewardsHandlerFn(cliCtx),
	).Methods("POST")

	// Replace the rewards withdrawal address
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/withdraw_address",
		setDelegatorWithdrawalAddrHandlerFn(cliCtx),
	).Methods("POST")

	// Withdraw validator rewards and commission
	r.HandleFunc(
		"/distribution/validators/{validatorAddr}/rewards",
		withdrawValidatorRewardsHandlerFn(cliCtx),
	).Methods("POST")

	// Fund the community pool
	r.HandleFunc(
		"/distribution/community_pool",
		fundCommunityPoolHandlerFn(cliCtx),
	).Methods("POST")

}

// Withdraw delegator rewards
func withdrawDelegatorRewardsHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req withdrawRewardsReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// read and validate URL's variables
		delAddr, ok := checkDelegatorAddressVar(w, r)
		if !ok {
			return
		}

		msgs, err := common.WithdrawAllDelegatorRewards(cliCtx, queryRoute, delAddr)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		authclient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, msgs)
	}
}

// Withdraw delegation rewards
func withdrawDelegationRewardsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req withdrawRewardsReq

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// read and validate URL's variables
		delAddr, ok := checkDelegatorAddressVar(w, r)
		if !ok {
			return
		}

		valAddr, ok := checkValidatorAddressVar(w, r)
		if !ok {
			return
		}

		msg := types.NewMsgWithdrawDelegatorReward(delAddr, valAddr)
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		authclient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// Replace the rewards withdrawal address
func setDelegatorWithdrawalAddrHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req setWithdrawalAddrReq

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// read and validate URL's variables
		delAddr, ok := checkDelegatorAddressVar(w, r)
		if !ok {
			return
		}

		msg := types.NewMsgSetWithdrawAddress(delAddr, req.WithdrawAddress)
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		authclient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// Withdraw validator rewards and commission
func withdrawValidatorRewardsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req withdrawRewardsReq

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// read and validate URL's variable
		valAddr, ok := checkValidatorAddressVar(w, r)
		if !ok {
			return
		}

		// prepare multi-message transaction
		msgs, err := common.WithdrawValidatorRewardsAndCommission(valAddr)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		authclient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, msgs)
	}
}

func fundCommunityPoolHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req fundCommunityPoolReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		msg := types.NewMsgFundCommunityPool(req.Amount, fromAddr)
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		authclient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// Auxiliary

func checkDelegatorAddressVar(w http.ResponseWriter, r *http.Request) (sdk.AccAddress, bool) {
	addr, err := sdk.AccAddressFromBech32(mux.Vars(r)["delegatorAddr"])
	if rest.CheckBadRequestError(w, err) {
		return nil, false
	}

	return addr, true
}

func checkValidatorAddressVar(w http.ResponseWriter, r *http.Request) (sdk.ValAddress, bool) {
	addr, err := sdk.ValAddressFromBech32(mux.Vars(r)["validatorAddr"])
	if rest.CheckBadRequestError(w, err) {
		return nil, false
	}

	return addr, true
}
