package gov

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"time"
)

type GenesisState struct {
	StartingProposalID uint64                `json:"starting_proposal_id"`
	Deposits           []DepositWithMetadata `json:"deposits"`
	Votes              []VoteWithMetadata    `json:"votes"`
	Proposals          []Proposal            `json:"proposals"`
	DepositParams      DepositParams         `json:"deposit_params"`
	VotingParams       VotingParams          `json:"voting_params"`
	TallyParams        TallyParams           `json:"tally_params"`
}

type DepositWithMetadata struct {
	ProposalID uint64  `json:"proposal_id"`
	Deposit    Deposit `json:"deposit"`
}

type VoteWithMetadata struct {
	ProposalID uint64 `json:"proposal_id"`
	Vote       Vote   `json:"vote"`
}

type Proposal struct {
	Content `json:"content"` // Proposal content interface

	ProposalID       uint64             `json:"proposal_id"`        //  ID of the proposal
	Status           gov.ProposalStatus `json:"proposal_status"`    // Status of the Proposal {Pending, Active, Passed, Rejected}
	FinalTallyResult TallyResult        `json:"final_tally_result"` // Result of Tallys

	SubmitTime     time.Time `json:"submit_time"`      // Time of the block where TxGovSubmitProposal was included
	DepositEndTime time.Time `json:"deposit_end_time"` // Time that the Proposal would expire if deposit amount isn't met
	TotalDeposit   sdk.Coins `json:"total_deposit"`    // Current deposit on this proposal. Initial value is set at InitialDeposit

	VotingStartTime time.Time `json:"voting_start_time"` // Time of the block where MinDeposit was reached. -1 if MinDeposit is not reached
	VotingEndTime   time.Time `json:"voting_end_time"`   // Time that the VotingPeriod for this proposal will end and votes will be tallied
}

type Proposals []Proposal

type Deposit struct {
	ProposalID uint64         `json:"proposal_id"` //  proposalID of the proposal
	Depositor  sdk.AccAddress `json:"depositor"`   //  Address of the depositor
	Amount     sdk.Coins      `json:"amount"`      //  Deposit amount
}

type Deposits []Deposit

type Vote struct {
	ProposalID uint64         `json:"proposal_id"` //  proposalID of the proposal
	Voter      sdk.AccAddress `json:"voter"`       //  address of the voter
	Option     gov.VoteOption `json:"option"`      //  option from OptionSet chosen by the voter
}

type Votes []Vote

type VoteOption byte

// Vote options
const (
	OptionEmpty      VoteOption = 0x00
	OptionYes        VoteOption = 0x01
	OptionAbstain    VoteOption = 0x02
	OptionNo         VoteOption = 0x03
	OptionNoWithVeto VoteOption = 0x04
)

// Param around deposits for governance
type DepositParams struct {
	MinDeposit       sdk.Coins     `json:"min_deposit,omitempty"`        //  Minimum deposit for a proposal to enter voting period.
	MaxDepositPeriod time.Duration `json:"max_deposit_period,omitempty"` //  Maximum period for Atom holders to deposit on a proposal. Initial value: 2 months
}

type Content interface {
	GetTitle() string
	GetDescription() string
	ProposalRoute() string
	ProposalType() string
	ValidateBasic() sdk.Error
	String() string
}

type (
	// ProposalQueue
	ProposalQueue []uint64

	// ProposalStatus is a type alias that represents a proposal status as a byte
	ProposalStatus byte
)

//nolint
const (
	StatusNil           ProposalStatus = 0x00
	StatusDepositPeriod ProposalStatus = 0x01
	StatusVotingPeriod  ProposalStatus = 0x02
	StatusPassed        ProposalStatus = 0x03
	StatusRejected      ProposalStatus = 0x04
	StatusFailed        ProposalStatus = 0x05
)

type TallyParams struct {
	Quorum    sdk.Dec `json:"quorum,omitempty"`    //  Minimum percentage of total stake needed to vote for a result to be considered valid
	Threshold sdk.Dec `json:"threshold,omitempty"` //  Minimum proportion of Yes votes for proposal to pass. Initial value: 0.5
	Veto      sdk.Dec `json:"veto,omitempty"`      //  Minimum value of Veto votes to Total votes ratio for proposal to be vetoed. Initial value: 1/3
}

type VotingParams struct {
	VotingPeriod time.Duration `json:"voting_period,omitempty"` //  Length of the voting period.
}

type TallyResult struct {
	Yes        sdk.Int `json:"yes"`
	Abstain    sdk.Int `json:"abstain"`
	No         sdk.Int `json:"no"`
	NoWithVeto sdk.Int `json:"no_with_veto"`
}

func RegisterCodec(codec *codec.Codec) {
	codec.RegisterInterface((*Content)(nil), nil)
}
