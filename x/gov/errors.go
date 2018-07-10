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
	CodeInvalidGenesis          sdk.CodeType = 10
)

//----------------------------------------
// Error constructors

func ErrUnknownProposal(codespace sdk.CodespaceType, proposalID int64) sdk.Error {
	return sdk.NewError(codespace, CodeUnknownProposal, fmt.Sprintf("Unknown proposal - %d", proposalID))
}

func ErrInactiveProposal(codespace sdk.CodespaceType, proposalID int64) sdk.Error {
	return sdk.NewError(codespace, CodeInactiveProposal, fmt.Sprintf("Inactive proposal - %d", proposalID))
}

func ErrAlreadyActiveProposal(codespace sdk.CodespaceType, proposalID int64) sdk.Error {
	return sdk.NewError(codespace, CodeAlreadyActiveProposal, fmt.Sprintf("Proposal %d has been already active", proposalID))
}

func ErrAlreadyFinishedProposal(codespace sdk.CodespaceType, proposalID int64) sdk.Error {
	return sdk.NewError(codespace, CodeAlreadyFinishedProposal, fmt.Sprintf("Proposal %d has already passed its voting period", proposalID))
}

func ErrAddressNotStaked(codespace sdk.CodespaceType, address sdk.AccAddress) sdk.Error {
	return sdk.NewError(codespace, CodeAddressNotStaked, fmt.Sprintf("Address %s is not staked and is thus ineligible to vote", address))
}

func ErrInvalidTitle(codespace sdk.CodespaceType, title string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidTitle, fmt.Sprintf("Proposal Title '%s' is not valid", title))
}

func ErrInvalidDescription(codespace sdk.CodespaceType, description string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDescription, fmt.Sprintf("Proposal Desciption '%s' is not valid", description))
}

func ErrInvalidProposalType(codespace sdk.CodespaceType, strProposalType string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidProposalType, fmt.Sprintf("Proposal Type '%s' is not valid", strProposalType))
}

func ErrInvalidVote(codespace sdk.CodespaceType, strOption string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidVote, fmt.Sprintf("'%s' is not a valid voting option", strOption))
}

func ErrInvalidGenesis(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidVote, msg)
}
