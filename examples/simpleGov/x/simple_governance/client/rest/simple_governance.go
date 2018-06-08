package rest

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/gorilla/mux"
	"github.com/tendermint/go-crypto/keys"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/proposals/{id}", GetProposalHandlerFn(cdc, kb, ctx)).Methods("GET")
	r.HandleFunc("/proposals/{id}/vote", VoteProposalHandlerFn(cdc, kb, ctx)).Methods("POST")
	r.HandleFunc("/accounts/{address}/proposals/{id}", GetAccountProposalHandlerFn(cdc, kb, ctx)).Methods("GET")
	r.HandleFunc("/accounts/{address}/proposals", SubmitProposalHandlerFn(cdc, kb, ctx)).Methods("POST")
}

type getProposalBody struct {
	ProposalID int64 `json:"proposal_id"`
}

type proposeBody struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Deposit     string `json:"deposit"`
}

type voteBody struct {
	ProposalID int64  `json:"proposal_id"`
	Option     string `json:"option"`
	Voter      string `json:"voter"`
}

// GetProposalHandlerFn - http request handler to get a current proposal
func GetProposalHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		proposalID := vars["id"]

		var getProp getProposalBody

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		err = json.Unmarshal(body, &getProp)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		// TODO handle obtain a proposal

		// w.Write(output)

	}
}

// SubmitProposalHandlerFn - http request handler to create a proposal
func SubmitProposalHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// collect data
		vars := mux.Vars(r)
		address := vars["address"]

		var m proposeBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		err = json.Unmarshal(body, &m)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		// Get address of proposer
		bz, err := hex.DecodeString(address)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		proposerAddr := sdk.Address(bz)

		// // build message
		// msg := client.BuildMsg(info.PubKey.Address(), to, m.Amount)
		// if err != nil { // XXX rechecking same error ?
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	w.Write([]byte(err.Error()))
		// 	return
		// }

		// // sign
		// ctx = ctx.WithSequence(m.Sequence)
		// txBytes, err := ctx.SignAndBuild(m.LocalAccountName, m.Password, msg, cdc)
		// if err != nil {
		// 	w.WriteHeader(http.StatusUnauthorized)
		// 	w.Write([]byte(err.Error()))
		// 	return
		// }

		// send
		// res, err := ctx.BroadcastTx(txBytes)
		// ctx.
		// if err != nil {
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	w.Write([]byte(err.Error()))
		// 	return
		// }

		// output, err := json.MarshalIndent(res, "", "  ")
		// if err != nil {
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	w.Write([]byte(err.Error()))
		// 	return
		// }

		w.Write(output)
	}
}

// VoteProposalHandlerFn - http request handler to vote on a proposal
func VoteProposalHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		proposalID := vars["id"]

		var vote voteBody

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		err = json.Unmarshal(body, &vote)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		// Get address of proposer
		bz, err := hex.DecodeString(vote.Voter)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		voterAddr := sdk.Address(bz)

		w.Write(output)
	}
}
