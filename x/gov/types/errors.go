//nolint
package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultCodespace sdk.CodespaceType = "gov"

	CodeUnknownProposal          sdk.CodeType = 1
	CodeInactiveProposal         sdk.CodeType = 2
	CodeAlreadyActiveProposal    sdk.CodeType = 3
	CodeAlreadyFinishedProposal  sdk.CodeType = 4
	CodeAddressNotStaked         sdk.CodeType = 5
	CodeInvalidContent           sdk.CodeType = 6
	CodeInvalidProposalType      sdk.CodeType = 7
	CodeInvalidVote              sdk.CodeType = 8
	CodeInvalidGenesis           sdk.CodeType = 9
	CodeInvalidProposalStatus    sdk.CodeType = 10
	CodeProposalHandlerNotExists sdk.CodeType = 11
)

func ErrUnknownProposal(codespace sdk.CodespaceType, proposalID uint64) sdk.Error {
	return sdk.NewError(codespace, CodeUnknownProposal, fmt.Sprintf("unknown proposal with id %d", proposalID))
}

func ErrInactiveProposal(codespace sdk.CodespaceType, proposalID uint64) sdk.Error {
	return sdk.NewError(codespace, CodeInactiveProposal, fmt.Sprintf("inactive proposal with id %d", proposalID))
}

func ErrAlreadyActiveProposal(codespace sdk.CodespaceType, proposalID uint64) sdk.Error {
	return sdk.NewError(codespace, CodeAlreadyActiveProposal, fmt.Sprintf("proposal %d has been already active", proposalID))
}

func ErrAlreadyFinishedProposal(codespace sdk.CodespaceType, proposalID uint64) sdk.Error {
	return sdk.NewError(codespace, CodeAlreadyFinishedProposal, fmt.Sprintf("proposal %d has already passed its voting period", proposalID))
}

func ErrAddressNotStaked(codespace sdk.CodespaceType, address sdk.AccAddress) sdk.Error {
	return sdk.NewError(codespace, CodeAddressNotStaked, fmt.Sprintf("address %s is not staked and is thus ineligible to vote", address))
}

func ErrInvalidProposalContent(cs sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(cs, CodeInvalidContent, fmt.Sprintf("invalid proposal content: %s", msg))
}

func ErrInvalidProposalType(codespace sdk.CodespaceType, proposalType string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidProposalType, fmt.Sprintf("proposal type '%s' is not valid", proposalType))
}

func ErrInvalidVote(codespace sdk.CodespaceType, voteOption VoteOption) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidVote, fmt.Sprintf("'%v' is not a valid voting option", voteOption.String()))
}

func ErrInvalidGenesis(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidVote, msg)
}

func ErrNoProposalHandlerExists(codespace sdk.CodespaceType, content interface{}) sdk.Error {
	return sdk.NewError(codespace, CodeProposalHandlerNotExists, fmt.Sprintf("'%T' does not have a corresponding handler", content))
}
