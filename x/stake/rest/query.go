package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/query"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/stack"

	"github.com/cosmos/gaia/modules/stake"
	scmds "github.com/cosmos/gaia/modules/stake/commands"

	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/tmlibs/common"
)

// RegisterQueryCandidate is a mux.Router handler that exposes GET
// method access on route /query/stake/candidate/{pubkey} to query a candidate
func RegisterQueryCandidate(r *mux.Router) error {
	r.HandleFunc("/query/stake/candidate/{pubkey}", queryCandidate).Methods("GET")
	return nil
}

// RegisterQueryCandidates is a mux.Router handler that exposes GET
// method access on route /query/stake/candidate to query the group of all candidates
func RegisterQueryCandidates(r *mux.Router) error {
	r.HandleFunc("/query/stake/candidates", queryCandidates).Methods("GET")
	return nil
}

// RegisterQueryDelegatorBond is a mux.Router handler that exposes GET
// method access on route /query/stake/candidate/{pubkey} to query a candidate
func RegisterQueryDelegatorBond(r *mux.Router) error {
	r.HandleFunc("/query/stake/delegator/{address}/{pubkey}", queryDelegatorBond).Methods("GET")
	return nil
}

// RegisterQueryDelegatorCandidates is a mux.Router handler that exposes GET
// method access on route /query/stake/candidate to query the group of all candidates
func RegisterQueryDelegatorCandidates(r *mux.Router) error {
	r.HandleFunc("/query/stake/delegator_candidates/{address}", queryDelegatorCandidates).Methods("GET")
	return nil
}

//---------------------------------------------------------------------

// queryCandidate is the HTTP handlerfunc to query a candidate
// it expects a query string
func queryCandidate(w http.ResponseWriter, r *http.Request) {

	// get the arguments object
	args := mux.Vars(r)
	prove := !viper.GetBool(commands.FlagTrustNode) // from viper because defined when starting server

	// get the pubkey
	pkArg := args["pubkey"]
	pk, err := scmds.GetPubKey(pkArg)
	if err != nil {
		common.WriteError(w, err)
		return
	}

	// get the candidate
	var candidate stake.Candidate
	key := stack.PrefixedKey(stake.Name(), stake.GetCandidateKey(pk))
	height, err := query.GetParsed(key, &candidate, query.GetHeight(), prove)
	if client.IsNoDataErr(err) {
		err := fmt.Errorf("candidate bytes are empty for pubkey: %q", pkArg)
		common.WriteError(w, err)
		return
	} else if err != nil {
		common.WriteError(w, err)
		return
	}

	// write the output
	err = query.FoutputProof(w, candidate, height)
	if err != nil {
		common.WriteError(w, err)
	}
}

// queryCandidates is the HTTP handlerfunc to query the group of all candidates
func queryCandidates(w http.ResponseWriter, r *http.Request) {

	var pks []crypto.PubKey

	prove := !viper.GetBool(commands.FlagTrustNode) // from viper because defined when starting server
	key := stack.PrefixedKey(stake.Name(), stake.CandidatesPubKeysKey)
	height, err := query.GetParsed(key, &pks, query.GetHeight(), prove)
	if err != nil {
		common.WriteError(w, err)
		return
	}

	err = query.FoutputProof(w, pks, height)
	if err != nil {
		common.WriteError(w, err)
	}
}

// queryDelegatorBond is the HTTP handlerfunc to query a delegator bond it
// expects a query string
func queryDelegatorBond(w http.ResponseWriter, r *http.Request) {

	// get the arguments object
	args := mux.Vars(r)
	prove := !viper.GetBool(commands.FlagTrustNode) // from viper because defined when starting server

	// get the pubkey
	pkArg := args["pubkey"]
	pk, err := scmds.GetPubKey(pkArg)
	if err != nil {
		common.WriteError(w, err)
		return
	}

	// get the delegator actor
	delegatorAddr := args["address"]
	delegator, err := commands.ParseActor(delegatorAddr)
	if err != nil {
		common.WriteError(w, err)
		return
	}
	delegator = coin.ChainAddr(delegator)

	// get the bond
	var bond stake.DelegatorBond
	key := stack.PrefixedKey(stake.Name(), stake.GetDelegatorBondKey(delegator, pk))
	height, err := query.GetParsed(key, &bond, query.GetHeight(), prove)
	if client.IsNoDataErr(err) {
		err := fmt.Errorf("bond bytes are empty for pubkey: %q, address: %q", pkArg, delegatorAddr)
		common.WriteError(w, err)
		return
	} else if err != nil {
		common.WriteError(w, err)
		return
	}

	// write the output
	err = query.FoutputProof(w, bond, height)
	if err != nil {
		common.WriteError(w, err)
	}
}

// queryDelegatorCandidates is the HTTP handlerfunc to query a delegator bond it
// expects a query string
func queryDelegatorCandidates(w http.ResponseWriter, r *http.Request) {

	// get the arguments object
	args := mux.Vars(r)
	prove := !viper.GetBool(commands.FlagTrustNode) // from viper because defined when starting server

	// get the delegator actor
	delegatorAddr := args["address"]
	delegator, err := commands.ParseActor(delegatorAddr)
	if err != nil {
		common.WriteError(w, err)
		return
	}
	delegator = coin.ChainAddr(delegator)

	// get the bond
	var bond stake.DelegatorBond
	key := stack.PrefixedKey(stake.Name(), stake.GetDelegatorBondsKey(delegator))
	height, err := query.GetParsed(key, &bond, query.GetHeight(), prove)
	if client.IsNoDataErr(err) {
		err := fmt.Errorf("bond bytes are empty for address: %q", delegatorAddr)
		common.WriteError(w, err)
		return
	} else if err != nil {
		common.WriteError(w, err)
		return
	}

	// write the output
	err = query.FoutputProof(w, bond, height)
	if err != nil {
		common.WriteError(w, err)
	}
}
