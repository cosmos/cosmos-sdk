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
	"github.com/pkg/errors"
	"strconv"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/gov/proposal", postProposalHandlerFn(cdc, kb, ctx)).Methods("POST")
	r.HandleFunc("/gov/deposit", depositHandlerFn(cdc, kb, ctx)).Methods("POST")
	r.HandleFunc("/gov/vote", voteHandlerFn(cdc, kb, ctx)).Methods("POST")
	r.HandleFunc("/gov/{proposalId}/proposal", queryProposalHandlerFn(gov.MsgType,cdc, kb, ctx)).Methods("GET")
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
func queryProposalHandlerFn(storeName string,cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		proposalId := vars["proposalId"]

		if len(proposalId) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			err := errors.Errorf("proposalId required but not specified")
			w.Write([]byte(err.Error()))
			return
		}

		id, err := strconv.ParseInt(proposalId, 10, 64)
		if err != nil {
			err := errors.Errorf("proposalID [%s] is not positive", proposalId)
			w.Write([]byte(err.Error()))
			return
		}

		ctx := context.NewCoreContextFromViper()

		key, _ := cdc.MarshalBinary(id)
		res, err := ctx.Query(key, storeName)
		if len(res) == 0 || err != nil {
			err := errors.Errorf("proposalID [%d] is not existed", proposalId)
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
