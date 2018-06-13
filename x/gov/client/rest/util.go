package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/pkg/errors"
)

//type request interface {
//	Validate(w http.ResponseWriter) bool
//}

type baseReq struct {
	Name          string `json:"name"`
	Password      string `json:"password"`
	ChainID       string `json:"chain_id"`
	AccountNumber int64  `json:"account_number"`
	Sequence      int64  `json:"sequence"`
	Gas           int64  `json:"gas"`
}

type postProposalReq struct {
	Title          string    `json:"title"`           //  Title of the proposal
	Description    string    `json:"description"`     //  Description of the proposal
	ProposalType   string    `json:"proposal_type"`   //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
	Proposer       string    `json:"proposer"`        //  Address of the proposer
	InitialDeposit sdk.Coins `json:"initial_deposit"` // Coins to add to the proposal's deposit
	BaseReq        baseReq   `json:"base_req"`
}

type depositReq struct {
	ProposalID int64     `json:"proposalID"` // ID of the proposal
	Depositer  string    `json:"depositer"`  // Address of the depositer
	Amount     sdk.Coins `json:"amount"`     // Coins to add to the proposal's deposit
	BaseReq    baseReq   `json:"base_req"`
}

type voteReq struct {
	Voter      string  `json:"voter"`      //  address of the voter
	ProposalID int64   `json:"proposalID"` //  proposalID of the proposal
	Option     string  `json:"option"`     //  option from OptionSet chosen by the voter
	BaseReq    baseReq `json:"base_req"`
}

func buildReq(w http.ResponseWriter, r *http.Request, req interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return err
	}
	err = json.Unmarshal(body, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return err
	}
	return nil
}

func (req baseReq) baseReqValidate(w http.ResponseWriter) bool {
	if len(req.Name) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		err := errors.Errorf("Name required but not specified")
		w.Write([]byte(err.Error()))
		return false
	}

	if len(req.Password) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		err := errors.Errorf("Password required but not specified")
		w.Write([]byte(err.Error()))
		return false
	}

	// if len(req.ChainID) == 0 {
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	err := errors.Errorf("ChainID required but not specified")
	// 	w.Write([]byte(err.Error()))
	// 	return false
	// }

	if req.AccountNumber < 0 {
		w.WriteHeader(http.StatusUnauthorized)
		err := errors.Errorf("Account Number required but not specified")
		w.Write([]byte(err.Error()))
		return false
	}

	if req.Sequence < 0 {
		w.WriteHeader(http.StatusUnauthorized)
		err := errors.Errorf("Sequence required but not specified")
		w.Write([]byte(err.Error()))
		return false
	}
	return true
}

func (req postProposalReq) Validate(w http.ResponseWriter) bool {
	if len(req.Title) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		err := errors.Errorf("Title required but not specified")
		w.Write([]byte(err.Error()))
		return false
	}

	if len(req.ProposalType) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		err := errors.Errorf("ProposalType required but not specified")
		w.Write([]byte(err.Error()))
		return false
	}

	if len(req.Proposer) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		err := errors.Errorf("Proposer required but not specified")
		w.Write([]byte(err.Error()))
		return false
	}

	if len(req.InitialDeposit) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		err := errors.Errorf("InitialDeposit required but not specified")
		w.Write([]byte(err.Error()))
		return false
	}
	return req.BaseReq.baseReqValidate(w)
}

func (req depositReq) Validate(w http.ResponseWriter) bool {
	if len(req.Depositer) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		err := errors.Errorf("Depositer required but not specified")
		w.Write([]byte(err.Error()))
		return false
	}

	if len(req.Amount) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		err := errors.Errorf("Amount required but not specified")
		w.Write([]byte(err.Error()))
		return false
	}
	return req.BaseReq.baseReqValidate(w)
}

func (req voteReq) Validate(w http.ResponseWriter) bool {
	if len(req.Voter) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		err := errors.Errorf("Voter required but not specified")
		w.Write([]byte(err.Error()))
		return false
	}

	if len(req.Option) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		err := errors.Errorf("Option required but not specified")
		w.Write([]byte(err.Error()))
		return false
	}
	return req.BaseReq.baseReqValidate(w)
}

func signAndBuild(w http.ResponseWriter, ctx context.CoreContext, baseReq baseReq, msg sdk.Msg, cdc *wire.Codec) {
	ctx = ctx.WithAccountNumber(baseReq.AccountNumber)
	ctx = ctx.WithSequence(baseReq.Sequence)

	// add gas to context
	ctx = ctx.WithGas(baseReq.Gas)

	txBytes, err := ctx.SignAndBuild(baseReq.Name, baseReq.Password, msg, cdc)
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
