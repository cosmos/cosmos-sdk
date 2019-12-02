package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
)

// RegisterRoutes register distribution REST routes.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, queryRoute string) {
	registerQueryRoutes(cliCtx, r, queryRoute)
	registerTxRoutes(cliCtx, r, queryRoute)
}

// Used for documentation/swagger purposes
// nolint: unused,deadcode
type communityPoolSpendProposal struct {
	Msgs       []types.CommunityPoolSpendProposal `json:"msg" yaml:"msg"`
	Fee        auth.StdFee                        `json:"fee" yaml:"fee"`
	Signatures []auth.StdSignature                `json:"signatures" yaml:"signatures"`
	Memo       string                             `json:"memo" yaml:"memo"`
}

// ProposalRESTHandler returns a ProposalRESTHandler that exposes the community pool spend REST handler with a given sub-route.
//
// @Summary	Spend the community pool based on governance proposals
// @Description Spend the community pool based on governance proposals
// @Tags distribution
// @Accept json
// @Produce json
// @Param body body rest.CommunityPoolSpendProposalReq true "The data required to spend the community pool"
// @Success 200 {object} rest.communityPoolSpendProposal
// @Failure 400 {object} rest.ErrorResponse "Returned if the request is invalid"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /distribution/community_pool_spend [post]
func ProposalRESTHandler(cliCtx context.CLIContext) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "community_pool_spend",
		Handler:  postProposalHandlerFn(cliCtx),
	}
}

func postProposalHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CommunityPoolSpendProposalReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		content := types.NewCommunityPoolSpendProposal(req.Title, req.Description, req.Recipient, req.Amount)

		msg := gov.NewMsgSubmitProposal(content, req.Deposit, req.Proposer)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
