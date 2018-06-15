package simpleGovernance

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Proposal defines the basic propierties of a staking proposal
type Proposal struct {
	Title       string      `json:"title"`        // Title of the proposal
	Description string      `json:"description"`  // Description of the proposal
	Submitter   sdk.Address `json:"submitter"`    // Account address of the proposer
	SubmitBlock int64       `json:"submit_block"` // Block height from which the proposal is open for votations
	State       string      `json:"state"`        // One of Open, Accepted, Rejected
	Deposit     sdk.Coins   `json:"deposit"`      // Coins deposited in escrow

	YesVotes     int64 `json:"yes_votes"`     // Total Yes votes
	NoVotes      int64 `json:"no_votes"`      // Total No votes
	AbstainVotes int64 `json:"abstain_votes"` // Total Abstain votes
}

// NewProposal validates deposit and creates a new proposal
func NewProposal(
	title string,
	description string,
	submitter sdk.Address,
	blockHeight int64,
	deposit sdk.Coins) Proposal {
	return Proposal{
		Title:        title,
		Description:  description,
		Submitter:    submitter,
		SubmitBlock:  blockHeight,
		State:        "Open",
		Deposit:      deposit,
		YesVotes:     0,
		NoVotes:      0,
		AbstainVotes: 0,
	}
}

func (p *Proposal) updateTally(option string, amount int64) sdk.Error {
	switch option {
	case "Yes":
		p.YesVotes += amount
		return nil
	case "No":
		p.NoVotes += amount
		return nil
	case "Abstain":
		p.AbstainVotes += amount
		return nil
	default:
		return ErrInvalidOption("Invalid option: " + option)
	}
}

// IsOpen checks if proposal is open for votations
func (p Proposal) IsOpen() bool {
	if p.State == "Open" {
		return true
	}
	return false
}

// ProposalQueue stores the proposals IDs
type ProposalQueue []int64

// IsEmpty checks if the ProposalQueue is empty
func (pq ProposalQueue) isEmpty() bool {
	if len(pq) == 0 {
		return true
	}
	return false
}

//--------------------------------------------------------
//--------------------------------------------------------

//SubmitProposalMsg defines a
type SubmitProposalMsg struct {
	Title       string
	Description string
	Deposit     sdk.Coins
	Submitter   sdk.Address
}

// NewSubmitProposalMsg submits a message with a new proposal
func NewSubmitProposalMsg(title string, description string, votingWindow int64, deposit sdk.Coins, submitter sdk.Address) SubmitProposalMsg {
	return SubmitProposalMsg{
		Title:       title,
		Description: description,
		Deposit:     deposit,
		Submitter:   submitter,
	}
}

// Implements Msg
func (msg SubmitProposalMsg) Type() string {
	return "simpleGov"
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
		return sdk.ErrInvalidAddress("Invalid address: " + msg.Submitter.String())
	}
	if len(strings.TrimSpace(msg.Title)) <= 0 {
		return ErrInvalidTitle()
	}

	if len(strings.TrimSpace(msg.Description)) <= 0 {
		return ErrInvalidDescription()
	}

	if !msg.Deposit.IsValid() {
		return sdk.ErrInvalidCoins("Deposit is not valid")
	}

	if !msg.Deposit.IsPositive() {
		return sdk.ErrInvalidCoins("Deposit cannot be negative")
	}

	return nil
}

func (msg SubmitProposalMsg) String() string {
	return fmt.Sprintf("SubmitProposalMsg{%v, %v}", msg.Title, msg.Description)
}

//--------------------------------------------------------
//--------------------------------------------------------

// VoteMsg defines the msg of a staker containing the vote option to an
// specific proposal
type VoteMsg struct {
	ProposalID int64
	Option     string
	Voter      sdk.Address
}

// NewVoteMsg creates a VoteMsg instance
func NewVoteMsg(proposalID int64, option string, voter sdk.Address) VoteMsg {
	// by default a nil option is an abstention
	if option == "" {
		option = "Abstain"
	}
	return VoteMsg{
		ProposalID: proposalID,
		Option:     option,
		Voter:      voter,
	}
}

// Implements Msg
func (msg VoteMsg) Type() string {
	return "simpleGov"
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

func isValidOption(option string) bool {
	options := []string{"Yes", "No", "Abstain"}
	for _, value := range options {
		if value == option {
			return true
		}
	}
	return false
}

// Implements Msg
func (msg VoteMsg) ValidateBasic() sdk.Error {
	if len(msg.Voter) == 0 {
		return sdk.ErrInvalidAddress("Invalid address: " + msg.Voter.String())
	}
	if msg.ProposalID <= 0 {
		return ErrInvalidProposalID("ProposalID cannot be negative")
	}
	if !isValidOption(msg.Option) {
		return ErrInvalidOption("Invalid voting option: " + msg.Option)
	}
	if len(strings.TrimSpace(msg.Option)) <= 0 {
		return ErrInvalidOption("Option can't be blank")
	}

	return nil
}

// Implements Msg
func (msg VoteMsg) String() string {
	return fmt.Sprintf("VoteMsg{%v, %v, %v}", msg.ProposalID, msg.Voter, msg.Option)
}
