package simpleGovernance

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	DefaultCodespace sdk.CodespaceType = 7

	// Simple Gov errors reserve 700 ~ 799.
	CodeInvalidOption      CodeType = 701
	CodeInvalidProposalID  CodeType = 702
	CodeVotingPeriodClosed CodeType = 703
	CodeEmptyProposalQueue CodeType = 704
	CodeInvalidTitle       CodeType = 705
	CodeInvalidDescription CodeType = 706
)

// NOTE: Don't stringer this, we'll put better messages in later.
func codeToDefaultMsg(code CodeType) string {
	switch code {
	case CodeInvalidOption:
		return "Invalid option"
	case CodeInvalidProposalID:
		return "Invalid proposalID"
	case CodeVotingPeriodClosed:
		return "Voting Period Closed"
	case CodeEmptyProposalQueue:
		return "ProposalQueue is empty"
	case CodeInvalidTitle:
		return "Invalid proposal title"
	case CodeInvalidDescription:
		return "Invalid proposal description"
	default:
		return sdk.CodeToDefaultMsg(code)
	}
}

//----------------------------------------
// Error constructors

// nolint
func ErrInvalidOption(msg) sdk.Error {
	if msg {
		return newError(CodeInvalidOption, msg)
	}
	return newError(CodeInvalidOption, "The chosen option is invalid")
}

// nolint
func ErrInvalidProposalID(msg) sdk.Error {
	if msg {
		return newError(CodeInvalidProposalID, msg)
	}
	return newError(CodeInvalidProposalID, "ProposalID is not valid")
}

// nolint
func ErrInvalidTitle() sdk.Error {
	return newError(CodeInvalidTitle, "Cannot submit a proposal with empty title")
}

// nolint
func ErrInvalidDescription() sdk.Error {
	return newError(CodeInvalidDescription, "Cannot submit a proposal with empty description")
}

// nolint
func ErrVotingPeriodClosed() sdk.Error {
	return newError(CodeVotingPeriodClosed, "Voting period is closed for this proposal")
}

// nolint
func ErrEmptyProposalQueue() sdk.Error {
	return newError(CodeEmptyProposalQueue, "Can't get element from an empty proposal queue")
}

//----------------------------------------

func msgOrDefaultMsg(msg string, code CodeType) string {
	if msg != "" {
		return msg
	}
	return codeToDefaultMsg(code)
}

func newError(codespace sdk.CodespaceType, code CodeType, msg string) sdk.Error {
	msg = msgOrDefaultMsg(msg, code)
	return sdk.NewError(codespace, code, msg)
}
