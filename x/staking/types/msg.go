package types

import (
	"bytes"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// staking message types
//
//nolint:gosec // these are not hard coded credentials
const (
	TypeMsgUndelegate                  = "begin_unbonding"
	TypeMsgUnbondValidator             = "unbond_validator"
	TypeMsgEditValidator               = "edit_validator"
	TypeMsgCreateValidator             = "create_validator"
	TypeMsgDelegate                    = "delegate"
	TypeMsgBeginRedelegate             = "begin_redelegate"
	TypeMsgCancelUnbondingDelegation   = "cancel_unbond"
	TypeMsgTokenizeShares              = "tokenize_shares"
	TypeMsgRedeemTokensForShares       = "redeem_tokens_for_shares"
	TypeMsgTransferTokenizeShareRecord = "transfer_tokenize_share_record"
	TypeMsgDisableTokenizeShares       = "disable_tokenize_shares"
	TypeMsgEnableTokenizeShares        = "enable_tokenize_shares"
	TypeMsgValidatorBond               = "validator_bond"
)

var (
	_ sdk.Msg                            = &MsgCreateValidator{}
	_ codectypes.UnpackInterfacesMessage = (*MsgCreateValidator)(nil)
	_ sdk.Msg                            = &MsgCreateValidator{}
	_ sdk.Msg                            = &MsgEditValidator{}
	_ sdk.Msg                            = &MsgDelegate{}
	_ sdk.Msg                            = &MsgUndelegate{}
	_ sdk.Msg                            = &MsgUnbondValidator{}
	_ sdk.Msg                            = &MsgBeginRedelegate{}
	_ sdk.Msg                            = &MsgTokenizeShares{}
	_ sdk.Msg                            = &MsgRedeemTokensForShares{}
	_ sdk.Msg                            = &MsgTransferTokenizeShareRecord{}
	_ sdk.Msg                            = &MsgDisableTokenizeShares{}
	_ sdk.Msg                            = &MsgEnableTokenizeShares{}
	_ sdk.Msg                            = &MsgCancelUnbondingDelegation{}
	_ sdk.Msg                            = &MsgValidatorBond{}
)

// NewMsgCreateValidator creates a new MsgCreateValidator instance.
// Delegator address and validator address are the same.
func NewMsgCreateValidator(
	valAddr sdk.ValAddress, pubKey cryptotypes.PubKey, //nolint:interfacer
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
		DelegatorAddress: sdk.AccAddress(valAddr).String(),
		ValidatorAddress: valAddr.String(),
		Pubkey:           pkAny,
		Value:            selfDelegation,
		Commission:       commission,
	}, nil
}

// Route implements the sdk.Msg interface.
func (msg MsgCreateValidator) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgCreateValidator) Type() string { return TypeMsgCreateValidator }

// GetSigners implements the sdk.Msg interface. It returns the address(es) that
// must sign over msg.GetSignBytes().
// If the validator address is not same as delegator's, then the validator must
// sign the msg as well.
func (msg MsgCreateValidator) GetSigners() []sdk.AccAddress {
	// delegator is first signer so delegator pays fees
	delAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		panic(err)
	}
	addrs := []sdk.AccAddress{delAddr}
	addr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(delAddr.Bytes(), addr.Bytes()) {
		addrs = append(addrs, sdk.AccAddress(addr))
	}

	return addrs
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgCreateValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgCreateValidator) ValidateBasic() error {
	// note that unmarshaling from bech32 ensures either empty or valid
	delAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return err
	}
	if delAddr.Empty() {
		return ErrEmptyDelegatorAddr
	}

	if msg.ValidatorAddress == "" {
		return ErrEmptyValidatorAddr
	}

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return err
	}
	if !sdk.AccAddress(valAddr).Equals(delAddr) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "validator address is invalid")
	}

	if msg.Pubkey == nil {
		return ErrEmptyValidatorPubKey
	}

	if !msg.Value.IsValid() || !msg.Value.Amount.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid delegation amount")
	}

	if msg.Description == (Description{}) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty description")
	}

	if msg.Commission == (CommissionRates{}) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty commission")
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
//
//nolint:interfacer
func NewMsgEditValidator(valAddr sdk.ValAddress, description Description, newRate *sdk.Dec) *MsgEditValidator {
	return &MsgEditValidator{
		Description:      description,
		CommissionRate:   newRate,
		ValidatorAddress: valAddr.String(),
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgEditValidator) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgEditValidator) Type() string { return TypeMsgEditValidator }

// GetSigners implements the sdk.Msg interface.
func (msg MsgEditValidator) GetSigners() []sdk.AccAddress {
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{valAddr.Bytes()}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgEditValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgEditValidator) ValidateBasic() error {
	if msg.ValidatorAddress == "" {
		return ErrEmptyValidatorAddr
	}

	if msg.Description == (Description{}) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty description")
	}

	if msg.CommissionRate != nil {
		if msg.CommissionRate.GT(sdk.OneDec()) || msg.CommissionRate.IsNegative() {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "commission rate must be between 0 and 1 (inclusive)")
		}
	}

	return nil
}

// NewMsgDelegate creates a new MsgDelegate instance.
//
//nolint:interfacer
func NewMsgDelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) *MsgDelegate {
	return &MsgDelegate{
		DelegatorAddress: delAddr.String(),
		ValidatorAddress: valAddr.String(),
		Amount:           amount,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgDelegate) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgDelegate) Type() string { return TypeMsgDelegate }

// GetSigners implements the sdk.Msg interface.
func (msg MsgDelegate) GetSigners() []sdk.AccAddress {
	delAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delAddr}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgDelegate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgDelegate) ValidateBasic() error {
	if msg.DelegatorAddress == "" {
		return ErrEmptyDelegatorAddr
	}

	if msg.ValidatorAddress == "" {
		return ErrEmptyValidatorAddr
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid delegation amount",
		)
	}

	return nil
}

// NewMsgBeginRedelegate creates a new MsgBeginRedelegate instance.
//
//nolint:interfacer
func NewMsgBeginRedelegate(
	delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress, amount sdk.Coin,
) *MsgBeginRedelegate {
	return &MsgBeginRedelegate{
		DelegatorAddress:    delAddr.String(),
		ValidatorSrcAddress: valSrcAddr.String(),
		ValidatorDstAddress: valDstAddr.String(),
		Amount:              amount,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgBeginRedelegate) Route() string { return RouterKey }

// Type implements the sdk.Msg interface
func (msg MsgBeginRedelegate) Type() string { return TypeMsgBeginRedelegate }

// GetSigners implements the sdk.Msg interface
func (msg MsgBeginRedelegate) GetSigners() []sdk.AccAddress {
	delAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delAddr}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgBeginRedelegate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgBeginRedelegate) ValidateBasic() error {
	if msg.DelegatorAddress == "" {
		return ErrEmptyDelegatorAddr
	}

	if msg.ValidatorSrcAddress == "" {
		return ErrEmptyValidatorAddr
	}

	if msg.ValidatorDstAddress == "" {
		return ErrEmptyValidatorAddr
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid shares amount",
		)
	}

	return nil
}

// NewMsgUndelegate creates a new MsgUndelegate instance.
//
//nolint:interfacer
func NewMsgUndelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) *MsgUndelegate {
	return &MsgUndelegate{
		DelegatorAddress: delAddr.String(),
		ValidatorAddress: valAddr.String(),
		Amount:           amount,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgUndelegate) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgUndelegate) Type() string { return TypeMsgUndelegate }

// GetSigners implements the sdk.Msg interface.
func (msg MsgUndelegate) GetSigners() []sdk.AccAddress {
	delAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delAddr}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgUndelegate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgUndelegate) ValidateBasic() error {
	if msg.DelegatorAddress == "" {
		return ErrEmptyDelegatorAddr
	}

	if msg.ValidatorAddress == "" {
		return ErrEmptyValidatorAddr
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid shares amount",
		)
	}

	return nil
}

// NewMsgUnbondValidator creates a new MsgUnbondValidator instance.
//
//nolint:interfacer
func NewMsgUnbondValidator(valAddr sdk.ValAddress) *MsgUnbondValidator {
	return &MsgUnbondValidator{
		ValidatorAddress: valAddr.String(),
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgUnbondValidator) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgUnbondValidator) Type() string { return TypeMsgUnbondValidator }

// GetSigners implements the sdk.Msg interface.
func (msg MsgUnbondValidator) GetSigners() []sdk.AccAddress {
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{valAddr.Bytes()}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgUnbondValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgUnbondValidator) ValidateBasic() error {
	if _, err := sdk.ValAddressFromBech32(msg.ValidatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	return nil
}

// NewMsgTokenizeShares creates a new MsgTokenizeShares instance.
//
//nolint:interfacer
func NewMsgTokenizeShares(delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin, owner sdk.AccAddress) *MsgTokenizeShares {
	return &MsgTokenizeShares{
		DelegatorAddress:    delAddr.String(),
		ValidatorAddress:    valAddr.String(),
		Amount:              amount,
		TokenizedShareOwner: owner.String(),
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgTokenizeShares) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgTokenizeShares) Type() string { return TypeMsgTokenizeShares }

// GetSigners implements the sdk.Msg interface.
func (msg MsgTokenizeShares) GetSigners() []sdk.AccAddress {
	delegator, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delegator}
}

// MsgTokenizeShares implements the sdk.Msg interface.
func (msg MsgTokenizeShares) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgTokenizeShares) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DelegatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}
	if _, err := sdk.ValAddressFromBech32(msg.ValidatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.TokenizedShareOwner); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid tokenize share owner address: %s", err)
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid shares amount",
		)
	}

	return nil
}

// NewMsgRedeemTokensForShares creates a new MsgRedeemTokensForShares instance.
//
//nolint:interfacer
func NewMsgRedeemTokensForShares(delAddr sdk.AccAddress, amount sdk.Coin) *MsgRedeemTokensForShares {
	return &MsgRedeemTokensForShares{
		DelegatorAddress: delAddr.String(),
		Amount:           amount,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgRedeemTokensForShares) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgRedeemTokensForShares) Type() string { return TypeMsgRedeemTokensForShares }

// GetSigners implements the sdk.Msg interface.
func (msg MsgRedeemTokensForShares) GetSigners() []sdk.AccAddress {
	delegator, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delegator}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgRedeemTokensForShares) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgRedeemTokensForShares) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DelegatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid shares amount",
		)
	}

	return nil
}

// NewMsgTransferTokenizeShareRecord creates a new MsgTransferTokenizeShareRecord instance.
//
//nolint:interfacer
func NewMsgTransferTokenizeShareRecord(recordId uint64, sender, newOwner sdk.AccAddress) *MsgTransferTokenizeShareRecord {
	return &MsgTransferTokenizeShareRecord{
		TokenizeShareRecordId: recordId,
		Sender:                sender.String(),
		NewOwner:              newOwner.String(),
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgTransferTokenizeShareRecord) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgTransferTokenizeShareRecord) Type() string { return TypeMsgTransferTokenizeShareRecord }

// GetSigners implements the sdk.Msg interface.
func (msg MsgTransferTokenizeShareRecord) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgTransferTokenizeShareRecord) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgTransferTokenizeShareRecord) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.NewOwner); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid new owner address: %s", err)
	}

	return nil
}

// NewMsgDisableTokenizeShares creates a new MsgDisableTokenizeShares instance.
//
//nolint:interfacer
func NewMsgDisableTokenizeShares(delAddr sdk.AccAddress) *MsgDisableTokenizeShares {
	return &MsgDisableTokenizeShares{
		DelegatorAddress: delAddr.String(),
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgDisableTokenizeShares) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgDisableTokenizeShares) Type() string { return TypeMsgDisableTokenizeShares }

// GetSigners implements the sdk.Msg interface.
func (msg MsgDisableTokenizeShares) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgDisableTokenizeShares) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgDisableTokenizeShares) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DelegatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	return nil
}

// NewMsgEnableTokenizeShares creates a new MsgEnableTokenizeShares instance.
//
//nolint:interfacer
func NewMsgEnableTokenizeShares(delAddr sdk.AccAddress) *MsgEnableTokenizeShares {
	return &MsgEnableTokenizeShares{
		DelegatorAddress: delAddr.String(),
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgEnableTokenizeShares) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgEnableTokenizeShares) Type() string { return TypeMsgEnableTokenizeShares }

// GetSigners implements the sdk.Msg interface.
func (msg MsgEnableTokenizeShares) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgEnableTokenizeShares) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgEnableTokenizeShares) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DelegatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	return nil
}

// NewMsgCancelUnbondingDelegation creates a new MsgCancelUnbondingDelegation instance.
//
//nolint:interfacer
func NewMsgCancelUnbondingDelegation(delAddr sdk.AccAddress, valAddr sdk.ValAddress, creationHeight int64, amount sdk.Coin) *MsgCancelUnbondingDelegation {
	return &MsgCancelUnbondingDelegation{
		DelegatorAddress: delAddr.String(),
		ValidatorAddress: valAddr.String(),
		Amount:           amount,
		CreationHeight:   creationHeight,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgCancelUnbondingDelegation) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgCancelUnbondingDelegation) Type() string { return TypeMsgCancelUnbondingDelegation }

// GetSigners implements the sdk.Msg interface.
func (msg MsgCancelUnbondingDelegation) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	return []sdk.AccAddress{delegator}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgCancelUnbondingDelegation) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgCancelUnbondingDelegation) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DelegatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}
	if _, err := sdk.ValAddressFromBech32(msg.ValidatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid amount",
		)
	}

	if msg.CreationHeight <= 0 {
		return sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid height",
		)
	}

	return nil
}

// NewMsgValidatorBond creates a new MsgValidatorBond instance.
//
//nolint:interfacer
func NewMsgValidatorBond(delAddr sdk.AccAddress, valAddr sdk.ValAddress) *MsgValidatorBond {
	return &MsgValidatorBond{
		DelegatorAddress: delAddr.String(),
		ValidatorAddress: valAddr.String(),
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgValidatorBond) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgValidatorBond) Type() string { return TypeMsgValidatorBond }

// GetSigners implements the sdk.Msg interface.
func (msg MsgValidatorBond) GetSigners() []sdk.AccAddress {
	delegator, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delegator}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgValidatorBond) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgValidatorBond) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DelegatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}
	if _, err := sdk.ValAddressFromBech32(msg.ValidatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	return nil
}
