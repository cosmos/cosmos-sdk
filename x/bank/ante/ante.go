package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewAnteHandler returns an AnteHandler that performs the token transfer validation
func NewAnteHandler(
	bankKeeper BankKeeper,
	tokenKeeper TokenKeeper,
) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		NewValidateTokenTransferDecorator(bankKeeper, tokenKeeper),
	)
}
