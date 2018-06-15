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

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/gov/submitproposal", postProposalHandlerFn(cdc, kb, ctx)).Methods("POST")
	r.HandleFunc("/gov/deposit", depositHandlerFn(cdc, kb, ctx)).Methods("POST")
	r.HandleFunc("/gov/vote", voteHandlerFn(cdc, kb, ctx)).Methods("POST")
	r.HandleFunc("/gov/proposals/{proposalID}", queryProposalHandlerFn("gov", cdc, kb, ctx)).Methods("GET")
	r.HandleFunc("/gov/votes/{proposalID}/{voterAddress}", queryVoteHandlerFn("gov", cdc, kb, ctx)).Methods("GET")
}

func postProposalHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req postProposalReq
		err := buildReq(w, r, &req)
		if err != nil {
			return
		}

		if !req.Validate(w) {
			return
		}

		proposer, err := sdk.GetAccAddressBech32(req.Proposer)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		// create the message
		msg := gov.NewMsgSubmitProposal(req.Title, req.Description, req.ProposalType, proposer, req.InitialDeposit)

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

		if !req.Validate(w) {
			return
		}

		depositer, err := sdk.GetAccAddressBech32(req.Depositer)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		// create the message
		msg := gov.NewMsgDeposit(depositer, req.ProposalID, req.Amount)

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

		if !req.Validate(w) {
			return
		}

		voter, err := sdk.GetAccAddressBech32(req.Voter)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		// create the message
		msg := gov.NewMsgVote(voter, req.ProposalID, req.Option)
		// sign
		signAndBuild(w, ctx, req.BaseReq, msg, cdc)
	}
}
func queryProposalHandlerFn(storeName string, cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars["proposalID"]

		if len(strProposalID) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			err := errors.Errorf("proposalId required but not specified")
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

		res, err := ctx.Query(gov.KeyProposal(proposalID), storeName)
		if len(res) == 0 || err != nil {
			err := errors.Errorf("proposalID [%d] does not exist", proposalID)
			w.Write([]byte(err.Error()))
			return
		}

		proposal := new(gov.Proposal)
		cdc.MustUnmarshalBinary(res, proposal)
		output, err := wire.MarshalJSONIndent(cdc, proposal)
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
		strProposalID := vars["proposalID"]
		bechVoterAddr := vars["voterAddress"]

		if len(strProposalID) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			err := errors.Errorf("proposalId required but not specified")
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
			err := errors.Errorf("voterAddress needs to be bech32 encoded")
			w.Write([]byte(err.Error()))
			return
		}

		ctx := context.NewCoreContextFromViper()

		key := []byte(gov.KeyVote(proposalID, voterAddr))
		res, err := ctx.Query(key, storeName)
		if len(res) == 0 || err != nil {

			key := []byte(fmt.Sprintf("%d", proposalID) + ":proposal")
			res, err := ctx.Query(key, storeName)
			if len(res) == 0 || err != nil {
				err := errors.Errorf("proposalID [%d] does not exist", proposalID)
				w.Write([]byte(err.Error()))
				return
			}
			err = errors.Errorf("voter [%s] did not vote on proposalID [%d]", bechVoterAddr, proposalID)
			w.Write([]byte(err.Error()))
			return
		}

		vote := new(gov.Vote)
		cdc.MustUnmarshalBinary(res, vote)

		output, err := wire.MarshalJSONIndent(cdc, vote)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(output)
	}
}
