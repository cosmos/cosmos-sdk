package types

import (
	"bytes"
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg = &MsgCreateValidator{}
	_ sdk.Msg = &MsgEditValidator{}
	_ sdk.Msg = &MsgDelegate{}
	_ sdk.Msg = &MsgUndelegate{}
	_ sdk.Msg = &MsgBeginRedelegate{}
)

//______________________________________________________________________

// MsgCreateValidator - struct for bonding transactions
type MsgCreateValidator struct {
	Description       Description    `json:"description"`
	Commission        CommissionMsg  `json:"commission"`
	MinSelfDelegation sdk.Int        `json:"min_self_delegation"`
	DelegatorAddr     sdk.AccAddress `json:"delegator_address"`
	ValidatorAddr     sdk.ValAddress `json:"validator_address"`
	PubKey            crypto.PubKey  `json:"pubkey"`
	Value             sdk.Coin       `json:"value"`
}

type msgCreateValidatorJSON struct {
	Description       Description    `json:"description"`
	Commission        CommissionMsg  `json:"commission"`
	MinSelfDelegation sdk.Int        `json:"min_self_delegation"`
	DelegatorAddr     sdk.AccAddress `json:"delegator_address"`
	ValidatorAddr     sdk.ValAddress `json:"validator_address"`
	PubKey            string         `json:"pubkey"`
	Value             sdk.Coin       `json:"value"`
}

// Default way to create validator. Delegator address and validator address are the same
func NewMsgCreateValidator(valAddr sdk.ValAddress, pubkey crypto.PubKey,
	selfDelegation sdk.Coin, description Description, commission CommissionMsg, minSelfDelegation sdk.Int) MsgCreateValidator {

	return NewMsgCreateValidatorOnBehalfOf(
		sdk.AccAddress(valAddr), valAddr, pubkey, selfDelegation, description, commission, minSelfDelegation,
	)
}

// Creates validator msg by delegator address on behalf of validator address
func NewMsgCreateValidatorOnBehalfOf(delAddr sdk.AccAddress, valAddr sdk.ValAddress,
	pubkey crypto.PubKey, value sdk.Coin, description Description, commission CommissionMsg, minSelfDelegation sdk.Int) MsgCreateValidator {
	return MsgCreateValidator{
		Description:       description,
		DelegatorAddr:     delAddr,
		ValidatorAddr:     valAddr,
		PubKey:            pubkey,
		Value:             value,
		Commission:        commission,
		MinSelfDelegation: minSelfDelegation,
	}
}

//nolint
func (msg MsgCreateValidator) Route() string { return RouterKey }
func (msg MsgCreateValidator) Type() string  { return "create_validator" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgCreateValidator) GetSigners() []sdk.AccAddress {
	// delegator is first signer so delegator pays fees
	addrs := []sdk.AccAddress{msg.DelegatorAddr}

	if !bytes.Equal(msg.DelegatorAddr.Bytes(), msg.ValidatorAddr.Bytes()) {
		// if validator addr is not same as delegator addr, validator must sign
		// msg as well
		addrs = append(addrs, sdk.AccAddress(msg.ValidatorAddr))
	}
	return addrs
}

// MarshalJSON implements the json.Marshaler interface to provide custom JSON
// serialization of the MsgCreateValidator type.
func (msg MsgCreateValidator) MarshalJSON() ([]byte, error) {
	return json.Marshal(msgCreateValidatorJSON{
		Description:       msg.Description,
		Commission:        msg.Commission,
		DelegatorAddr:     msg.DelegatorAddr,
		ValidatorAddr:     msg.ValidatorAddr,
		PubKey:            sdk.MustBech32ifyConsPub(msg.PubKey),
		Value:             msg.Value,
		MinSelfDelegation: msg.MinSelfDelegation,
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface to provide custom
// JSON deserialization of the MsgCreateValidator type.
func (msg *MsgCreateValidator) UnmarshalJSON(bz []byte) error {
	var msgCreateValJSON msgCreateValidatorJSON
	if err := json.Unmarshal(bz, &msgCreateValJSON); err != nil {
		return err
	}

	msg.Description = msgCreateValJSON.Description
	msg.Commission = msgCreateValJSON.Commission
	msg.DelegatorAddr = msgCreateValJSON.DelegatorAddr
	msg.ValidatorAddr = msgCreateValJSON.ValidatorAddr
	msg.PubKey = sdk.MustGetConsPubKeyBech32(msgCreateValJSON.PubKey)
	msg.Value = msgCreateValJSON.Value
	msg.MinSelfDelegation = msgCreateValJSON.MinSelfDelegation

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgCreateValidator) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgCreateValidator) ValidateBasic() sdk.Error {
	// note that unmarshaling from bech32 ensures either empty or valid
	if msg.DelegatorAddr.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddr.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.Value.Amount.LTE(sdk.ZeroInt()) {
		return ErrBadDelegationAmount(DefaultCodespace)
	}
	if msg.Description == (Description{}) {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "description must be included")
	}
	if msg.Commission == (CommissionMsg{}) {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "commission must be included")
	}
	if !msg.MinSelfDelegation.GT(sdk.ZeroInt()) {
		return ErrMinSelfDelegationInvalid(DefaultCodespace)
	}
	if msg.Value.Amount.LT(msg.MinSelfDelegation) {
		return ErrSelfDelegationBelowMinimum(DefaultCodespace)
	}

	return nil
}

// MsgEditValidator - struct for editing a validator
type MsgEditValidator struct {
	Description
	ValidatorAddr sdk.ValAddress `json:"address"`

	// We pass a reference to the new commission rate and min self delegation as it's not mandatory to
	// update. If not updated, the deserialized rate will be zero with no way to
	// distinguish if an update was intended.
	//
	// REF: #2373
	CommissionRate    *sdk.Dec `json:"commission_rate"`
	MinSelfDelegation *sdk.Int `json:"min_self_delegation"`
}

func NewMsgEditValidator(valAddr sdk.ValAddress, description Description, newRate *sdk.Dec, newMinSelfDelegation *sdk.Int) MsgEditValidator {
	return MsgEditValidator{
		Description:       description,
		CommissionRate:    newRate,
		ValidatorAddr:     valAddr,
		MinSelfDelegation: newMinSelfDelegation,
	}
}

//nolint
func (msg MsgEditValidator) Route() string { return RouterKey }
func (msg MsgEditValidator) Type() string  { return "edit_validator" }
func (msg MsgEditValidator) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.ValidatorAddr)}
}

// get the bytes for the message signer to sign on
func (msg MsgEditValidator) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgEditValidator) ValidateBasic() sdk.Error {
	if msg.ValidatorAddr.Empty() {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "nil validator address")
	}

	if msg.Description == (Description{}) {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "transaction must include some information to modify")
	}

	if msg.MinSelfDelegation != nil && !(*msg.MinSelfDelegation).GT(sdk.ZeroInt()) {
		return ErrMinSelfDelegationInvalid(DefaultCodespace)
	}

	if msg.CommissionRate != nil {
		if msg.CommissionRate.GT(sdk.OneDec()) || msg.CommissionRate.LT(sdk.ZeroDec()) {
			return sdk.NewError(DefaultCodespace, CodeInvalidInput, "commission rate must be between 0 and 1, inclusive")
		}
	}

	return nil
}

// MsgDelegate - struct for bonding transactions
type MsgDelegate struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
	ValidatorAddr sdk.ValAddress `json:"validator_addr"`
	Value         sdk.Coin       `json:"value"`
}

func NewMsgDelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, value sdk.Coin) MsgDelegate {
	return MsgDelegate{
		DelegatorAddr: delAddr,
		ValidatorAddr: valAddr,
		Value:         value,
	}
}

//nolint
func (msg MsgDelegate) Route() string { return RouterKey }
func (msg MsgDelegate) Type() string  { return "delegate" }
func (msg MsgDelegate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddr}
}

// get the bytes for the message signer to sign on
func (msg MsgDelegate) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgDelegate) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddr.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.Value.Amount.LTE(sdk.ZeroInt()) {
		return ErrBadDelegationAmount(DefaultCodespace)
	}
	return nil
}

//______________________________________________________________________

// MsgDelegate - struct for bonding transactions
type MsgBeginRedelegate struct {
	DelegatorAddr    sdk.AccAddress `json:"delegator_addr"`
	ValidatorSrcAddr sdk.ValAddress `json:"validator_src_addr"`
	ValidatorDstAddr sdk.ValAddress `json:"validator_dst_addr"`
	SharesAmount     sdk.Dec        `json:"shares_amount"`
}

func NewMsgBeginRedelegate(delAddr sdk.AccAddress, valSrcAddr,
	valDstAddr sdk.ValAddress, sharesAmount sdk.Dec) MsgBeginRedelegate {

	return MsgBeginRedelegate{
		DelegatorAddr:    delAddr,
		ValidatorSrcAddr: valSrcAddr,
		ValidatorDstAddr: valDstAddr,
		SharesAmount:     sharesAmount,
	}
}

//nolint
func (msg MsgBeginRedelegate) Route() string { return RouterKey }
func (msg MsgBeginRedelegate) Type() string  { return "begin_redelegate" }
func (msg MsgBeginRedelegate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddr}
}

// get the bytes for the message signer to sign on
func (msg MsgBeginRedelegate) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgBeginRedelegate) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorSrcAddr.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.ValidatorDstAddr.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.SharesAmount.LTE(sdk.ZeroDec()) {
		return ErrBadSharesAmount(DefaultCodespace)
	}
	return nil
}

// MsgUndelegate - struct for unbonding transactions
type MsgUndelegate struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
	ValidatorAddr sdk.ValAddress `json:"validator_addr"`
	SharesAmount  sdk.Dec        `json:"shares_amount"`
}

func NewMsgUndelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdk.Dec) MsgUndelegate {
	return MsgUndelegate{
		DelegatorAddr: delAddr,
		ValidatorAddr: valAddr,
		SharesAmount:  sharesAmount,
	}
}

//nolint
func (msg MsgUndelegate) Route() string                { return RouterKey }
func (msg MsgUndelegate) Type() string                 { return "begin_unbonding" }
func (msg MsgUndelegate) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.DelegatorAddr} }

// get the bytes for the message signer to sign on
func (msg MsgUndelegate) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgUndelegate) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddr.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.SharesAmount.LTE(sdk.ZeroDec()) {
		return ErrBadSharesAmount(DefaultCodespace)
	}
	return nil
}
