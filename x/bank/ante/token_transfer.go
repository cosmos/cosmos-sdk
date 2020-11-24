package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// ValidateTokenTransferDecorator checks if the parties of the token transfer satisfy the underlying constraint
type ValidateTokenTransferDecorator struct {
	bankKeeper  BankKeeper
	tokenKeeper TokenKeeper
}

// NewValidateTokenTransferDecorator constructs a new ValidateTokenTransferDecorator instance
func NewValidateTokenTransferDecorator(bk BankKeeper, tk TokenKeeper) ValidateTokenTransferDecorator {
	return ValidateTokenTransferDecorator{
		bankKeeper:  bk,
		tokenKeeper: tk,
	}
}

// AnteHandle implements AnteHandler
func (vtd ValidateTokenTransferDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// check only when the UnrestrictedTokenTransfer param is set to false
	if !vtd.bankKeeper.GetParams(ctx).UnrestrictedTokenTransfer {
		for _, msg := range tx.GetMsgs() {
			if msg.Route() == types.ModuleName {

				switch msg := msg.(type) {
				case *types.MsgSend:
					err := vtd.validateMsgSend(ctx, msg)
					if err != nil {
						return ctx, err
					}

				case *types.MsgMultiSend:
					err := vtd.validateMsgMultiSend(ctx, msg)
					if err != nil {
						return ctx, err
					}
				}
			}
		}
	}

	return next(ctx, tx, simulate)
}

// validateMsgSend validates the MsgSend msg
func (vtd ValidateTokenTransferDecorator) validateMsgSend(ctx sdk.Context, msg *types.MsgSend) error {
	for _, coin := range msg.Amount {
		ownerAddr, err := vtd.tokenKeeper.GetOwner(ctx, coin.Denom)
		if err == nil {
			owner := ownerAddr.String()

			if msg.FromAddress != owner && msg.ToAddress != owner {
				return sdkerrors.Wrapf(
					types.ErrUnauthorizedTransfer,
					"either the sender or recipient must be the owner %s for token %s",
					owner, coin.Denom,
				)
			}
		}
	}

	return nil
}

// validateMsgSend validates the MsgMultiSend msg
func (vtd ValidateTokenTransferDecorator) validateMsgMultiSend(ctx sdk.Context, msg *types.MsgMultiSend) error {
	inputMap := getInputMap(msg.Inputs)
	outputMap := getOutputMap(msg.Outputs)

	for denom, addresses := range inputMap {
		ownerAddr, err := vtd.tokenKeeper.GetOwner(ctx, denom)
		if err == nil {
			owner := ownerAddr.String()

			if !owned(owner, addresses) && !owned(owner, outputMap[denom]) {
				return sdkerrors.Wrapf(
					types.ErrUnauthorizedTransfer,
					"either the sender or recipient must be the owner %s for token %s",
					owner, denom,
				)
			}
		}
	}

	return nil
}

// owned returns false if any address is not the owner of the denom among the given non-empty addresses
// True otherwise
func owned(owner string, addresses []string) bool {
	for _, addr := range addresses {
		if addr != owner {
			return false
		}
	}

	return true
}

// getInputMap maps input denoms to addresses
func getInputMap(inputs []types.Input) map[string][]string {
	inputMap := make(map[string][]string)

	for _, input := range inputs {
		for _, coin := range input.Coins {
			inputMap[coin.Denom] = append(inputMap[coin.Denom], input.Address)
		}
	}

	return inputMap
}

// getOutputMap maps output denoms to addresses
func getOutputMap(outputs []types.Output) map[string][]string {
	outputMap := make(map[string][]string)

	for _, output := range outputs {
		for _, coin := range output.Coins {
			outputMap[coin.Denom] = append(outputMap[coin.Denom], output.Address)
		}
	}

	return outputMap
}
