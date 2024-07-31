package types

import (
	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg                            = &MsgCreateValidator{}
	_ codectypes.UnpackInterfacesMessage = (*MsgCreateValidator)(nil)
	_ sdk.Msg                            = &MsgEditValidator{}
	_ sdk.Msg                            = &MsgDelegate{}
	_ sdk.Msg                            = &MsgUndelegate{}
	_ sdk.Msg                            = &MsgUnbondValidator{}
	_ sdk.Msg                            = &MsgBeginRedelegate{}
	_ sdk.Msg                            = &MsgCancelUnbondingDelegation{}
	_ sdk.Msg                            = &MsgUpdateParams{}
	_ sdk.Msg                            = &MsgTokenizeShares{}
	_ sdk.Msg                            = &MsgRedeemTokensForShares{}
	_ sdk.Msg                            = &MsgTransferTokenizeShareRecord{}
	_ sdk.Msg                            = &MsgDisableTokenizeShares{}
	_ sdk.Msg                            = &MsgEnableTokenizeShares{}
	_ sdk.Msg                            = &MsgValidatorBond{}
)

// NewMsgCreateValidator creates a new MsgCreateValidator instance.
// Delegator address and validator address are the same.
func NewMsgCreateValidator(
	valAddr string, pubKey cryptotypes.PubKey,
	selfDelegation sdk.Coin, description Description, commission CommissionRates,
) (*MsgCreateValidator, error) {
	var pkAny *codectypes.Any
	if pubKey != nil {
		var err error
		if pkAny, err = codectypes.NewAnyWithValue(pubKey); err != nil {
			return nil, err
		}
	}
	return &MsgCreateValidator{
		Description:      description,
		ValidatorAddress: valAddr,
		Pubkey:           pkAny,
		Value:            selfDelegation,
		Commission:       commission,
	}, nil
}

// Validate validates the MsgCreateValidator sdk msg.
func (msg MsgCreateValidator) Validate(ac address.Codec) error {
	// note that unmarshaling from bech32 ensures both non-empty and valid
	_, err := ac.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	if msg.Pubkey == nil {
		return ErrEmptyValidatorPubKey
	}

	if !msg.Value.IsValid() || !msg.Value.Amount.IsPositive() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid delegation amount")
	}

	if msg.Description == (Description{}) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty description")
	}

	if msg.Commission == (CommissionRates{}) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty commission")
	}

	if err := msg.Commission.Validate(); err != nil {
		return err
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgCreateValidator) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey cryptotypes.PubKey
	return unpacker.UnpackAny(msg.Pubkey, &pubKey)
}

// NewMsgEditValidator creates a new MsgEditValidator instance
func NewMsgEditValidator(valAddr string, description Description, newRate *math.LegacyDec) *MsgEditValidator {
	return &MsgEditValidator{
		Description:      description,
		CommissionRate:   newRate,
		ValidatorAddress: valAddr,
	}
}

// NewMsgDelegate creates a new MsgDelegate instance.
func NewMsgDelegate(delAddr, valAddr string, amount sdk.Coin) *MsgDelegate {
	return &MsgDelegate{
		DelegatorAddress: delAddr,
		ValidatorAddress: valAddr,
		Amount:           amount,
	}
}

// NewMsgBeginRedelegate creates a new MsgBeginRedelegate instance.
func NewMsgBeginRedelegate(
	delAddr, valSrcAddr, valDstAddr string, amount sdk.Coin,
) *MsgBeginRedelegate {
	return &MsgBeginRedelegate{
		DelegatorAddress:    delAddr,
		ValidatorSrcAddress: valSrcAddr,
		ValidatorDstAddress: valDstAddr,
		Amount:              amount,
	}
}

// NewMsgUndelegate creates a new MsgUndelegate instance.
func NewMsgUndelegate(delAddr, valAddr string, amount sdk.Coin) *MsgUndelegate {
	return &MsgUndelegate{
		DelegatorAddress: delAddr,
		ValidatorAddress: valAddr,
		Amount:           amount,
	}
}

// NewMsgCancelUnbondingDelegation creates a new MsgCancelUnbondingDelegation instance.
func NewMsgCancelUnbondingDelegation(delAddr, valAddr string, creationHeight int64, amount sdk.Coin) *MsgCancelUnbondingDelegation {
	return &MsgCancelUnbondingDelegation{
		DelegatorAddress: delAddr,
		ValidatorAddress: valAddr,
		Amount:           amount,
		CreationHeight:   creationHeight,
	}
}

// NewMsgUnbondValidator creates a new MsgUnbondValidator instance.
// Validator account address must be provided - do not use valoper address.
func NewMsgUnbondValidator(valAccountAddr string) *MsgUnbondValidator {
	return &MsgUnbondValidator{
		ValidatorAddress: valAccountAddr,
	}
}

// NewMsgTokenizeShares creates a new MsgTokenizeShares instance.
func NewMsgTokenizeShares(delAddr, valAddr string, amount sdk.Coin, owner string) *MsgTokenizeShares {
	return &MsgTokenizeShares{
		DelegatorAddress:    delAddr,
		ValidatorAddress:    valAddr,
		Amount:              amount,
		TokenizedShareOwner: owner,
	}
}

// NewMsgRedeemTokensForShares creates a new MsgRedeemTokensForShares instance.
func NewMsgRedeemTokensForShares(delAddr string, amount sdk.Coin) *MsgRedeemTokensForShares {
	return &MsgRedeemTokensForShares{
		DelegatorAddress: delAddr,
		Amount:           amount,
	}
}

// NewMsgTransferTokenizeShareRecord creates a new MsgTransferTokenizeShareRecord instance.
func NewMsgTransferTokenizeShareRecord(recordID uint64, sender, newOwner string) *MsgTransferTokenizeShareRecord {
	return &MsgTransferTokenizeShareRecord{
		TokenizeShareRecordId: recordID,
		Sender:                sender,
		NewOwner:              newOwner,
	}
}

// NewMsgDisableTokenizeShares creates a new MsgDisableTokenizeShares instance.
func NewMsgDisableTokenizeShares(delAddr string) *MsgDisableTokenizeShares {
	return &MsgDisableTokenizeShares{
		DelegatorAddress: delAddr,
	}
}

// NewMsgEnableTokenizeShares creates a new MsgEnableTokenizeShares instance.
func NewMsgEnableTokenizeShares(delAddr string) *MsgEnableTokenizeShares {
	return &MsgEnableTokenizeShares{
		DelegatorAddress: delAddr,
	}
}

// NewMsgValidatorBond creates a new MsgValidatorBond instance.
func NewMsgValidatorBond(delAddr, valAddr string) *MsgValidatorBond {
	return &MsgValidatorBond{
		DelegatorAddress: delAddr,
		ValidatorAddress: valAddr,
	}
}
