//nolint
package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	// Simple Gov errors reserve 700 ~ 799.
	CodeInvalidOption      CodeType = 701
	CodeInvalidProposalID  CodeType = 702
	CodeVotingPeriodClosed CodeType = 703
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
	default:
		return sdk.CodeToDefaultMsg(code)
	}
}

//----------------------------------------
// Error constructors

func ErrInvalidOption() sdk.Error {
	return newError(CodeInvalidOption, "The chosen option is invalid")
}

func ErrInvalidProposalID() sdk.Error {
	return newError(CodeInvalidProposalID, "There is no Proposal for this ProposalID")
}

func ErrVotingPeriodClosed() sdk.Error {
	return newError(CodeVotingPeriodClosed, "Voting period is closed for this proposal")
}

//----------------------------------------

func msgOrDefaultMsg(msg string, code CodeType) string {
	if msg != "" {
		return msg
	} else {
		return codeToDefaultMsg(code)
	}
}

func newError(code CodeType, msg string) sdk.Error {
	msg = msgOrDefaultMsg(msg, code)
	return sdk.NewError(code, msg)
}
