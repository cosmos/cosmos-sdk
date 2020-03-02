package types

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewVote creates a new Vote instance
func NewVote(proposalID uint64, voter sdk.AccAddress, option VoteOption) Vote {
	return Vote{proposalID, voter, option}
}

func (v Vote) String() string {
	out, _ := yaml.Marshal(v)
	return string(out)
}

// Votes is a collection of Vote objects
type Votes []Vote

// Equal returns true if two slices (order-dependant) of votes are equal.
func (v Votes) Equal(other Votes) bool {
	if len(v) != len(other) {
		return false
	}

	for i, vote := range v {
		if !vote.Equal(other[i]) {
			return false
		}
	}

	return true
}

func (v Votes) String() string {
	if len(v) == 0 {
		return "[]"
	}
	out := fmt.Sprintf("Votes for Proposal %d:", v[0].ProposalID)
	for _, vot := range v {
		out += fmt.Sprintf("\n  %s: %s", vot.Voter, vot.Option)
	}
	return out
}

// Empty returns whether a vote is empty.
func (v Vote) Empty() bool {
	return v.Equal(Vote{})
}

// VoteOptionFromString returns a VoteOption from a string. It returns an error
// if the string is invalid.
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
		return VoteOption(0xff), fmt.Errorf("'%s' is not a valid vote option", str)
	}
}

// ValidVoteOption returns true if the vote option is valid and false otherwise.
func ValidVoteOption(option VoteOption) bool {
	if option == OptionYes ||
		option == OptionAbstain ||
		option == OptionNo ||
		option == OptionNoWithVeto {
		return true
	}
	return false
}

// Marshal needed for protobuf compatibility.
func (vo VoteOption) Marshal() ([]byte, error) {
	return []byte{byte(vo)}, nil
}

// Unmarshal needed for protobuf compatibility.
func (vo *VoteOption) Unmarshal(data []byte) error {
	*vo = VoteOption(data[0])
	return nil
}

// Marshals to JSON using string.
func (vo VoteOption) MarshalJSON() ([]byte, error) {
	return json.Marshal(vo.String())
}

// UnmarshalJSON decodes from JSON assuming Bech32 encoding.
func (vo *VoteOption) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	bz2, err := VoteOptionFromString(s)
	if err != nil {
		return err
	}

	*vo = bz2
	return nil
}

// String implements the Stringer interface.
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

// Format implements the fmt.Formatter interface.
func (vo VoteOption) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(vo.String()))
	default:
		s.Write([]byte(fmt.Sprintf("%v", byte(vo))))
	}
}
