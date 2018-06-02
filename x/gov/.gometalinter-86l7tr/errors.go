//nolint
package gov

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const ( // TODO TODO TODO TODO TODO TODO
	// Gov errors reserve 200 ~ 299.
	CodeUnknownProposal          sdk.CodeType = 201
	CodeInactiveProposal         sdk.CodeType = 202
	CodeAlreadyActiveProposal    sdk.CodeType = 203
	CodeAddressChangedDelegation sdk.CodeType = 204
	CodeAddressNotStaked         sdk.CodeType = 205
	CodeInvalidTitle             sdk.CodeType = 206
	CodeInvalidDescription       sdk.CodeType = 207
	CodeInvalidProposalType      sdk.CodeType = 208
	CodeInvalidVote              sdk.CodeType = 209
)

//----------------------------------------
// Error constructors

func ErrUnknownProposal(proposalID int64) sdk.Error {
	return sdk.NewError(CodeUnknownProposal, "Unknown proposal - "+strconv.FormatInt(proposalID, 10))
}

func ErrInactiveProposal(proposalID int64) sdk.Error {
	return sdk.NewError(CodeInactiveProposal, "Unknown proposal - "+strconv.FormatInt(proposalID, 10))
}

func ErrAlreadyActiveProposal(proposalID int64) sdk.Error {
	return sdk.NewError(CodeAlreadyActiveProposal, "Proposal "+strconv.FormatInt(proposalID, 10)+" already active")
}

func ErrAddressChangedDelegation(address sdk.Address) sdk.Error {
	return sdk.NewError(CodeAddressChangedDelegation, "Address "+string(address)+" has redelegated since vote began and is thus ineligible to vote")
}

func ErrAddressNotStaked(address sdk.Address) sdk.Error {
	return sdk.NewError(CodeAddressNotStaked, "Address "+string(address)+" is not staked and is thus ineligible to vote")
}

func ErrInvalidTitle(title string) sdk.Error {
	return sdk.NewError(CodeInvalidTitle, "Proposal Title '"+title+"' is not valid")
}

func ErrInvalidDescription(description string) sdk.Error {
	return sdk.NewError(CodeInvalidDescription, "Proposal Desciption '"+description+"' is not valid")
}

func ErrInvalidProposalType(proposalType string) sdk.Error {
	return sdk.NewError(CodeInvalidProposalType, "Proposal Type '"+proposalType+"' is not valid")
}

func ErrInvalidVote(voteOption string) sdk.Error {
	return sdk.NewError(CodeInvalidVote, "'"+voteOption+"' is not a valid voting option")
}
