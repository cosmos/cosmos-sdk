// Package v040 is copy-pasted from:
// https://github.com/cosmos/cosmos-sdk/blob/v0.41.1/x/gov/types/vote.go
package v040

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// NewVote creates a new Vote instance
//nolint:interfacer
func NewVote(proposalID uint64, voter sdk.AccAddress, option types.VoteOption) Vote {
	return Vote{proposalID, voter.String(), option}
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
		out += fmt.Sprintf("\n  %s: %s", vot.Voter, vot.Option)
	}
	return out
}

// Empty returns whether a vote is empty.
func (v Vote) Empty() bool {
	return v.String() == Vote{}.String()
}

// VoteOptionFromString returns a VoteOption from a string. It returns an error
// if the string is invalid.
func VoteOptionFromString(str string) (types.VoteOption, error) {
	option, ok := types.VoteOption_value[str]
	if !ok {
		return types.OptionEmpty, fmt.Errorf("'%s' is not a valid vote option", str)
	}
	return types.VoteOption(option), nil
}

// ValidVoteOption returns true if the vote option is valid and false otherwise.
func ValidVoteOption(option types.VoteOption) bool {
	if option == types.OptionYes ||
		option == types.OptionAbstain ||
		option == types.OptionNo ||
		option == types.OptionNoWithVeto {
		return true
	}
	return false
}
