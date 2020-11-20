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
					// TODO
				}
			}
		}
	}

	return next(ctx, tx, simulate)
}

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

func (vtd ValidateTokenTransferDecorator) validateMsgMultiSend(ctx sdk.Context, msg *types.MsgMultiSend) error {
	// TODO
	return nil
}
