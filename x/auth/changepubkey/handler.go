package changepubkey

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	changepubkeykeeper "github.com/cosmos/cosmos-sdk/x/auth/changepubkey/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/types"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
)

// NewHandler returns a handler for x/auth message types.
func NewHandler(ak keeper.AccountKeeper, pk changepubkeykeeper.ChangePubKeyKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgChangePubKey:
			return handleMsgChangePubKey(ctx, ak, pk, msg)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}

func handleMsgChangePubKey(ctx sdk.Context, ak keeper.AccountKeeper, pk changepubkeykeeper.ChangePubKeyKeeper, msg *types.MsgChangePubKey) (*sdk.Result, error) {
	acc := ak.GetAccount(ctx, msg.Address)
	if acc == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s does not exist", msg.Address)
	}
	acc.SetPubKey(msg.GetPubKey())
	ak.SetAccount(ctx, acc)

	// handle additional fee logic inside MsgChangePubKey handler
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "signers should exist")
	}

	amount := pk.GetParams(ctx).PubKeyChangeCost
	ctx.GasMeter().ConsumeGas(amount, "pubkey change fee")

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}
