package rest

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/gamarin2/cosmos-sdk/examples/simpleGov/x/simpleGovernance"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/gorilla/mux"
	"github.com/tendermint/go-crypto/keys"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	// GET /proposals
	r.HandleFunc("/proposals",
		GetProposalsHandlerFn(cdc, kb, ctx)).Methods("GET")
	// POST /proposals
	r.HandleFunc("/proposals",
		SubmitProposalHandlerFn(cdc, kb, ctx)).Methods("POST")
	// GET /proposals/{id}
	r.HandleFunc("/proposals/{id}",
		GetProposalHandlerFn(cdc, kb, ctx)).Methods("GET")
	// GET /proposals/{id}/votes
	r.HandleFunc("/proposals/{id}/votes",
		GetProposalVotesHandlerFn(cdc, kb, ctx)).Methods("GET")
	// POST /proposals/{id}/votes
	r.HandleFunc("/proposals/{id}/votes",
		SubmitVoteHandlerFn(cdc, kb, ctx)).Methods("POST")
	// GET /proposals/{id}/votes/{address}
	r.HandleFunc("/proposals/{id}/votes/{address}",
		GetProposalVoteHandlerFn(cdc, kb, ctx)).Methods("GET")
}

type proposeBody struct {
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Deposit          sdk.Coins `json:"deposit"` // use
	BlockLimit       int64     `json:"block_limit"`
	LocalAccountName string    `json:"name"`
	Password         string    `json:"password"`
	ChainID          string    `json:"chain_id"`
	Sequence         int64     `json:"sequence"`
}

type voteBody struct {
	Option           string `json:"option"`
	LocalAccountName string `json:"name"`
	Password         string `json:"password"`
	ChainID          string `json:"chain_id"`
}

// GetProposalsHandlerFn - http request handler to get a proposal
func GetProposalsHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		key := []byte("proposals")

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

// GetProposalHandlerFn - http request handler to get a proposal
func GetProposalHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		proposalID := vars["id"]
		proposalID, err := strconv.Atoi(vars["id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		key := simpleGovernance.GenerateProposalKey(int64(proposalID))

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

		// get local account Name
		info, err := kb.Get(proposeBody.LocalAccountName)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}
		proposerAddr := info.PubKey.Address()

		// build message
		msg := simpleGovernace.NewSubmitProposalMsg(title, description, proposeBody.BlockLimit, proposeBody.Deposit, proposerAddr)

		// sign Msg
		ctx = ctx.WithSequence(proposeBody.Sequence)
		ctx = ctx.WithChainID(proposeBody.ChainID)
		txBytes, err := ctx.SignAndBuild(proposeBody.LocalAccountName, proposeBody.Password, msg, cdc)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		// send Tx
		res, err := ctx.BroadcastTx(txBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
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

// GetProposalVotesHandlerFn - http request handler to get all proposals from an account
func GetProposalVotesHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// collect data
		vars := mux.Vars(r)
		address := vars["id"]

		proposalID, err := strconv.Atoi(vars["id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		key := simpleGovernance.GenerateProposalVotesVoteKey(int64(proposalID))

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

		// Get proposalID
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

		// get local account Name
		info, err := kb.Get(vote.LocalAccountName)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}
		voter := info.PubKey.Address()

		// TODO submit vote

		// build message
		msg := simpleGovernance.NewVoteMsg(int64(proposalID), vote.Option, voter)

		// sign
		ctx = ctx.WithChainID(vote.ChainID)
		txBytes, err := ctx.SignAndBuild(vote.LocalAccountName, vote.Password, msg, cdc)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		// send
		res, err := ctx.BroadcastTx(txBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
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

// GetAccountProposalsVoteHandlerFn -  http request handler to get an account vote on a proposal
func GetAccountProposalsVoteHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		proposalID := vars["id"]
		address := vars["address"]

		proposalID, err := strconv.Atoi(proposalID)
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

		key := simpleGovernance.GenerateAccountProposalsVoteKey(int64(proposalID), voterAddr)
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
