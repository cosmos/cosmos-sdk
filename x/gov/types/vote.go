package types

import (
	"encoding/json"
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewVote creates a new Vote instance
//nolint:interfacer
func NewVote(proposalID uint64, voter sdk.AccAddress, subvotes []SubVote) Vote {
	return Vote{proposalID, voter.String(), subvotes}
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
		if vote.String() != other[i].String() {
			return false
		}
	}

	return true
}

func (v Votes) String() string {
	if len(v) == 0 {
		return "[]"
	}
	out := fmt.Sprintf("Votes for Proposal %d:", v[0].ProposalId)
	for _, vot := range v {
		out += fmt.Sprintf("\n  %s: %s", vot.Voter, vot.SubVotes)
	}
	return out
}

// Empty returns whether a vote is empty.
func (v Vote) Empty() bool {
	return v.String() == Vote{}.String()
}

// NewSubVote creates a new Vote instance
//nolint:interfacer
func NewSubVote(option VoteOption, rate int64) SubVote {
	return SubVote{option, sdk.NewDec(rate)}
}

func (v SubVote) String() string {
	out, _ := json.Marshal(v)
	return string(out)
}

// SubVotes describes array of SubVote
type SubVotes []SubVote

func (v SubVotes) String() string {
	out, _ := json.Marshal(v)
	return string(out)
}

// ValidSubVote returns true if the sub vote is valid and false otherwise.
func ValidSubVote(subvote SubVote) bool {
	if !subvote.Rate.IsPositive() {
		return false
	}
	return ValidVoteOption(subvote.Option)
}

// VoteOptionFromString returns a VoteOption from a string. It returns an error
// if the string is invalid.
func VoteOptionFromString(str string) (VoteOption, error) {
	option, ok := VoteOption_value[str]
	if !ok {
		return OptionEmpty, fmt.Errorf("'%s' is not a valid vote option", str)
	}
	return VoteOption(option), nil
}

// SubVotesFromString returns a SubVotes from a string. It returns an error
// if the string is invalid.
func SubVotesFromString(str string) (SubVotes, error) {
	subvotes := SubVotes{}
	for _, subvote := range strings.Split(str, ",") {
		fields := strings.Split(subvote, "=")
		option, err := VoteOptionFromString(fields[0])
		if err != nil {
			return subvotes, err
		}
		if len(fields) < 2 {
			return subvotes, fmt.Errorf("rate field does not exist for %s opion", fields[0])
		}
		rate, err := sdk.NewDecFromStr(fields[1])
		if err != nil {
			return subvotes, err
		}
		subvotes = append(subvotes, SubVote{
			option,
			rate,
		})
	}
	return subvotes, nil
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

// Format implements the fmt.Formatter interface.
func (vo VoteOption) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(vo.String()))
	default:
		s.Write([]byte(fmt.Sprintf("%v", byte(vo))))
	}
}
