package params

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultCodespace sdk.CodespaceType = "params"
	//
	CodeInvalidString sdk.CodeType = 0

	CodeInvalidMinDepositDenom  sdk.CodeType = 100
	CodeInvalidMinDepositAmount sdk.CodeType = 101
	CodeInvalidDepositPeriod    sdk.CodeType = 102
	CodeInvalidVotingPeriod     sdk.CodeType = 103
	CodeInvalidThreshold        sdk.CodeType = 104
	CodeInvalidVeto             sdk.CodeType = 105
	CodeInvalidMaxProposalNum   sdk.CodeType = 106
	CodeInvalidKey              sdk.CodeType = 107
)

func ErrInvalidString(valuestr string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidString, fmt.Sprintf("%s can't convert to a specific type", valuestr))
}
