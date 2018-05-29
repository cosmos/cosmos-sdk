//nolint
package gov

import (
	"strconv"

	"encoding/hex"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const ( // TODO TODO TODO TODO TODO TODO

	DefaultCodespace sdk.CodespaceType = 4

	// Gov errors reserve 401 ~ 499.
	CodeUnknownProposal          sdk.CodeType = 401
	CodeInactiveProposal         sdk.CodeType = 402
	CodeAlreadyActiveProposal    sdk.CodeType = 403
	CodeAddressChangedDelegation sdk.CodeType = 404
	CodeAddressNotStaked         sdk.CodeType = 405
	CodeInvalidTitle             sdk.CodeType = 406
	CodeInvalidDescription       sdk.CodeType = 407
	CodeInvalidProposalType      sdk.CodeType = 408
	CodeInvalidVote              sdk.CodeType = 409
	CodeAddressAlreadyVote       sdk.CodeType = 410
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
	return sdk.NewError(DefaultCodespace, CodeAlreadyActiveProposal, "Proposal "+strconv.FormatInt(proposalID, 10)+" already active")
}

func ErrAddressChangedDelegation(address sdk.Address) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeAddressChangedDelegation, "Address "+hex.EncodeToString(address)+" has redelegated since vote began and is thus ineligible to vote")
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

func ErrAlreadyVote(address sdk.Address) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeAddressAlreadyVote, "Address "+hex.EncodeToString(address)+" has already voted")
}
