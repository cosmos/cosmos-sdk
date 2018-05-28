package module_tutorial

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stake "github.coms/cosmos/cosmos-sdk/x/stake"
	"reflect"
)

type Proposal struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Submitter   sdk.Address `json:"submitter"`
	SubmitBlock int64       `json:"submit_block"`
	State       string      `json:"state"`
	Deposit     sdk.Coins   `json:"deposit"`

	YesVotes     int64 `json:"yes_votes"`
	NoVotes      int64 `json:"no_votes"`
	AbstainVotes int64 `json:"abstain_votes"`
}

type ProposalQueue []int64

func (p Proposal) updateTally(option string, amount int64) {
	switch option {
	case "Yes":
		proposal.YesVotes += amount
		return nil
	case "No":
		proposal.NoVotes += amount
		return nil
	case "Abstain":
		proposal.AbstainVotes += amount
	default:
		panic("Should not happen, update tally only takes option that comes from vote_msg, options should be checked in ValidateBasic()")
	}
}

//--------------------------------------------------------
//--------------------------------------------------------

type SubmitProposalMsg struct {
	Title       string
	Description string
	Deposit     sdk.Coins
	Submitter   sdk.Address
}

func NewSubmitProposalMsg(title string, description string, deposit sdk.Coins, submitter sdk.Address) SubmitProposalMsg {
	return SubmitProposalMsg{
		Title:       title,
		Description: description,
		Deposit:     deposit,
		Submitter:   submitter,
	}
}

// Implements Msg
func (msg SubmitProposalMsg) Type() string {
	return "simple_gov"
}

// Implements Msg
func (msg SubmitProposalMsg) Get(key interface{}) (value interface{}) {
	return nil
}

// Implements Msg
func (msg SubmitProposalMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// Implements Msg
func (msg SubmitProposalMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Submitter}
}

// Implements Msg
func (msg SubmitProposalMsg) ValidateBasic() sdk.Error {
	if len(msg.Submitter) == 0 {
		return sdk.ErrUnrecognizedAddress(msg.Submitter).Trace("")
	}

	if len(msg.Title) <= 0 {
		return sdk.ErrUnauthorized("").Trace("Title cannot be empty")
	}

	if len(msg.Description) <= 0 {
		return sdk.ErrUnauthorized("").Trace("Description cannot be empty")
	}

	if !msg.Deposit.IsValid() {
		return sdk.ErrUnauthorized("").Trace("Deposit is not valid")
	}

	if !msg.Deposit.IsPositive() {
		return sdk.ErrUnauthorized("").Trace("Deposit cannot be negative")
	}

	return nil
}

func (msg SubmitProposalMsg) String() string {
	return fmt.Sprintf("SubmitProposalMsg{%v, %v}", msg.Title, msg.Description)
}

//--------------------------------------------------------
//--------------------------------------------------------

type VoteMsg struct {
	ProposalID int64
	Option     string
	Voter      sdk.Address
}

func VoteMsg(proposalID int64, option string, voter sdk.Address) VoteMsg {
	return VoteMsg{
		ProposalID: proposalID,
		Option:     option,
		Voter:      voter,
	}
}

// Implements Msg
func (msg VoteMsg) Type() string {
	return "simple_gov"
}

// Implements Msg
func (msg VoteMsg) Get(key interface{}) (value interface{}) {
	return nil
}

// Implements Msg
func (msg VoteMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// Implements Msg
func (msg VoteMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Voter}
}

// Implements Msg
func (msg VoteMsg) ValidateBasic() sdk.Error {
	if len(msg.Voter) == 0 {
		return sdk.ErrUnrecognizedAddress(msg.Voter).Trace("")
	}

	if len(msg.ProposalID) <= 0 {
		return sdk.ErrUnauthorized("").Trace("ProposalID cannot be negative")
	}

	if msg.Option != "Yes" || msg.Option != "No" || msg.Option != "Abstain" {
		return ErrInvalidOption()
	}

	return nil
}

func (msg VoteMsg) String() string {
	return fmt.Sprintf("VoteMsg{%v, %v}", msg.Title, msg.Description)
}

