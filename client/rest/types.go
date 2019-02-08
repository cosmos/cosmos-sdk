package rest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// GasEstimateResponse defines a response definition for tx gas estimation.
type GasEstimateResponse struct {
	GasEstimate uint64 `json:"gas_estimate"`
}

// BaseReq defines a structure that can be embedded in other request structures
// that all share common "base" fields.
type BaseReq struct {
	From          string       `json:"from"`
	Password      string       `json:"password"`
	Memo          string       `json:"memo"`
	ChainID       string       `json:"chain_id"`
	AccountNumber uint64       `json:"account_number"`
	Sequence      uint64       `json:"sequence"`
	Fees          sdk.Coins    `json:"fees"`
	GasPrices     sdk.DecCoins `json:"gas_prices"`
	Gas           string       `json:"gas"`
	GasAdjustment string       `json:"gas_adjustment"`
	GenerateOnly  bool         `json:"generate_only"`
	Simulate      bool         `json:"simulate"`
}

// NewBaseReq creates a new basic request instance and sanitizes its values
func NewBaseReq(
	from, password, memo, chainID string, gas, gasAdjustment string,
	accNumber, seq uint64, fees sdk.Coins, gasPrices sdk.DecCoins, genOnly, simulate bool,
) BaseReq {

	return BaseReq{
		From:          strings.TrimSpace(from),
		Password:      password,
		Memo:          strings.TrimSpace(memo),
		ChainID:       strings.TrimSpace(chainID),
		Fees:          fees,
		GasPrices:     gasPrices,
		Gas:           strings.TrimSpace(gas),
		GasAdjustment: strings.TrimSpace(gasAdjustment),
		AccountNumber: accNumber,
		Sequence:      seq,
		GenerateOnly:  genOnly,
		Simulate:      simulate,
	}
}

// Sanitize performs basic sanitization on a BaseReq object.
func (br BaseReq) Sanitize() BaseReq {
	return NewBaseReq(
		br.From, br.Password, br.Memo, br.ChainID, br.Gas, br.GasAdjustment,
		br.AccountNumber, br.Sequence, br.Fees, br.GasPrices, br.GenerateOnly, br.Simulate,
	)
}

// ValidateBasic performs basic validation of a BaseReq. If custom validation
// logic is needed, the implementing request handler should perform those
// checks manually.
func (br BaseReq) ValidateBasic(w http.ResponseWriter) bool {
	if !br.GenerateOnly && !br.Simulate {
		switch {
		case len(br.Password) == 0:
			WriteErrorResponse(w, http.StatusUnauthorized, "password required but not specified")
			return false

		case len(br.ChainID) == 0:
			WriteErrorResponse(w, http.StatusUnauthorized, "chain-id required but not specified")
			return false

		case !br.Fees.IsZero() && !br.GasPrices.IsZero():
			// both fees and gas prices were provided
			WriteErrorResponse(w, http.StatusBadRequest, "cannot provide both fees and gas prices")
			return false

		case !br.Fees.IsValid() && !br.GasPrices.IsValid():
			// neither fees or gas prices were provided
			WriteErrorResponse(w, http.StatusPaymentRequired, "invalid fees or gas prices provided")
			return false
		}
	}

	if len(br.From) == 0 {
		WriteErrorResponse(w, http.StatusUnauthorized, "name or address required but not specified")
		return false
	}

	return true
}

/*
ReadRESTReq is a simple convenience wrapper that reads the body and
unmarshals to the req interface. Returns false if errors occurred.

  Usage:
    type SomeReq struct {
      BaseReq            `json:"base_req"`
      CustomField string `json:"custom_field"`
		}

    req := new(SomeReq)
    if ok := ReadRESTReq(w, r, cdc, req); !ok {
        return
    }
*/
func ReadRESTReq(w http.ResponseWriter, r *http.Request, cdc *codec.Codec, req interface{}) bool {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return false
	}

	err = cdc.UnmarshalJSON(body, req)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to decode JSON payload: %s", err))
		return false
	}

	return true
}

// AddrSeed combines an Address with the mnemonic of the private key to that address
type AddrSeed struct {
	Address  sdk.AccAddress
	Seed     string
	Name     string
	Password string
}

// SendReq requests sending an amount of coins
type SendReq struct {
	Amount  sdk.Coins `json:"amount"`
	BaseReq BaseReq   `json:"base_req"`
}

// MsgBeginRedelegateInput request to begin a redelegation
type MsgBeginRedelegateInput struct {
	BaseReq          BaseReq        `json:"base_req"`
	DelegatorAddr    sdk.AccAddress `json:"delegator_addr"`     // in bech32
	ValidatorSrcAddr sdk.ValAddress `json:"validator_src_addr"` // in bech32
	ValidatorDstAddr sdk.ValAddress `json:"validator_dst_addr"` // in bech32
	SharesAmount     sdk.Dec        `json:"shares"`
}

// PostProposalReq requests a proposals
type PostProposalReq struct {
	BaseReq        BaseReq        `json:"base_req"`
	Title          string         `json:"title"`           //  Title of the proposal
	Description    string         `json:"description"`     //  Description of the proposal
	ProposalType   string         `json:"proposal_type"`   //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
	Proposer       sdk.AccAddress `json:"proposer"`        //  Address of the proposer
	InitialDeposit sdk.Coins      `json:"initial_deposit"` // Coins to add to the proposal's deposit
}

// BroadcastReq requests broadcasting a transaction
type BroadcastReq struct {
	Tx     auth.StdTx `json:"tx"`
	Return string     `json:"return"`
}

// DepositReq requests a deposit of an amount of coins
type DepositReq struct {
	BaseReq   BaseReq        `json:"base_req"`
	Depositor sdk.AccAddress `json:"depositor"` // Address of the depositor
	Amount    sdk.Coins      `json:"amount"`    // Coins to add to the proposal's deposit
}

// VoteReq requests sending a vote
type VoteReq struct {
	BaseReq BaseReq        `json:"base_req"`
	Voter   sdk.AccAddress `json:"voter"`  //  address of the voter
	Option  string         `json:"option"` //  option from OptionSet chosen by the voter
}

// UnjailReq request unjailing
type UnjailReq struct {
	BaseReq BaseReq `json:"base_req"`
}
