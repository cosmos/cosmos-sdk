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
)

// REST Variable names
// nolint
const (
	RestProposalID = "proposalID"
	RestDepositer  = "depositer"
	RestVoter      = "voter"
	storeName      = "gov"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec) {
	r.HandleFunc("/gov/proposals", postProposalHandlerFn(cdc, ctx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/deposits", RestProposalID), depositHandlerFn(cdc, ctx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/votes", RestProposalID), voteHandlerFn(cdc, ctx)).Methods("POST")

	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}", RestProposalID), queryProposalHandlerFn(cdc)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/deposits/{%s}", RestProposalID, RestDepositer), queryDepositHandlerFn(cdc)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/votes/{%s}", RestProposalID, RestVoter), queryVoteHandlerFn(cdc)).Methods("GET")

	r.HandleFunc("/gov/proposals", queryProposalsWithParameterFn(cdc)).Methods("GET")
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
	BaseReq   baseReq   `json:"base_req"`
	Depositer string    `json:"depositer"` // Address of the depositer
	Amount    sdk.Coins `json:"amount"`    // Coins to add to the proposal's deposit
}

type voteReq struct {
	BaseReq baseReq `json:"base_req"`
	Voter   string  `json:"voter"`  //  address of the voter
	Option  string  `json:"option"` //  option from OptionSet chosen by the voter
}

func postProposalHandlerFn(cdc *wire.Codec, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req postProposalReq
		err := buildReq(w, r, cdc, &req)
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

func depositHandlerFn(cdc *wire.Codec, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]

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

		var req depositReq
		err = buildReq(w, r, cdc, &req)
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
		msg := gov.NewMsgDeposit(depositer, proposalID, req.Amount)
		err = msg.ValidateBasic()
		if err != nil {
			writeErr(&w, http.StatusBadRequest, err.Error())
			return
		}

		// sign
		signAndBuild(w, ctx, req.BaseReq, msg, cdc)
	}
}

func voteHandlerFn(cdc *wire.Codec, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]

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

		var req voteReq
		err = buildReq(w, r, cdc, &req)
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
		msg := gov.NewMsgVote(voter, proposalID, voteOptionByte)
		err = msg.ValidateBasic()
		if err != nil {
			writeErr(&w, http.StatusBadRequest, err.Error())
			return
		}

		// sign
		signAndBuild(w, ctx, req.BaseReq, msg, cdc)
	}
}

func queryProposalHandlerFn(cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]

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
		if err != nil || len(res) == 0 {
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

func queryDepositHandlerFn(cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]
		bechDepositerAddr := vars[RestDepositer]

		if len(strProposalID) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			err := errors.New("proposalId required but not specified")
			w.Write([]byte(err.Error()))
			return
		}

		proposalID, err := strconv.ParseInt(strProposalID, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			err := errors.Errorf("proposalID [%d] is not positive", proposalID)
			w.Write([]byte(err.Error()))
			return
		}

		if len(bechDepositerAddr) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			err := errors.New("depositer address required but not specified")
			w.Write([]byte(err.Error()))
			return
		}

		depositerAddr, err := sdk.GetAccAddressBech32(bechDepositerAddr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			err := errors.Errorf("'%s' needs to be bech32 encoded", RestDepositer)
			w.Write([]byte(err.Error()))
			return
		}

		ctx := context.NewCoreContextFromViper()

		res, err := ctx.QueryStore(gov.KeyDeposit(proposalID, depositerAddr), storeName)
		if err != nil || len(res) == 0 {
			res, err := ctx.QueryStore(gov.KeyProposal(proposalID), storeName)
			if err != nil || len(res) == 0 {
				w.WriteHeader(http.StatusNotFound)
				err := errors.Errorf("proposalID [%d] does not exist", proposalID)
				w.Write([]byte(err.Error()))
				return
			}
			w.WriteHeader(http.StatusNotFound)
			err = errors.Errorf("depositer [%s] did not deposit on proposalID [%d]", bechDepositerAddr, proposalID)
			w.Write([]byte(err.Error()))
			return
		}

		var deposit gov.Deposit
		cdc.MustUnmarshalBinary(res, &deposit)
		depositRest := gov.DepositToRest(deposit)
		output, err := wire.MarshalJSONIndent(cdc, depositRest)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(output)
	}
}

func queryVoteHandlerFn(cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]
		bechVoterAddr := vars[RestVoter]

		if len(strProposalID) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			err := errors.New("proposalId required but not specified")
			w.Write([]byte(err.Error()))
			return
		}

		proposalID, err := strconv.ParseInt(strProposalID, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			err := errors.Errorf("proposalID [%s] is not positive", proposalID)
			w.Write([]byte(err.Error()))
			return
		}

		if len(bechVoterAddr) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			err := errors.New("voter address required but not specified")
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

		res, err := ctx.QueryStore(gov.KeyVote(proposalID, voterAddr), storeName)
		if err != nil || len(res) == 0 {

			res, err := ctx.QueryStore(gov.KeyProposal(proposalID), storeName)
			if err != nil || len(res) == 0 {
				w.WriteHeader(http.StatusNotFound)
				err := errors.Errorf("proposalID [%d] does not exist", proposalID)
				w.Write([]byte(err.Error()))
				return
			}
			w.WriteHeader(http.StatusNotFound)
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

func queryProposalsWithParameterFn(cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bechVoterAddr := r.URL.Query().Get(RestVoter)
		bechDepositerAddr := r.URL.Query().Get(RestDepositer)

		var err error
		var voterAddr sdk.AccAddress
		var depositerAddr sdk.AccAddress

		if len(bechVoterAddr) != 0 {
			voterAddr, err = sdk.GetAccAddressBech32(bechVoterAddr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				err := errors.Errorf("'%s' needs to be bech32 encoded", RestVoter)
				w.Write([]byte(err.Error()))
				return
			}
		}

		if len(bechDepositerAddr) != 0 {
			depositerAddr, err = sdk.GetAccAddressBech32(bechDepositerAddr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				err := errors.Errorf("'%s' needs to be bech32 encoded", RestDepositer)
				w.Write([]byte(err.Error()))
				return
			}
		}

		ctx := context.NewCoreContextFromViper()

		res, err := ctx.QueryStore(gov.KeyNextProposalID, storeName)
		if err != nil {
			err = errors.New("no proposals exist yet and proposalID has not been set")
			w.Write([]byte(err.Error()))
			return
		}
		var maxProposalID int64
		cdc.MustUnmarshalBinary(res, &maxProposalID)

		matchingProposals := []gov.ProposalRest{}

		for proposalID := int64(0); proposalID < maxProposalID; proposalID++ {
			if voterAddr != nil {
				res, err = ctx.QueryStore(gov.KeyVote(proposalID, voterAddr), storeName)
				if err != nil || len(res) == 0 {
					continue
				}
			}

			if depositerAddr != nil {
				res, err = ctx.QueryStore(gov.KeyDeposit(proposalID, depositerAddr), storeName)
				if err != nil || len(res) == 0 {
					continue
				}
			}

			res, err = ctx.QueryStore(gov.KeyProposal(proposalID), storeName)
			if err != nil || len(res) == 0 {
				continue
			}
			var proposal gov.Proposal
			cdc.MustUnmarshalBinary(res, &proposal)

			matchingProposals = append(matchingProposals, gov.ProposalToRest(proposal))
		}

		output, err := wire.MarshalJSONIndent(cdc, matchingProposals)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(output)
	}
}
