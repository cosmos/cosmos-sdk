package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	ak AccountKeeper
}

// NewMsgServerImpl returns an implementation of the x/auth MsgServer interface.
func NewMsgServerImpl(ak AccountKeeper) types.MsgServer {
	return &msgServer{
		ak: ak,
	}
}

func (ms msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := sdk.ValidateAuthority(ctx, ms.ak.authority, msg.Authority); err != nil {
		return nil, err
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}
	if err := ms.ak.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// ChangePubKey replaces the public key of an account while keeping its
// address. The msg is signed by the account's current key (checked by the ante
// handler); the proof shows that the holder of the new key agreed to control
// this account on this chain.
func (ms msgServer) ChangePubKey(goCtx context.Context, msg *types.MsgChangePubKey) (*types.MsgChangePubKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params, err := ms.ak.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	if !params.PubKeyChangeEnabled {
		return nil, types.ErrPubKeyChangeDisabled
	}

	addr, err := ms.ak.addressCodec.StringToBytes(msg.Address)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address: %s", err)
	}

	acc := ms.ak.GetAccount(ctx, addr)
	if acc == nil {
		return nil, errorsmod.Wrapf(types.ErrAccountNotRekeyable, "account %s does not exist", msg.Address)
	}
	if _, ok := acc.(sdk.ModuleAccountI); ok {
		return nil, errorsmod.Wrapf(types.ErrAccountNotRekeyable, "%s is a module account", msg.Address)
	}
	// A user signer's pubkey is stored by SetPubKeyDecorator earlier in the
	// same tx, so a nil pubkey means the msg was dispatched by a module (a
	// contract or interchain account) for an account that never had a key.
	oldPk := acc.GetPubKey()
	if oldPk == nil {
		return nil, errorsmod.Wrapf(types.ErrAccountNotRekeyable, "%s has no pubkey", msg.Address)
	}
	// This check is what keeps x/group policy accounts (and other
	// ModuleCredential-derived accounts) from being rekeyed; see
	// enterprise/group/x/group/module/rekey_test.go.
	if _, ok := oldPk.(*types.ModuleCredential); ok {
		return nil, errorsmod.Wrapf(types.ErrAccountNotRekeyable, "%s is a module credential account", msg.Address)
	}

	// Validate the new key before anything calls Address() or GetPubKeys() on
	// it; both panic on malformed or un-unpacked keys.
	if msg.NewPubKey == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidNewPubKey, "new pubkey is nil")
	}
	newPk, ok := msg.NewPubKey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, errorsmod.Wrapf(types.ErrInvalidNewPubKey, "new pubkey %s is not an unpacked pubkey", msg.NewPubKey.TypeUrl)
	}
	if err := types.ValidateRekeyPubKey(oldPk, newPk, params); err != nil {
		return nil, err
	}

	// Charge before decoding and verifying the proof, so a tx whose proof
	// fails still pays for the signature verification it caused.
	ctx.GasMeter().ConsumeGas(params.PubKeyChangeCost, "change pubkey")

	doc := types.ChangePubKeyProofDoc{
		ChainId:       ctx.ChainID(),
		AccountNumber: acc.GetAccountNumber(),
		Address:       msg.Address,
		NewPubKey:     msg.NewPubKey,
	}
	signBytes, err := types.ChangePubKeyProofSignBytes(ms.ak.addressCodec, doc, newPk)
	if err != nil {
		return nil, err
	}
	proof, err := types.DecodeChangePubKeyProof(msg.Proof)
	if err != nil {
		return nil, err
	}
	if err := types.VerifyChangePubKeyProof(newPk, signBytes, proof); err != nil {
		return nil, err
	}

	oldKeyAddr, err := ms.ak.addressCodec.BytesToString(oldPk.Address())
	if err != nil {
		return nil, err
	}
	newKeyAddr, err := ms.ak.addressCodec.BytesToString(newPk.Address())
	if err != nil {
		return nil, err
	}

	if err := ms.ak.applyRekey(ctx, acc, newPk); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeChangePubKey,
		sdk.NewAttribute(types.AttributeKeyAddress, msg.Address),
		sdk.NewAttribute(types.AttributeKeyOldPubKeyAddress, oldKeyAddr),
		sdk.NewAttribute(types.AttributeKeyNewPubKeyAddress, newKeyAddr),
	))

	return &types.MsgChangePubKeyResponse{}, nil
}
