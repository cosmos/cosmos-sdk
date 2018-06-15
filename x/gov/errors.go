//nolint
package gov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultCodespace sdk.CodespaceType = 5

	CodeUnknownProposal         sdk.CodeType = 1
	CodeInactiveProposal        sdk.CodeType = 2
	CodeAlreadyActiveProposal   sdk.CodeType = 3
	CodeAlreadyFinishedProposal sdk.CodeType = 4
	CodeAddressNotStaked        sdk.CodeType = 5
	CodeInvalidTitle            sdk.CodeType = 6
	CodeInvalidDescription      sdk.CodeType = 7
	CodeInvalidProposalType     sdk.CodeType = 8
	CodeInvalidVote             sdk.CodeType = 9
)

//----------------------------------------
// Error constructors

func ErrUnknownProposal(proposalID int64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeUnknownProposal, fmt.Sprintf("Unknown proposal - %d", proposalID))
}

func ErrInactiveProposal(proposalID int64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInactiveProposal, fmt.Sprintf("Inactive proposal - %d", proposalID))
}

func ErrAlreadyActiveProposal(proposalID int64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeAlreadyActiveProposal, fmt.Sprintf("Proposal %d has been already active", proposalID))
}

func ErrAlreadyFinishedProposal(proposalID int64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeAlreadyFinishedProposal, fmt.Sprintf("Proposal %d has already passed its voting period", proposalID))
}

func ErrAddressNotStaked(address sdk.Address) sdk.Error {
	bechAddr, _ := sdk.Bech32ifyAcc(address)
	return sdk.NewError(DefaultCodespace, CodeAddressNotStaked, fmt.Sprintf("Address %s is not staked and is thus ineligible to vote", bechAddr))
}

func ErrInvalidTitle(title string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidTitle, fmt.Sprintf("Proposal Title '%s' is not valid", title))
}

func ErrInvalidDescription(description string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidDescription, fmt.Sprintf("Proposal Desciption '%s' is not valid", description))
}

func ErrInvalidProposalType(proposalType string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidProposalType, fmt.Sprintf("Proposal Type '%s' is not valid", proposalType))
}

func ErrInvalidVote(voteOption string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidVote, fmt.Sprintf("'%s' is not a valid voting option", voteOption))
}
