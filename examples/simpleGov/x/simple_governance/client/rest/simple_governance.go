package rest

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/simpleGov"
	"github.com/gorilla/mux"
	"github.com/tendermint/go-crypto/keys"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	// GET /proposals/{id}
	r.HandleFunc("/proposals/{id}",
		GetProposalHandlerFn(cdc, kb, ctx)).Methods("GET")
	// GET /accounts/{address}/proposals
	r.HandleFunc("/accounts/{address}/proposals",
		GetAccountProposalHandlerFn(cdc, kb, ctx)).Methods("GET")
	// POST /accounts/{address}/proposals
	r.HandleFunc("/accounts/{address}/proposals",
		SubmitProposalHandlerFn(cdc, kb, ctx)).Methods("POST")
	// GET /accounts/{address}/proposals/{id}/vote
	r.HandleFunc("/accounts/{address}/proposals/{id}/vote",
		GetAccountProposalsVoteHandlerFn(cdc, kb, ctx)).Methods("GET")
	// POST /accounts/{address}/proposals/{id}/vote
	r.HandleFunc("/accounts/{address}/proposals/{id}/vote",
		SubmitVoteHandlerFn(cdc, kb, ctx)).Methods("POST")
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

// GetProposalHandlerFn - http request handler to get a proposal
func GetProposalHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		proposalID := vars["id"]

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		key := simpleGov.GenerateProposalKey(int64(proposalID))

		res, err := ctx.Query(key, "proposal")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		// the query will return empty if there is no data for this bond
		if len(res) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		output, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)

	}
}

// GetAccountProposalHandlerFn - http request handler to get all proposals from an account
func GetAccountProposalHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// collect data
		vars := mux.Vars(r)
		address := vars["address"]

		// Get address of proposer
		bz, err := hex.DecodeString(address)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		proposerAddr := sdk.Address(bz)

		key := simpleGov.GenerateAccountProposalsKey(proposerAddr)

		res, err := ctx.Query(key, "proposal")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		// the query will return empty if there is no data for this bond
		if len(res) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		output, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
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

		// TODO submit proposal
		// TODO check which key corresponds

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

		// w.Write(output)
	}
}

// GetAccountProposalsVoteHandlerFn -  http request handler to get an account vote on a proposal
func GetAccountProposalsVoteHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		address := vars["address"]
		proposalID, err := strconv.Atoi(vars["id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		bz, err := hex.DecodeString(address)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		voterAddr := sdk.Address(bz)

		key := simpleGov.GenerateAccountProposalsVoteKey(int64(proposalID), voterAddr)
		res, err := ctx.Query(key, "proposal")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		// the query will return empty if there is no data for this bond
		if len(res) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		output, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}

// SubmitVoteHandlerFn - http request handler to vote on a proposal
func SubmitVoteHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		address := vars["address"]
		proposalID, err := strconv.Atoi(vars["id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

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
		bz, err := hex.DecodeString(address)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		voterAddr := sdk.Address(bz)

		key := simpleGov.GenerateAccountProposalsVoteKey(int64(proposalID), voterAddr)

		// TODO submit vote
		// w.Write(output)
	}
}
