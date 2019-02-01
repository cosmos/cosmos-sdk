package client

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

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

// SendReq requests sending an amount of coins
type SendReq struct {
	Amount  sdk.Coins `json:"amount"`
	BaseReq BaseReq   `json:"base_req"`
}

// MsgDelegationsInput requests a delegation
type MsgDelegationsInput struct {
	BaseReq       BaseReq        `json:"base_req"`
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"` // in bech32
	ValidatorAddr sdk.ValAddress `json:"validator_addr"` // in bech32
	Delegation    sdk.Coin       `json:"delegation"`
}

// MsgUndelegateInput request an undelegation
type MsgUndelegateInput struct {
	BaseReq       BaseReq        `json:"base_req"`
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"` // in bech32
	ValidatorAddr sdk.ValAddress `json:"validator_addr"` // in bech32
	SharesAmount  sdk.Dec        `json:"shares"`
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

// DepositReq requests a deposit of an amount of coins
type DepositReq struct {
	BaseReq   BaseReq        `json:"base_req"`
	Depositor sdk.AccAddress `json:"depositor"` // Address of the depositor
	Amount    sdk.Coins      `json:"amount"`    // Coins to add to the proposal's deposit
}

// AddrSeed combines an Address with the mnemonic of the private key to that address
type AddrSeed struct {
	Address  sdk.AccAddress
	Seed     string
	Name     string
	Password string
}

// NewKeyBody request a new key
type NewKeyBody struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Mnemonic string `json:"mnemonic"`
	Account  int    `json:"account,string,omitempty"`
	Index    int    `json:"index,string,omitempty"`
}

// RecoverKeyBody recovers a key
type RecoverKeyBody struct {
	Password string `json:"password"`
	Mnemonic string `json:"mnemonic"`
	Account  int    `json:"account,string,omitempty"`
	Index    int    `json:"index,string,omitempty"`
}

// UpdateKeyReq requests updating a key
type UpdateKeyReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// DeleteKeyReq requests deleting a key
type DeleteKeyReq struct {
	Password string `json:"password"`
}

// BroadcastReq requests broadcasting a transaction
type BroadcastReq struct {
	Tx     auth.StdTx `json:"tx"`
	Return string     `json:"return"`
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

// NewBaseReq creates a new basic request instance and sanitizes its values
func NewBaseReq(from, password, memo, chainID string, gas, gasAdjustment string,
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
