package rest

import (
	"encoding/hex"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/gorilla/mux"
	"github.com/tendermint/go-crypto/keys"
	"net/http"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/gov/proposal", postProposalHandlerFn(cdc, kb, ctx)).Methods("POST")
	r.HandleFunc("/gov/deposit", depositHandlerFn(cdc, kb, ctx)).Methods("POST")
	r.HandleFunc("/gov/vote", voteHandlerFn(cdc, kb, ctx)).Methods("POST")
}

func postProposalHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req postProposalReq
		err := buildReq(w,r,&req)
		if err != nil {
			return
		}

		if !req.Validate(w) {
			return
		}

		bz, err := hex.DecodeString(req.Proposer)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		proposer := sdk.Address(bz)
		// create the message
		msg := gov.NewMsgSubmitProposal(req.Title, req.Description, req.ProposalType, proposer, req.InitialDeposit)

		// sign
		signAndBuild(w, ctx, req.BaseReq, msg, cdc)
	}
}

func depositHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req depositReq
		err := buildReq(w,r,&req)
		if err != nil {
			return
		}

		if !req.Validate(w) {
			return
		}

		bz, err := hex.DecodeString(req.Depositer)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		Depositer := sdk.Address(bz)
		// create the message
		msg := gov.NewMsgDeposit(req.ProposalID, Depositer, req.Amount)

		// sign
		signAndBuild(w, ctx, req.BaseReq, msg, cdc)
	}
}

func voteHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req voteReq
		err := buildReq(w,r,&req)
		if err != nil {
			return
		}

		if !req.Validate(w) {
			return
		}

		bz, err := hex.DecodeString(req.Voter)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		voter := sdk.Address(bz)
		// create the message
		msg := gov.NewMsgVote(voter, req.ProposalID, req.Option)
		// sign
		signAndBuild(w, ctx, req.BaseReq, msg, cdc)
	}
}
