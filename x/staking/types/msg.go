package types

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/proto"

	rosettatypes "github.com/coinbase/rosetta-sdk-go/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// staking message types
const (
	TypeMsgUndelegate      = "begin_unbonding"
	TypeMsgEditValidator   = "edit_validator"
	TypeMsgCreateValidator = "create_validator"
	TypeMsgDelegate        = "delegate"
	TypeMsgBeginRedelegate = "begin_redelegate"

	// These are used for querying events by action.
	TypeSvcMsgUndelegate      = "/cosmos.staking.v1beta1.Msg/Undelegate"
	TypeSvcMsgEditValidator   = "/cosmos.staking.v1beta1.Msg/EditValidator"
	TypeSvcMsgCreateValidator = "/cosmos.staking.v1beta1.Msg/CreateValidator"
	TypeSvcMsgDelegate        = "/cosmos.staking.v1beta1.Msg/Deledate"
	TypeSvcMsgBeginRedelegate = "/cosmos.staking.v1beta1.Msg/BeginRedelegate"
)

var (
	_ sdk.Msg                            = &MsgCreateValidator{}
	_ codectypes.UnpackInterfacesMessage = (*MsgCreateValidator)(nil)
	_ sdk.Msg                            = &MsgCreateValidator{}
	_ sdk.Msg                            = &MsgEditValidator{}
	_ sdk.Msg                            = &MsgDelegate{}
	_ sdk.Msg                            = &MsgUndelegate{}
	_ sdk.Msg                            = &MsgBeginRedelegate{}
)

// NewMsgCreateValidator creates a new MsgCreateValidator instance.
// Delegator address and validator address are the same.
func NewMsgCreateValidator(
	valAddr sdk.ValAddress, pubKey cryptotypes.PubKey, //nolint:interfacer
	selfDelegation sdk.Coin, description Description, commission CommissionRates, minSelfDelegation sdk.Int,
) (*MsgCreateValidator, error) {
	var pkAny *codectypes.Any
	if pubKey != nil {
		var err error
		if pkAny, err = codectypes.NewAnyWithValue(pubKey); err != nil {
			return nil, err
		}
	}
	return &MsgCreateValidator{
		Description:       description,
		DelegatorAddress:  sdk.AccAddress(valAddr).String(),
		ValidatorAddress:  valAddr.String(),
		Pubkey:            pkAny,
		Value:             selfDelegation,
		Commission:        commission,
		MinSelfDelegation: minSelfDelegation,
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
		return ErrBadValidatorAddr
	}

	if msg.Pubkey == nil {
		return ErrEmptyValidatorPubKey
	}

	if !msg.Value.IsValid() || !msg.Value.Amount.IsPositive() {
		return ErrBadDelegationAmount
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

	if !msg.MinSelfDelegation.IsPositive() {
		return ErrMinSelfDelegationInvalid
	}

	if msg.Value.Amount.LT(msg.MinSelfDelegation) {
		return ErrSelfDelegationBelowMinimum
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgCreateValidator) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey cryptotypes.PubKey
	return unpacker.UnpackAny(msg.Pubkey, &pubKey)
}

// NewMsgEditValidator creates a new MsgEditValidator instance
//nolint:interfacer
func NewMsgEditValidator(valAddr sdk.ValAddress, description Description, newRate *sdk.Dec, newMinSelfDelegation *sdk.Int) *MsgEditValidator {
	return &MsgEditValidator{
		Description:       description,
		CommissionRate:    newRate,
		ValidatorAddress:  valAddr.String(),
		MinSelfDelegation: newMinSelfDelegation,
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

	if msg.MinSelfDelegation != nil && !msg.MinSelfDelegation.IsPositive() {
		return ErrMinSelfDelegationInvalid
	}

	if msg.CommissionRate != nil {
		if msg.CommissionRate.GT(sdk.OneDec()) || msg.CommissionRate.IsNegative() {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "commission rate must be between 0 and 1 (inclusive)")
		}
	}

	return nil
}

// NewMsgDelegate creates a new MsgDelegate instance.
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
		return ErrBadDelegationAmount
	}

	return nil
}

// Rosetta Msg interface.
func (msg *MsgDelegate) ToOperations(withStatus bool, hasError bool) []*rosettatypes.Operation {
	var operations []*rosettatypes.Operation
	delAddr := msg.DelegatorAddress
	valAddr := msg.ValidatorAddress
	coin := msg.Amount
	delOp := func(account *rosettatypes.AccountIdentifier, amount string, index int) *rosettatypes.Operation {
		var status string
		if withStatus {
			status = rosetta.StatusSuccess
			if hasError {
				status = rosetta.StatusReverted
			}
		}
		return &rosettatypes.Operation{
			OperationIdentifier: &rosettatypes.OperationIdentifier{
				Index: int64(index),
			},
			Type:    proto.MessageName(msg),
			Status:  status,
			Account: account,
			Amount: &rosettatypes.Amount{
				Value: amount,
				Currency: &rosettatypes.Currency{
					Symbol: coin.Denom,
				},
			},
		}
	}
	delAcc := &rosettatypes.AccountIdentifier{
		Address: delAddr,
	}
	valAcc := &rosettatypes.AccountIdentifier{
		Address: "staking_account",
		SubAccount: &rosettatypes.SubAccountIdentifier{
			Address: valAddr,
		},
	}
	operations = append(operations,
		delOp(delAcc, "-"+coin.Amount.String(), 0),
		delOp(valAcc, coin.Amount.String(), 1),
	)
	return operations
}

func (msg *MsgDelegate) FromOperations(ops []*rosettatypes.Operation) (sdk.Msg, error) {
	var (
		delAddr sdk.AccAddress
		valAddr sdk.ValAddress
		sendAmt sdk.Coin
		err     error
	)

	for _, op := range ops {
		if strings.HasPrefix(op.Amount.Value, "-") {
			if op.Account == nil {
				return nil, fmt.Errorf("account identifier must be specified")
			}
			delAddr, err = sdk.AccAddressFromBech32(op.Account.Address)
			if err != nil {
				return nil, err
			}
			continue
		}

		if op.Account.SubAccount == nil {
			return nil, fmt.Errorf("account identifier subaccount must be specified")
		}
		valAddr, err = sdk.ValAddressFromBech32(op.Account.SubAccount.Address)
		if err != nil {
			return nil, err
		}

		amount, err := strconv.ParseInt(op.Amount.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid amount: %w", err)
		}

		sendAmt = sdk.NewCoin(op.Amount.Currency.Symbol, sdk.NewInt(amount))
	}

	return NewMsgDelegate(delAddr, valAddr, sendAmt), nil
}

// NewMsgBeginRedelegate creates a new MsgBeginRedelegate instance.
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
		return ErrBadSharesAmount
	}

	return nil
}

// Rosetta Msg interface.
func (msg *MsgBeginRedelegate) ToOperations(withStatus bool, hasError bool) []*rosettatypes.Operation {
	var operations []*rosettatypes.Operation
	delAddr := msg.DelegatorAddress
	srcValAddr := msg.ValidatorSrcAddress
	destValAddr := msg.ValidatorDstAddress
	coin := msg.Amount
	delOp := func(account *rosettatypes.AccountIdentifier, amount string, index int) *rosettatypes.Operation {
		var status string
		if withStatus {
			status = rosetta.StatusSuccess
			if hasError {
				status = rosetta.StatusReverted
			}
		}
		return &rosettatypes.Operation{
			OperationIdentifier: &rosettatypes.OperationIdentifier{
				Index: int64(index),
			},
			Type:    proto.MessageName(msg),
			Status:  status,
			Account: account,
			Amount: &rosettatypes.Amount{
				Value: amount,
				Currency: &rosettatypes.Currency{
					Symbol: coin.Denom,
				},
			},
		}
	}
	srcValAcc := &rosettatypes.AccountIdentifier{
		Address: delAddr,
		SubAccount: &rosettatypes.SubAccountIdentifier{
			Address: srcValAddr,
		},
	}
	destValAcc := &rosettatypes.AccountIdentifier{
		Address: "staking_account",
		SubAccount: &rosettatypes.SubAccountIdentifier{
			Address: destValAddr,
		},
	}
	operations = append(operations,
		delOp(srcValAcc, "-"+coin.Amount.String(), 0),
		delOp(destValAcc, coin.Amount.String(), 1),
	)
	return operations
}

func (msg *MsgBeginRedelegate) FromOperations(ops []*rosettatypes.Operation) (sdk.Msg, error) {
	var (
		delAddr     sdk.AccAddress
		srcValAddr  sdk.ValAddress
		destValAddr sdk.ValAddress
		sendAmt     sdk.Coin
		err         error
	)

	for _, op := range ops {
		if strings.HasPrefix(op.Amount.Value, "-") {
			if op.Account == nil {
				return nil, fmt.Errorf("account identifier must be specified")
			}
			delAddr, err = sdk.AccAddressFromBech32(op.Account.Address)
			if err != nil {
				return nil, err
			}

			if op.Account.SubAccount == nil {
				return nil, fmt.Errorf("account identifier subaccount must be specified")
			}
			srcValAddr, err = sdk.ValAddressFromBech32(op.Account.SubAccount.Address)
			if err != nil {
				return nil, err
			}
			continue
		}

		if op.Account.SubAccount == nil {
			return nil, fmt.Errorf("account identifier subaccount must be specified")
		}
		destValAddr, err = sdk.ValAddressFromBech32(op.Account.SubAccount.Address)
		if err != nil {
			return nil, err
		}

		amount, err := strconv.ParseInt(op.Amount.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid amount: %w", err)
		}

		sendAmt = sdk.NewCoin(op.Amount.Currency.Symbol, sdk.NewInt(amount))
	}

	return NewMsgBeginRedelegate(delAddr, srcValAddr, destValAddr, sendAmt), nil
}

// NewMsgUndelegate creates a new MsgUndelegate instance.
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
		return ErrBadSharesAmount
	}

	return nil
}

// Rosetta Msg interface.
func (msg *MsgUndelegate) ToOperations(withStatus bool, hasError bool) []*rosettatypes.Operation {
	var operations []*rosettatypes.Operation
	delAddr := msg.DelegatorAddress
	valAddr := msg.ValidatorAddress
	coin := msg.Amount
	delOp := func(account *rosettatypes.AccountIdentifier, amount string, index int) *rosettatypes.Operation {
		var status string
		if withStatus {
			status = rosetta.StatusSuccess
			if hasError {
				status = rosetta.StatusReverted
			}
		}
		return &rosettatypes.Operation{
			OperationIdentifier: &rosettatypes.OperationIdentifier{
				Index: int64(index),
			},
			Type:    proto.MessageName(msg),
			Status:  status,
			Account: account,
			Amount: &rosettatypes.Amount{
				Value: amount,
				Currency: &rosettatypes.Currency{
					Symbol: coin.Denom,
				},
			},
		}
	}
	delAcc := &rosettatypes.AccountIdentifier{
		Address: delAddr,
	}
	valAcc := &rosettatypes.AccountIdentifier{
		Address: "staking_account",
		SubAccount: &rosettatypes.SubAccountIdentifier{
			Address: valAddr,
		},
	}
	operations = append(operations,
		delOp(valAcc, "-"+coin.Amount.String(), 0),
		delOp(delAcc, coin.Amount.String(), 1),
	)
	return operations
}

func (msg *MsgUndelegate) FromOperations(ops []*rosettatypes.Operation) (sdk.Msg, error) {
	var (
		delAddr  sdk.AccAddress
		valAddr  sdk.ValAddress
		undelAmt sdk.Coin
		err      error
	)

	for _, op := range ops {
		if strings.HasPrefix(op.Amount.Value, "-") {
			if op.Account.SubAccount == nil {
				return nil, fmt.Errorf("account identifier subaccount must be specified")
			}
			valAddr, err = sdk.ValAddressFromBech32(op.Account.SubAccount.Address)
			if err != nil {
				return nil, err
			}
			continue
		}

		if op.Account == nil {
			return nil, fmt.Errorf("account identifier must be specified")
		}

		delAddr, err = sdk.AccAddressFromBech32(op.Account.Address)
		if err != nil {
			return nil, err
		}

		amount, err := strconv.ParseInt(op.Amount.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid amount")
		}

		undelAmt = sdk.NewCoin(op.Amount.Currency.Symbol, sdk.NewInt(amount))
	}

	return NewMsgUndelegate(delAddr, valAddr, undelAmt), nil
}
