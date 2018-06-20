package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/tendermint/go-crypto/keys"
)

// REST Variable names
// nolint
const (
	ProposalRestID = "proposalID"
	RestVoter      = "voterAddress"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/gov/submitproposal", postProposalHandlerFn(cdc, kb, ctx)).Methods("POST")
	r.HandleFunc("/gov/deposit", depositHandlerFn(cdc, kb, ctx)).Methods("POST")
	r.HandleFunc("/gov/vote", voteHandlerFn(cdc, kb, ctx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}", ProposalRestID), queryProposalHandlerFn("gov", cdc, kb, ctx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/gov/votes/{%s}/{%s}", ProposalRestID, RestVoter), queryVoteHandlerFn("gov", cdc, kb, ctx)).Methods("GET")
}

type postProposalReq struct {
	BaseReq        baseReq   `json:"base_req"`
	Title          string    `json:"title"`           //  Title of the proposal
	Description    string    `json:"description"`     //  Description of the proposal
	ProposalType   string    `json:"proposal_type"`   //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
	Proposer       string    `json:"proposer"`        //  Address of the proposer
	InitialDeposit sdk.Coins `json:"initial_deposit"` // Coins to add to the proposal's deposit
}

type depositReq struct {
	BaseReq    baseReq   `json:"base_req"`
	ProposalID int64     `json:"proposalID"` // ID of the proposal
	Depositer  string    `json:"depositer"`  // Address of the depositer
	Amount     sdk.Coins `json:"amount"`     // Coins to add to the proposal's deposit
}

type voteReq struct {
	BaseReq    baseReq `json:"base_req"`
	Voter      string  `json:"voter"`      //  address of the voter
	ProposalID int64   `json:"proposalID"` //  proposalID of the proposal
	Option     string  `json:"option"`     //  option from OptionSet chosen by the voter
}

func postProposalHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req postProposalReq
		err := buildReq(w, r, &req)
		if err != nil {
			return
		}

		if !req.BaseReq.baseReqValidate(w) {
			return
		}

		proposer, err := sdk.GetAccAddressBech32(req.Proposer)
		if err != nil {
			writeErr(&w, http.StatusBadRequest, err.Error())
			return
		}

		proposalTypeByte, err := gov.StringToProposalType(req.ProposalType)
		if err != nil {
			writeErr(&w, http.StatusBadRequest, err.Error())
			return
		}

		// create the message
		msg := gov.NewMsgSubmitProposal(req.Title, req.Description, proposalTypeByte, proposer, req.InitialDeposit)
		err = msg.ValidateBasic()
		if err != nil {
			writeErr(&w, http.StatusBadRequest, err.Error())
			return
		}

		// sign
		signAndBuild(w, ctx, req.BaseReq, msg, cdc)
	}
}

func depositHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req depositReq
		err := buildReq(w, r, &req)
		if err != nil {
			return
		}

		if !req.BaseReq.baseReqValidate(w) {
			return
		}

		depositer, err := sdk.GetAccAddressBech32(req.Depositer)
		if err != nil {
			writeErr(&w, http.StatusBadRequest, err.Error())
			return
		}

		// create the message
		msg := gov.NewMsgDeposit(depositer, req.ProposalID, req.Amount)
		err = msg.ValidateBasic()
		if err != nil {
			writeErr(&w, http.StatusBadRequest, err.Error())
			return
		}

		// sign
		signAndBuild(w, ctx, req.BaseReq, msg, cdc)
	}
}

func voteHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req voteReq
		err := buildReq(w, r, &req)
		if err != nil {
			return
		}

		if !req.BaseReq.baseReqValidate(w) {
			return
		}

		voter, err := sdk.GetAccAddressBech32(req.Voter)
		if err != nil {
			writeErr(&w, http.StatusBadRequest, err.Error())
			return
		}

		voteOptionByte, err := gov.StringToVoteOption(req.Option)
		if err != nil {
			writeErr(&w, http.StatusBadRequest, err.Error())
			return
		}

		// create the message
		msg := gov.NewMsgVote(voter, req.ProposalID, voteOptionByte)
		err = msg.ValidateBasic()
		if err != nil {
			writeErr(&w, http.StatusBadRequest, err.Error())
			return
		}

		// sign
		signAndBuild(w, ctx, req.BaseReq, msg, cdc)
	}
}

func queryProposalHandlerFn(storeName string, cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[ProposalRestID]

		if len(strProposalID) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			err := errors.New("proposalId required but not specified")
			w.Write([]byte(err.Error()))
			return
		}

		proposalID, err := strconv.ParseInt(strProposalID, 10, 64)
		if err != nil {
			err := errors.Errorf("proposalID [%d] is not positive", proposalID)
			w.Write([]byte(err.Error()))
			return
		}

		ctx := context.NewCoreContextFromViper()

		res, err := ctx.QueryStore(gov.KeyProposal(proposalID), storeName)
		if len(res) == 0 || err != nil {
			err := errors.Errorf("proposalID [%d] does not exist", proposalID)
			w.Write([]byte(err.Error()))
			return
		}

		var proposal gov.Proposal
		cdc.MustUnmarshalBinary(res, &proposal)
		proposalRest := gov.ProposalToRest(proposal)
		output, err := wire.MarshalJSONIndent(cdc, proposalRest)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(output)
	}
}

func queryVoteHandlerFn(storeName string, cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[ProposalRestID]
		bechVoterAddr := vars[RestVoter]

		if len(strProposalID) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			err := errors.New("proposalId required but not specified")
			w.Write([]byte(err.Error()))
			return
		}

		proposalID, err := strconv.ParseInt(strProposalID, 10, 64)
		if err != nil {
			err := errors.Errorf("proposalID [%s] is not positive", proposalID)
			w.Write([]byte(err.Error()))
			return
		}

		voterAddr, err := sdk.GetAccAddressBech32(bechVoterAddr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			err := errors.Errorf("'%s' needs to be bech32 encoded", RestVoter)
			w.Write([]byte(err.Error()))
			return
		}

		ctx := context.NewCoreContextFromViper()

		key := []byte(gov.KeyVote(proposalID, voterAddr))
		res, err := ctx.QueryStore(key, storeName)
		if len(res) == 0 || err != nil {

			res, err := ctx.QueryStore(gov.KeyProposal(proposalID), storeName)
			if len(res) == 0 || err != nil {
				err := errors.Errorf("proposalID [%d] does not exist", proposalID)
				w.Write([]byte(err.Error()))
				return
			}
			err = errors.Errorf("voter [%s] did not vote on proposalID [%d]", bechVoterAddr, proposalID)
			w.Write([]byte(err.Error()))
			return
		}

		var vote gov.Vote
		cdc.MustUnmarshalBinary(res, &vote)
		voteRest := gov.VoteToRest(vote)
		output, err := wire.MarshalJSONIndent(cdc, voteRest)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(output)
	}
}
