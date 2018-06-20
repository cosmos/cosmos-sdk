package simpleGovernance

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	DefaultCodespace sdk.CodespaceType = 7

	// Simple Gov errors reserve 700 ~ 799.
	CodeInvalidOption         CodeType = 701
	CodeInvalidProposalID     CodeType = 702
	CodeVotingPeriodClosed    CodeType = 703
	CodeEmptyProposalQueue    CodeType = 704
	CodeInvalidTitle          CodeType = 705
	CodeInvalidDescription    CodeType = 706
	CodeProposalNotFound      CodeType = 707
	CodeVoteNotFound          CodeType = 708
	CodeProposalQueueNotFound CodeType = 709
	CodeInvalidDeposit        CodeType = 710
)

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
	case CodeProposalNotFound:
		return "Proposal not found"
	case CodeVoteNotFound:
		return "Vote not found"
	case CodeProposalQueueNotFound:
		return "Proposal Queue not found"
	case CodeInvalidDeposit:
		return "Invalid deposit"
	default:
		return sdk.CodeToDefaultMsg(code)
	}
}

//----------------------------------------
// Error constructors

// ErrInvalidOption throws an error on invalid option
func ErrInvalidOption(msg string) sdk.Error {
	return newError(DefaultCodespace, CodeInvalidOption, msg)
}

// ErrInvalidProposalID throws an error on invalid proposaID
func ErrInvalidProposalID(msg string) sdk.Error {
	return newError(DefaultCodespace, CodeInvalidProposalID, msg)
}

// ErrInvalidTitle throws an error on invalid title
func ErrInvalidTitle(msg string) sdk.Error {
	return newError(DefaultCodespace, CodeInvalidTitle, msg)
}

// ErrInvalidDescription throws an error on invalid description
func ErrInvalidDescription(msg string) sdk.Error {
	return newError(DefaultCodespace, CodeInvalidDescription, msg)
}

// ErrVotingPeriodClosed throws an error when voting period is closed
func ErrVotingPeriodClosed() sdk.Error {
	return newError(DefaultCodespace, CodeVotingPeriodClosed, "Voting period is closed for this proposal")
}

// ErrEmptyProposalQueue throws an error when ProposalQueue is empty
func ErrEmptyProposalQueue(msg string) sdk.Error {
	return newError(DefaultCodespace, CodeEmptyProposalQueue, msg)
}

// ErrProposalNotFound throws an error when the searched proposal is not found
func ErrProposalNotFound(proposalID int64) sdk.Error {
	return newError(DefaultCodespace, CodeProposalNotFound, "Proposal with id "+
		strconv.Itoa(int(proposalID))+" not found")
}

// ErrVoteNotFound throws an error when the searched vote is not found
func ErrVoteNotFound(msg string) sdk.Error {
	return newError(DefaultCodespace, CodeVoteNotFound, msg)
}

// ErrProposalQueueNotFound throws an error on when the searched ProposalQueue is not found
func ErrProposalQueueNotFound(msg string) sdk.Error {
	return newError(DefaultCodespace, CodeProposalQueueNotFound, msg)
}

// ErrMinimumDeposit throws an error when deposit is less than the default minimum
func ErrMinimumDeposit() sdk.Error {
	return newError(DefaultCodespace, CodeInvalidDeposit, "Deposit is lower than the minimum")
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
