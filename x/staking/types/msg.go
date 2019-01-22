package types

import (
	"bytes"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Verify interface at compile time
var _, _, _ sdk.Msg = &MsgCreateValidator{}, &MsgEditValidator{}, &MsgDelegate{}

//______________________________________________________________________

// MsgCreateValidator - struct for bonding transactions
type MsgCreateValidator struct {
	Description   Description    `json:"description"`
	Commission    CommissionMsg  `json:"commission"`
	DelegatorAddr sdk.AccAddress `json:"delegator_address"`
	ValidatorAddr sdk.ValAddress `json:"validator_address"`
	PubKey        crypto.PubKey  `json:"pubkey"`
	Value         sdk.Coin       `json:"value"`
}

// Default way to create validator. Delegator address and validator address are the same
func NewMsgCreateValidator(valAddr sdk.ValAddress, pubkey crypto.PubKey,
	selfDelegation sdk.Coin, description Description, commission CommissionMsg) MsgCreateValidator {

	return NewMsgCreateValidatorOnBehalfOf(
		sdk.AccAddress(valAddr), valAddr, pubkey, selfDelegation, description, commission,
	)
}

// Creates validator msg by delegator address on behalf of validator address
func NewMsgCreateValidatorOnBehalfOf(delAddr sdk.AccAddress, valAddr sdk.ValAddress,
	pubkey crypto.PubKey, value sdk.Coin, description Description, commission CommissionMsg) MsgCreateValidator {
	return MsgCreateValidator{
		Description:   description,
		DelegatorAddr: delAddr,
		ValidatorAddr: valAddr,
		PubKey:        pubkey,
		Value:         value,
		Commission:    commission,
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

// TODO Remove use of custom struct (no longer necessary)
// get the bytes for the message signer to sign on
func (msg MsgCreateValidator) GetSignBytes() []byte {
	b, err := MsgCdc.MarshalJSON(struct {
		Description   Description    `json:"description"`
		Commission    CommissionMsg  `json:"commission"`
		DelegatorAddr sdk.AccAddress `json:"delegator_address"`
		ValidatorAddr sdk.ValAddress `json:"validator_address"`
		PubKey        string         `json:"pubkey"`
		Value         sdk.Coin       `json:"value"`
	}{
		Description:   msg.Description,
		ValidatorAddr: msg.ValidatorAddr,
		PubKey:        sdk.MustBech32ifyConsPub(msg.PubKey),
		Value:         msg.Value,
	})
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// quick validity check
func (msg MsgCreateValidator) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr == nil {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddr == nil {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if !(msg.Value.Amount.GT(sdk.ZeroInt())) {
		return ErrBadDelegationAmount(DefaultCodespace)
	}
	if msg.Description == (Description{}) {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "description must be included")
	}
	if msg.Commission == (CommissionMsg{}) {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "commission must be included")
	}

	return nil
}

//______________________________________________________________________

// MsgEditValidator - struct for editing a validator
type MsgEditValidator struct {
	Description
	ValidatorAddr sdk.ValAddress `json:"address"`

	// We pass a reference to the new commission rate as it's not mandatory to
	// update. If not updated, the deserialized rate will be zero with no way to
	// distinguish if an update was intended.
	//
	// REF: #2373
	CommissionRate *sdk.Dec `json:"commission_rate"`
}

func NewMsgEditValidator(valAddr sdk.ValAddress, description Description, newRate *sdk.Dec) MsgEditValidator {
	return MsgEditValidator{
		Description:    description,
		CommissionRate: newRate,
		ValidatorAddr:  valAddr,
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
	b, err := MsgCdc.MarshalJSON(struct {
		Description
		ValidatorAddr sdk.ValAddress `json:"address"`
	}{
		Description:   msg.Description,
		ValidatorAddr: msg.ValidatorAddr,
	})
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// quick validity check
func (msg MsgEditValidator) ValidateBasic() sdk.Error {
	if msg.ValidatorAddr == nil {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "nil validator address")
	}

	if msg.Description == (Description{}) {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "transaction must include some information to modify")
	}

	return nil
}

//______________________________________________________________________

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
	b, err := MsgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// quick validity check
func (msg MsgDelegate) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr == nil {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddr == nil {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if !(msg.Value.Amount.GT(sdk.ZeroInt())) {
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
	b, err := MsgCdc.MarshalJSON(struct {
		DelegatorAddr    sdk.AccAddress `json:"delegator_addr"`
		ValidatorSrcAddr sdk.ValAddress `json:"validator_src_addr"`
		ValidatorDstAddr sdk.ValAddress `json:"validator_dst_addr"`
		SharesAmount     string         `json:"shares"`
	}{
		DelegatorAddr:    msg.DelegatorAddr,
		ValidatorSrcAddr: msg.ValidatorSrcAddr,
		ValidatorDstAddr: msg.ValidatorDstAddr,
		SharesAmount:     msg.SharesAmount.String(),
	})
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// quick validity check
func (msg MsgBeginRedelegate) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr == nil {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorSrcAddr == nil {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.ValidatorDstAddr == nil {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.SharesAmount.LTE(sdk.ZeroDec()) {
		return ErrBadSharesAmount(DefaultCodespace)
	}
	return nil
}

//______________________________________________________________________

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
	b, err := MsgCdc.MarshalJSON(struct {
		DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
		ValidatorAddr sdk.ValAddress `json:"validator_addr"`
		SharesAmount  string         `json:"shares_amount"`
	}{
		DelegatorAddr: msg.DelegatorAddr,
		ValidatorAddr: msg.ValidatorAddr,
		SharesAmount:  msg.SharesAmount.String(),
	})
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// quick validity check
func (msg MsgUndelegate) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr == nil {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddr == nil {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.SharesAmount.LTE(sdk.ZeroDec()) {
		return ErrBadSharesAmount(DefaultCodespace)
	}
	return nil
}
