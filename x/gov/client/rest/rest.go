package rest

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	gcutils "github.com/cosmos/cosmos-sdk/x/gov/client/utils"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// REST Variable names
// nolint
const (
	RestParamsType     = "type"
	RestProposalID     = "proposal-id"
	RestDepositor      = "depositor"
	RestVoter          = "voter"
	RestProposalStatus = "status"
	RestNumLimit       = "limit"
)

// ProposalRESTHandler defines a REST handler implemented in another module. The
// sub-route is mounted on the governance REST handler.
type ProposalRESTHandler struct {
	SubRoute string
	Handler  func(http.ResponseWriter, *http.Request)
}

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, phs []ProposalRESTHandler) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r, phs)
}

// PostProposalReq defines the properties of a proposal request's body.
type PostProposalReq struct {
	BaseReq        rest.BaseReq   `json:"base_req"`
	Title          string         `json:"title"`           // Title of the proposal
	Description    string         `json:"description"`     // Description of the proposal
	ProposalType   string         `json:"proposal_type"`   // Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
	Proposer       sdk.AccAddress `json:"proposer"`        // Address of the proposer
	InitialDeposit sdk.Coins      `json:"initial_deposit"` // Coins to add to the proposal's deposit
}

// DepositReq defines the properties of a deposit request's body.
type DepositReq struct {
	BaseReq   rest.BaseReq   `json:"base_req"`
	Depositor sdk.AccAddress `json:"depositor"` // Address of the depositor
	Amount    sdk.Coins      `json:"amount"`    // Coins to add to the proposal's deposit
}

// VoteReq defines the properties of a vote request's body.
type VoteReq struct {
	BaseReq rest.BaseReq   `json:"base_req"`
	Voter   sdk.AccAddress `json:"voter"`  // address of the voter
	Option  string         `json:"option"` // option from OptionSet chosen by the voter
}
