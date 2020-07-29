package rest

import (
	"net/http"
	"time"

	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/gorilla/mux"

	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// nolint
func newRegisterTxRoutes(
	clientCtx client.Context,
	txg client.TxConfig,
	newMsgFn func() gov.MsgSubmitProposalI,
	r *mux.Router) {
	r.HandleFunc("/upgrade/plan", newPostPlanHandler(clientCtx, txg, newMsgFn)).Methods("POST")
	r.HandleFunc("/upgrade/cancel", newCancelPlanHandler(clientCtx, txg, newMsgFn)).Methods("POST")
}

func registerTxRoutes(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc("/upgrade/plan", postPlanHandler(clientCtx)).Methods("POST")
	r.HandleFunc("/upgrade/cancel", cancelPlanHandler(clientCtx)).Methods("POST")
}

// PlanRequest defines a proposal for a new upgrade plan.
type PlanRequest struct {
	BaseReq       rest.BaseReq `json:"base_req" yaml:"base_req"`
	Title         string       `json:"title" yaml:"title"`
	Description   string       `json:"description" yaml:"description"`
	Deposit       sdk.Coins    `json:"deposit" yaml:"deposit"`
	UpgradeName   string       `json:"upgrade_name" yaml:"upgrade_name"`
	UpgradeHeight int64        `json:"upgrade_height" yaml:"upgrade_height"`
	UpgradeTime   string       `json:"upgrade_time" yaml:"upgrade_time"`
	UpgradeInfo   string       `json:"upgrade_info" yaml:"upgrade_info"`
}

// CancelRequest defines a proposal to cancel a current plan.
type CancelRequest struct {
	BaseReq     rest.BaseReq `json:"base_req" yaml:"base_req"`
	Title       string       `json:"title" yaml:"title"`
	Description string       `json:"description" yaml:"description"`
	Deposit     sdk.Coins    `json:"deposit" yaml:"deposit"`
}

func ProposalRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "upgrade",
		Handler:  postPlanHandler(clientCtx),
	}
}

func ProposalCancelRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "upgrade",
		Handler:  cancelPlanHandler(clientCtx),
	}
}

// nolint
func newPostPlanHandler(clientCtx client.Context, txg client.TxConfig, newMsgFn func() gov.MsgSubmitProposalI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PlanRequest

		if !rest.ReadRESTReq(w, r, clientCtx.Codec, &req) {
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

		var t time.Time
		if req.UpgradeTime != "" {
			t, err = time.Parse(time.RFC3339, req.UpgradeTime)
			if rest.CheckBadRequestError(w, err) {
				return
			}
		}

		plan := types.Plan{Name: req.UpgradeName, Time: t, Height: req.UpgradeHeight, Info: req.UpgradeInfo}
		content := types.NewSoftwareUpgradeProposal(req.Title, req.Description, plan)

		msg := newMsgFn()
		err = msg.SetContent(content)
		if rest.CheckBadRequestError(w, err) {
			return
		}
		msg.SetInitialDeposit(req.Deposit)
		msg.SetProposer(fromAddr)
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}

// nolint
func newCancelPlanHandler(clientCtx client.Context, txg client.TxConfig, newMsgFn func() gov.MsgSubmitProposalI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CancelRequest

		if !rest.ReadRESTReq(w, r, clientCtx.Codec, &req) {
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

		content := types.NewCancelSoftwareUpgradeProposal(req.Title, req.Description)

		msg := newMsgFn()
		err = msg.SetContent(content)
		if rest.CheckBadRequestError(w, err) {
			return
		}
		msg.SetInitialDeposit(req.Deposit)
		msg.SetProposer(fromAddr)
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}

func postPlanHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PlanRequest

		if !rest.ReadRESTReq(w, r, clientCtx.Codec, &req) {
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

		var t time.Time
		if req.UpgradeTime != "" {
			t, err = time.Parse(time.RFC3339, req.UpgradeTime)
			if rest.CheckBadRequestError(w, err) {
				return
			}
		}

		plan := types.Plan{Name: req.UpgradeName, Time: t, Height: req.UpgradeHeight, Info: req.UpgradeInfo}
		content := types.NewSoftwareUpgradeProposal(req.Title, req.Description, plan)
		msg, err := gov.NewMsgSubmitProposal(content, req.Deposit, fromAddr)
		if rest.CheckBadRequestError(w, err) {
			return
		}
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		authclient.WriteGenerateStdTxResponse(w, clientCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func cancelPlanHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CancelRequest

		if !rest.ReadRESTReq(w, r, clientCtx.Codec, &req) {
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

		content := types.NewCancelSoftwareUpgradeProposal(req.Title, req.Description)
		msg, err := gov.NewMsgSubmitProposal(content, req.Deposit, fromAddr)
		if rest.CheckBadRequestError(w, err) {
			return
		}
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		authclient.WriteGenerateStdTxResponse(w, clientCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
