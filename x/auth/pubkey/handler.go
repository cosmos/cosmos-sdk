package pubkey

import (
	authtypes "github.com/cosmos/cosmos-sdk/auth/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/pubkey/types"
)

// NewHandler returns a handler for x/auth message types.
func NewHandler(ak keeper.AccountKeeper, bk types.BankKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgChangePubKey:
			return handleMsgChangePubKey(ctx, ak, bk, msg)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}

func handleMsgChangePubKey(ctx sdk.Context, ak keeper.AccountKeeper, bk types.BankKeeper, msg *types.MsgChangePubKey) (*sdk.Result, error) {
	acc := ak.GetAccount(ctx, msg.Address)
	if acc == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s does not exist", msg.Address)
	}
	acc.SetPubKey(msg.PubKey)
	ak.SetAccount(ctx, acc)

	// TODO is this correct to handle gas logic here or should do on ante handler ?
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "signers should exist")
	}
	feePayer := signers[0]
	bondDenom := k.paramstore.Get(ctx, types.KeyBondDenom, &res)

	amount := uint64(5000) // should get from auth params
	fees := sdk.Coins{sdk.NewInt64Coin(bondDenom, int64(amount))}

	err := bk.SendCoinsFromAccountToModule(ctx, feePayer, authtypes.FeeCollectorName, fees)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err)
	}
	ctx.GasMeter().ConsumeGas(amount, "pubkey change fee")

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}
