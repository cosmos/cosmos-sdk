package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/gogo/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// bank message types
const (
	TypeMsgSend      = "send"
	TypeMsgMultiSend = "multisend"
)

var _ sdk.Msg = &MsgSend{}

// NewMsgSend - construct a msg to send coins from one account to another.
//nolint:interfacer
func NewMsgSend(fromAddr, toAddr sdk.AccAddress, amount sdk.Coins) *MsgSend {
	return &MsgSend{FromAddress: fromAddr.String(), ToAddress: toAddr.String(), Amount: amount}
}

// Route Implements Msg.
func (msg MsgSend) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgSend) Type() string { return TypeMsgSend }

// ValidateBasic Implements Msg.
func (msg MsgSend) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid recipient address (%s)", err)
	}

	if !msg.Amount.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	if !msg.Amount.IsAllPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgSend) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners Implements Msg.
func (msg MsgSend) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

// Rosetta interface
func (msg *MsgSend) ToOperations(withStatus bool, hasError bool) []*types.Operation {
	var operations []*types.Operation

	fromAddress := msg.FromAddress
	toAddress := msg.ToAddress
	amounts := msg.Amount
	if len(amounts) == 0 {
		return []*types.Operation{}
	}

	coin := amounts[0]
	sendOp := func(account, amount string, index int) *types.Operation {
		var status string
		if withStatus {
			status = "Success"
			if hasError {
				status = "Reverted"
			}
		}
		return &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: int64(index),
			},
			Type:   proto.MessageName(msg),
			Status: status,
			Account: &types.AccountIdentifier{
				Address: account,
			},
			Amount: &types.Amount{
				Value: amount,
				Currency: &types.Currency{
					Symbol: coin.Denom,
				},
			},
		}
	}
	operations = append(operations,
		sendOp(fromAddress, "-"+coin.Amount.String(), 0),
		sendOp(toAddress, coin.Amount.String(), 1),
	)

	return operations
}

func (msg MsgSend) FromOperations(ops []*types.Operation) (sdk.Msg, error) {
	var (
		from, to sdk.AccAddress
		sendAmt  sdk.Coin
		err      error
	)

	for _, op := range ops {
		if strings.HasPrefix(op.Amount.Value, "-") {
			from, err = sdk.AccAddressFromBech32(op.Account.Address)
			if err != nil {
				return nil, err
			}
			continue
		}

		to, err = sdk.AccAddressFromBech32(op.Account.Address)
		if err != nil {
			return nil, err
		}

		amount, err := strconv.ParseInt(op.Amount.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid amount")
		}

		sendAmt = sdk.NewCoin(op.Amount.Currency.Symbol, sdk.NewInt(amount))
	}

	return NewMsgSend(from, to, sdk.NewCoins(sendAmt)), nil
}

var _ sdk.Msg = &MsgMultiSend{}

// NewMsgMultiSend - construct arbitrary multi-in, multi-out send msg.
func NewMsgMultiSend(in []Input, out []Output) *MsgMultiSend {
	return &MsgMultiSend{Inputs: in, Outputs: out}
}

// Route Implements Msg
func (msg MsgMultiSend) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgMultiSend) Type() string { return TypeMsgMultiSend }

// ValidateBasic Implements Msg.
func (msg MsgMultiSend) ValidateBasic() error {
	// this just makes sure all the inputs and outputs are properly formatted,
	// not that they actually have the money inside
	if len(msg.Inputs) == 0 {
		return ErrNoInputs
	}

	if len(msg.Outputs) == 0 {
		return ErrNoOutputs
	}

	return ValidateInputsOutputs(msg.Inputs, msg.Outputs)
}

// GetSignBytes Implements Msg.
func (msg MsgMultiSend) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners Implements Msg.
func (msg MsgMultiSend) GetSigners() []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, len(msg.Inputs))
	for i, in := range msg.Inputs {
		addr, _ := sdk.AccAddressFromBech32(in.Address)
		addrs[i] = addr
	}

	return addrs
}

// ValidateBasic - validate transaction input
func (in Input) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(in.Address)
	if err != nil {
		return err
	}

	if !in.Coins.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, in.Coins.String())
	}

	if !in.Coins.IsAllPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, in.Coins.String())
	}

	return nil
}

// NewInput - create a transaction input, used with MsgMultiSend
//nolint:interfacer
func NewInput(addr sdk.AccAddress, coins sdk.Coins) Input {
	return Input{
		Address: addr.String(),
		Coins:   coins,
	}
}

// ValidateBasic - validate transaction output
func (out Output) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(out.Address)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid output address (%s)", err)
	}

	if !out.Coins.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, out.Coins.String())
	}

	if !out.Coins.IsAllPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, out.Coins.String())
	}

	return nil
}

// NewOutput - create a transaction output, used with MsgMultiSend
//nolint:interfacer
func NewOutput(addr sdk.AccAddress, coins sdk.Coins) Output {
	return Output{
		Address: addr.String(),
		Coins:   coins,
	}
}

// ValidateInputsOutputs validates that each respective input and output is
// valid and that the sum of inputs is equal to the sum of outputs.
func ValidateInputsOutputs(inputs []Input, outputs []Output) error {
	var totalIn, totalOut sdk.Coins

	for _, in := range inputs {
		if err := in.ValidateBasic(); err != nil {
			return err
		}

		totalIn = totalIn.Add(in.Coins...)
	}

	for _, out := range outputs {
		if err := out.ValidateBasic(); err != nil {
			return err
		}

		totalOut = totalOut.Add(out.Coins...)
	}

	// make sure inputs and outputs match
	if !totalIn.IsEqual(totalOut) {
		return ErrInputOutputMismatch
	}

	return nil
}
