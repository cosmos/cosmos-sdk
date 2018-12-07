package gov

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

// Vote
type Vote struct {
	Voter      sdk.AccAddress `json:"voter"`       //  address of the voter
	ProposalID uint64         `json:"proposal_id"` //  proposalID of the proposal
	Option     VoteOption     `json:"option"`      //  option from OptionSet chosen by the voter
}

// Returns whether 2 votes are equal
func (voteA Vote) Equals(voteB Vote) bool {
	return voteA.Voter.Equals(voteB.Voter) && voteA.ProposalID == voteB.ProposalID && voteA.Option == voteB.Option
}

// Returns whether a vote is empty
func (voteA Vote) Empty() bool {
	voteB := Vote{}
	return voteA.Equals(voteB)
}

// Deposit
type Deposit struct {
	Depositor  sdk.AccAddress `json:"depositor"`   //  Address of the depositor
	ProposalID uint64         `json:"proposal_id"` //  proposalID of the proposal
	Amount     sdk.Coins      `json:"amount"`      //  Deposit amount
}

// Returns whether 2 deposits are equal
func (depositA Deposit) Equals(depositB Deposit) bool {
	return depositA.Depositor.Equals(depositB.Depositor) && depositA.ProposalID == depositB.ProposalID && depositA.Amount.IsEqual(depositB.Amount)
}

// Returns whether a deposit is empty
func (depositA Deposit) Empty() bool {
	depositB := Deposit{}
	return depositA.Equals(depositB)
}

// Type that represents VoteOption as a byte
type VoteOption byte

//nolint
const (
	OptionEmpty      VoteOption = 0x00
	OptionYes        VoteOption = 0x01
	OptionAbstain    VoteOption = 0x02
	OptionNo         VoteOption = 0x03
	OptionNoWithVeto VoteOption = 0x04
)

// String to proposalType byte.  Returns ff if invalid.
func VoteOptionFromString(str string) (VoteOption, error) {
	switch str {
	case "Yes":
		return OptionYes, nil
	case "Abstain":
		return OptionAbstain, nil
	case "No":
		return OptionNo, nil
	case "NoWithVeto":
		return OptionNoWithVeto, nil
	default:
		return VoteOption(0xff), errors.Errorf("'%s' is not a valid vote option", str)
	}
}

// Is defined VoteOption
func validVoteOption(option VoteOption) bool {
	if option == OptionYes ||
		option == OptionAbstain ||
		option == OptionNo ||
		option == OptionNoWithVeto {
		return true
	}
	return false
}

// Marshal needed for protobuf compatibility
func (vo VoteOption) Marshal() ([]byte, error) {
	return []byte{byte(vo)}, nil
}

// Unmarshal needed for protobuf compatibility
func (vo *VoteOption) Unmarshal(data []byte) error {
	*vo = VoteOption(data[0])
	return nil
}

// Marshals to JSON using string
func (vo VoteOption) MarshalJSON() ([]byte, error) {
	return json.Marshal(vo.String())
}

// Unmarshals from JSON assuming Bech32 encoding
func (vo *VoteOption) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return nil
	}

	bz2, err := VoteOptionFromString(s)
	if err != nil {
		return err
	}
	*vo = bz2
	return nil
}

// Turns VoteOption byte to String
func (vo VoteOption) String() string {
	switch vo {
	case OptionYes:
		return "Yes"
	case OptionAbstain:
		return "Abstain"
	case OptionNo:
		return "No"
	case OptionNoWithVeto:
		return "NoWithVeto"
	default:
		return ""
	}
}

// For Printf / Sprintf, returns bech32 when using %s
// nolint: errcheck
func (vo VoteOption) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(fmt.Sprintf("%s", vo.String())))
	default:
		s.Write([]byte(fmt.Sprintf("%v", byte(vo))))
	}
}
