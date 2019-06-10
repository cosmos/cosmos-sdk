package types

import (
	"strings"
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
type MsgSwapOrder struct {
	SwapDenom  string         `json:"swap_denom"`   // The desired denomination either to be bought or sold
	Amount     sdk.Coins      `json:"amount"`       // The specified amount to be either bought or sold
	Bound      sdk.Int        `json:"bound"`        // If buy order, maximum amount of coins to be sold; otherwise minimum amount of coins to be bought
	Deadline   time.Time      `json:"deadline"`     // deadline for the transaction to still be considered valid
	Sender     sdk.AccAddress `json:"sender"`       // address swapping coin
	Recipient  sdk.AccAddress `json:"recipient"`    // address output coin is being sent to
	IsBuyOrder bool           `json:"is_buy_order"` // boolean indicating whether the order should be treated as a buy or sell
}

// NewMsgSwapOrder is a constructor function for MsgSwapOrder
func NewMsgSwapOrder(
	swapDenom string, amt sdk.Coins, bound sdk.Int, deadline time.Time,
	sender, recipient sdk.AccAddress, isBuyOrder bool,
) MsgSwapOrder {

	return MsgSwapOrder{
		SwapDenom:  swapDenom,
		Amount:     amt,
		Bound:      bound,
		Deadline:   deadline,
		Sender:     sender,
		Recipient:  recipient,
		IsBuyOrder: isBuyOrder,
	}
}

// Route Implements Msg
func (msg MsgSwapOrder) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgSwapOrder) Type() string { return "swap_order" }

// ValidateBasic Implements Msg.
func (msg MsgSwapOrder) ValidateBasic() sdk.Error {
	if strings.TrimSpace(msg.SwapDenom) == "" {
		return ErrNoDenom(DefaultCodespace)
	}
	// initially only support trading 1 coin only
	if len(msg.Amount) != 1 {
		return sdk.ErrInvalidCoins("must provide a single coin")
	}
	if !msg.Amount.IsValid() {
		return sdk.ErrInvalidCoins("coin is invalid: " + msg.Amount.String())
	}
	if msg.Amount[0].Denom == msg.SwapDenom {
		return ErrEqualDenom(DefaultCodespace)
	}
	if !msg.Bound.IsPositive() {
		return ErrInvalidBound(DefaultCodespace, "")
	}
	if msg.Deadline.IsZero() {
		return ErrInvalidDeadline(DefaultCodespace)
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

// MsgAddLiquidity - struct for adding liquidity to an exchange
type MsgAddLiquidity struct {
	ExchangeDenom string         `json:"exchange_denom"` // denomination of the exchange being added to
	DepositAmount sdk.Int        `json:"deposit_amount"` // exact amount of native asset being add to the liquidity pool
	MinLiquidity  sdk.Int        `json:"min_liquidity"`  // lower bound UNI sender is willing to accept for deposited coins
	MaxCoins      sdk.Int        `json:"max_coins"`      // maximum amount of the coin the sender is willing to deposit.
	Deadline      time.Time      `json:"deadline"`
	Sender        sdk.AccAddress `json:"sender"`
}

// NewMsgAddLiquidity is a constructor function for MsgAddLiquidity
func NewMsgAddLiquidity(
	exchangeDenom string, depositAmount, minLiquidity, maxCoins sdk.Int,
	deadline time.Time, sender sdk.AccAddress,
) MsgAddLiquidity {

	return MsgAddLiquidity{
		DepositAmount: depositAmount,
		ExchangeDenom: exchangeDenom,
		MinLiquidity:  minLiquidity,
		MaxCoins:      maxCoins,
		Deadline:      deadline,
		Sender:        sender,
	}
}

// Type Implements Msg
func (msg MsgAddLiquidity) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgAddLiquidity) Type() string { return "add_liquidity" }

// ValidateBasic Implements Msg.
func (msg MsgAddLiquidity) ValidateBasic() sdk.Error {
	if !msg.DepositAmount.IsPositive() {
		return ErrInsufficientAmount(DefaultCodespace, "deposit amount provided is not positive")
	}
	if strings.TrimSpace(msg.ExchangeDenom) == "" {
		return ErrNoDenom(DefaultCodespace)
	}
	if !msg.MinLiquidity.IsPositive() {
		return ErrInvalidBound(DefaultCodespace, "minimum liquidity is not positive")
	}
	if !msg.MaxCoins.IsPositive() {
		return ErrInvalidBound(DefaultCodespace, "maxmimum coins is not positive")
	}
	if msg.Deadline.IsZero() {
		return ErrInvalidDeadline(DefaultCodespace)
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

// MsgRemoveLiquidity - struct for removing liquidity from an exchange
type MsgRemoveLiquidity struct {
	ExchangeDenom  string         `json:"exchange_denom"`  // denomination of the exchange being withdrawn from
	WithdrawAmount sdk.Int        `json:"withdraw_amount"` // amount of UNI to be burned to withdraw liquidity from an exchange
	MinNative      sdk.Int        `json:"min_native"`      // minimum amount of the native asset the sender is willing to accept
	MinCoins       sdk.Int        `json:"min_coins"`       // minimum amount of the exchange coin the sender is willing to accept
	Deadline       time.Time      `json:"deadline"`
	Sender         sdk.AccAddress `json:"sender"`
}

// NewMsgRemoveLiquidity is a contructor function for MsgRemoveLiquidity
func NewMsgRemoveLiquidity(
	exchangeDenom string, withdrawAmount, minNative, minCoins sdk.Int,
	deadline time.Time, sender sdk.AccAddress,
) MsgRemoveLiquidity {

	return MsgRemoveLiquidity{
		WithdrawAmount: withdrawAmount,
		ExchangeDenom:  exchangeDenom,
		MinNative:      minNative,
		MinCoins:       minCoins,
		Deadline:       deadline,
		Sender:         sender,
	}
}

// Type Implements Msg
func (msg MsgRemoveLiquidity) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgRemoveLiquidity) Type() string { return "remove_liquidity" }

// ValidateBasic Implements Msg.
func (msg MsgRemoveLiquidity) ValidateBasic() sdk.Error {
	if !msg.WithdrawAmount.IsPositive() {
		return ErrInsufficientAmount(DefaultCodespace, "withdraw amount is not positive")
	}
	if strings.TrimSpace(msg.ExchangeDenom) == "" {
		return ErrNoDenom(DefaultCodespace)
	}
	if !msg.MinNative.IsPositive() {
		return ErrInvalidBound(DefaultCodespace, "minimum native is not positive")
	}
	if !msg.MinCoins.IsPositive() {
		return ErrInvalidBound(DefaultCodespace, "minimum coins is not positive")
	}
	if msg.Deadline.IsZero() {
		return ErrInvalidDeadline(DefaultCodespace)
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
