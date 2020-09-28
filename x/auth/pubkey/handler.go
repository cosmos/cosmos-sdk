package pubkey

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/pubkey/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// NewHandler returns a handler for x/auth message types.
func NewHandler(ak keeper.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgChangePubKey:
			return handleMsgChangePubKey(ctx, ak, bk, sk, msg)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}

func handleMsgChangePubKey(ctx sdk.Context, ak keeper.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper, msg *types.MsgChangePubKey) (*sdk.Result, error) {
	acc := ak.GetAccount(ctx, msg.Address)
	if acc == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s does not exist", msg.Address)
	}
	acc.SetPubKey(msg.PubKey)
	ak.SetAccount(ctx, acc)

	// handle additional fee logic inside MsgChangePubKey handler
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "signers should exist")
	}
	feePayer := signers[0]
	amount := ak.GetParams(ctx).PubKeyChangeCost // should get from auth params
	fees := sdk.Coins{sdk.NewInt64Coin(sk.BondDenom(ctx), int64(amount))}
	err := bk.SendCoinsFromAccountToModule(ctx, feePayer, authtypes.FeeCollectorName, fees)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
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
