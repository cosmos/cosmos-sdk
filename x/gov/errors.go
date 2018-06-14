//nolint
package gov

import (
	"strconv"

	"encoding/hex"

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
	return sdk.NewError(DefaultCodespace, CodeUnknownProposal, "Unknown proposal - "+strconv.FormatInt(proposalID, 10))
}

func ErrInactiveProposal(proposalID int64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInactiveProposal, "Unknown proposal - "+strconv.FormatInt(proposalID, 10))
}

func ErrAlreadyActiveProposal(proposalID int64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeAlreadyActiveProposal, "Proposal "+strconv.FormatInt(proposalID, 10)+" has been already active")
}

func ErrAlreadyFinishedProposal(proposalID int64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeAlreadyFinishedProposal, "Proposal "+strconv.FormatInt(proposalID, 10)+" has already passed its voting period")
}

func ErrAddressNotStaked(address sdk.Address) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeAddressNotStaked, "Address "+hex.EncodeToString(address)+" is not staked and is thus ineligible to vote")
}

func ErrInvalidTitle(title string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidTitle, "Proposal Title '"+title+"' is not valid")
}

func ErrInvalidDescription(description string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidDescription, "Proposal Desciption '"+description+"' is not valid")
}

func ErrInvalidProposalType(proposalType string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidProposalType, "Proposal Type '"+proposalType+"' is not valid")
}

func ErrInvalidVote(voteOption string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidVote, "'"+voteOption+"' is not a valid voting option")
}
