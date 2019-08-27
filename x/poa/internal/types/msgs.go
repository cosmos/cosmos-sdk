package types

import (
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg = &MsgCreateValidator{}
	_ sdk.Msg = &MsgEditValidator{}
)

//______________________________________________________________________

// MsgCreateValidator - struct for bonding transactions
// creation of a validator is defaulted to weight of 10
type MsgCreateValidator struct {
	Description      stakingtypes.Description `json:"description" yaml:"description"`
	ValidatorAddress sdk.ValAddress           `json:"validator_address" yaml:"validator_address"`
	PubKey           crypto.PubKey            `json:"pubkey" yaml:"pubkey"`
}

type msgCreateValidatorJSON struct {
	Description      stakingtypes.Description `json:"description" yaml:"description"`
	ValidatorAddress sdk.ValAddress           `json:"validator_address" yaml:"validator_address"`
	PubKey           string                   `json:"pubkey" yaml:"pubkey"`
}

// Default way to create validator. Delegator address and validator address are the same
func NewMsgCreateValidator(
	valAddr sdk.ValAddress, pubKey crypto.PubKey,
	description stakingtypes.Description,
) MsgCreateValidator {

	return MsgCreateValidator{
		Description:      description,
		ValidatorAddress: valAddr,
		PubKey:           pubKey,
	}
}

//nolint
func (msg MsgCreateValidator) Route() string { return RouterKey }
func (msg MsgCreateValidator) Type() string  { return "create_validator" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgCreateValidator) GetSigners() []sdk.AccAddress {
	// delegator is first signer so delegator pays fees
	addrs := []sdk.AccAddress{}

	addrs = append(addrs, sdk.AccAddress(msg.ValidatorAddress))

	return addrs
}

// MarshalJSON implements the json.Marshaler interface to provide custom JSON
// serialization of the MsgCreateValidator type.
func (msg MsgCreateValidator) MarshalJSON() ([]byte, error) {
	return json.Marshal(msgCreateValidatorJSON{
		Description:      msg.Description,
		ValidatorAddress: msg.ValidatorAddress,
		PubKey:           sdk.MustBech32ifyConsPub(msg.PubKey),
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
	msg.ValidatorAddress = msgCreateValJSON.ValidatorAddress
	var err error
	msg.PubKey, err = sdk.GetConsPubKeyBech32(msgCreateValJSON.PubKey)
	if err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgCreateValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgCreateValidator) ValidateBasic() sdk.Error {
	// note that unmarshaling from bech32 ensures either empty or valid
	// if Params.AcceptAllValidators = false {
	// 	return
	// }
	if msg.ValidatorAddress.Empty() {
		return stakingtypes.ErrNilValidatorAddr(DefaultCodeSpace)
	}
	if msg.Description == (stakingtypes.Description{}) {
		return sdk.NewError(stakingtypes.DefaultCodespace, stakingtypes.CodeInvalidInput, "description must be included")
	}

	return nil
}

// MsgEditValidator - struct for editing a validator
type MsgEditValidator struct {
	ValidatorAddress sdk.ValAddress `json:"address" yaml:"address"`
	stakingtypes.Description
}

func NewMsgEditValidator(valAddr sdk.ValAddress, description stakingtypes.Description) MsgEditValidator {
	return MsgEditValidator{
		ValidatorAddress: valAddr,
		Description:      description,
	}
}

//nolint
func (msg MsgEditValidator) Route() string { return RouterKey }
func (msg MsgEditValidator) Type() string  { return "edit_validator" }
func (msg MsgEditValidator) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.ValidatorAddress)}
}

// get the bytes for the message signer to sign on
func (msg MsgEditValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgEditValidator) ValidateBasic() sdk.Error {
	if msg.ValidatorAddress.Empty() {
		return sdk.NewError(DefaultCodeSpace, stakingtypes.CodeInvalidInput, "nil validator address")
	}

	if msg.Description == (stakingtypes.Description{}) {
		return sdk.NewError(DefaultCodeSpace, stakingtypes.CodeInvalidInput, "transaction must include some information to modify")
	}

	return nil
}
