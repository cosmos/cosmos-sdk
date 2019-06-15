package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = MsgSwapOrder{}
	_ sdk.Msg = MsgAddLiquidity{}
	_ sdk.Msg = MsgRemoveLiquidity{}
)

/* --------------------------------------------------------------------------- */
// MsgSwapOrder
/* --------------------------------------------------------------------------- */

// MsgSwap Order - struct for swapping a coin
// Input and Output can either be exact or calculated.
// An exact coin has the senders desired buy or sell amount.
// A calculated coin has the desired denomination and bounded amount
// the sender is willing to buy or sell in this order.
type MsgSwapOrder struct {
	Input      sdk.Coin       `json:"input"`    // the amount the sender is trading
	Output     sdk.Coin       `json:"output"`   // the amount the sender is recieivng
	Deadline   time.Time      `json:"deadline"` // deadline for the transaction to still be considered valid
	Sender     sdk.AccAddress `json:"sender"`
	Recipient  sdk.AccAddress `json:"recipient"`
	IsBuyOrder bool           `json:"is_buy_order"` // boolean indicating whether the order should be treated as a buy or sell
}

// NewMsgSwapOrder creates a new MsgSwapOrder object.
func NewMsgSwapOrder(
	input, output sdk.Coin, deadline time.Time,
	sender, recipient sdk.AccAddress, isBuyOrder bool,
) MsgSwapOrder {

	return MsgSwapOrder{
		Input:      input,
		Output:     output,
		Deadline:   deadline,
		Sender:     sender,
		Recipient:  recipient,
		IsBuyOrder: isBuyOrder,
	}
}

// Route Implements Msg.
func (msg MsgSwapOrder) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgSwapOrder) Type() string { return "swap_order" }

// ValidateBasic Implements Msg.
func (msg MsgSwapOrder) ValidateBasic() sdk.Error {
	if !msg.Input.IsValid() {
		return sdk.ErrInvalidCoins("coin is invalid: " + msg.Input.String())
	}
	if !msg.Output.IsValid() {
		return sdk.ErrInvalidCoins("coin is invalid: " + msg.Output.String())
	}
	if msg.Input.Denom == msg.Output.Denom {
		return ErrEqualDenom(DefaultCodespace)
	}
	if msg.Deadline.IsZero() {
		return ErrInvalidDeadline(DefaultCodespace, "deadline for MsgSwapOrder not initialized")
	}
	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress("invalid sender address")
	}
	if msg.Recipient.Empty() {
		return sdk.ErrInvalidAddress("invalid recipient address")
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgSwapOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgSwapOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

/* --------------------------------------------------------------------------- */
// MsgAddLiquidity
/* --------------------------------------------------------------------------- */

// MsgAddLiquidity - struct for adding liquidity to a reserve pool
type MsgAddLiquidity struct {
	Deposit       sdk.Coin       `json:"deposit"`        // coin to be deposited as liquidity with an upper bound for its amount
	DepositAmount sdk.Int        `json:"deposit_amount"` // exact amount of native asset being add to the liquidity pool
	MinReward     sdk.Int        `json:"min_reward"`     // lower bound UNI sender is willing to accept for deposited coins
	Deadline      time.Time      `json:"deadline"`
	Sender        sdk.AccAddress `json:"sender"`
}

// NewMsgAddLiquidity creates a new MsgAddLiquidity object.
func NewMsgAddLiquidity(
	deposit sdk.Coin, depositAmount, minReward sdk.Int,
	deadline time.Time, sender sdk.AccAddress,
) MsgAddLiquidity {

	return MsgAddLiquidity{
		Deposit:       deposit,
		DepositAmount: depositAmount,
		MinReward:     minReward,
		Deadline:      deadline,
		Sender:        sender,
	}
}

// Type Implements Msg.
func (msg MsgAddLiquidity) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgAddLiquidity) Type() string { return "add_liquidity" }

// ValidateBasic Implements Msg.
func (msg MsgAddLiquidity) ValidateBasic() sdk.Error {
	if !msg.Deposit.IsValid() {
		return sdk.ErrInvalidCoins("coin is invalid: " + msg.Deposit.String())
	}
	if !msg.DepositAmount.IsPositive() {
		return ErrNotPositive(DefaultCodespace, "deposit amount provided is not positive")
	}
	if !msg.MinReward.IsPositive() {
		return ErrNotPositive(DefaultCodespace, "minimum liquidity is not positive")
	}
	if msg.Deadline.IsZero() {
		return ErrInvalidDeadline(DefaultCodespace, "deadline for MsgAddLiquidity not initialized")
	}
	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress("invalid sender address")
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgAddLiquidity) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgAddLiquidity) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

/* --------------------------------------------------------------------------- */
// MsgRemoveLiquidity
/* --------------------------------------------------------------------------- */

// MsgRemoveLiquidity - struct for removing liquidity from a reserve pool
type MsgRemoveLiquidity struct {
	Withdraw       sdk.Coin       `json:"withdraw"`        // coin to be withdrawn with a lower bound for its amount
	WithdrawAmount sdk.Int        `json:"withdraw_amount"` // amount of UNI to be burned to withdraw liquidity from a reserve pool
	MinNative      sdk.Int        `json:"min_native"`      // minimum amount of the native asset the sender is willing to accept
	Deadline       time.Time      `json:"deadline"`
	Sender         sdk.AccAddress `json:"sender"`
}

// NewMsgRemoveLiquidity creates a new MsgRemoveLiquidity object
func NewMsgRemoveLiquidity(
	withdraw sdk.Coin, withdrawAmount, minNative sdk.Int,
	deadline time.Time, sender sdk.AccAddress,
) MsgRemoveLiquidity {

	return MsgRemoveLiquidity{
		Withdraw:       withdraw,
		WithdrawAmount: withdrawAmount,
		MinNative:      minNative,
		Deadline:       deadline,
		Sender:         sender,
	}
}

// Type Implements Msg.
func (msg MsgRemoveLiquidity) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgRemoveLiquidity) Type() string { return "remove_liquidity" }

// ValidateBasic Implements Msg.
func (msg MsgRemoveLiquidity) ValidateBasic() sdk.Error {
	if !msg.WithdrawAmount.IsPositive() {
		return ErrNotPositive(DefaultCodespace, "withdraw amount is not positive")
	}
	if !msg.Withdraw.IsValid() {
		return sdk.ErrInvalidCoins("coin is invalid: " + msg.Withdraw.String())
	}
	if !msg.MinNative.IsPositive() {
		return ErrNotPositive(DefaultCodespace, "minimum native amount is not positive")
	}
	if msg.Deadline.IsZero() {
		return ErrInvalidDeadline(DefaultCodespace, "deadline for MsgRemoveLiquidity not initialized")
	}
	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress("invalid sender address")
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgRemoveLiquidity) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgRemoveLiquidity) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
